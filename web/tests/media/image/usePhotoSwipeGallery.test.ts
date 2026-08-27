// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { defineComponent, ref } from 'vue'
import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { usePhotoSwipeGallery } from '@/components/advanced/media/image/composables/usePhotoSwipeGallery'

type MockPhotoSwipeInstance = {
  options: Record<string, unknown>
  init: ReturnType<typeof vi.fn>
  destroy: ReturnType<typeof vi.fn>
  on: ReturnType<typeof vi.fn>
  addFilter: ReturnType<typeof vi.fn>
}

const photoSwipeInstances = vi.hoisted(() => [] as MockPhotoSwipeInstance[])
const requestedImageSrcs = vi.hoisted(() => [] as string[])
const imageSizeMap = vi.hoisted(
  () => new Map<string, { width: number; height: number; delayMs?: number }>(),
)

vi.mock('photoswipe', () => ({
  default: class MockPhotoSwipe {
    options: Record<string, unknown>
    init = vi.fn()
    destroy = vi.fn()
    on = vi.fn()
    addFilter = vi.fn()

    constructor(options: Record<string, unknown>) {
      this.options = options
      photoSwipeInstances.push(this as MockPhotoSwipeInstance)
    }
  },
}))

class MockImage {
  onload: (() => void) | null = null
  onerror: (() => void) | null = null
  naturalWidth = 0
  naturalHeight = 0

  set src(value: string) {
    requestedImageSrcs.push(value)
    const resolved = imageSizeMap.get(value)

    if (!resolved) {
      this.onerror?.()
      return
    }

    const trigger = () => {
      this.naturalWidth = resolved.width
      this.naturalHeight = resolved.height
      this.onload?.()
    }

    if (resolved.delayMs && resolved.delayMs > 0) {
      setTimeout(trigger, resolved.delayMs)
      return
    }

    trigger()
  }
}

const createHost = (items: Array<{ src: string; width?: number; height?: number; alt?: string }>) =>
  mount(
    defineComponent({
      setup() {
        const { open } = usePhotoSwipeGallery(ref(items))
        return { open }
      },
      template: '<div />',
    }),
  )

const waitForOpen = async () => {
  await new Promise<void>((resolve) => setTimeout(resolve, 0))
  await Promise.resolve()
}

describe('usePhotoSwipeGallery', () => {
  beforeEach(() => {
    photoSwipeInstances.length = 0
    requestedImageSrcs.length = 0
    imageSizeMap.clear()
    vi.stubGlobal('Image', MockImage)
  })

  it('prefetches all missing dimensions before opening', async () => {
    imageSizeMap.set('/a.jpg', { width: 640, height: 360 })
    imageSizeMap.set('/c.jpg', { width: 1200, height: 900 })

    const wrapper = createHost([
      { src: '/a.jpg', width: 0, height: undefined, alt: 'a' },
      { src: '/b.jpg', width: 800, height: 600, alt: 'b' },
      { src: '/c.jpg', alt: 'c' },
    ])
    ;(wrapper.vm as { open: (index: number) => void }).open(99)
    await waitForOpen()

    const options = photoSwipeInstances[0]?.options as {
      index: number
      dataSource: Array<{ src: string; width?: number; height?: number }>
    }

    expect(options.index).toBe(2)
    expect(options.dataSource[0]).toMatchObject({
      src: '/a.jpg',
      width: 640,
      height: 360,
    })
    expect(options.dataSource[1]).toMatchObject({
      src: '/b.jpg',
      width: 800,
      height: 600,
    })
    expect(options.dataSource[2]).toMatchObject({
      src: '/c.jpg',
      width: 1200,
      height: 900,
    })

    expect(requestedImageSrcs).toEqual(['/a.jpg', '/c.jpg'])
    expect(photoSwipeInstances[0]?.init).toHaveBeenCalledTimes(1)

    wrapper.unmount()
  })

  it('uses thumbnail natural size for current slide and keeps msrc placeholder behavior', async () => {
    imageSizeMap.set('/first.jpg', { width: 999, height: 777 })
    imageSizeMap.set('/second.jpg', { width: 700, height: 400 })

    const wrapper = createHost([
      { src: '/first.jpg' },
      { src: '/second.jpg' },
    ])

    const sourceElement = document.createElement('button')
    const img = document.createElement('img')
    img.src = '/thumb.jpg'
    Object.defineProperty(img, 'naturalWidth', { value: 500, configurable: true })
    Object.defineProperty(img, 'naturalHeight', { value: 300, configurable: true })
    sourceElement.appendChild(img)

    ;(wrapper.vm as { open: (index: number, sourceElement?: HTMLElement | null) => void }).open(
      0,
      sourceElement,
    )
    await waitForOpen()

    const options = photoSwipeInstances[0]?.options as {
      dataSource: Array<Record<string, unknown> & { width?: number; height?: number }>
    }
    const currentSlide = options.dataSource[0]

    expect(currentSlide.width).toBe(500)
    expect(currentSlide.height).toBe(300)
    expect(currentSlide.element).toBe(sourceElement)
    expect(String(currentSlide.msrc)).toContain('/thumb.jpg')

    const addFilter = photoSwipeInstances[0]?.addFilter
    expect(addFilter).toHaveBeenCalledWith('placeholderSrc', expect.any(Function))
    const placeholderFilter = addFilter?.mock.calls[0]?.[1] as (
      placeholderSrc: string,
      content: { slide?: { index?: number } },
    ) => string

    expect(placeholderFilter('fallback', { slide: { index: 0 } })).toContain('/thumb.jpg')
    expect(placeholderFilter('fallback', { slide: { index: 1 } })).toBe('fallback')

    wrapper.unmount()
  })

  it('only keeps the latest open request when multiple opens overlap', async () => {
    imageSizeMap.set('/slow.jpg', { width: 1200, height: 800, delayMs: 30 })
    imageSizeMap.set('/fast.jpg', { width: 800, height: 600 })

    const wrapper = createHost([{ src: '/slow.jpg' }, { src: '/fast.jpg' }])
    const open = (wrapper.vm as { open: (index: number) => void }).open

    open(0)
    open(1)

    await new Promise<void>((resolve) => setTimeout(resolve, 50))
    await waitForOpen()

    expect(photoSwipeInstances).toHaveLength(1)
    const options = photoSwipeInstances[0]?.options as { index: number }
    expect(options.index).toBe(1)

    wrapper.unmount()
  })
})
