// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useEditorStore } from '@/stores/editor'
import { fetchGetEchoById } from '@/service/api'
import { ExtensionType } from '@/enums/enums'

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

const fetchByIdMock = vi.mocked(fetchGetEchoById)

const echo = (id: string): App.Api.Ech0.Echo =>
  ({
    id,
    content: `content-${id}`,
    username: 'tester',
    private: false,
    user_id: 'user-1',
    created_at: 1,
    echo_files: [],
    tags: [],
  }) as unknown as App.Api.Ech0.Echo

describe('editor update session', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    fetchByIdMock.mockReset()
  })

  it('loads the fetched echo into a single update session', async () => {
    fetchByIdMock.mockResolvedValue({
      code: 1,
      msg: '',
      data: echo('echo-1'),
    })

    const editorStore = useEditorStore()
    await editorStore.startUpdate(echo('echo-1'))

    expect(fetchByIdMock).toHaveBeenCalledWith('echo-1')
    expect(editorStore.isUpdateMode).toBe(true)
    expect(editorStore.updateEcho?.id).toBe('echo-1')
    expect(editorStore.echoToAdd.content).toBe('content-echo-1')
  })

  it('seeds attachments and extension drafts from the update target', async () => {
    const target = echo('echo-2')
    target.echo_files = [
      {
        file_id: 'file-1',
        sort_order: 0,
        file: {
          id: 'file-1',
          url: 'https://example.test/a.png',
          storage_type: 'local',
          category: 'image',
        },
      },
    ] as unknown as App.Api.Ech0.Echo['echo_files']
    target.extension = {
      type: ExtensionType.WEBSITE,
      payload: { title: 'Example', site: 'https://example.com' },
    }
    fetchByIdMock.mockResolvedValue({ code: 1, msg: '', data: target })

    const editorStore = useEditorStore()
    await editorStore.startUpdate(echo('echo-2'))

    expect(editorStore.filesToAdd.map((file) => file.id)).toEqual(['file-1'])
    expect(editorStore.currentExtensionType).toBe(ExtensionType.WEBSITE)
    expect(editorStore.websiteToAdd.title).toBe('Example')
    expect(editorStore.websiteToAdd.site).toBe('https://example.com')
  })

  it('returns to compose after exiting the update session', async () => {
    fetchByIdMock.mockResolvedValue({ code: 1, msg: '', data: echo('echo-3') })

    const editorStore = useEditorStore()
    await editorStore.startUpdate(echo('echo-3'))
    editorStore.handleExitUpdateMode()

    expect(editorStore.isUpdateMode).toBe(false)
    expect(editorStore.updateEcho).toBeNull()
    expect(editorStore.echoToAdd.content).toBe('')
  })
})
