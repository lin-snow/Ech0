// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { nextTick } from 'vue'
import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import TheEditorPreview from '@/views/home/modules/TheEditor/TheEditorPreview.vue'
import { ExtensionType, Mode } from '@/enums/enums'

type Preview = App.Api.Ech0.EchoExtension | null

type StoreState = {
  filesToAdd: Array<{ id: string }>
  extensionPreview: Preview
  currentMode: Mode
}

const shared = vi.hoisted(() => ({
  store: null as unknown as StoreState,
}))

vi.mock('@/stores', async () => {
  // vi.mock factories are hoisted above static imports, so Vue must be loaded
  // inside the factory; a static import would be uninitialized here.
  const { reactive } = await import('vue')
  const store = reactive({ filesToAdd: [], extensionPreview: null, currentMode: 0 })
  shared.store = store as unknown as StoreState
  return { useEditorStore: () => store }
})

vi.mock('@/views/home/modules/TheEditor/TheEditorMedia.vue', () => ({
  default: { name: 'TheEditorMedia', template: '<div data-test="media" />' },
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

const websitePreview = (title: string): Preview => ({
  type: ExtensionType.WEBSITE,
  payload: { title, site: 'https://example.com' },
})

const mountPreview = () => mount(TheEditorPreview)

const flush = async () => {
  await nextTick()
  await nextTick()
}

describe('TheEditorPreview', () => {
  it('shows only the file preview without a switch', async () => {
    shared.store.filesToAdd = [{ id: 'file-1' }]
    shared.store.extensionPreview = null
    shared.store.currentMode = Mode.ECH0

    const wrapper = mountPreview()
    await flush()

    expect(wrapper.find('[data-test="media"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="renderer"]').exists()).toBe(false)
    expect(wrapper.find('.seg').exists()).toBe(false)
  })

  it('shows only the extension preview without a switch', async () => {
    shared.store.filesToAdd = []
    shared.store.extensionPreview = websitePreview('Example')
    shared.store.currentMode = Mode.ECH0

    const wrapper = mountPreview()
    await flush()

    expect(wrapper.find('.sketch-frame').exists()).toBe(true)
    expect(wrapper.find('[data-test="media"]').exists()).toBe(false)
    expect(wrapper.find('.seg').exists()).toBe(false)
  })

  it('offers a switch when both previews exist and defaults to files', async () => {
    shared.store.filesToAdd = [{ id: 'file-1' }]
    shared.store.extensionPreview = websitePreview('Example')
    shared.store.currentMode = Mode.ECH0

    const wrapper = mountPreview()
    await flush()

    const tabs = wrapper.findAll('.seg__btn')
    expect(tabs).toHaveLength(2)
    expect(wrapper.find('[data-test="media"]').exists()).toBe(true)

    await tabs[1]!.trigger('click')
    await flush()

    expect(wrapper.find('[data-test="media"]').exists()).toBe(false)
    expect(wrapper.find('.sketch-frame').exists()).toBe(true)
  })

  it('defaults to the extension preview while editing an extension', async () => {
    shared.store.filesToAdd = [{ id: 'file-1' }]
    shared.store.extensionPreview = websitePreview('Example')
    shared.store.currentMode = Mode.EXTEN

    const wrapper = mountPreview()
    await flush()

    expect(wrapper.find('.seg').exists()).toBe(true)
    expect(wrapper.find('.sketch-frame').exists()).toBe(true)
    expect(wrapper.find('[data-test="media"]').exists()).toBe(false)
  })
})
