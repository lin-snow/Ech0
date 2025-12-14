import { defineStore } from 'pinia'
import { ref, computed, watch } from 'vue'
import { localStg } from '@/utils/storage'

export type ThemeMode = 'light' | 'dark' | 'auto'
type ThemeType = 'light' | 'dark'

const isValidThemeMode = (value: unknown): value is ThemeMode => {
  return value === 'light' || value === 'dark' || value === 'auto'
}

export const useThemeStore = defineStore('themeStore', () => {
  const savedMode = localStg.getItem('themeMode')
  const mode = ref<ThemeMode>(isValidThemeMode(savedMode) ? savedMode : 'auto')
  
  const systemTheme = ref<ThemeType>(
    window.matchMedia?.('(prefers-color-scheme: dark)').matches ? 'dark' : 'light',
  )

  const theme = computed<ThemeType>(() => {
    return mode.value === 'auto' ? systemTheme.value : mode.value
  })

  const toggleTheme = () => {
    mode.value = mode.value === 'light' ? 'dark' : mode.value === 'dark' ? 'auto' : 'light'
    localStg.setItem('themeMode', mode.value)
  }

  const applyTheme = (t: ThemeType) => {
    document.documentElement.classList.remove('light', 'dark')
    document.documentElement.classList.add(t)
  }

  const init = () => {
    applyTheme(theme.value)
    watch(theme, applyTheme)

    if (window.matchMedia) {
      const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
      mediaQuery.addEventListener('change', (e: MediaQueryListEvent) => {
        systemTheme.value = e.matches ? 'dark' : 'light'
      })
    }
  }

  return {
    theme,
    mode,
    toggleTheme,
    applyTheme,
    init,
  }
})
