// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

export async function pMapLimit<T, R>(
  items: readonly T[],
  limit: number,
  fn: (item: T, index: number) => Promise<R>,
  options?: { onSettled?: (result: PromiseSettledResult<R>, index: number) => void },
): Promise<PromiseSettledResult<R>[]> {
  const results: PromiseSettledResult<R>[] = new Array(items.length)
  let next = 0
  const onSettled = options?.onSettled

  const worker = async () => {
    while (true) {
      const i = next++
      if (i >= items.length) return
      let result: PromiseSettledResult<R>
      try {
        result = { status: 'fulfilled', value: await fn(items[i]!, i) }
      } catch (reason) {
        result = { status: 'rejected', reason }
      }
      results[i] = result
      onSettled?.(result, i)
    }
  }

  const n = Math.max(1, Math.min(limit, items.length))
  await Promise.all(Array.from({ length: n }, worker))
  return results
}

export const HUB_FAN_OUT_LIMIT = 8
