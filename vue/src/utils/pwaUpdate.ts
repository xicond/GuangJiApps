import { registerSW } from 'virtual:pwa-register'
import type { Router } from 'vue-router'

let updateSW: ((reloadPage?: boolean) => Promise<void>) | undefined
let hasPendingUpdate = false

export function initPwaUpdate(router: Router) {
  if (typeof window !== 'undefined' && 'serviceWorker' in navigator) {
    updateSW = registerSW({
      immediate: true,
      onNeedRefresh() {
        // console.log('[PWA Update] Asset update detected')
        hasPendingUpdate = true
      },
      onOfflineReady() {
        // console.log('[PWA Update] App ready offline')
      }
    })

    // Listen for SW_UPDATED message from Service Worker activate event
    navigator.serviceWorker.addEventListener('message', (event) => {
      if (event.data && event.data.type === 'SW_UPDATED') {
        hasPendingUpdate = true
      }
    })

    // Apply pending SW update seamlessly on route change (page navigation)
    router.beforeEach((_to, from, next) => {
      if (from.name && hasPendingUpdate && updateSW) {
        hasPendingUpdate = false
        updateSW(true).catch(() => {
          window.location.reload()
        })
        return
      }
      next()
    })

    // Check for SW asset updates on every page navigation
    router.afterEach(() => {
      if (navigator.serviceWorker.controller) {
        navigator.serviceWorker.getRegistration().then((reg) => {
          reg?.update().catch(() => {})
        })
      }
    })
  }
}
