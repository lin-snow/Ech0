// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import type { Ref } from 'vue'
import { watch } from 'vue'
import { useRoute } from 'vue-router'

type SeoSystemSetting = {
  site_title?: string
  server_url?: string
}

const DEFAULT_SITE_TITLE = 'Ech0'
const DEFAULT_OG_IMAGE = '/Ech0.png'
const WEBSITE_JSONLD_ID = 'website-jsonld'
const DEFAULT_SEO_DESCRIPTION =
  'Ech0 is a next-generation open-source self-hosted platform built for individuals. It is lightweight and low-cost, making it easy to publish and share your ideas, writing, and links.'

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

const updateWebsiteJsonLd = (name: string, description: string, canonicalUrl: string) => {
  const script = document.getElementById(WEBSITE_JSONLD_ID)
  if (!script) return

  try {
    const raw = script.textContent || '{}'
    const jsonLd = JSON.parse(raw) as Record<string, unknown>
    jsonLd.name = name
    jsonLd.description = description
    jsonLd.url = canonicalUrl
    script.textContent = JSON.stringify(jsonLd)
  } catch {}
}

export const useSeoHead = (systemSetting: Ref<SeoSystemSetting>) => {
  const route = useRoute()

  const resolveSiteBaseUrl = () => {
    const configuredUrl = (systemSetting.value.server_url || '').trim()
    if (configuredUrl) {
      try {
        return new URL(configuredUrl).origin
      } catch {}
    }
    return window.location.origin
  }

  const updateSeoMeta = () => {
    const siteTitle = (systemSetting.value.site_title || DEFAULT_SITE_TITLE).trim()
    const pageTitle = resolvePageTitle(siteTitle, route.meta.title)
    const routeDescription =
      typeof route.meta.description === 'string' ? route.meta.description.trim() : ''
    const description = routeDescription || DEFAULT_SEO_DESCRIPTION
    const baseUrl = resolveSiteBaseUrl()
    const canonicalUrl = new URL(route.path || '/', `${baseUrl}/`).toString()
    const noIndex = route.meta.noindex === true
    const ogImage = new URL(DEFAULT_OG_IMAGE, `${baseUrl}/`).toString()

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
    upsertMetaTag('meta[property="og:type"]', { property: 'og:type', content: 'website' })
    upsertMetaTag('meta[property="og:image"]', { property: 'og:image', content: ogImage })
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
    updateWebsiteJsonLd(siteTitle, description, canonicalUrl)
  }

  watch(
    () => [route.fullPath, route.meta.title, route.meta.description, route.meta.noindex],
    () => {
      updateSeoMeta()
    },
    { immediate: true },
  )

  watch(
    () => [systemSetting.value.site_title, systemSetting.value.server_url],
    () => {
      updateSeoMeta()
    },
  )
}
