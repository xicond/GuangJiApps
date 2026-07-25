// src/sw.ts
import { precacheAndRoute, cleanupOutdatedCaches } from 'workbox-precaching'
import { clientsClaim } from 'workbox-core'
import { registerRoute } from 'workbox-routing'
import { StaleWhileRevalidate } from 'workbox-strategies'
import { CacheableResponsePlugin } from 'workbox-cacheable-response'

declare let self: ServiceWorkerGlobalScope


self.skipWaiting()
clientsClaim()

cleanupOutdatedCaches()

// Fallback to [] so dev mode doesn't crash on undefined __WB_MANIFEST
precacheAndRoute(self.__WB_MANIFEST || [])


// Match GET requests to guangji.id or /v1/ or /api/
registerRoute(
    ({ url, request }) => {
        const isMatch = request.method === 'GET' &&
            (url.origin.includes('guangji.id') || url.pathname.includes('/v1/') || url.pathname.includes('/api/'))

        return isMatch
    },
    new StaleWhileRevalidate({
        cacheName: 'api-cache',
        plugins: [
            new CacheableResponsePlugin({
                statuses: [0, 200]
            })
        ]
    })
)





