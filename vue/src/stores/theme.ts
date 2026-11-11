import { defineStore } from 'pinia'
import { ref } from 'vue'

export type ThemeMode = 'system' | 'dark' | 'light'

const STORAGE_KEY = 'theme_mode'

export const useThemeStore = defineStore('theme', () => {
  const getStoredMode = (): ThemeMode => {
    if (typeof localStorage === 'undefined') return 'system'
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored === 'system' || stored === 'dark' || stored === 'light') {
      return stored
    }
    return 'system'
  }

  const mediaQuery = typeof window !== 'undefined' ? window.matchMedia('(prefers-color-scheme: dark)') : null
  const currentMode = ref<ThemeMode>(getStoredMode())
  const isDark = ref<boolean>(false)

  function updateTheme() {
    if (typeof document === 'undefined') return
    if (!mediaQuery) return
    if (currentMode.value === 'system') {
      isDark.value = mediaQuery.matches
    } else {
      isDark.value = currentMode.value === 'dark'
    }

    if (isDark.value) {
      document.documentElement.classList.add('dark')
    } else {
      document.documentElement.classList.remove('dark')
    }
  }

  let mediaQueryListenerAttached = false

  function initTheme() {
    updateTheme()
    if (mediaQuery && !mediaQueryListenerAttached) {
      mediaQueryListenerAttached = true
      mediaQuery.addEventListener('change', () => {
        if (currentMode.value === 'system') {
          updateTheme()
        }
      })
    }
  }

  function setMode(mode: ThemeMode) {
    currentMode.value = mode
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem(STORAGE_KEY, mode)
    }
    updateTheme()
  }

  return {
    currentMode,
    isDark,
    updateTheme,
    initTheme,
    setMode
  }
})
