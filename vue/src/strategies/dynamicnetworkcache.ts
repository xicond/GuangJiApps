import { Strategy, StrategyOptions } from 'workbox-strategies';

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

    // Menggunakan 'any' untuk handler guna menghindari isu versi workbox-core
    protected async _handle(request: Request, handler: any): Promise<Response> {
        const cachedResponse = await handler.cacheMatch(request);

        if (!cachedResponse) {
            try {
                const networkResponse = await handler.fetch(request);
                if (networkResponse && networkResponse.ok) {
                    await handler.cachePut(request, networkResponse.clone());
                }
                return networkResponse;
            } catch (error) {
                throw error;
            }
        }

        try {
            const networkPromise = handler.fetch(request);
            const timeoutPromise = new Promise<never>((_, reject) =>
                setTimeout(() => reject(new Error('Network timeout')), this.timeoutMs)
            );

            const networkResponse = await Promise.race([networkPromise, timeoutPromise]);

            if (networkResponse && networkResponse.ok) {
                await handler.cachePut(request, networkResponse.clone());
            }
            return networkResponse;
        } catch (error) {
            this.triggerDebouncedBackgroundUpdate(request, handler);
            return cachedResponse;
        }
    }

    private triggerDebouncedBackgroundUpdate(request: Request, handler: any): void {
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
        }, this.debounceMs) as unknown as number;

        this.debounceMap.set(url, timerId);
    }
}