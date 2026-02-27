import { defineStore } from 'pinia'
import { ref } from 'vue'
import { localStg } from '@/utils/storage'

export const useZenStore = defineStore('zenStore', () => {
  const isZenMode = ref<boolean>(localStg.getItem<boolean>('zenMode') ?? false)
  let isTransitioning = false
  const isMobileViewport = () =>
    typeof window !== 'undefined' && window.matchMedia('(max-width: 639px)').matches
  const mobileQuery = typeof window !== 'undefined' ? window.matchMedia('(max-width: 639px)') : null

  const setZenMode = (value: boolean) => {
    const nextValue = isMobileViewport() ? false : value
    isZenMode.value = nextValue
    localStg.setItem('zenMode', nextValue)
  }

  if (isZenMode.value && isMobileViewport()) {
    setZenMode(false)
  }

  mobileQuery?.addEventListener('change', (event) => {
    if (event.matches && isZenMode.value) {
      setZenMode(false)
    }
  })

  const toggleZenMode = async () => {
    if (isTransitioning) return
    if (!isZenMode.value && isMobileViewport()) return

    const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches
    if (prefersReducedMotion) {
      setZenMode(!isZenMode.value)
      return
    }

    isTransitioning = true
    const root = document.documentElement

    try {
      const fadeOut = root.animate([{ opacity: 1 }, { opacity: 0.9 }], {
        duration: 360,
        easing: 'cubic-bezier(0.22, 1, 0.36, 1)',
        fill: 'forwards',
      })
      await fadeOut.finished

      setZenMode(!isZenMode.value)

      const fadeIn = root.animate([{ opacity: 0.9 }, { opacity: 1 }], {
        duration: 380,
        easing: 'cubic-bezier(0.22, 1, 0.36, 1)',
        fill: 'forwards',
      })
      await fadeIn.finished
    } finally {
      root.style.opacity = ''
      isTransitioning = false
    }
  }

  return {
    isZenMode,
    setZenMode,
    toggleZenMode,
  }
})
