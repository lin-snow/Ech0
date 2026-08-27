// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { onBeforeUnmount, shallowRef, type Ref } from 'vue'
import PhotoSwipe from 'photoswipe'
import 'photoswipe/style.css'

type PhotoSwipeItem = {
  src: string
  width?: number
  height?: number
  alt?: string
}

type ResolvedDimensions = {
  width?: number
  height?: number
}

const resolveThumbImageSrc = (sourceElement?: HTMLElement | null) => {
  if (!sourceElement) return undefined

  const thumbnail =
    sourceElement instanceof HTMLImageElement
      ? sourceElement
      : (sourceElement.querySelector('img') as HTMLImageElement | null)

  if (!thumbnail) return undefined

  return thumbnail.currentSrc || thumbnail.src || undefined
}

const resolveThumbNaturalSize = (sourceElement?: HTMLElement | null): ResolvedDimensions => {
  if (!sourceElement) return {}

  const thumbnail =
    sourceElement instanceof HTMLImageElement
      ? sourceElement
      : (sourceElement.querySelector('img') as HTMLImageElement | null)

  if (!thumbnail) return {}

  const width = thumbnail.naturalWidth
  const height = thumbnail.naturalHeight
  if (width > 0 && height > 0) {
    return { width, height }
  }

  return {}
}

const isValidDimension = (value?: number) =>
  typeof value === 'number' && Number.isFinite(value) && value > 0

const hasValidDimensions = (item: PhotoSwipeItem) =>
  isValidDimension(item.width) && isValidDimension(item.height)

const loadImageNaturalSize = (src: string) =>
  new Promise<ResolvedDimensions>((resolve) => {
    const image = new Image()
    image.onload = () => {
      const width = image.naturalWidth
      const height = image.naturalHeight
      if (width > 0 && height > 0) {
        resolve({ width, height })
        return
      }
      resolve({})
    }
    image.onerror = () => resolve({})
    image.src = src
  })

export const usePhotoSwipeGallery = (items: Ref<PhotoSwipeItem[]>) => {
  const instance = shallowRef<PhotoSwipe | null>(null)
  const dimensionCache = new Map<string, Promise<ResolvedDimensions>>()
  let latestOpenToken = 0

  const destroy = () => {
    instance.value?.destroy()
    instance.value = null
  }

  const resolveItemDimensions = (
    item: PhotoSwipeItem,
    override?: ResolvedDimensions,
  ): Promise<ResolvedDimensions> => {
    if (override?.width && override?.height) {
      return Promise.resolve(override)
    }

    if (hasValidDimensions(item)) {
      return Promise.resolve({
        width: item.width,
        height: item.height,
      })
    }

    if (!item.src) return Promise.resolve({})

    const cached = dimensionCache.get(item.src)
    if (cached) return cached

    const request = loadImageNaturalSize(item.src)
    dimensionCache.set(item.src, request)
    return request
  }

  const resolveAllImageDimensions = async (
    currentIndex: number,
    sourceElement?: HTMLElement | null,
  ) => {
    const currentSlideNaturalSize = resolveThumbNaturalSize(sourceElement)

    return Promise.all(
      items.value.map((item, index) => {
        const override = index === currentIndex ? currentSlideNaturalSize : undefined
        return resolveItemDimensions(item, override)
      }),
    )
  }

  const open = (startIndex: number, sourceElement?: HTMLElement | null) => {
    const openToken = ++latestOpenToken
    void (async () => {
      if (!items.value.length) return

      destroy()

      const normalizedIndex = Math.min(Math.max(startIndex, 0), items.value.length - 1)
      const thumbImageSrc = resolveThumbImageSrc(sourceElement)
      const resolvedDimensions = await resolveAllImageDimensions(normalizedIndex, sourceElement)

      if (openToken !== latestOpenToken || !items.value.length) return

      const dataSource = items.value.map((item, index) => {
        const resolved = resolvedDimensions[index] || {}
        const baseItem = {
          src: item.src,
          width: resolved.width,
          height: resolved.height,
          alt: item.alt,
        }

        if (index === normalizedIndex && sourceElement) {
          return {
            ...baseItem,
            element: sourceElement,
            msrc: thumbImageSrc || item.src,
          }
        }

        if (index === normalizedIndex && thumbImageSrc) {
          return {
            ...baseItem,
            msrc: thumbImageSrc,
          }
        }

        return baseItem
      })

      const pswp = new PhotoSwipe({
        dataSource,
        index: normalizedIndex,
        bgOpacity: 0.85,
        showHideAnimationType: 'zoom',
        preloadFirstSlide: true,
        escKey: true,
        arrowKeys: true,
        bgClickAction: 'close',
        closeOnVerticalDrag: true,
        clickToCloseNonZoomable: true,
      })

      pswp.on('destroy', () => {
        if (instance.value === pswp) {
          instance.value = null
        }
      })

      if (thumbImageSrc) {
        pswp.addFilter('placeholderSrc', (placeholderSrc, content) => {
          if (content.slide?.index === normalizedIndex) {
            return thumbImageSrc
          }
          return placeholderSrc
        })
      }

      instance.value = pswp
      pswp.init()
    })()
  }

  onBeforeUnmount(() => {
    destroy()
  })

  return {
    open,
    destroy,
  }
}
