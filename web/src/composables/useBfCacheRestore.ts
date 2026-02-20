import { onBeforeUnmount, onMounted, ref } from 'vue'

type UseBfCacheRestoreOptions = {
  clearDelayMs?: number
  debug?: boolean
  onRestore?: (event: PageTransitionEvent) => void
}

export function useBfCacheRestore(options: UseBfCacheRestoreOptions = {}) {
  const { clearDelayMs = 250, debug = false, onRestore } = options

  const isBfCacheRestore = ref(false)
  let restoreTimer: number | null = null
  let guardedActionTimer: number | null = null

  const clearBfCacheRestoreFlag = () => {
    if (restoreTimer !== null) {
      window.clearTimeout(restoreTimer)
      restoreTimer = null
    }
    isBfCacheRestore.value = false
  }

  const clearGuardedActionTimer = () => {
    if (guardedActionTimer !== null) {
      window.clearTimeout(guardedActionTimer)
      guardedActionTimer = null
    }
  }

  const onPageShow = (event: PageTransitionEvent) => {
    if (debug && import.meta.env.DEV) {
      console.debug('[bfcache] pageshow', { persisted: event.persisted })
    }

    if (!event.persisted) return

    isBfCacheRestore.value = true
    onRestore?.(event)

    if (restoreTimer !== null) {
      window.clearTimeout(restoreTimer)
    }
    restoreTimer = window.setTimeout(() => {
      isBfCacheRestore.value = false
      restoreTimer = null
    }, clearDelayMs)
  }

  const onPageHide = (event: PageTransitionEvent) => {
    if (debug && import.meta.env.DEV) {
      console.debug('[bfcache] pagehide', { persisted: event.persisted })
    }
  }

  const runWithBfCacheGuard = (action: () => void, delayMs = 120) => {
    if (!isBfCacheRestore.value) {
      action()
      return
    }

    clearGuardedActionTimer()
    guardedActionTimer = window.setTimeout(() => {
      isBfCacheRestore.value = false
      guardedActionTimer = null
      action()
    }, delayMs)
  }

  onMounted(() => {
    window.addEventListener('pageshow', onPageShow)
    window.addEventListener('pagehide', onPageHide)
  })

  onBeforeUnmount(() => {
    window.removeEventListener('pageshow', onPageShow)
    window.removeEventListener('pagehide', onPageHide)
    clearBfCacheRestoreFlag()
    clearGuardedActionTimer()
  })

  return {
    isBfCacheRestore,
    clearBfCacheRestoreFlag,
    runWithBfCacheGuard,
  }
}
