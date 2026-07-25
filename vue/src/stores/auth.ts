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
  [key: string]: any
}

const API_BASE = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string>(localStorage.getItem('token') || '')
  const user = ref<User | null>(
    localStorage.getItem('user') ? JSON.parse(localStorage.getItem('user')!) : null
  )
  const mainMenus = ref<MainMenuItem[]>(
    localStorage.getItem('main_menu') ? JSON.parse(localStorage.getItem('main_menu')!) : []
  )

  const isAuthenticated = computed(() => !!token.value)

  function hasMenu(menuName: string): boolean {
    if (!mainMenus.value || mainMenus.value.length === 0) return false
    return mainMenus.value.some(
      (item) => item.main_menu?.trim().toLowerCase() === menuName.trim().toLowerCase()
    )
  }

  function hasSubMenu(parentMenuName: string, subMenuName: string): boolean {
    if (!mainMenus.value || mainMenus.value.length === 0) return false
    return mainMenus.value.some(
      (item) =>
        item.main_menu?.trim().toLowerCase() === parentMenuName.trim().toLowerCase() &&
      item.sub_menu.some((subItem) => subItem.level2?.trim().toLowerCase() === subMenuName.trim().toLowerCase())
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
        user.value = response.data.user || { username }
        mainMenus.value = response.data.main_menu || []
        localStorage.setItem('token', token.value)
        localStorage.setItem('user', JSON.stringify(user.value))
        localStorage.setItem('main_menu', JSON.stringify(mainMenus.value))
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
    logout
  }
})
