// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useEditorStore } from '@/stores/editor'
import { useEchoStore } from '@/stores/echo'

vi.mock('@/service/api', () => ({
  fetchAddEcho: vi.fn(),
  fetchUpdateEcho: vi.fn(),
  fetchGetTags: vi.fn(),
  fetchGetEchoById: vi.fn(),
  fetchQueryEchos: vi.fn(),
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

const echoToUpdate = (id: string): App.Api.Ech0.EchoToUpdate => ({
  id,
  content: id,
  username: 'tester',
  private: false,
  user_id: 'user-1',
  created_at: 1,
  echo_files: [],
})

describe('editor update session', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('seeds a new update target only once per session', () => {
    const editorStore = useEditorStore()
    const echoStore = useEchoStore()
    const target = echoToUpdate('echo-1')

    editorStore.beginUpdate(target)

    expect(echoStore.echoToUpdate?.id).toBe('echo-1')
    expect(editorStore.isUpdateMode).toBe(true)
    expect(editorStore.updateDraftSeeded).toBe(false)

    // The editor marks the draft as seeded after populating it; remounting the
    // editor must not reset it.
    editorStore.updateDraftSeeded = true
    expect(editorStore.updateDraftSeeded).toBe(true)

    editorStore.beginUpdate(echoToUpdate('echo-2'))
    expect(editorStore.updateDraftSeeded).toBe(false)
  })

  it('clears the update session when exiting', () => {
    const editorStore = useEditorStore()
    const echoStore = useEchoStore()

    editorStore.beginUpdate(echoToUpdate('echo-1'))
    editorStore.updateDraftSeeded = true

    editorStore.handleExitUpdateMode()

    expect(editorStore.isUpdateMode).toBe(false)
    expect(editorStore.updateDraftSeeded).toBe(false)
    expect(echoStore.echoToUpdate).toBeNull()
  })
})
