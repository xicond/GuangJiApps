import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import axios from 'axios'

export interface User {
  username?: string
  name?: string
  role?: string
  [key: string]: any
}

const API_BASE = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string>(localStorage.getItem('token') || '')
  const user = ref<User | null>(
    localStorage.getItem('user') ? JSON.parse(localStorage.getItem('user')!) : null
  )

  const isAuthenticated = computed(() => !!token.value)

  async function login(username: string, password: string): Promise<boolean> {
    try {
      const response = await axios.post(`${API_BASE}/login`, {
        username,
        password
      })

      if (response.data && response.data.token) {
        token.value = response.data.token
        user.value = response.data.user || { username }
        localStorage.setItem('token', token.value)
        localStorage.setItem('user', JSON.stringify(user.value))
        return true
      }
      return false
    } catch (error: any) {
      console.error('Login error:', error)
      const msg = error.response?.data?.error || error.message || 'Login failed'
      throw new Error(msg)
    }
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  return {
    token,
    user,
    isAuthenticated,
    login,
    logout
  }
})
