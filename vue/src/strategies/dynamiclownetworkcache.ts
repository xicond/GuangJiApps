
import { StrategyHandler } from 'workbox-strategies';
import {
    DynamicNetworkCacheStrategyOptions, DynamicNetworkCacheStrategy
} from './dynamicnetworkcache';



export interface DynamicLowNetworkCacheStrategyOptions extends DynamicNetworkCacheStrategyOptions {
    lowNetworkDebounceMs?: number;
}

export class DynamicLowNetworkCacheStrategy extends DynamicNetworkCacheStrategy {
    private lowNetworkDebounceMs: number;
    private lastFetchMap: Map<string, number>;

    constructor(options: DynamicLowNetworkCacheStrategyOptions = {}) {
        super(options);
        // Default 30 menit (30 * 60 * 1000 ms)
        this.lowNetworkDebounceMs = options.lowNetworkDebounceMs ?? 30 * 60 * 1000;
        this.lastFetchMap = new Map<string, number>();
    }

    private isLowNetwork(): boolean {
        if (typeof navigator === 'undefined') return false;
        const nav = navigator as unknown as {
            connection?: {
                effectiveType?: string;
                saveData?: boolean;
            };
            mozConnection?: {
                effectiveType?: string;
                saveData?: boolean;
            };
            webkitConnection?: {
                effectiveType?: string;
                saveData?: boolean;
            };
        };
        const conn = nav.connection || nav.mozConnection || nav.webkitConnection;
        if (!conn) return false;

        if (conn.saveData === true) {
            return true;
        }

        const lowTypes = ['slow-2g', '2g', '3g'];
        if (conn.effectiveType && lowTypes.includes(conn.effectiveType.toLowerCase())) {
            return true;
        }

        return false;
    }

    protected async _handle(request: Request, handler: StrategyHandler): Promise<Response> {
        if (!this.isLowNetwork()) {
            return await super._handle(request, handler);
        }

        // Low network mode:
        // 1. Prioritaskan cache jika hadir
        const cachedResponse = await handler.cacheMatch(request).catch(() => undefined);

        if (cachedResponse) {
            const lastFetch = this.lastFetchMap.get(request.url) ?? 0;
            const now = Date.now();

            // Background fetch hanya jika jeda debounce (default 30m) telah terpenuhi
            if (now - lastFetch >= this.lowNetworkDebounceMs) {
                this.lastFetchMap.set(request.url, now);
                const backgroundFetch = (async () => {
                    try {
                        const networkResponse = await this.fetchDeduplicated(request, handler);
                        if (networkResponse && networkResponse.ok) {
                            await handler.cachePut(request, networkResponse.clone());
                        }
                    } catch (err) {
                        // Fail silently for background update on low network
                    }
                })();

                if (typeof handler.waitUntil === 'function') {
                    handler.waitUntil(backgroundFetch);
                }
            }

            return cachedResponse;
        }

        // 2. Jika cache tidak hadir, baru fetch dari network
        const networkResponse = await this.fetchDeduplicated(request, handler);
        if (networkResponse && networkResponse.ok) {
            const putPromise = handler.cachePut(request, networkResponse.clone()).catch(() => { });
            if (typeof handler.waitUntil === 'function') {
                handler.waitUntil(putPromise);
            }
            this.lastFetchMap.set(request.url, Date.now());
        }

        return networkResponse;
    }
}