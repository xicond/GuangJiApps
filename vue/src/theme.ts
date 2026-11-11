import { useThemeStore, type ThemeMode } from './stores/theme'

export type { ThemeMode }

export function initTheme() {
  const themeStore = useThemeStore()
  themeStore.initTheme()
}

export function updateTheme() {
  const themeStore = useThemeStore()
  themeStore.updateTheme()
}

export function setMode(mode: ThemeMode) {
  const themeStore = useThemeStore()
  themeStore.setMode(mode)
}

export { useThemeStore }
