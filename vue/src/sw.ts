// src/sw.ts
import { precacheAndRoute, cleanupOutdatedCaches } from 'workbox-precaching'
import { clientsClaim } from 'workbox-core'
import { registerRoute } from 'workbox-routing'
import {
    StaleWhileRevalidate, CacheFirst,/* , NetworkOnly */
    NetworkFirst
} from 'workbox-strategies'
import { CacheableResponsePlugin } from 'workbox-cacheable-response'
import { ExpirationPlugin } from 'workbox-expiration'
import {
    DynamicNetworkCacheStrategy
} from './strategies/dynamicnetworkcache'
// import { BackgroundSyncPlugin } from 'workbox-background-sync'
declare let self: ServiceWorkerGlobalScope


self.skipWaiting()
clientsClaim()

cleanupOutdatedCaches()

// Fallback to [] so dev mode doesn't crash on undefined __WB_MANIFEST
precacheAndRoute(self.__WB_MANIFEST || [])

registerRoute(
    ({ url, request }) => request.method === 'GET' && url.pathname.includes('/v1/lookup/'),
    new DynamicNetworkCacheStrategy({
        cacheName: 'lookup-cache',
        timeoutMs: 280,
        debounceMs: 2000,
        plugins: [
            new CacheableResponsePlugin({
                statuses: [0, 200]
            }),
            new ExpirationPlugin({
                maxAgeSeconds: 7 * 60 * 60 // 24 Jam
            })
        ]
    })
)

registerRoute(
    ({ url, request }) => request.method === 'GET' && (url.pathname.includes('/v1/lookup/') || url.pathname.includes('/umat')),
    new DynamicNetworkCacheStrategy({
        cacheName: 'api-cache',
        timeoutMs: 220,
        debounceMs: 2200,
        plugins: [
            new CacheableResponsePlugin({
                statuses: [0, 200]
            }),
            new ExpirationPlugin({
                maxAgeSeconds: 7 * 60 * 60 // 1 week
            })
        ]
    })
)

registerRoute(
    ({ url, request }) => request.method === 'GET' && url.pathname.includes('/v1/') && url.pathname.includes('/report'),
    new DynamicNetworkCacheStrategy({
        cacheName: 'report-cache',
        timeoutMs: 2200,
        debounceMs: 10000,
        plugins: [
            new CacheableResponsePlugin({
                statuses: [0, 200]
            }),
            new ExpirationPlugin({
                maxAgeSeconds: 2 * 60 * 60 // 2 Jam
            })
        ]
    })
)

// Match GET requests to /v1/ or /api/
registerRoute(
    ({ url, request }) => {
        const isMatch = request.method === 'GET' &&
            (url.pathname.includes('/v1/'))

        return isMatch
    },
    new NetworkFirst({
        cacheName: 'api-cache',
        plugins: [
            new CacheableResponsePlugin({
                statuses: [0, 200]
            }),
            new ExpirationPlugin({
                maxAgeSeconds: 1 * 60 * 60 // 1 Jam
            })
        ]
    })
)



// const bgSyncPlugin = new BackgroundSyncPlugin('api-queue', {
//   maxRetentionTime: 24 * 60, // Waktu maksimal request disimpan dalam antrean (dalam menit = 24 jam)
// })

// // 2. Daftarkan route untuk method POST/PUT/DELETE
// registerRoute(
//   ({ url, request }) => {
//     return (request.method === 'POST' || request.method === 'PUT') &&
//            url.pathname.startsWith('/v1/submit')
//   },
//   new NetworkOnly({
//     plugins: [bgSyncPlugin],
//   })
// )

