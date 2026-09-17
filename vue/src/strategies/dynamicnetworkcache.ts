import { Strategy, StrategyOptions, StrategyHandler } from 'workbox-strategies';

export interface DynamicNetworkCacheStrategyOptions extends StrategyOptions {
    timeoutMs?: number;
    debounceMs?: number;
}

export class DynamicNetworkCacheStrategy extends Strategy {
    private timeoutMs: number;
    protected inFlightRequests: Map<string, Promise<Response>>;

    constructor(options: DynamicNetworkCacheStrategyOptions = {}) {
        super(options);
        this.timeoutMs = options.timeoutMs ?? 500;
        this.inFlightRequests = new Map<string, Promise<Response>>();
    }

    /**
     * Executes network fetch and cache with deduplication.
     * If an identical GET/HEAD request is already in-flight, returns a clone of the existing promise.
     */
    protected fetchDeduplicated(request: Request, handler: StrategyHandler): Promise<Response> {
        if (request.method !== 'GET' && request.method !== 'HEAD') {
            return handler.fetchAndCachePut(request);
        }

        const key = `${request.method}:${request.url}`;

        const existingPromise = this.inFlightRequests.get(key);
        if (existingPromise) {
            return existingPromise.then((response) => response.clone());
        }

        let cleanupTimer: ReturnType<typeof setTimeout> | undefined;

        const fetchPromise = handler.fetchAndCachePut(request)
            .finally(() => {
                if (cleanupTimer) clearTimeout(cleanupTimer);
                this.inFlightRequests.delete(key);
            });

        // Safety timeout: ensure inFlightRequests key is deleted within 15s no matter what
        cleanupTimer = setTimeout(() => {
            this.inFlightRequests.delete(key);
        }, 15000);

        // Attach passive catch handler so inFlightRequests never causes unhandledrejection in SW
        fetchPromise.catch(() => { });

        this.inFlightRequests.set(key, fetchPromise);

        return fetchPromise;
    }

    protected async _handle(request: Request, handler: StrategyHandler): Promise<Response> {
        // 1. Start deduplicated network fetch + automatic background cache update via Workbox
        const networkFetchPromise = this.fetchDeduplicated(request, handler);

        // 2. Setup timeoutMs timer promise
        let timeoutTimer: ReturnType<typeof setTimeout> | undefined;
        const timeoutPromise = new Promise<'TIMEOUT'>((resolve) => {
            timeoutTimer = setTimeout(() => resolve('TIMEOUT'), this.timeoutMs);
        });

        // 3. Race network fetch against timeoutMs
        try {
            const raceResult = await Promise.race([
                networkFetchPromise,
                timeoutPromise
            ]);

            clearTimeout(timeoutTimer);

            if (raceResult !== 'TIMEOUT') {
                // Network fetch finished BEFORE timeoutMs! Return network response directly.
                return raceResult;
            }
        } catch (err: any) {
            clearTimeout(timeoutTimer);
            if (request.signal?.aborted || err?.name === 'AbortError' || (err?.message && String(err.message).toLowerCase().includes('aborted'))) {
                throw err;
            }
            // Network fetch failed before timeoutMs -> try cache fallback
            const cachedResponse = await handler.cacheMatch(request).catch(() => undefined);
            if (cachedResponse) {
                return cachedResponse;
            }
            throw err;
        }

        // 4. Network did NOT finish within timeoutMs (timed out): Check cache fallback
        const cachedResponse = await handler.cacheMatch(request).catch(() => undefined);

        if (cachedResponse) {
            // Cache hit! Return cached response immediately (< 500ms response time).
            // networkFetchPromise is already running fetchAndCachePut in background with handler.waitUntil!
            return cachedResponse;
        }

        // 5. Cache miss after timeoutMs: Wait for network fetch to finish
        try {
            return await networkFetchPromise;
        } catch (err: any) {
            if (request.signal?.aborted || err?.name === 'AbortError' || (err?.message && String(err.message).toLowerCase().includes('aborted'))) {
                throw err;
            }
            throw err;
        }
    }
}