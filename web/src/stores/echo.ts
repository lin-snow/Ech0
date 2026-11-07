// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { defineStore } from 'pinia'
import { ref, computed, watch } from 'vue'
import { fetchQueryEchos, fetchGetTags, fetchGetEchoById, fetchCreateTag } from '@/service/api'
import { TIMELINE_BASE_PAGE_SIZE } from '@/utils/timeline'

/** Backend clamps to [1, 100]; mirror it so a restored depth cannot overflow. */
const MAX_PAGE_SIZE = 100
const PAGE_SIZE = TIMELINE_BASE_PAGE_SIZE

type LoadPhase = 'first' | 'more'

export const useEchoStore = defineStore('echoStore', () => {
  const normalizeEchoId = (echo: App.Api.Ech0.Echo): string => String(echo?.id ?? '').trim()

  const echoList = ref<App.Api.Ech0.Echo[]>([])
  const echoIndexMap = ref(new Map<string, number>())
  const isLoading = ref<boolean>(true)
  const hasFetched = ref<boolean>(false)
  const isLoadingMore = ref<boolean>(false)
  const loadError = ref<LoadPhase | null>(null)
  const total = ref<number>(0)
  const pageSize = ref<number>(PAGE_SIZE)

  /**
   * `undefined` = the current query has not been fetched yet, a string = the
   * next keyset position, `null` = the feed is exhausted. This single ref is
   * the only pagination state the view has to understand.
   */
  const nextCursor = ref<string | null | undefined>(undefined)
  const loadedSignature = ref<string | null>(null)
  /** Page size the current stream was opened with, so a first-load retry restores the same depth. */
  const refreshPageSize = ref<number | null>(null)
  let generation = 0

  const searchValue = ref<string>('')
  const searchingMode = computed(() => searchValue.value.length > 0)
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

  const canLoadMore = computed(
    () =>
      nextCursor.value !== undefined &&
      nextCursor.value !== null &&
      !isLoading.value &&
      !isLoadingMore.value &&
      loadError.value === null,
  )

  /**
   * True whenever the timeline is showing a narrowed result set. `total` and
   * `echoList` then describe the query, not the instance, so site-level UI
   * (onboarding) must not read them while this is true.
   */
  const hasActiveQuery = computed(
    () =>
      searchingMode.value ||
      isFilteringMode.value ||
      isDateRangeActive.value ||
      isTagSelectionActive.value ||
      isVisibilityFilterActive.value,
  )

  watch(searchingMode, (newValue, oldValue) => {
    if (newValue === false && oldValue === true) {
      refreshEchos()
    }
  })

  function buildQueryParams(): App.Api.Ech0.EchoQueryParams {
    const params: App.Api.Ech0.EchoQueryParams = {
      page: 1,
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

  /**
   * Identifies the filter set currently backing the loaded stream. Mounting the
   * timeline again (tab switches, back navigation) reuses the loaded pages only
   * while this still matches.
   */
  function querySignature(): string {
    const params = buildQueryParams()
    return JSON.stringify([
      params.search ?? '',
      [...(params.tagIds ?? [])].sort(),
      params.dateFrom ?? null,
      params.dateTo ?? null,
      visibilityFilter.value,
    ])
  }

  const isTimelineCurrent = computed(
    () => loadedSignature.value !== null && loadedSignature.value === querySignature(),
  )

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

  const normalizeItems = (items: App.Api.Ech0.Echo[] | undefined): App.Api.Ech0.Echo[] =>
    (items ?? []).map((item) => ({ ...item, id: normalizeEchoId(item) }))

  const rebuildIndex = () => {
    const nextIndex = new Map<string, number>()
    echoList.value.forEach((item, idx) => {
      if (item.id) nextIndex.set(item.id, idx)
    })
    echoIndexMap.value = nextIndex
  }

  const appendDeduped = (items: App.Api.Ech0.Echo[]) => {
    const seen = new Set(echoList.value.map((item) => item.id))
    items.forEach((item) => {
      if (!item.id || seen.has(item.id)) return
      seen.add(item.id)
      echoList.value.push(item)
    })
    rebuildIndex()
  }

  const clampPageSize = (value: number): number =>
    Math.max(1, Math.min(MAX_PAGE_SIZE, Math.trunc(value) || PAGE_SIZE))

  async function refreshEchos(initialPageSize?: number): Promise<void> {
    const requestGeneration = ++generation
    const requestedPageSize = clampPageSize(initialPageSize ?? pageSize.value)
    refreshPageSize.value = requestedPageSize
    hasFetched.value = true
    isLoading.value = true
    isLoadingMore.value = false
    loadError.value = null
    loadedSignature.value = null
    nextCursor.value = undefined
    echoList.value = []
    echoIndexMap.value = new Map()

    try {
      const requestQuery = buildQueryParams()
      // Describe the request with the filters it was built from, not with
      // whatever the store holds when the response lands.
      const requestSignature = querySignature()
      const res = await fetchQueryEchos({
        ...requestQuery,
        pageSize: requestedPageSize,
        cursor: '',
      })
      if (requestGeneration !== generation) return
      if (res.code !== 1 || !res.data) {
        loadError.value = 'first'
        return
      }
      total.value = res.data.total ?? 0
      appendDeduped(normalizeItems(res.data.items))
      nextCursor.value = res.data.nextCursor ?? null
      loadedSignature.value = requestSignature
    } catch {
      if (requestGeneration === generation) {
        loadError.value = 'first'
      }
    } finally {
      if (requestGeneration === generation) {
        isLoading.value = false
      }
    }
  }

  async function loadMore(): Promise<void> {
    const cursor = nextCursor.value
    if (!cursor || isLoading.value || isLoadingMore.value || loadError.value !== null) return
    if (!isTimelineCurrent.value) {
      await refreshEchos()
      return
    }

    const requestGeneration = generation
    isLoadingMore.value = true
    try {
      const res = await fetchQueryEchos({
        ...buildQueryParams(),
        cursor,
      })
      if (requestGeneration !== generation) return
      if (res.code !== 1 || !res.data) {
        loadError.value = 'more'
        return
      }
      appendDeduped(normalizeItems(res.data.items))
      nextCursor.value = res.data.nextCursor ?? null
      total.value = res.data.total ?? total.value
    } catch {
      if (requestGeneration === generation) {
        loadError.value = 'more'
      }
    } finally {
      if (requestGeneration === generation) {
        isLoadingMore.value = false
      }
    }
  }

  const retryLoad = async (): Promise<void> => {
    if (loadError.value === null) return
    loadError.value = null
    if (nextCursor.value === undefined) {
      // Re-open at the depth the timeline was mounted with: a shallow first
      // page would drop the restored anchor row from the result set.
      await refreshEchos(refreshPageSize.value ?? undefined)
      return
    }
    await loadMore()
  }

  const clearEchos = () => {
    generation++
    nextCursor.value = undefined
    loadedSignature.value = null
    refreshPageSize.value = null
    loadError.value = null
    total.value = 0
    echoList.value = []
    echoIndexMap.value = new Map()
    isLoading.value = false
    isLoadingMore.value = false
    hasFetched.value = false
  }

  const updateEcho = (echo: App.Api.Ech0.Echo) => {
    const idx = echoIndexMap.value.get(echo.id)
    if (idx !== undefined) {
      echoList.value[idx] = echo
    }
  }

  const replaceEcho = (echo: App.Api.Ech0.Echo) => {
    const idx = echoIndexMap.value.get(echo.id)
    if (idx === undefined) return
    echoList.value[idx] = { ...echo, id: normalizeEchoId(echo) }
    rebuildIndex()
  }

  const removeEcho = (echoId: string) => {
    const idx = echoIndexMap.value.get(echoId)
    if (idx === undefined) return
    echoList.value.splice(idx, 1)
    total.value = Math.max(0, total.value - 1)
    rebuildIndex()
  }

  const setEchoReactions = (echoId: string, reactions: App.Api.Ech0.ReactionCount[]) => {
    const idx = echoIndexMap.value.get(echoId)
    if (idx === undefined) return
    const targetEcho = echoList.value[idx]
    if (!targetEcho) return
    echoList.value[idx] = { ...targetEcho, reactions }
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
    hasFetched,
    isLoadingMore,
    loadError,
    total,
    pageSize,
    nextCursor,
    canLoadMore,
    isTimelineCurrent,
    loadedSignature,
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
    hasActiveQuery,

    refreshEchos,
    loadMore,
    retryLoad,
    clearEchos,
    resetDateRange,
    resetSelectedTags,
    resetVisibilityFilter,
    removeSelectedTag,
    updateEcho,
    replaceEcho,
    removeEcho,
    setEchoReactions,
    prefetchEcho,
    getTags,
    ensureTagsLoaded,
    createTag,
  }
})
