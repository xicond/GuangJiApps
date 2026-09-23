import axios, { type InternalAxiosRequestConfig, type AxiosRequestConfig } from 'axios'
import { useAuthStore } from '../stores/auth'
import router from '../router'
import { useQuickStreamSpeedTest } from './useQuickStreamSpeedTest'

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
  (config) => {
    const authStore = useAuthStore()
    const token = authStore.token || localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
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
      if (config.fetchOptions?.priority !== 'low' && canRunTest()) {
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

apiClient.interceptors.response.use(
  (response) => {
    clearRequestTimer(response.config)
    return response
  },
  (error) => {
    clearRequestTimer(error.config)
    if (error.response && error.response.status === 401) {
      if (error.response.data?.error === 'invalid token') {
        error.response.data.error = 'login expired'
      }
      const authStore = useAuthStore()
      authStore.logout()
      if (router.currentRoute.value.name !== 'login') {
        router.push({ name: 'login' })
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
