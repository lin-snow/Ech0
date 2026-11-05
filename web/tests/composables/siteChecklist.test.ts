// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { Ref } from 'vue'
import { createPinia, setActivePinia } from 'pinia'
import { useEchoStore } from '@/stores/echo'
import { useSiteChecklist } from '@/composables/useSiteChecklist'

vi.mock('@/service/api', () => ({
  fetchQueryEchos: vi.fn(),
  fetchGetTags: vi.fn(),
  fetchGetEchoById: vi.fn(),
  fetchCreateTag: vi.fn(),
}))

const stubs = vi.hoisted(() => ({
  userState: null as Ref<{ id: string } | null> | null,
  settingState: null as Ref<{ site_title: string; server_url: string }> | null,
}))

// The factory is hoisted above every import, so `vue` and the real echo store
// cannot be imported statically here; the dynamic import IS the boundary.
vi.mock('@/stores', async () => {
  const { ref, computed, reactive } = await import('vue')
  const { useEchoStore: realEchoStore } = await import('@/stores/echo')

  const userState = ref<{ id: string } | null>({ id: 'owner' })
  const settingState = ref({ site_title: '我的站点', server_url: 'https://blog.example.org' })
  stubs.userState = userState
  stubs.settingState = settingState

  return {
    useEchoStore: realEchoStore,
    useSettingStore: () => reactive({ SystemSetting: settingState }),
    useUserStore: () => reactive({ isLogin: computed(() => userState.value !== null) }),
  }
})

const setUser = (signedIn: boolean) => {
  if (stubs.userState) stubs.userState.value = signedIn ? { id: 'owner' } : null
}

describe('useSiteChecklist', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    setUser(true)
  })

  it('stays hidden while a search narrows the timeline to zero results', () => {
    const echoStore = useEchoStore()
    echoStore.isLoading = false
    echoStore.total = 300
    const { visible } = useSiteChecklist()
    expect(visible.value).toBe(false)

    echoStore.searchValue = 'missing-keyword'
    echoStore.total = 0
    expect(visible.value).toBe(false)
  })

  it('stays hidden while a non-keyword filter is active', () => {
    const echoStore = useEchoStore()
    echoStore.isLoading = false
    echoStore.total = 300
    const { visible } = useSiteChecklist()

    echoStore.selectedTagIds = ['tag-1']
    echoStore.total = 0
    expect(visible.value).toBe(false)

    echoStore.resetSelectedTags()
    echoStore.visibilityFilter = 'private'
    expect(visible.value).toBe(false)
  })

  it('shows for the owner once the query clears and steps stay pending', () => {
    const echoStore = useEchoStore()
    echoStore.isLoading = false
    echoStore.total = 0
    const { visible } = useSiteChecklist()

    echoStore.searchValue = 'missing-keyword'
    expect(visible.value).toBe(false)

    echoStore.searchValue = ''
    expect(visible.value).toBe(true)
  })

  it('hides for guests', () => {
    const echoStore = useEchoStore()
    echoStore.isLoading = false
    echoStore.total = 0
    setUser(false)
    const { visible } = useSiteChecklist()
    expect(visible.value).toBe(false)
  })
})
