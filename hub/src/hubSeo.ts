// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import type { RouteLocationNormalized } from 'vue-router'

export const HUB_SITE_ORIGIN = (
  import.meta.env.VITE_HUB_SITE_ORIGIN ?? 'https://hub.ech0.app'
).replace(/\/+$/, '')

const DEFAULT_DESCRIPTION =
  'Ech0 Hub — discover and connect with the resonating voices of the web.'

const KEYWORDS =
  'Ech0, Ech0 Hub, microblog, timeline, self-hosted, open source, RSS, feed, aggregate, decentralized'

const ROUTES = {
  home: {
    title: 'Ech0 Hub — where echoes meet and ideas resonate',
    description: DEFAULT_DESCRIPTION,
    path: '/',
    jsonLdType: 'WebSite',
  },
  explore: {
    title: 'Explore — Ech0 Hub',
    description:
      'Browse public Ech0 instances in one aggregated timeline — discover voices from across the web.',
    path: '/explore',
    jsonLdType: 'CollectionPage',
  },
} as const

const OG_IMAGE = `${HUB_SITE_ORIGIN}/android-chrome-512x512.png`

function ensureMeta(attr: 'name' | 'property', key: string, content: string) {
  const selector = attr === 'name' ? `meta[name="${key}"]` : `meta[property="${key}"]`
  let el = document.head.querySelector<HTMLMetaElement>(selector)
  if (!el) {
    el = document.createElement('meta')
    el.setAttribute(attr, key)
    document.head.appendChild(el)
  }
  el.setAttribute('content', content)
}

function ensureLinkCanonical(href: string) {
  let el = document.head.querySelector<HTMLLinkElement>('link[rel="canonical"]')
  if (!el) {
    el = document.createElement('link')
    el.setAttribute('rel', 'canonical')
    document.head.appendChild(el)
  }
  el.setAttribute('href', href)
}

function ensureJsonLd(content: string) {
  let el = document.head.querySelector<HTMLScriptElement>('script[data-hub-jsonld="route"]')
  if (!el) {
    el = document.createElement('script')
    el.type = 'application/ld+json'
    el.dataset.hubJsonld = 'route'
    document.head.appendChild(el)
  }
  el.textContent = content
}

export function applyHubRouteMeta(to: RouteLocationNormalized): void {
  const isExplore = to.name === 'explore' || to.path.startsWith('/explore')
  const cfg = isExplore ? ROUTES.explore : ROUTES.home
  const url = `${HUB_SITE_ORIGIN}${to.path === '/' ? '/' : to.path}`

  document.title = cfg.title

  ensureMeta('name', 'description', cfg.description)
  ensureMeta('name', 'keywords', KEYWORDS)

  ensureMeta('property', 'og:type', 'website')
  ensureMeta('property', 'og:site_name', 'Ech0 Hub')
  ensureMeta('property', 'og:title', cfg.title)
  ensureMeta('property', 'og:description', cfg.description)
  ensureMeta('property', 'og:url', url)
  ensureMeta('property', 'og:locale', 'en_US')
  ensureMeta('property', 'og:image', OG_IMAGE)

  ensureMeta('name', 'twitter:card', 'summary_large_image')
  ensureMeta('name', 'twitter:title', cfg.title)
  ensureMeta('name', 'twitter:description', cfg.description)
  ensureMeta('name', 'twitter:image', OG_IMAGE)

  ensureLinkCanonical(url)
  ensureJsonLd(
    JSON.stringify({
      '@context': 'https://schema.org',
      '@type': cfg.jsonLdType,
      name: cfg.title,
      description: cfg.description,
      url,
      isPartOf: {
        '@type': 'WebSite',
        name: 'Ech0 Hub',
        url: `${HUB_SITE_ORIGIN}/`,
      },
      image: OG_IMAGE,
      inLanguage: 'en-US',
    }),
  )
}
