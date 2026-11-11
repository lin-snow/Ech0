// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useEditorStore } from '@/stores/editor'
import { fetchAddEcho } from '@/service/api'
import { ExtensionType, ImageLayout } from '@/enums/enums'

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

const fetchAddEchoMock = vi.mocked(fetchAddEcho)

describe('editor submit payload', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    fetchAddEchoMock.mockReset()
    fetchAddEchoMock.mockResolvedValue({ code: 0, msg: '' } as never)
  })

  it('derives the attachment order, tags and extension from one draft', async () => {
    const editorStore = useEditorStore()

    editorStore.echoToAdd.content = 'hello'
    editorStore.tagToAdd = [' one ', '', 'two']
    editorStore.setFilesToAdd([
      {
        id: 'file-b',
        url: 'https://example.test/b.png',
        storage_type: 'local',
        category: 'image',
      },
      {
        id: 'file-a',
        url: 'https://example.test/a.png',
        storage_type: 'local',
        category: 'image',
      },
    ])
    editorStore.extensionToAdd = {
      extension: '',
      extension_type: ExtensionType.WEBSITE,
    }
    editorStore.websiteToAdd = { title: 'Example', site: 'https://example.com' }

    await editorStore.handleAddOrUpdateEcho(false)

    expect(fetchAddEchoMock).toHaveBeenCalledWith({
      content: 'hello',
      private: false,
      layout: ImageLayout.WATERFALL,
      extension: {
        type: ExtensionType.WEBSITE,
        payload: { title: 'Example', site: 'https://example.com' },
      },
      tags: [{ name: 'one' }, { name: 'two' }],
      echo_files: [
        { file_id: 'file-b', sort_order: 0 },
        { file_id: 'file-a', sort_order: 1 },
      ],
    })
    expect(editorStore.echoToAdd.echo_files).toBeUndefined()
    expect(editorStore.echoToAdd.extension).toBeUndefined()
  })

  it('refuses to submit an incomplete extension', async () => {
    const editorStore = useEditorStore()

    editorStore.echoToAdd.content = 'hello'
    editorStore.extensionToAdd = {
      extension: '',
      extension_type: ExtensionType.WEBSITE,
    }
    editorStore.websiteToAdd = { title: 'Example', site: '' }

    await editorStore.handleAddOrUpdateEcho(false)

    expect(fetchAddEchoMock).not.toHaveBeenCalled()
  })
})
