import { defineStore } from 'pinia'
import { ref, computed, watch } from 'vue'
import axios from 'axios'

export interface MainMenuItem {
  menu_id: number
  main_menu: string
  sub_menu: SubMenuItem[]
}

export interface SubMenuItem {
  menu_id: number
  parent_id: number
  level2: string
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

const STORAGE_KEYS = {
  TOKEN: 'token',
  REFRESH_TOKEN: 'refresh_token',
  USER: 'user',
  MAIN_MENU: 'main_menu'
} as const

const API_BASE = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

export const useAuthStore = defineStore('auth', () => {
  const getStorageItem = (key: string): string => {
    if (typeof localStorage === 'undefined') return ''
    return localStorage.getItem(key) || ''
  }

  const safeParse = <T>(raw: string | null, fallback: T): T => {
    if (!raw) return fallback
    try {
      return JSON.parse(raw) as T
    } catch {
      return fallback
    }
  }

  const token = ref<string>(getStorageItem(STORAGE_KEYS.TOKEN))
  const refreshToken = ref<string>(getStorageItem(STORAGE_KEYS.REFRESH_TOKEN))
  const user = ref<User | null>(safeParse<User | null>(getStorageItem(STORAGE_KEYS.USER) || null, null))
  const mainMenus = ref<MainMenuItem[]>(safeParse<MainMenuItem[]>(getStorageItem(STORAGE_KEYS.MAIN_MENU) || null, []))

  const isAuthenticated = computed(() => !!token.value)

  function setToken(newToken: string) {
    token.value = newToken
    if (typeof localStorage !== 'undefined') {
      if (newToken) {
        localStorage.setItem(STORAGE_KEYS.TOKEN, newToken)
      } else {
        localStorage.removeItem(STORAGE_KEYS.TOKEN)
      }
    }
  }

  function setRefreshToken(newRefreshToken: string) {
    refreshToken.value = newRefreshToken
    if (typeof localStorage !== 'undefined') {
      if (newRefreshToken) {
        localStorage.setItem(STORAGE_KEYS.REFRESH_TOKEN, newRefreshToken)
      } else {
        localStorage.removeItem(STORAGE_KEYS.REFRESH_TOKEN)
      }
    }
  }

  function setTokens(newToken: string, newRefreshToken?: string) {
    setToken(newToken)
    if (newRefreshToken !== undefined) {
      setRefreshToken(newRefreshToken)
    }
  }

  function setUser(newUser: User | null) {
    user.value = newUser
    if (typeof localStorage !== 'undefined') {
      if (newUser) {
        localStorage.setItem(STORAGE_KEYS.USER, JSON.stringify(newUser))
      } else {
        localStorage.removeItem(STORAGE_KEYS.USER)
      }
    }
  }

  function setMainMenus(menus: MainMenuItem[]) {
    mainMenus.value = menus
    if (typeof localStorage !== 'undefined') {
      if (menus && menus.length > 0) {
        localStorage.setItem(STORAGE_KEYS.MAIN_MENU, JSON.stringify(menus))
      } else {
        localStorage.removeItem(STORAGE_KEYS.MAIN_MENU)
      }
    }
  }

  watch(token, (val) => {
    if (typeof localStorage === 'undefined') return
    if (val) {
      if (localStorage.getItem(STORAGE_KEYS.TOKEN) !== val) {
        localStorage.setItem(STORAGE_KEYS.TOKEN, val)
      }
    } else {
      localStorage.removeItem(STORAGE_KEYS.TOKEN)
    }
  })

  watch(refreshToken, (val) => {
    if (typeof localStorage === 'undefined') return
    if (val) {
      if (localStorage.getItem(STORAGE_KEYS.REFRESH_TOKEN) !== val) {
        localStorage.setItem(STORAGE_KEYS.REFRESH_TOKEN, val)
      }
    } else {
      localStorage.removeItem(STORAGE_KEYS.REFRESH_TOKEN)
    }
  })

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
        setToken(response.data.token)
        if (response.data.refresh_token) {
          setRefreshToken(response.data.refresh_token)
        }
        const userData = response.data.user || {}
        setUser({
          ...userData,
          username: userData.username || userData.Username || username
        })
        setMainMenus(response.data.main_menu || [])
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
    refreshToken.value = ''
    user.value = null
    mainMenus.value = []
    if (typeof localStorage !== 'undefined') {
      localStorage.removeItem(STORAGE_KEYS.TOKEN)
      localStorage.removeItem(STORAGE_KEYS.REFRESH_TOKEN)
      localStorage.removeItem(STORAGE_KEYS.USER)
      localStorage.removeItem(STORAGE_KEYS.MAIN_MENU)
    }
  }

  return {
    token,
    refreshToken,
    user,
    mainMenus,
    isAuthenticated,
    hasMenu,
    hasSubMenu,
    setToken,
    setRefreshToken,
    setTokens,
    setUser,
    setMainMenus,
    login,
    changePassword,
    logout
  }
})
