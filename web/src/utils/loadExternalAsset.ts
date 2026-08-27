// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

const scriptLoadPromises = new Map<string, Promise<void>>()

const scriptSelector = (src: string) => `script[data-external-script="${src}"],script[src="${src}"]`
const DEFAULT_SCRIPT_TIMEOUT_MS = 8_000
const DEFAULT_SCRIPT_RETRIES = 0

type ScriptLoadOptions = {
  timeoutMs?: number
  retries?: number
}

const resolveFromReadyState = (node: HTMLScriptElement, url: string, timeoutMs?: number) => {
  if (node.getAttribute('data-external-script-ready') === 'true') {
    return Promise.resolve()
  }

  return new Promise<void>((resolve, reject) => {
    let timer: ReturnType<typeof setTimeout> | null = null
    const cleanup = () => {
      node.removeEventListener('load', onLoad)
      node.removeEventListener('error', onError)
      if (timer) {
        clearTimeout(timer)
        timer = null
      }
    }

    const onLoad = () => {
      node.setAttribute('data-external-script-ready', 'true')
      cleanup()
      resolve()
    }

    const onError = () => {
      cleanup()
      reject(new Error(`Failed to load external asset: ${url}`))
    }

    node.addEventListener('load', onLoad, { once: true })
    node.addEventListener('error', onError, { once: true })
    if (timeoutMs && timeoutMs > 0) {
      timer = setTimeout(() => {
        cleanup()
        reject(new Error(`Timed out loading external asset: ${url}`))
      }, timeoutMs)
    }
  })
}

const createExternalScript = (src: string) => {
  const script = document.createElement('script')
  script.src = src
  script.defer = true
  script.setAttribute('data-external-script', src)
  script.removeAttribute('data-external-script-failed')
  document.head.appendChild(script)
  return script
}

export const loadExternalScript = (src: string, options: ScriptLoadOptions = {}): Promise<void> => {
  if (typeof document === 'undefined') return Promise.resolve()
  const cached = scriptLoadPromises.get(src)
  if (cached) return cached

  const timeoutMs = options.timeoutMs ?? DEFAULT_SCRIPT_TIMEOUT_MS
  const retries = Math.max(0, options.retries ?? DEFAULT_SCRIPT_RETRIES)

  const promise = (async () => {
    let latestError: unknown = null

    for (let attempt = 0; attempt <= retries; attempt += 1) {
      let script = document.querySelector<HTMLScriptElement>(scriptSelector(src))

      if (script?.getAttribute('data-external-script-failed') === 'true') {
        script.remove()
        script = null
      }
      if (!script) {
        script = createExternalScript(src)
      }

      try {
        await resolveFromReadyState(script, src, timeoutMs)
        return
      } catch (error) {
        latestError = error
        script.setAttribute('data-external-script-failed', 'true')
        script.removeAttribute('data-external-script-ready')
        script.remove()
      }
    }

    throw latestError ?? new Error(`Failed to load external asset: ${src}`)
  })().catch((error) => {
    scriptLoadPromises.delete(src)
    throw error
  })

  scriptLoadPromises.set(src, promise)
  return promise
}
