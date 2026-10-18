import axios, { type InternalAxiosRequestConfig, type AxiosRequestConfig } from 'axios'
import { useAuthStore } from '../stores/auth'
import router from '../router'
import { useQuickStreamSpeedTest } from './useQuickStreamSpeedTest'
import { renewBiometricSession } from '../services/passkey'

declare module 'axios' {
  export interface InternalAxiosRequestConfig {
    fetchOptions?: Record<string, any> | Omit<RequestInit, 'body' | 'signal' | 'headers' | 'method'>
    metadata?: {
      requestId: string
      startTime: number
    }
  }
  export interface AxiosRequestConfig {
    fetchOptions?: Record<string, any> | Omit<RequestInit, 'body' | 'signal' | 'headers' | 'method'>
    metadata?: {
      requestId: string
      startTime: number
    }
  }
}

const API_BASE = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

export const apiClient = axios.create({
  baseURL: API_BASE,
  adapter: 'fetch',
  headers: {
    'Content-Type': 'application/json'
  }
})

const { canRunTest,
  measureSpeedInOneSecond } = useQuickStreamSpeedTest()

// Map untuk menyimpan referensi timer tiap request aktif
const activeTimers = new Map<string, ReturnType<typeof setTimeout>>()

apiClient.interceptors.request.use(
  async (config) => {
    // Jika proses biometric renewal sedang berlangsung, tunggu sampai selesai
    // agar request ini langsung memakai token baru tanpa memicu 401 tambahan
    if (isRenewing && renewPromise) {
      try {
        await renewPromise
      } catch {
        // Renewal gagal, biarkan request berjalan (atau ditolak oleh response interceptor)
      }
    }

    const authStore = useAuthStore()
    const token = (authStore.token || localStorage.getItem('token') || '').trim()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    } else if (config.headers.Authorization) {
      delete config.headers.Authorization
    }

    const requestId = Math.random().toString(36).substring(2, 9)
    config.metadata = { requestId, startTime: Date.now() }

    const signal = config.signal as AbortSignal | undefined
    if (signal) {
      const onAbort = () => {
        if (typeof navigator !== 'undefined' && 'serviceWorker' in navigator) {
          const fullUrl = apiClient.getUri(config)
          const msg = {
            type: 'ABORT_REQUEST',
            url: fullUrl
          }
          if (navigator.serviceWorker.controller) {
            navigator.serviceWorker.controller.postMessage(msg)
          } else {
            navigator.serviceWorker.ready.then((reg) => {
              reg.active?.postMessage(msg)
            }).catch(() => { })
          }
        }
      }
      if (signal.aborted) {
        onAbort()
      } else if (typeof signal.addEventListener === 'function') {
        signal.addEventListener('abort', onAbort, { once: true })
      }
    }

    const timer = setTimeout(() => {
      // Pastikan speed test tidak sedang berjalan dan sudah melewati throttle 5s
      const url = config.url || ''
      const isHeavyDataEndpoint = url.includes('/report') || url.includes('/excel') || url.includes('/download') || url.includes('/export')
      if (!isHeavyDataEndpoint && config.fetchOptions?.priority !== 'low' && canRunTest()) {
        console.warn(`[Network Monitor] Request ke ${config.url} berjalan >= 1s. Memicu speed test...`)
        measureSpeedInOneSecond()
      }
    }, 1000)

    activeTimers.set(requestId, timer)
    if (['POST', 'PUT', 'PATCH', 'DELETE'].includes(config.method || 'GET')) {
      config.fetchOptions = { priority: 'high' }
    }
    return config
  },
  (error) => Promise.reject(error)
)
// State untuk sinkronisasi concurrent renewal session biometrik
let isRenewing = false
let renewPromise: Promise<string> | null = null

apiClient.interceptors.response.use(
  (response) => {
    clearRequestTimer(response.config)
    return response
  },
  async (error) => {
    clearRequestTimer(error.config)

    if (error.response && error.response.status === 401) {
      if (error.response.data?.error === 'invalid token') {
        error.response.data.error = 'login expired'
      }

      const originalRequest = error.config as (InternalAxiosRequestConfig & { _retry?: boolean }) | undefined
      const url = originalRequest?.url || ''
      const isAuthEndpoint = url.includes('/login-renew') || url.includes('/login')

      // 1. Jika sedang ada proses renewal berjalan di request lain, TUNGGU proses tersebut
      // JANGAN langsung logout atau memicu prompt biometrik ganda!
      if (isRenewing && renewPromise) {
        try {
          const freshToken = await renewPromise
          if (originalRequest && freshToken) {
            originalRequest.headers.Authorization = `Bearer ${freshToken}`
            return apiClient(originalRequest)
          }
        } catch (retryErr) {
          return Promise.reject(retryErr)
        }
      }

      const refreshToken = (localStorage.getItem('refresh_token') || '').trim()

      // 2. Jika bukan endpoint auth, belum pernah di-retry, dan ada simpanan refresh token:
      // jalankan renewal biometrik secara atomik (single flight)
      if (originalRequest && !originalRequest._retry && !isAuthEndpoint && refreshToken) {
        originalRequest._retry = true

        const authStore = useAuthStore()
        const currentActiveToken = (authStore.token || localStorage.getItem('token') || '').trim()
        const requestToken = originalRequest.headers.Authorization?.toString().replace(/^Bearer\s+/i, '').trim() || ''

        // Cek jika token sudah diperbarui sebelumnya oleh request lain
        if (currentActiveToken && requestToken && currentActiveToken !== requestToken) {
          originalRequest.headers.Authorization = `Bearer ${currentActiveToken}`
          return apiClient(originalRequest)
        }

        isRenewing = true

        renewPromise = (async () => {
          try {
            const result = await renewBiometricSession(refreshToken)
            const freshToken = (result.token || '').trim()
            if (!freshToken) {
              throw new Error('Access token baru kosong dari server.')
            }
            authStore.token = freshToken
            localStorage.setItem('token', freshToken)
            if (result.refresh_token && result.refresh_token.trim()) {
              localStorage.setItem('refresh_token', result.refresh_token.trim())
            }
            return freshToken
          } catch (renewErr) {
            // Hanya logout jika autentikasi biometrik dibatalkan / gagal
            authStore.logout()
            if (router.currentRoute.value.name !== 'login') {
              router.push({ name: 'login' })
            }
            throw renewErr
          } finally {
            isRenewing = false
            renewPromise = null
          }
        })()

        try {
          const newToken = await renewPromise
          if (newToken) {
            originalRequest.headers.Authorization = `Bearer ${newToken}`
            return apiClient(originalRequest)
          }
        } catch (renewErr) {
          return Promise.reject(renewErr)
        }
      }

      // 3. Hanya lakukan fallback logout jika TIDAK ada proses renewal yang sedang berjalan
      if (!isRenewing && !renewPromise) {
        const authStore = useAuthStore()
        authStore.logout()
        if (router.currentRoute.value.name !== 'login') {
          router.push({ name: 'login' })
        }
      }
    }

    return Promise.reject(error)
  }
)

// Helper untuk membersihkan timer saat request selesai < 1s
function clearRequestTimer(config?: InternalAxiosRequestConfig | AxiosRequestConfig) {
  if (config?.metadata?.requestId) {
    const requestId = config.metadata.requestId
    const timer = activeTimers.get(requestId)
    if (timer) {
      clearTimeout(timer) // Hapus timer karena request selesai sebelum 1 detik
      activeTimers.delete(requestId)
    }
  }
}

export default apiClient
