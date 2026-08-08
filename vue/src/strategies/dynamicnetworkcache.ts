import { Strategy, StrategyOptions, StrategyHandler } from 'workbox-strategies';

export interface DynamicNetworkCacheStrategyOptions extends StrategyOptions {
    timeoutMs?: number;
    debounceMs?: number;
}

export class DynamicNetworkCacheStrategy extends Strategy {
    private timeoutMs: number;
    private debounceMs: number;
    private debounceMap: Map<string, number>; // Diubah ke number (tipe setTimeout di browser/SW)

    constructor(options: DynamicNetworkCacheStrategyOptions = {}) {
        super(options);
        this.timeoutMs = options.timeoutMs ?? 500;
        this.debounceMs = options.debounceMs ?? 5000;
        this.debounceMap = new Map<string, number>();
    }

    protected async _handle(request: Request, handler: StrategyHandler): Promise<Response> {
        // console.log('[SW DynamicNetworkCacheStrategy] Intercepting:', request.method, request.url);
        // 1. Check cache and fetch network asynchronously in parallel without initial await
        const cachePromise: Promise<Response | undefined> = handler.cacheMatch(request).catch(() => undefined);
        const networkPromise: Promise<Response> = handler.fetch(request);

        // Attach non-blocking background cache update once network fetch completes successfully
        networkPromise.then((networkResponse) => {
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

            // If network finished BEFORE timeoutMs, return fast network response
            if (!result.isTimeout && result.res) {
                clearTimeout(timeoutTimer);
                return result.res;
            }
        } catch (err) {
            // Network fetch failed or errored before timeoutMs
            this.triggerDebouncedBackgroundUpdate(request, handler);
        } finally {
            clearTimeout(timeoutTimer);
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
        } catch (err) {
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
        const url = request.url;

        if (this.debounceMap.has(url)) {
            clearTimeout(this.debounceMap.get(url)!);
        }

        // Menggunakan global setTimeout di Service Worker scope
        const timerId = setTimeout(async () => {
            this.debounceMap.delete(url);
            try {
                const networkResponse = await handler.fetch(request);
                if (networkResponse && networkResponse.ok) {
                    await handler.cachePut(request, networkResponse.clone());
                }
            } catch (err) {
                console.warn('Background update failed for:', url, err);
            }
        }, this.debounceMs) as number;

        this.debounceMap.set(url, timerId);
    }
}