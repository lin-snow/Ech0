// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { defineStore } from 'pinia'
import { ref, computed, watch } from 'vue'
import { fetchQueryEchos, fetchGetTags, fetchGetEchoById, fetchCreateTag } from '@/service/api'

export const useEchoStore = defineStore('echoStore', () => {
  const normalizeEchoId = (echo: App.Api.Ech0.Echo): string => String(echo?.id ?? '').trim()

  const echoList = ref<App.Api.Ech0.Echo[]>([])
  const echoIndexMap = ref(new Map<string, number>())
  const isLoading = ref<boolean>(true)
  const total = ref<number>(0)
  const pageSize = ref<number>(7)
  const currentPage = ref<number>(1)
  const searchValue = ref<string>('')
  const searchingMode = computed(() => searchValue.value.length > 0)
  const totalPages = computed(() =>
    total.value <= 0 ? 1 : Math.max(1, Math.ceil(total.value / pageSize.value)),
  )
  const echoToUpdate = ref<App.Api.Ech0.EchoToUpdate | null>(null)

  const tagList = ref<App.Api.Ech0.Tag[]>([])
  const tagOptions = computed<string[]>(() => tagList.value.map((tag) => tag.name))

  const isFilteringMode = ref<boolean>(false)
  const filteredTag = ref<App.Api.Ech0.Tag | null>(null)

  const dateFrom = ref<number | null>(null)
  const dateTo = ref<number | null>(null)
  const isDateRangeActive = computed(() => dateFrom.value !== null || dateTo.value !== null)

  const selectedTagIds = ref<string[]>([])
  const isTagSelectionActive = computed(() => selectedTagIds.value.length > 0)

  const visibilityFilter = ref<App.Api.Ech0.EchoVisibilityFilter>('all')
  const isVisibilityFilterActive = computed(() => visibilityFilter.value !== 'all')

  watch(searchingMode, (newValue, oldValue) => {
    if (newValue === false && oldValue === true) {
      refreshEchos()
    }
  })

  function buildQueryParams(): App.Api.Ech0.EchoQueryParams {
    const params: App.Api.Ech0.EchoQueryParams = {
      page: currentPage.value,
      pageSize: pageSize.value,
      search: searchValue.value || undefined,
    }
    const tagIds = new Set<string>()
    if (isFilteringMode.value && filteredTag.value) {
      tagIds.add(filteredTag.value.id)
    }
    selectedTagIds.value.forEach((id) => tagIds.add(id))
    if (tagIds.size > 0) {
      params.tagIds = Array.from(tagIds)
    }
    if (dateFrom.value !== null) {
      params.dateFrom = dateFrom.value
    }
    if (dateTo.value !== null) {
      params.dateTo = dateTo.value
    }
    if (visibilityFilter.value !== 'all') {
      params.private = visibilityFilter.value === 'private'
    }
    return params
  }

  const resetDateRange = () => {
    dateFrom.value = null
    dateTo.value = null
  }

  const resetSelectedTags = () => {
    selectedTagIds.value = []
  }

  const resetVisibilityFilter = () => {
    visibilityFilter.value = 'all'
  }

  const removeSelectedTag = (tagId: string) => {
    selectedTagIds.value = selectedTagIds.value.filter((id) => id !== tagId)
    if (filteredTag.value?.id === tagId && isFilteringMode.value) {
      isFilteringMode.value = false
      filteredTag.value = null
    }
  }

  let pendingFetch: Promise<void> | null = null
  async function fetchCurrentPage() {
    if (pendingFetch) return pendingFetch

    isLoading.value = true
    pendingFetch = fetchQueryEchos(buildQueryParams())
      .then((res) => {
        if (res.code === 1) {
          total.value = res.data.total
          const items = (res.data.items ?? []).map((item) => ({
            ...item,
            id: normalizeEchoId(item),
          }))
          echoList.value = items
          const nextIndex = new Map<string, number>()
          items.forEach((item, idx) => {
            if (item.id) nextIndex.set(item.id, idx)
          })
          echoIndexMap.value = nextIndex
        }
      })
      .finally(() => {
        isLoading.value = false
        pendingFetch = null
      })

    return pendingFetch
  }

  async function goToPage(page: number) {
    const target = Math.max(1, Math.floor(page) || 1)
    if (target === currentPage.value && echoList.value.length > 0) return
    currentPage.value = target
    await fetchCurrentPage()
  }

  const refreshEchos = () => {
    currentPage.value = 1
    total.value = 0
    echoList.value = []
    echoIndexMap.value = new Map()
    return fetchCurrentPage()
  }

  const clearEchos = () => {
    currentPage.value = 1
    total.value = 0
    echoList.value = []
    echoIndexMap.value = new Map()
  }

  const refreshForSearch = () => {
    currentPage.value = 1
    total.value = 0
    echoList.value = []
    echoIndexMap.value = new Map()
  }

  const updateEcho = (echo: App.Api.Ech0.Echo) => {
    const idx = echoIndexMap.value.get(echo.id)
    if (idx !== undefined) {
      echoList.value[idx] = echo
    }
  }

  const updateLikeCount = (echoId: string, delta: number = 1) => {
    const idx = echoIndexMap.value.get(echoId)
    if (idx !== undefined) {
      const targetEcho = echoList.value[idx]
      if (targetEcho) {
        targetEcho.fav_count = (targetEcho.fav_count || 0) + delta
        echoList.value[idx] = { ...targetEcho }
      }
    }
  }

  const pendingEchoMap = new Map<string, Promise<App.Api.Ech0.Echo | null>>()

  const prefetchEcho = (echoId: string): Promise<App.Api.Ech0.Echo | null> => {
    const id = String(echoId ?? '').trim()
    if (!id) return Promise.resolve(null)

    const idx = echoIndexMap.value.get(id)
    if (idx !== undefined && echoList.value[idx]) {
      return Promise.resolve(echoList.value[idx]!)
    }

    const existing = pendingEchoMap.get(id)
    if (existing) return existing

    const promise: Promise<App.Api.Ech0.Echo | null> = fetchGetEchoById(id)
      .then((res) => {
        if (res.code === 1 && res.data) {
          return res.data
        }
        return null
      })
      .catch(() => null)
      .finally(() => {
        pendingEchoMap.delete(id)
      }) as Promise<App.Api.Ech0.Echo | null>

    pendingEchoMap.set(id, promise)
    return promise
  }

  const getTags = async () => {
    const res = await fetchGetTags()
    if (res.code === 1) {
      tagList.value.splice(0, tagList.value.length, ...res.data)
    }
  }

  const createTag = async (name: string) => {
    const res = await fetchCreateTag(name)
    if (res.code === 1) {
      await getTags()
      return res.data
    }
    return null
  }

  let tagsLoadPromise: Promise<void> | null = null
  const ensureTagsLoaded = (): Promise<void> => {
    if (tagList.value.length > 0) return Promise.resolve()
    if (tagsLoadPromise) return tagsLoadPromise
    tagsLoadPromise = getTags().finally(() => {
      tagsLoadPromise = null
    })
    return tagsLoadPromise
  }

  return {
    echoList,
    echoIndexMap,
    isLoading,
    total,
    pageSize,
    currentPage,
    totalPages,
    searchValue,
    searchingMode,
    echoToUpdate,
    tagList,
    tagOptions,

    isFilteringMode,
    filteredTag,

    dateFrom,
    dateTo,
    isDateRangeActive,

    selectedTagIds,
    isTagSelectionActive,

    visibilityFilter,
    isVisibilityFilterActive,

    fetchCurrentPage,
    goToPage,
    refreshEchos,
    clearEchos,
    refreshForSearch,
    resetDateRange,
    resetSelectedTags,
    resetVisibilityFilter,
    removeSelectedTag,
    updateEcho,
    updateLikeCount,
    prefetchEcho,
    getTags,
    ensureTagsLoaded,
    createTag,
  }
})
