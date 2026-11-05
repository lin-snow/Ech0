// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import type { Ref } from 'vue'
import { watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { DEFAULT_SITE_TITLE } from '@/constants/site'
import { seoOverride } from '@/composables/useSeoOverride'

type SeoSystemSetting = {
  site_title?: string
}

export const DEFAULT_SEO_DESCRIPTION =
  'Ech0 is a next-generation open-source self-hosted platform built for individuals. It is lightweight and low-cost, making it easy to publish and share your ideas, writing, and links.'

const DEFAULT_OG_IMAGE = '/og-cover.jpg'
const OG_IMAGE_WIDTH = 1200
const OG_IMAGE_HEIGHT = 630
const WEBSITE_JSONLD_ID = 'website-jsonld'
const PROJECT_URL = 'https://github.com/lin-snow/Ech0'

const resolvePageTitle = (siteTitle: string, routeTitle: unknown) => {
  if (typeof routeTitle !== 'string') return siteTitle
  const normalizedRouteTitle = routeTitle.trim()
  if (!normalizedRouteTitle || normalizedRouteTitle === siteTitle) return siteTitle
  return `${normalizedRouteTitle} | ${siteTitle}`
}

const upsertMetaTag = (selector: string, attrs: Record<string, string>) => {
  const head = document.head
  if (!head) return

  let meta = head.querySelector<HTMLMetaElement>(selector)
  if (!meta) {
    meta = document.createElement('meta')
    head.appendChild(meta)
  }

  Object.entries(attrs).forEach(([key, value]) => {
    meta!.setAttribute(key, value)
  })
}

const removeMetaTag = (selector: string) => {
  document.head?.querySelector(selector)?.remove()
}

const upsertCanonicalLink = (href: string) => {
  const head = document.head
  if (!head) return

  let canonical = head.querySelector<HTMLLinkElement>('link[rel="canonical"]')
  if (!canonical) {
    canonical = document.createElement('link')
    canonical.rel = 'canonical'
    head.appendChild(canonical)
  }
  canonical.href = href
}

const updateJsonLd = (payload: Record<string, unknown>) => {
  const script = document.getElementById(WEBSITE_JSONLD_ID)
  if (!script) return
  script.textContent = JSON.stringify(payload)
}

export const useSeoHead = (systemSetting: Ref<SeoSystemSetting>) => {
  const route = useRoute()
  const { locale } = useI18n()

  // The server already rendered article meta for this document; rewriting it
  // before the echo payload arrives would replace a correct card with a
  // generic one.
  const serverRenderedEchoPath =
    route.name === 'echo' &&
    document.querySelector('meta[property="og:type"][content="article"]') !== null
      ? route.path
      : ''
  let awaitingEchoOverride = serverRenderedEchoPath !== ''

  const resolveSiteBaseUrl = () => window.location.origin

  const updateSeoMeta = () => {
    const activeOverride = seoOverride.value?.path === route.path ? seoOverride.value : null

    if (awaitingEchoOverride) {
      if (route.path !== serverRenderedEchoPath) {
        awaitingEchoOverride = false
      } else if (!activeOverride) {
        return
      } else {
        awaitingEchoOverride = false
      }
    }

    const siteTitle = (systemSetting.value.site_title || DEFAULT_SITE_TITLE).trim()
    const pageTitle = resolvePageTitle(siteTitle, activeOverride?.title ?? route.meta.title)
    const routeDescription =
      typeof route.meta.description === 'string' ? route.meta.description.trim() : ''
    const description = activeOverride?.description || routeDescription || DEFAULT_SEO_DESCRIPTION
    const baseUrl = resolveSiteBaseUrl()
    const canonicalUrl = new URL(route.path || '/', `${baseUrl}/`).toString()
    const noIndex = route.meta.noindex === true
    const ogType = activeOverride?.type ?? 'website'
    const ogImage = activeOverride?.image ?? new URL(DEFAULT_OG_IMAGE, `${baseUrl}/`).toString()
    const ogImageAlt = activeOverride?.title || siteTitle
    const ogLocale = locale.value.replace('-', '_')

    document.documentElement.lang = locale.value
    document.title = pageTitle
    upsertCanonicalLink(canonicalUrl)
    upsertMetaTag('meta[name="description"]', { name: 'description', content: description })
    upsertMetaTag('meta[name="robots"]', {
      name: 'robots',
      content: noIndex ? 'noindex,nofollow' : 'index,follow,max-image-preview:large',
    })
    upsertMetaTag('meta[property="og:title"]', { property: 'og:title', content: pageTitle })
    upsertMetaTag('meta[property="og:description"]', {
      property: 'og:description',
      content: description,
    })
    upsertMetaTag('meta[property="og:site_name"]', { property: 'og:site_name', content: siteTitle })
    upsertMetaTag('meta[property="og:url"]', { property: 'og:url', content: canonicalUrl })
    upsertMetaTag('meta[property="og:type"]', { property: 'og:type', content: ogType })
    upsertMetaTag('meta[property="og:image"]', { property: 'og:image', content: ogImage })
    upsertMetaTag('meta[property="og:image:alt"]', {
      property: 'og:image:alt',
      content: ogImageAlt,
    })
    upsertMetaTag('meta[property="og:locale"]', { property: 'og:locale', content: ogLocale })

    if (activeOverride?.image) {
      // A post image has unknown dimensions; a wrong pair makes unfurlers
      // reserve the wrong box.
      removeMetaTag('meta[property="og:image:width"]')
      removeMetaTag('meta[property="og:image:height"]')
      removeMetaTag('meta[property="og:image:type"]')
    } else {
      upsertMetaTag('meta[property="og:image:type"]', {
        property: 'og:image:type',
        content: 'image/jpeg',
      })
      upsertMetaTag('meta[property="og:image:width"]', {
        property: 'og:image:width',
        content: String(OG_IMAGE_WIDTH),
      })
      upsertMetaTag('meta[property="og:image:height"]', {
        property: 'og:image:height',
        content: String(OG_IMAGE_HEIGHT),
      })
    }

    if (ogType === 'article' && activeOverride?.publishedAt) {
      upsertMetaTag('meta[property="article:published_time"]', {
        property: 'article:published_time',
        content: activeOverride.publishedAt,
      })
    } else {
      removeMetaTag('meta[property="article:published_time"]')
    }

    upsertMetaTag('meta[name="twitter:card"]', {
      name: 'twitter:card',
      content: 'summary_large_image',
    })
    upsertMetaTag('meta[name="twitter:title"]', { name: 'twitter:title', content: pageTitle })
    upsertMetaTag('meta[name="twitter:description"]', {
      name: 'twitter:description',
      content: description,
    })
    upsertMetaTag('meta[name="twitter:image"]', { name: 'twitter:image', content: ogImage })
    upsertMetaTag('meta[name="twitter:image:alt"]', {
      name: 'twitter:image:alt',
      content: ogImageAlt,
    })

    const publisher = {
      '@type': 'Organization',
      name: siteTitle,
      url: baseUrl,
      logo: new URL('/Ech0.png', `${baseUrl}/`).toString(),
    }

    updateJsonLd(
      activeOverride?.jsonLd
        ? {
            ...activeOverride.jsonLd,
            url: canonicalUrl,
            mainEntityOfPage: { '@type': 'WebPage', '@id': canonicalUrl },
            publisher,
          }
        : {
            '@context': 'https://schema.org',
            '@type': 'WebSite',
            name: siteTitle,
            description,
            url: `${baseUrl}/`,
            publisher,
            sameAs: [PROJECT_URL],
          },
    )
  }

  watch(
    () => [
      route.fullPath,
      route.meta.title,
      route.meta.description,
      route.meta.noindex,
      seoOverride.value,
      locale.value,
    ],
    () => {
      updateSeoMeta()
    },
    { immediate: true },
  )

  watch(
    () => systemSetting.value.site_title,
    () => {
      updateSeoMeta()
    },
  )
}
