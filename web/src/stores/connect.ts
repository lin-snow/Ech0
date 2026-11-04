// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { ref } from 'vue'
import { defineStore } from 'pinia'
import { fetchGetConnectList, fetchGetAllConnectInfo } from '@/service/api'

const CONNECT_CACHE_TTL_MS = 30 * 60 * 1000

export const useConnectStore = defineStore('connectStore', () => {
  const connects = ref<App.Api.Connect.Connected[]>([])
  const connectsInfo = ref<App.Api.Connect.Connect[]>([])
  const loading = ref<boolean>(true)
  const connectsFetchedAt = ref<number>(0)
  const connectsInfoFetchedAt = ref<number>(0)
  const connectsInFlight = ref<Promise<void> | null>(null)
  const connectsInfoInFlight = ref<Promise<void> | null>(null)

  const requestConnects = () => {
    if (connectsInFlight.value) return connectsInFlight.value

    const task = fetchGetConnectList()
      .then((res) => {
        if (res.code === 1) {
          connects.value = res.data
          connectsFetchedAt.value = Date.now()
        }
      })
      .catch((err) => {
        console.error(err)
      })
      .finally(() => {
        connectsInFlight.value = null
      })

    connectsInFlight.value = task
    return task
  }

  const requestConnectsInfo = () => {
    if (connectsInfoInFlight.value) return connectsInfoInFlight.value

    loading.value = connectsInfoFetchedAt.value === 0
    const task = fetchGetAllConnectInfo()
      .then((res) => {
        if (res.code === 1) {
          connectsInfo.value = res.data
          connectsInfoFetchedAt.value = Date.now()
        }
      })
      .catch((err) => {
        console.error(err)
      })
      .finally(() => {
        loading.value = false
        connectsInfoInFlight.value = null
      })

    connectsInfoInFlight.value = task
    return task
  }

  // Stale-while-revalidate: cached data is returned immediately and refreshed
  // in the background, so a page open never waits on an expired cache.
  async function getConnect(options?: { force?: boolean }) {
    const force = Boolean(options?.force)
    const hasCache = connectsFetchedAt.value > 0

    if (
      !force &&
      connectsFetchedAt.value > 0 &&
      Date.now() - connectsFetchedAt.value < CONNECT_CACHE_TTL_MS
    ) {
      return
    }
    if (connectsInFlight.value) {
      if (hasCache && !force) return
      return connectsInFlight.value
    }

    const task = requestConnects()
    if (hasCache && !force) return
    return task
  }

  async function getConnectInfo(options?: { force?: boolean }) {
    const force = Boolean(options?.force)
    const hasCache = connectsInfoFetchedAt.value > 0

    if (
      !force &&
      connectsInfoFetchedAt.value > 0 &&
      Date.now() - connectsInfoFetchedAt.value < CONNECT_CACHE_TTL_MS
    ) {
      loading.value = false
      return
    }
    if (connectsInfoInFlight.value) {
      if (hasCache && !force) return
      return connectsInfoInFlight.value
    }

    const task = requestConnectsInfo()
    if (hasCache && !force) return
    return task
  }

  function invalidateConnectsInfo() {
    connectsInfo.value = []
    connectsInfoFetchedAt.value = 0
    loading.value = false
  }

  return {
    connects,
    connectsInfo,
    loading,
    getConnect,
    getConnectInfo,
    invalidateConnectsInfo,
  }
})
