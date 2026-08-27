// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { fetchQueryEchos } from '@/service/api'

export const useZenStore = defineStore('zenStore', () => {
  const normalizeId = (echo: App.Api.Ech0.Echo): string => String(echo?.id ?? '').trim()

  const echoList = ref<App.Api.Ech0.Echo[]>([])
  const isLoading = ref<boolean>(false)
  const total = ref<number>(0)
  const pageSize = ref<number>(20)
  const currentPage = ref<number>(1)

  const hasMore = computed(() => {
    if (currentPage.value === 1 && echoList.value.length === 0) return true
    return echoList.value.length < total.value
  })

  let pendingFetch: Promise<void> | null = null

  async function loadNextPage(): Promise<void> {
    if (pendingFetch) return pendingFetch
    if (echoList.value.length > 0 && !hasMore.value) return

    isLoading.value = true
    pendingFetch = fetchQueryEchos({ page: currentPage.value, pageSize: pageSize.value })
      .then((res) => {
        if (res.code !== 1) return
        total.value = res.data.total
        const incoming = (res.data.items ?? []).map((item) => ({
          ...item,
          id: normalizeId(item),
        }))
        const seen = new Set(echoList.value.map((e) => e.id))
        const fresh = incoming.filter((item) => item.id && !seen.has(item.id))
        echoList.value.push(...fresh)
        currentPage.value += 1
      })
      .finally(() => {
        isLoading.value = false
        pendingFetch = null
      })

    return pendingFetch
  }

  function reset(): void {
    echoList.value = []
    total.value = 0
    currentPage.value = 1
  }

  return {
    echoList,
    isLoading,
    total,
    pageSize,
    currentPage,
    hasMore,
    loadNextPage,
    reset,
  }
})
