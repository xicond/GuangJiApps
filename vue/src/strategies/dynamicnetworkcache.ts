import { Strategy, StrategyOptions, StrategyHandler } from 'workbox-strategies';

export interface DynamicNetworkCacheStrategyOptions extends StrategyOptions {
    timeoutMs?: number;
    debounceMs?: number;
}

interface InFlightEntry {
    url: string;
    promise: Promise<Response>;
    abortController: AbortController;
    subscribers: Set<AbortSignal>;
}

// Global registry of in-flight requests across strategy instances in the SW
const globalInFlightRequests = new Map<string, InFlightEntry>();

// Listen for ABORT_REQUEST messages sent from client windows (Axios / fetch abort)
if (typeof self !== 'undefined' && 'addEventListener' in self) {
    self.addEventListener('message', (event: any) => {
        if (event.data && event.data.type === 'ABORT_REQUEST') {
            const { url } = event.data;
            if (!url) return;
            for (const [key, entry] of globalInFlightRequests.entries()) {
                const matchesUrl =
                    entry.url === url ||
                    entry.url.endsWith(url) ||
                    url.endsWith(entry.url);

                if (matchesUrl) {
                    entry.abortController.abort(new DOMException('The user aborted a request.', 'AbortError'));
                    globalInFlightRequests.delete(key);
                }
            }
        }
    });
}

export class DynamicNetworkCacheStrategy extends Strategy {
    private timeoutMs: number;
    private debounceMs: number;
    private debounceMap: Map<string, ReturnType<typeof setTimeout>>; // Diubah ke number (tipe setTimeout di browser/SW)
    protected inFlightRequests: Map<string, InFlightEntry> = globalInFlightRequests;

    constructor(options: DynamicNetworkCacheStrategyOptions = {}) {
        super(options);
        this.timeoutMs = options.timeoutMs ?? 500;
        this.debounceMs = options.debounceMs ?? 5000;
        this.debounceMap = new Map<string, ReturnType<typeof setTimeout>>();
    }

    /**
     * Helper to identify if an error is an AbortError or CanceledError.
     */
    protected isAbortError(err: any): boolean {
        return Boolean(
            err && (
                err.name === 'AbortError' ||
                err.name === 'CanceledError' ||
                err.code === 20 // DOMException.ABORT_ERR
            )
        );
    }

    /**
     * Executes network fetch with deduplication and abort propagation.
     * If all subscribing requests for a given URL have aborted, aborts the underlying network fetch.
     */
    protected fetchDeduplicated(request: Request, handler: StrategyHandler): Promise<Response> {
        const signal = request.signal;

        // If this specific request was already aborted before execution, reject immediately
        if (signal?.aborted) {
            return Promise.reject(new DOMException('The user aborted a request.', 'AbortError'));
        }

        if (request.method !== 'GET' && request.method !== 'HEAD') {
            const reqWithSignal = signal ? new Request(request, { signal }) : request;
            return handler.fetch(reqWithSignal);
        }

        const key = `${request.method}:${request.url}`;
        const existingEntry = globalInFlightRequests.get(key);

        // If an active in-flight request exists and is not aborted, attach this subscriber
        if (existingEntry && !existingEntry.abortController.signal.aborted) {
            if (signal) {
                existingEntry.subscribers.add(signal);
                const onAbort = () => {
                    existingEntry.subscribers.delete(signal);
                    if (existingEntry.subscribers.size === 0) {
                        existingEntry.abortController.abort(signal.reason);
                        globalInFlightRequests.delete(key);
                    }
                };
                signal.addEventListener('abort', onAbort, { once: true });

                return existingEntry.promise
                    .finally(() => {
                        signal.removeEventListener('abort', onAbort);
                        existingEntry.subscribers.delete(signal);
                    })
                    .then((response) => response.clone());
            }

            return existingEntry.promise.then((response) => response.clone());
        }

        // Create a new in-flight fetch controlled by a dedicated AbortController
        const abortController = new AbortController();
        const subscribers = new Set<AbortSignal>();

        const onAbort = () => {
            if (signal) {
                subscribers.delete(signal);
            }
            if (subscribers.size === 0) {
                abortController.abort(signal?.reason);
                globalInFlightRequests.delete(key);
            }
        };

        if (signal) {
            subscribers.add(signal);
            signal.addEventListener('abort', onAbort, { once: true });
        }

        // Pass the AbortController's signal into a new Request instance
        const reqWithSignal = new Request(request, { signal: abortController.signal });

        const fetchPromise = handler.fetch(reqWithSignal)
            .finally(() => {
                if (signal) {
                    signal.removeEventListener('abort', onAbort);
                    subscribers.delete(signal);
                }
                globalInFlightRequests.delete(key);
            });

        const entry: InFlightEntry = {
            url: request.url,
            promise: fetchPromise,
            abortController,
            subscribers
        };

        globalInFlightRequests.set(key, entry);

        return fetchPromise.then((response) => response.clone());
    }

