// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { nextTick } from 'vue'
import { mount, type VueWrapper } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import TheEditorPreview from '@/views/home/modules/TheEditor/TheEditorPreview.vue'
import { ExtensionType, Mode } from '@/enums/enums'

type Preview = App.Api.Ech0.EchoExtension | null

type StoreState = {
  echoToAdd: { content: string }
  filesToAdd: Array<{ id: string }>
  extensionPreview: Preview
  currentMode: Mode
  isUpdateMode: boolean
  removeExtension: () => void
  handleAddOrUpdateEcho: (syncOnly: boolean) => Promise<void>
}

const shared = vi.hoisted(() => ({
  store: null as unknown as StoreState,
}))

vi.mock('@/stores', async () => {
  // vi.mock factories are hoisted above static imports, so Vue must be loaded
  // inside the factory; a static import would be uninitialized here.
  const { reactive } = await import('vue')
  const store = reactive({
    echoToAdd: { content: '' },
    filesToAdd: [],
    extensionPreview: null,
    currentMode: 0,
    isUpdateMode: false,
    removeExtension: vi.fn(),
    handleAddOrUpdateEcho: vi.fn(async () => {}),
  })
  shared.store = store as unknown as StoreState
  return { useEditorStore: () => store }
})

vi.mock('@/views/home/modules/TheEditor/TheEditorMedia.vue', () => ({
  default: { name: 'TheEditorMedia', template: '<div data-test="media" />' },
}))

vi.mock('@/composables/useBaseDialog', () => ({
  useBaseDialog: () => ({
    openConfirm: (options: { onConfirm?: () => void }) => options.onConfirm?.(),
  }),
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

const wrappers: VueWrapper[] = []

const mountPreview = () => {
  const wrapper = mount(TheEditorPreview, { global: { directives: { tooltip: {} } } })
  wrappers.push(wrapper)
  return wrapper
}

const flush = async () => {
  await nextTick()
  await nextTick()
}

describe('TheEditorPreview', () => {
  afterEach(async () => {
    // Let lazily imported preview cards settle before jsdom is torn down.
    await new Promise((resolve) => setTimeout(resolve, 30))
    wrappers.splice(0).forEach((wrapper) => wrapper.unmount())
  })

  beforeEach(() => {
    shared.store.echoToAdd.content = ''
    shared.store.isUpdateMode = false
    vi.mocked(shared.store.removeExtension).mockClear()
    vi.mocked(shared.store.handleAddOrUpdateEcho).mockClear()
  })

  it('shows only the file preview without a switch', async () => {
    shared.store.filesToAdd = [{ id: 'file-1' }]
    shared.store.extensionPreview = null
    shared.store.currentMode = Mode.ECH0

    const wrapper = mountPreview()
    await flush()

    expect(wrapper.find('[data-test="media"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="renderer"]').exists()).toBe(false)
    expect(wrapper.find('.sketch-frame').exists()).toBe(true)
    expect(wrapper.find('.seg').exists()).toBe(false)
    expect(wrapper.text()).toContain('editor.preview')
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
    expect(wrapper.find('.seg').classes()).toContain('seg--compact')
    expect(wrapper.find('[data-test="media"]').exists()).toBe(true)

    await tabs[1]!.trigger('click')
    await flush()

    expect(wrapper.find('[data-test="media"]').exists()).toBe(false)
    expect(wrapper.find('.sketch-frame').exists()).toBe(true)
  })

  it('shows the text preview by default when only content exists', async () => {
    shared.store.echoToAdd.content = 'Hello **world**'
    shared.store.filesToAdd = []
    shared.store.extensionPreview = null
    shared.store.currentMode = Mode.ECH0

    const wrapper = mountPreview()
    await flush()

    expect(wrapper.find('[data-test="content-preview"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="media"]').exists()).toBe(false)
    expect(wrapper.find('.seg').exists()).toBe(false)
  })

  it('switches to files when an attachment is added', async () => {
    shared.store.echoToAdd.content = 'Hello'
    shared.store.filesToAdd = []
    shared.store.extensionPreview = null
    shared.store.currentMode = Mode.ECH0

    const wrapper = mountPreview()
    await flush()
    expect(wrapper.find('[data-test="content-preview"]').exists()).toBe(true)

    shared.store.filesToAdd = [{ id: 'file-1' }]
    await flush()

    expect(wrapper.find('[data-test="media"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="content-preview"]').exists()).toBe(false)
    expect(wrapper.findAll('.seg__btn')).toHaveLength(2)
  })

  it('offers text and files tabs and can switch back to text', async () => {
    shared.store.echoToAdd.content = 'Hello'
    shared.store.filesToAdd = [{ id: 'file-1' }]
    shared.store.extensionPreview = null
    shared.store.currentMode = Mode.ECH0

    const wrapper = mountPreview()
    await flush()

    const tabs = wrapper.findAll('.seg__btn')
    expect(tabs.map((tab) => tab.text())).toEqual(['editor.previewContent', 'editor.previewFiles'])
    expect(wrapper.find('[data-test="media"]').exists()).toBe(true)

    await tabs[0]!.trigger('click')
    await flush()

    expect(wrapper.find('[data-test="content-preview"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="media"]').exists()).toBe(false)
  })

  it('removes the extension from its preview', async () => {
    shared.store.filesToAdd = []
    shared.store.extensionPreview = websitePreview('Example')
    shared.store.currentMode = Mode.ECH0

    const wrapper = mountPreview()
    await flush()

    await wrapper.find('button[aria-label="editor.removeExtension"]').trigger('click')

    expect(shared.store.removeExtension).toHaveBeenCalledTimes(1)
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
