// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { defineComponent } from 'vue'
import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import { ImageLayout } from '@/enums/enums'
import TheImageGallery from '@/components/advanced/media/image/TheImageGallery.vue'

vi.mock('@/components/advanced/media/image/composables/usePhotoSwipeGallery', () => ({
  usePhotoSwipeGallery: () => ({
    open: vi.fn(),
    destroy: vi.fn(),
  }),
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: { index?: number }) =>
        params?.index ? `${key}-${params.index}` : key,
    }),
  }
})

const GalleryWaterfallStub = defineComponent({
  name: 'GalleryWaterfall',
  props: {
    images: { type: Array, required: true },
    resolvedSrcs: { type: Array, required: true },
    getAlt: { type: Function, required: true },
    getImageKey: { type: Function, required: true },
    isLoaded: { type: Function, required: true },
    markLoaded: { type: Function, required: true },
    open: { type: Function, required: true },
    getAspectRatioStyle: { type: Function, required: true },
  },
  template: '<div data-test="waterfall" />',
})

const layoutStubs = {
  GalleryWaterfall: GalleryWaterfallStub,
  GalleryGrid: defineComponent({ name: 'GalleryGrid', template: '<div data-test="grid" />' }),
  GalleryCarousel: defineComponent({
    name: 'GalleryCarousel',
    template: '<div data-test="carousel" />',
  }),
  GalleryHorizontal: defineComponent({
    name: 'GalleryHorizontal',
    template: '<div data-test="horizontal" />',
  }),
  GalleryStack: defineComponent({
    name: 'GalleryStack',
    template: '<div data-test="stack" />',
  }),
}

const createImage = (overrides: Record<string, unknown> = {}) =>
  ({
    id: 'img-1',
    url: '/image-1.jpg',
    width: 300,
    height: 200,
    ...overrides,
  }) as App.Api.Ech0.FileObject

describe('TheImageGallery', () => {
  it('falls back to waterfall layout for unsupported layout values', () => {
    const wrapper = mount(TheImageGallery, {
      props: {
        images: [createImage()],
        layout: 'unexpected-layout',
      },
      global: {
        stubs: layoutStubs,
      },
    })

    expect(wrapper.find('[data-test="waterfall"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="grid"]').exists()).toBe(false)
    expect(wrapper.find('[data-test="carousel"]').exists()).toBe(false)
    expect(wrapper.find('[data-test="horizontal"]').exists()).toBe(false)
    expect(wrapper.find('[data-test="stack"]').exists()).toBe(false)
  })

  it('renders stack layout when requested', () => {
    const wrapper = mount(TheImageGallery, {
      props: {
        images: [createImage()],
        layout: ImageLayout.STACK,
      },
      global: {
        stubs: layoutStubs,
      },
    })

    expect(wrapper.find('[data-test="stack"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="waterfall"]').exists()).toBe(false)
  })

  it('cleans loaded state when image dataset changes', async () => {
    const firstImage = createImage({ id: 'img-a', url: '/a.jpg' })
    const secondImage = createImage({ id: 'img-b', url: '/b.jpg' })

    const wrapper = mount(TheImageGallery, {
      props: {
        images: [firstImage],
        layout: ImageLayout.WATERFALL,
      },
      global: {
        stubs: layoutStubs,
      },
    })

    const beforeChange = wrapper.findComponent(GalleryWaterfallStub)
    const markLoaded = beforeChange.props('markLoaded') as (
      image: App.Api.Ech0.FileObject,
      idx: number,
    ) => void
    const isLoadedBefore = beforeChange.props('isLoaded') as (
      image: App.Api.Ech0.FileObject,
      idx: number,
    ) => boolean

    expect(isLoadedBefore(firstImage, 0)).toBe(false)
    markLoaded(firstImage, 0)
    expect(isLoadedBefore(firstImage, 0)).toBe(true)

    await wrapper.setProps({ images: [secondImage] })

    const afterChange = wrapper.findComponent(GalleryWaterfallStub)
    const isLoadedAfter = afterChange.props('isLoaded') as (
      image: App.Api.Ech0.FileObject,
      idx: number,
    ) => boolean

    expect(isLoadedAfter(firstImage, 0)).toBe(false)
    expect(isLoadedAfter(secondImage, 0)).toBe(false)
  })
})
