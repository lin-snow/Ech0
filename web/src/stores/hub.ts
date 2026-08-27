// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useFetch } from '@vueuse/core'
import { theToast } from '@/utils/toast'
import { timeValueToMs } from '@/utils/timeValue'
import { useConnectStore } from './connect'
import { i18n } from '@/locales'

interface HubState {
  url: string
  buffer: App.Api.Hub.Echo[]
  currentPage: number
  hasMore: boolean
  isLoading: boolean
}

export const useHubStore = defineStore('hubStore', () => {
  const connectStore = useConnectStore()

  const hubList = ref<App.Api.Hub.HubList>([])
  const hubinfoList = ref<App.Api.Hub.HubInfoList>([])
  const hubInfoMap = ref<Map<string, App.Api.Hub.HubItemInfo>>(new Map())
  const hubStates = ref<Map<string, HubState>>(new Map())

  const echoList = ref<App.Api.Hub.Echo[]>([])
  const existingIds = ref<Set<string>>(new Set())

  const isPreparing = ref<boolean>(true)
  const isLoading = ref<boolean>(false)
  const hasTriedInitialLoad = ref<boolean>(false)
  const pageSize = ref<number>(10)
  const batchSize = ref<number>(10)
  const hasMore = ref<boolean>(true)

  const getHubList = async () => {
    isPreparing.value = true
    hasTriedInitialLoad.value = false
    await connectStore.getConnect()

    hubList.value = connectStore.connects
  }

  const getHubInfoList = async () => {
    if (hubList.value.length === 0) {
      theToast.info(String(i18n.global.t('hub.emptyList')))
      isPreparing.value = false
      return
    }

    hubList.value = hubList.value.map((item) => {
      return typeof item === 'string'
        ? item.endsWith('/')
          ? item.slice(0, -1)
          : item
        : item.connect_url.endsWith('/')
          ? {
              ...item,
              connect_url: item.connect_url.slice(0, -1),
            }
          : item
    })

    const fetchWithTimeout = async (
      url: string,
      timeout: number = 5000,
    ): Promise<App.Api.Hub.HubItemInfo | null> => {
      return new Promise((resolve) => {
        let isResolved = false

        const timeoutId = setTimeout(() => {
          if (!isResolved) {
            isResolved = true
            console.warn(`[Hub] 请求超时: ${url}`)
            resolve(null)
          }
        }, timeout)

        ;(async () => {
          try {
            const { error, data } = await useFetch<App.Api.Response<App.Api.Hub.HubItemInfo>>(
              `${url}/api/connect`,
            ).json()

            clearTimeout(timeoutId)
            if (!isResolved) {
              isResolved = true
              if (error.value || data.value?.code !== 1) {
                console.warn(`[Hub] 请求失败: ${url}`, error.value)
                resolve(null)
              } else {
                resolve(data.value?.data || null)
              }
            }
          } catch (err) {
            clearTimeout(timeoutId)
            if (!isResolved) {
              isResolved = true
              console.error(`[Hub] 请求异常: ${url}`, err)
              resolve(null)
            }
          }
        })()
      })
    }

    const promises = hubList.value.map(async (hub) => {
      const url = typeof hub === 'string' ? hub : hub.connect_url
      return await fetchWithTimeout(url, 5000)
    })

    const results = await Promise.allSettled(promises)

    const validHubs: typeof hubList.value = []
    const failedHubs: string[] = []

    results.forEach((result, index) => {
      const hub = hubList.value[index]
      if (!hub) return

      const hubUrl = typeof hub === 'string' ? hub : hub.connect_url

      if (result.status === 'fulfilled' && result.value) {
        hubinfoList.value.push(result.value)
        validHubs.push(hub)

        if (typeof hubUrl === 'string') {
          hubInfoMap.value.set(hubUrl, result.value)
        }
      } else {
        if (typeof hubUrl === 'string') {
          failedHubs.push(hubUrl)
          console.warn(`[Hub] 实例不可用，已排除: ${hubUrl}`)
        }
      }
    })

    hubList.value = validHubs

    if (hubList.value.length === 0) {
      theToast.info(String(i18n.global.t('hub.noAvailableInstance')))
      isPreparing.value = false
      return
    }

    hubStates.value.clear()
    for (const hub of hubList.value) {
      const url = typeof hub === 'string' ? hub : hub.connect_url
      hubStates.value.set(url, {
        url,
        buffer: [],
        currentPage: 1,
        hasMore: true,
        isLoading: false,
      })
    }

    isPreparing.value = false
    console.info(String(i18n.global.t('hub.connectedCount', { count: hubList.value.length })))

    await Promise.all(Array.from(hubStates.value.keys()).map((url) => fetchHubPage(url)))
  }

  const fetchHubPage = async (hubUrl: string): Promise<void> => {
    const state = hubStates.value.get(hubUrl)
    if (!state || state.isLoading || !state.hasMore) return

    state.isLoading = true
    try {
      const { error, data } = await useFetch<App.Api.Response<App.Api.Ech0.PaginationResult>>(
        hubUrl + '/api/echo/page',
      )
        .post({
          page: state.currentPage,
          pageSize: pageSize.value,
        })
        .json()

      if (error.value || data.value?.code !== 1) {
        console.warn(`[Hub] 请求失败: ${hubUrl}`, error.value)
        state.hasMore = false
        return
      }

      const items = (data.value?.data.items || []).map((echo: App.Api.Ech0.Echo) => ({
        ...echo,
        createdTs: timeValueToMs(echo.created_at),
        server_name: hubInfoMap.value.get(hubUrl)?.server_name || 'Ech0',
        server_url: hubUrl,
        virtual_key: `${hubUrl}-${echo.id}`,
        logo: hubInfoMap.value.get(hubUrl)?.logo || '/Ech0.svg',
      }))

      items.sort((a: App.Api.Hub.Echo, b: App.Api.Hub.Echo) => b.createdTs - a.createdTs)
      state.buffer.push(...items)
      state.currentPage++
      state.hasMore = items.length >= pageSize.value
    } catch (err) {
      console.error(`[Hub] 请求异常: ${hubUrl}`, err)
      state.hasMore = false
    } finally {
      state.isLoading = false
    }
  }

  const loadEchoListPage = async () => {
    if (isLoading.value || isPreparing.value) return

    const canLoadMore = Array.from(hubStates.value.values()).some(
      (s) => s.hasMore || s.buffer.length > 0,
    )
    if (!canLoadMore) {
      hasMore.value = false
      hasTriedInitialLoad.value = true
      return
    }

    isLoading.value = true
    try {
      const result: App.Api.Hub.Echo[] = []
      let attempts = 0
      const maxAttempts = batchSize.value * 3

      while (result.length < batchSize.value && attempts < maxAttempts) {
        attempts++

        let maxTs = -1
        let maxHubUrl: string | null = null

        for (const [url, state] of hubStates.value) {
          const head = state.buffer[0]
          if (head) {
            const headTs = head.createdTs
            if (headTs > maxTs) {
              maxTs = headTs
              maxHubUrl = url
            }
          }
        }

        if (maxHubUrl === null) {
          const emptyHubsWithMore = Array.from(hubStates.value.values()).filter(
            (s) => s.hasMore && !s.isLoading && s.buffer.length === 0,
          )

          if (emptyHubsWithMore.length === 0) {
            break
          }

          await Promise.all(emptyHubsWithMore.map((s) => fetchHubPage(s.url)))
          continue
        }

        const state = hubStates.value.get(maxHubUrl)!
        const echo = state.buffer.shift()!

        const key = `${echo.server_url}-${echo.id}`
        if (!existingIds.value.has(key)) {
          existingIds.value.add(key)
          result.push(echo)
        }

        if (state.buffer.length < 3 && state.hasMore && !state.isLoading) {
          fetchHubPage(maxHubUrl)
        }
      }

      echoList.value.push(...result)

      hasMore.value = Array.from(hubStates.value.values()).some(
        (s) => s.hasMore || s.buffer.length > 0,
      )

      if (!hasMore.value && echoList.value.length > 0) {
        theToast.info(String(i18n.global.t('hub.noMoreData')))
      }
    } finally {
      isLoading.value = false
      hasTriedInitialLoad.value = true
    }
  }

  return {
    echoList,
    hubList,
    hubInfoMap,
    hubinfoList,
    hubStates,
    isLoading,
    isPreparing,
    hasTriedInitialLoad,
    pageSize,
    batchSize,
    hasMore,
    getHubList,
    getHubInfoList,
    loadEchoListPage,
  }
})
