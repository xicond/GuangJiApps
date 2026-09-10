import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import axios from 'axios'

export interface MainMenuItem {
  menu_id: number
  main_menu: string,
  sub_menu: SubMenuItem[]
}

export interface SubMenuItem {
  menu_id: number
  parent_id: number
  level2: string,
  sequence?: string
}

export interface User {
  username?: string
  name?: string
  role?: string
  [key: string]: string | number | boolean | undefined
}

export interface LoginErrorResult extends Error {
  retryAfter?: number
  status?: number
}

const API_BASE = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string>(localStorage.getItem('token') || '')
  const safeParse = <T>(raw: string | null, fallback: T): T => {
    if (!raw) return fallback
    try {
      return JSON.parse(raw) as T
    } catch {
      return fallback
    }
  }

  const user = ref<User | null>(safeParse<User | null>(localStorage.getItem('user')!, null))
  const mainMenus = ref<MainMenuItem[]>(safeParse<MainMenuItem[]>(localStorage.getItem('main_menu')!, []))

  const isAuthenticated = computed(() => !!token.value)

  function hasMenu(menuName: string): boolean {
    if (!mainMenus.value || mainMenus.value.length === 0) return false
    return mainMenus.value.some(
      (item) => item.main_menu?.trim().toLowerCase() === menuName.trim().toLowerCase()
    )
  }

  function hasSubMenu(parentMenuName: string, subMenuName: string): boolean {
    if (!mainMenus.value || mainMenus.value.length === 0) return false
    const parent = parentMenuName.trim().toLowerCase()
    const sub = subMenuName.trim().toLowerCase()
    if (parent === 'master data' && sub === 'kelas') {
      return hasMenu('Master Data')
    }
    return mainMenus.value.some(
      (item) =>
        item.main_menu?.trim().toLowerCase() === parent &&
        item.sub_menu.some((subItem) => subItem.level2?.trim().toLowerCase() === sub)
    )
  }

  async function login(username: string, password: string): Promise<boolean> {
    try {
      const response = await axios.post(`${API_BASE}/login`, {
        username,
        password
      })

      if (response.data && response.data.token) {
        token.value = response.data.token
        const userData = response.data.user || {}
        user.value = {
          ...userData,
          username: userData.username || userData.Username || username
        }
        mainMenus.value = response.data.main_menu || []
        localStorage.setItem('token', token.value)
        localStorage.setItem('user', JSON.stringify(user.value))
        localStorage.setItem('main_menu', JSON.stringify(mainMenus.value))
        return true
      }
      return false
    } catch (error: unknown) {
      console.error('Login error:', error)
      if (axios.isAxiosError(error) && error.response) {
        const status = error.response.status
        const data = error.response.data as { error?: string; retry_after?: number } | undefined
        const retryAfterHeader = error.response.headers['retry-after'] || error.response.headers['x-retry-after']
        let retryAfter: number | undefined
        if (typeof data?.retry_after === 'number') {
          retryAfter = data.retry_after
        } else if (retryAfterHeader) {
          retryAfter = parseInt(String(retryAfterHeader), 10)
        }

        const msg = data?.error || (error instanceof Error ? error.message : 'Login failed')
        const errObj: LoginErrorResult = new Error(msg)
        errObj.retryAfter = retryAfter
        errObj.status = status
        throw errObj
      }
      throw new Error(error instanceof Error ? error.message : 'Login failed')
    }
  }

  async function changePassword(oldPassword: string, newPassword: string): Promise<boolean> {
    try {
      const response = await axios.post(`${API_BASE}/v1/change-password`, {
        username: user.value?.username,
        old_password: oldPassword,
        new_password: newPassword
      }, {
        headers: {
          Authorization: `Bearer ${token.value}`
        }
      })
      return response.status === 200
    } catch (error: unknown) {
      console.error('Change password error:', error)
      const msg = (axios.isAxiosError(error) && error.response?.data?.error) || (error instanceof Error ? error.message : 'Failed to change password')
      throw new Error(msg)
    }
  }

  function logout() {
    token.value = ''
    user.value = null
    mainMenus.value = []
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    localStorage.removeItem('main_menu')
  }

  return {
    token,
    user,
    mainMenus,
    isAuthenticated,
    hasMenu,
    hasSubMenu,
    login,
    changePassword,
    logout
  }
})