    protected async _handle(request: Request, handler: StrategyHandler): Promise<Response> {
        // Fast-path: Check if client request was already aborted
        if (request.signal?.aborted) {
            throw new DOMException('The user aborted a request.', 'AbortError');
        }

        // 1. Check cache and fetch network asynchronously in parallel without initial await
        const cachePromise: Promise<Response | undefined> = handler.cacheMatch(request).catch(() => undefined);
        const networkPromise: Promise<Response> = this.fetchDeduplicated(request, handler);

        // Attach non-blocking background cache update once network fetch completes successfully
        networkPromise.then((networkResponse) => {
            if (request.signal?.aborted) return;
            if (networkResponse && networkResponse.ok) {
                const putPromise = handler.cachePut(request, networkResponse.clone()).catch(() => { });
                if (typeof handler.waitUntil === 'function') {
                    handler.waitUntil(putPromise);
                }
            }
            return networkResponse;
        }).catch(() => { });

        // 2. Setup timeoutMs promise with clean timer reference
        let timeoutTimer: ReturnType<typeof setTimeout> | undefined;
        const timeoutPromise = new Promise<{ isTimeout: true }>((resolve) => {
            timeoutTimer = setTimeout(() => resolve({ isTimeout: true }), this.timeoutMs);
        });

        // 3. Race network fetch against timeoutMs
        try {
            const result = await Promise.race([
                networkPromise.then((res) => ({ isTimeout: false as const, res })),
                timeoutPromise
            ]);

            if (request.signal?.aborted) {
                throw new DOMException('The user aborted a request.', 'AbortError');
            }

            // If network finished BEFORE timeoutMs, return fast network response
            if (!result.isTimeout && result.res) {
                return result.res;
            }
        } catch (err: any) {
            // If the client aborted the request, re-throw immediately without triggering background update
            if (request.signal?.aborted || this.isAbortError(err)) {
                throw err;
            }
            // Network fetch failed or errored before timeoutMs (e.g. offline)
            this.triggerDebouncedBackgroundUpdate(request, handler);
        } finally {
            if (timeoutTimer) {
                clearTimeout(timeoutTimer);
            }
        }

        if (request.signal?.aborted) {
            throw new DOMException('The user aborted a request.', 'AbortError');
        }

        // 4. Network did NOT finish before timeoutMs (or network fetch failed):
        // Race checking cache vs waiting for network fetch
        type FallbackWinner =
            | { source: 'cache'; res: Response | undefined }
            | { source: 'network'; res: Response };

        let fallbackWinner: FallbackWinner | null = null;
        try {
            fallbackWinner = await Promise.race([
                cachePromise.then((res) => ({ source: 'cache' as const, res })),
                networkPromise.then((res) => ({ source: 'network' as const, res }))
            ]);

            if (request.signal?.aborted) {
                throw new DOMException('The user aborted a request.', 'AbortError');
            }

            if (fallbackWinner.source === 'cache') {
                if (fallbackWinner.res) {
                    // Cache exists! Return cache immediately; fetch will renew cache in background
                    return fallbackWinner.res;
                }
                // Cache does NOT exist, wait for fetch to return and add to cache
                return await networkPromise;
            } else {
                // networkPromise won the race! Return network response
                return fallbackWinner.res;
            }
        } catch (err: any) {
            if (request.signal?.aborted || this.isAbortError(err)) {
                throw err;
            }
            if (fallbackWinner && fallbackWinner.source === 'cache') {
                if (fallbackWinner.res) {
                    // Cache exists! Return cache immediately; fetch will renew cache in background
                    return fallbackWinner.res;
                }
            }
            const cachedResponse = await cachePromise;
            if (cachedResponse) {
                return cachedResponse;
            }
            return await networkPromise;
        }
    }

    private triggerDebouncedBackgroundUpdate(request: Request, handler: StrategyHandler): void {
        if (request.signal?.aborted) return;
        const url = request.url;

        if (this.debounceMap.has(url)) {
            clearTimeout(this.debounceMap.get(url)!);
        }

        // Menggunakan global setTimeout di Service Worker scope
        const timerId = setTimeout(async () => {
            this.debounceMap.delete(url);
            try {
                // Use a clean request decoupled from the original aborted client signal
                const bgRequest = new Request(request.url, {
                    method: request.method,
                    headers: request.headers
                });
                const networkResponse = await this.fetchDeduplicated(bgRequest, handler);
                if (networkResponse && networkResponse.ok) {
                    await handler.cachePut(bgRequest, networkResponse.clone());
                }
            } catch (err) {
                console.warn('Background update failed for:', url, err);
            }
        }, this.debounceMs);

        this.debounceMap.set(url, timerId);
    }
}