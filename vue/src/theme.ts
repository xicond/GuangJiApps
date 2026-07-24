import { ref } from 'vue'

export type ThemeMode = 'system' | 'dark' | 'light'

const mediaQuery = typeof window !== 'undefined' ? window.matchMedia('(prefers-color-scheme: dark)') : null
export const currentMode = ref<ThemeMode>((localStorage.getItem('theme_mode') as ThemeMode) || 'system')
export const isDark = ref<boolean>(false)

export function updateTheme() {
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

export function initTheme() {
  updateTheme()
  if (mediaQuery) {
    mediaQuery.addEventListener('change', () => {
      if (currentMode.value === 'system') {
        updateTheme()
      }
    })
  }
}

export function setMode(mode: ThemeMode) {
  currentMode.value = mode
  localStorage.setItem('theme_mode', mode)
  updateTheme()
}
