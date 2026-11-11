// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { nextTick } from 'vue'
import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import TheExtensionEditor from '@/views/home/modules/TheEditor/TheExtensionEditor.vue'
import { ExtensionType } from '@/enums/enums'

type Preview = App.Api.Ech0.EchoExtension | null

const shared = vi.hoisted(() => ({
  store: null as unknown as Record<string, unknown> & { extensionPreview: Preview },
}))

vi.mock('@/stores', async () => {
  // vi.mock factories are hoisted above static imports, so Vue must be loaded
  // inside the factory; a static import would be uninitialized here.
  const { reactive } = await import('vue')
  const store = reactive({
    currentExtensionType: 'WEBSITE',
    extensionToAdd: { extension_type: 'WEBSITE', extension: '' },
    websiteToAdd: { title: '', site: '' },
    videoURL: '',
    musicURL: '',
    githubRepo: '',
    locationToAdd: { latitude: null, longitude: null, placeholder: '' },
    tweetToAdd: { url: '', username: '', statusId: '' },
    extensionPreview: null as Preview,
  })
  shared.store = store
  return { useEditorStore: () => store }
})

vi.mock('@/service/api', () => ({
  fetchGetWebsiteTitle: vi.fn(),
}))

vi.mock('@/components/advanced/extension/shared/LocationPicker.vue', () => ({
  default: { name: 'LocationPicker', template: '<div />' },
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

describe('TheExtensionEditor preview', () => {
  it('shows the dashed preview frame only when a draft is previewable', async () => {
    const wrapper = mount(TheExtensionEditor)

    expect(wrapper.find('.sketch-frame').exists()).toBe(false)

    shared.store.extensionPreview = {
      type: ExtensionType.WEBSITE,
      payload: { title: 'Example', site: 'https://example.com' },
    }
    await nextTick()
    expect(wrapper.find('.sketch-frame').exists()).toBe(true)

    shared.store.extensionPreview = null
    await nextTick()
    expect(wrapper.find('.sketch-frame').exists()).toBe(false)
  })
})
