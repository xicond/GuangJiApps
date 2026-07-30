// src/sw.ts
import { precacheAndRoute, cleanupOutdatedCaches } from 'workbox-precaching'
import { clientsClaim } from 'workbox-core'
import { registerRoute } from 'workbox-routing'
import { StaleWhileRevalidate, CacheFirst } from 'workbox-strategies'
import { CacheableResponsePlugin } from 'workbox-cacheable-response'
import { ExpirationPlugin } from 'workbox-expiration'

declare let self: ServiceWorkerGlobalScope


self.skipWaiting()
clientsClaim()

cleanupOutdatedCaches()

// Fallback to [] so dev mode doesn't crash on undefined __WB_MANIFEST
precacheAndRoute(self.__WB_MANIFEST || [])

// registerRoute(
//     ({ url, request }) => request.method === 'GET' && url.pathname.includes('/v1/lookup/'),
//     new CacheFirst({
//         cacheName: 'lookup-cache',
//         plugins: [
//             new CacheableResponsePlugin({
//                 statuses: [0, 200]
//             }),
//             new ExpirationPlugin({
//                 maxAgeSeconds: 24 * 60 * 60 // 24 Jam
//             })
//         ]
//     })
// )

// Match GET requests to guangji.id or /v1/ or /api/
registerRoute(
    ({ url, request }) => {
        const isMatch = request.method === 'GET' &&
            (url.pathname.includes('/v1/'))

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





