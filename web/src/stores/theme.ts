import { defineStore } from 'pinia'
import { ref, watch } from 'vue'
import { localStg } from '@/utils/storage'

export type ThemeMode = 'system' | 'light' | 'dark'
type ThemeType = 'light' | 'dark'

export const useThemeStore = defineStore('themeStore', () => {
  const savedThemeMode = localStg.getItem('themeMode')
  const savedTheme = localStg.getItem('theme')

  // 初始化 themeMode
  const mode = ref<ThemeMode>(
    savedThemeMode === 'system' || savedThemeMode === 'light' || savedThemeMode === 'dark'
      ? savedThemeMode
      : 'system',
  )
  const theme = ref<ThemeType>(
    savedTheme === 'light' || savedTheme === 'dark' ? savedTheme : 'light',
  )

  // 内部切换主题逻辑
  const applyThemeToggle = () => {
    if (mode.value === 'system') {
      mode.value = 'light'
    } else if (mode.value === 'light') {
      mode.value = 'dark'
    } else {
      mode.value = 'system'
    }

    applyTheme()
    localStg.setItem('themeMode', mode.value)
  }

  // 带扩散动画的主题切换
  const toggleTheme = async (event?: MouseEvent) => {
    // 获取点击坐标，如果没有事件则从屏幕中心扩散
    const x = event?.clientX ?? window.innerWidth / 2
    const y = event?.clientY ?? window.innerHeight / 2

    // 计算到最远角的距离（用于确定圆形大小）
    const endRadius = Math.hypot(
      Math.max(x, window.innerWidth - x),
      Math.max(y, window.innerHeight - y),
    )

    // 检查浏览器是否支持 View Transitions API
    // @ts-expect-error View Transitions API 类型支持
    if (!document.startViewTransition) {
      // 降级处理：直接切换
      applyThemeToggle()
      return
    }

    // 使用 View Transitions API
    // @ts-expect-error View Transitions API 类型支持
    const transition = document.startViewTransition(() => {
      applyThemeToggle()
    })

    await transition.ready

    // 执行圆形扩散动画
    document.documentElement.animate(
      {
        clipPath: [`circle(0px at ${x}px ${y}px)`, `circle(${endRadius}px at ${x}px ${y}px)`],
      },
      {
        duration: 400,
        easing: 'ease-out',
        pseudoElement: '::view-transition-new(root)',
      },
    )
  }

  const applyTheme = () => {
    switch (mode.value) {
      case 'system':
        theme.value = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
        break
      case 'light':
        theme.value = 'light'
        break
      case 'dark':
        theme.value = 'dark'
        break
    }

    document.documentElement.classList.remove('light', 'dark')
    document.documentElement.classList.add(theme.value)
    localStg.setItem('theme', theme.value)
  }

  const init = () => {
    applyTheme()
    // 监听系统主题变化
    watch(theme, applyTheme)
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', applyTheme)
  }

  return {
    theme,
    mode,
    toggleTheme,
    applyTheme,
    init,
  }
})
