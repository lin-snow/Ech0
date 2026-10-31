// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { useEffect, useState, type ReactNode } from 'react'
import { Icon, type IconName } from './components/Icon'
import { Components } from './pages/Components'
import { Foundations } from './pages/Foundations'
import { Home } from './pages/Home'
import { Patterns } from './pages/Patterns'
import { Ux } from './pages/Ux'
import type { PageId, Route, RoutePage } from './routes'

type Theme = 'light' | 'dark'

const pages: { id: PageId; num: string; label: string; icon: IconName; element: ReactNode }[] = [
  { id: 'foundations', num: '00', label: '基础', icon: 'spark', element: <Foundations /> },
  { id: 'components', num: '01', label: '组件', icon: 'menu', element: <Components /> },
  { id: 'patterns', num: '02', label: '模式', icon: 'focus', element: <Patterns /> },
  { id: 'ux', num: '03', label: '结构与状态', icon: 'activity', element: <Ux /> },
]

function parseRoute(): Route {
  const raw = window.location.hash.replace(/^#\/?/, '')
  if (!raw || raw === 'home') return { page: 'home' }
  const [page, anchor] = raw.split('/')
  if (page === 'home') return { page: 'home' }
  return pages.some((item) => item.id === page)
    ? { page: page as PageId, anchor: anchor || undefined }
    : { page: 'home' }
}

export default function App() {
  const [route, setRoute] = useState<Route>(parseRoute)
  const [theme, setTheme] = useState<Theme>(() => {
    const query = new URLSearchParams(window.location.search).get('theme')
    if (query === 'light' || query === 'dark') return query
    return window.localStorage.getItem('ech0-design-theme') === 'dark' ? 'dark' : 'light'
  })

  useEffect(() => {
    window.localStorage.setItem('ech0-design-theme', theme)
  }, [theme])

  useEffect(() => {
    const onHashChange = () => setRoute(parseRoute())
    window.addEventListener('hashchange', onHashChange)
    return () => window.removeEventListener('hashchange', onHashChange)
  }, [])

  useEffect(() => {
    const active = route.page === 'home' ? null : pages.find((item) => item.id === route.page)
    document.title = active ? `Ech0 — ${active.label}` : 'Ech0 — Design Assets'

    if (!route.anchor) {
      window.scrollTo(0, 0)
      return
    }

    requestAnimationFrame(() => {
      const target = document.getElementById(route.anchor!)
      if (!target) return
      window.scrollTo(0, target.getBoundingClientRect().top + window.scrollY - 24)
    })
  }, [route])

  function navigate(nextPage: RoutePage, anchor?: string) {
    window.location.hash = anchor ? `/${nextPage}/${anchor}` : `/${nextPage}`
    setRoute({ page: nextPage, anchor })
  }

  const activePage =
    route.page === 'home'
      ? <Home onNavigate={navigate} />
      : (pages.find((item) => item.id === route.page)?.element ?? <Home onNavigate={navigate} />)

  return (
    <div className="gallery">
      <aside className="gallery-rail">
        <div className="brand">
          <div className="brand__row">
            <span className="brand__name">echo</span>
            <span className="square" aria-hidden="true" />
          </div>
          <span className="brand__meta">
            DESIGN SYSTEM
            <br />
            v0.1 / MAIN
          </span>
        </div>

        <nav className="nav" aria-label="设计系统页面">
          <button
            type="button"
            className="nav__item"
            data-active={route.page === 'home'}
            aria-current={route.page === 'home' ? 'page' : undefined}
            onClick={() => navigate('home')}
          >
            <span className="nav__num">IDX</span>
            <Icon name="home" size={15} />
            <span className="nav__label">资产库</span>
            <span className="nav__badge">HOME</span>
          </button>

          {pages.map((item) => (
            <button
              key={item.id}
              type="button"
              className="nav__item"
              data-active={item.id === route.page}
              aria-current={item.id === route.page ? 'page' : undefined}
              onClick={() => navigate(item.id)}
            >
              <span className="nav__num">{item.num}</span>
              <Icon name={item.icon} size={15} />
              <span className="nav__label">{item.label}</span>
              <span className="nav__badge">{item.id === 'ux' ? 'IA' : ''}</span>
            </button>
          ))}
        </nav>

        <div className="rail-footer">
          {route.page === 'home' ? null : (
            <div className="theme-toggle" role="group" aria-label="资产预览主题">
              <button
                type="button"
                aria-pressed={theme === 'light'}
                onClick={() => setTheme('light')}
              >
                <Icon name="sun" size={12} />
                LIGHT
              </button>
              <button
                type="button"
                aria-pressed={theme === 'dark'}
                onClick={() => setTheme('dark')}
              >
                <Icon name="moon" size={12} />
                DARK
              </button>
            </div>
          )}
          <span className="rail-meta">
            {route.page === 'home' ? (
              'NAVIGATION / INDEX'
            ) : (
              <>
                ASSET PREVIEW
                <br />
                TOKENS · PATTERNS
              </>
            )}
          </span>
        </div>
      </aside>

      <main className="gallery-main">
        <div className="asset-preview" data-theme={theme} key={route.page}>
          {activePage}
        </div>
      </main>
    </div>
  )
}
