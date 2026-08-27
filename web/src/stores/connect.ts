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

  const isFresh = (ts: number) => ts > 0 && Date.now() - ts < CONNECT_CACHE_TTL_MS

  async function getConnect(options?: { force?: boolean }) {
    const force = Boolean(options?.force)
    if (!force && isFresh(connectsFetchedAt.value)) return
    if (connectsInFlight.value) return connectsInFlight.value

    connectsInFlight.value = fetchGetConnectList()
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

    return connectsInFlight.value
  }

  const getConnectInfo = async (options?: { force?: boolean }) => {
    const force = Boolean(options?.force)
    if (!force && isFresh(connectsInfoFetchedAt.value)) {
      loading.value = false
      return
    }
    if (connectsInfoInFlight.value) return connectsInfoInFlight.value

    loading.value = true
    connectsInfoInFlight.value = fetchGetAllConnectInfo()
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

    return connectsInfoInFlight.value
  }

  return {
    connects,
    connectsInfo,
    loading,
    getConnect,
    getConnectInfo,
  }
})
