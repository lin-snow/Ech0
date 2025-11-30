import { defineStore } from 'pinia'
import { ref, computed, watch } from 'vue'
import { localStg } from '@/utils/storage'

type ThemeMode = 'light' | 'dark' | 'auto'
type ThemeType = 'light' | 'dark'

export const useThemeStore = defineStore('themeStore', () => {
  const savedMode = localStg.getItem('themeMode')
  const mode = ref<ThemeMode>(
    savedMode === 'light' || savedMode === 'dark' || savedMode === 'auto' ? savedMode : 'auto',
  )

  const systemTheme = ref<ThemeType>(
    window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches
      ? 'dark'
      : 'light',
  )

  const theme = computed<ThemeType>(() => {
    return mode.value === 'auto' ? systemTheme.value : mode.value
  })

  const toggleTheme = () => {
    if (mode.value === 'light') {
      mode.value = 'dark'
    } else if (mode.value === 'dark') {
      mode.value = 'auto'
    } else {
      mode.value = 'light'
    }
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
      mediaQuery.addEventListener('change', (e) => {
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
