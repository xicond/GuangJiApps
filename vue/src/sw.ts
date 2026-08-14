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

// Listen for SKIP_WAITING message from client
self.addEventListener('message', (event) => {
    if (event.data && event.data.type === 'SKIP_WAITING') {
        self.skipWaiting()
    }
})

// Notify open window clients when a new SW version is activated (assets update)
self.addEventListener('activate', (event) => {
    event.waitUntil(
        (async () => {
            await self.clients.claim()
            const clients = await self.clients.matchAll({ type: 'window' })
            for (const client of clients) {
                client.postMessage({ type: 'SW_UPDATED' })
            }
        })()
    )
})

// 1. Match GET requests to /v1/lookup/
registerRoute(
    ({ url, request }) => {
        const isMatch = request.method === 'GET' && url.pathname.includes('/v1/lookup/')
        // if (isMatch) console.log('[SW Route] Matched lookup:', url.pathname)
        return isMatch
    },
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

// 2. Match GET requests to /umat
registerRoute(
    ({ url, request }) => {
        const isMatch = request.method === 'GET' && url.pathname.includes('/umat')
        // if (isMatch) console.log('[SW Route] Matched umat:', url.pathname)
        return isMatch
    },
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

// 3. Match GET requests to /report
registerRoute(
    ({ url, request }) => {
        const isMatch = request.method === 'GET' && url.pathname.includes('/v1/') && url.pathname.includes('/report')
        // if (isMatch) console.log('[SW Route] Matched report:', url.pathname)
        return isMatch
    },
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

// 4. Catch-all GET requests to /v1/ or /api/
registerRoute(
    ({ url, request }) => {
        const isMatch = request.method === 'GET' && url.pathname.includes('/v1/')
        // if (isMatch) console.log('[SW Route] Matched /v1/ or /api/:', url.pathname)
        return isMatch
    },
    new DynamicNetworkCacheStrategy({
        cacheName: 'api-cache',
        timeoutMs: 500,
        debounceMs: 5000,
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

