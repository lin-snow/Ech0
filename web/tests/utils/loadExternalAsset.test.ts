// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { describe, expect, it, vi } from 'vitest'

import { loadExternalScript } from '../../src/utils/loadExternalAsset'

const scriptSelector = (src: string) => `script[data-external-script="${src}"]`

describe('loadExternalAsset', () => {
  it('deduplicates concurrent script loads and reuses loaded script', async () => {
    const src = `/others/scripts/APlayer.min.js?test=${Date.now()}`

    const promiseA = loadExternalScript(src)
    const promiseB = loadExternalScript(src)

    expect(promiseA).toBe(promiseB)
    expect(document.querySelectorAll(scriptSelector(src))).toHaveLength(1)

    const script = document.querySelector<HTMLScriptElement>(scriptSelector(src))
    expect(script).toBeTruthy()
    script?.dispatchEvent(new Event('load'))

    await expect(promiseA).resolves.toBeUndefined()
    await expect(loadExternalScript(src)).resolves.toBeUndefined()
    expect(document.querySelectorAll(scriptSelector(src))).toHaveLength(1)
  })

  it('retries script loading once and succeeds on second attempt', async () => {
    const src = `/others/scripts/Meting.min.js?retry=${Date.now()}`
    const appendChild = document.head.appendChild.bind(document.head)
    let attempts = 0

    const appendSpy = vi.spyOn(document.head, 'appendChild').mockImplementation((node: Node) => {
      const result = appendChild(node)
      if (node instanceof HTMLScriptElement && node.getAttribute('data-external-script') === src) {
        attempts += 1
        queueMicrotask(() => {
          node.dispatchEvent(new Event(attempts === 1 ? 'error' : 'load'))
        })
      }
      return result
    })

    await expect(loadExternalScript(src, { timeoutMs: 8_000, retries: 1 })).resolves.toBeUndefined()
    expect(attempts).toBe(2)
    appendSpy.mockRestore()
  })

  it('fails after timeout and retry exhaustion', async () => {
    vi.useFakeTimers()
    try {
      const src = `/others/scripts/APlayer.min.js?timeout=${Date.now()}`
      const expectation = expect(
        loadExternalScript(src, { timeoutMs: 10, retries: 1 }),
      ).rejects.toThrow('Timed out loading external asset')
      await vi.advanceTimersByTimeAsync(25)
      await expectation
      expect(document.querySelectorAll(scriptSelector(src))).toHaveLength(0)
    } finally {
      vi.useRealTimers()
    }
  })
})
