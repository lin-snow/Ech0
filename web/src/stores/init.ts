// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { defineStore } from 'pinia'
import { ref } from 'vue'
import { fetchGetInitStatus, fetchInitOwner } from '@/service/api'
import { i18n } from '@/locales'
import { localStg } from '@/utils/storage'

export const useInitStore = defineStore('initStore', () => {
  const initialized = ref<boolean>(false)
  const ownerExists = ref<boolean>(false)
  const ready = ref<boolean>(false)

  const saveCache = () => {
    localStg.setItem('initialized', initialized.value)
  }

  const clearCache = () => {
    localStg.removeItem('initialized')
  }

  const getStatus = async () => {
    const res = await fetchGetInitStatus()
    if (res) {
      initialized.value = res.initialized
      ownerExists.value = res.owner_exists
      ready.value = true
      saveCache()
    } else {
      ready.value = false
      clearCache()
    }
    return res
  }

  const initOwner = async (payload: App.Api.Init.InitOwnerParams) => {
    const enriched: App.Api.Init.InitOwnerParams = {
      ...payload,
      locale: payload.locale || String(i18n.global.locale.value),
    }
    const res = await fetchInitOwner(enriched)
    if (res) {
      initialized.value = true
      ownerExists.value = true
      ready.value = true
      saveCache()
    } else {
      // A refusing server may still have finished the setup (racing submit);
      // the status poll is what decides.
      await getStatus().catch(() => undefined)
    }
    return res
  }

  const init = async () => {
    const cached = localStg.getItem<boolean>('initialized')
    if (cached !== null) {
      initialized.value = cached
    }
    await getStatus()
  }

  return {
    initialized,
    ownerExists,
    ready,
    getStatus,
    initOwner,
    init,
  }
})
