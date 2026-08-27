// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import {
  fetchCancelMigration,
  fetchCleanupMigration,
  fetchGetMigrationStatus,
  fetchStartMigration,
  type MigrationStatusPayload,
  type StartMigrationPayload,
} from '@/service/api'

const POLL_INTERVAL_MS = 2000

export const useMigrationStore = defineStore('migrationStore', () => {
  const state = ref<MigrationStatusPayload>({
    version: 1,
    source_type: 'ech0',
    source_payload: {},
    status: 'idle',
    error_message: '',
  })
  const loading = ref(false)
  const pollTimer = ref<number | null>(null)

  const isRunning = computed(
    () => state.value.status === 'pending' || state.value.status === 'running',
  )
  const hasJob = computed(() => state.value.status !== 'idle')
  const canCleanup = computed(() => state.value.status !== 'idle' && !isRunning.value)
  const isSuccess = computed(() => state.value.status === 'success')

  function applyState(next: Partial<MigrationStatusPayload> | null | undefined) {
    if (!next) return
    state.value = {
      ...state.value,
      ...next,
    }
  }

  function stopPolling() {
    if (pollTimer.value !== null) {
      window.clearInterval(pollTimer.value)
      pollTimer.value = null
    }
  }

  function startPolling() {
    if (pollTimer.value !== null) return
    pollTimer.value = window.setInterval(async () => {
      await fetchStatus()
      if (!isRunning.value) {
        stopPolling()
      }
    }, POLL_INTERVAL_MS)
  }

  async function fetchStatus() {
    const res = await fetchGetMigrationStatus()
    if (res.code !== 1) {
      return false
    }
    applyState(res.data as Partial<MigrationStatusPayload>)
    return true
  }

  async function startMigration(payload: StartMigrationPayload) {
    loading.value = true
    try {
      const res = await fetchStartMigration(payload)
      if (res.code !== 1) {
        return res
      }
      applyState(res.data as Partial<MigrationStatusPayload>)
      if (isRunning.value) {
        startPolling()
      }
      return res
    } finally {
      loading.value = false
    }
  }

  async function cancelMigration() {
    const res = await fetchCancelMigration()
    if (res.code === 1) {
      applyState(res.data as Partial<MigrationStatusPayload>)
      if (isRunning.value) {
        startPolling()
      } else {
        stopPolling()
      }
    }
    return res
  }

  async function init() {
    const ok = await fetchStatus()
    if (ok && isRunning.value) {
      startPolling()
    }
  }

  async function cleanupMigration() {
    const res = await fetchCleanupMigration()
    if (res.code === 1) {
      state.value = {
        version: 1,
        source_type: 'ech0',
        source_payload: {},
        status: 'idle',
        error_message: '',
      }
      stopPolling()
    }
    return res
  }

  return {
    state,
    loading,
    hasJob,
    isRunning,
    canCleanup,
    isSuccess,
    init,
    fetchStatus,
    startMigration,
    cancelMigration,
    cleanupMigration,
    startPolling,
    stopPolling,
  }
})
