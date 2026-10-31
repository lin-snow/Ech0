// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { Tooltip } from '@base-ui/react/tooltip'
import { useEffect, useRef, useState } from 'react'
import app from './app.module.css'
import { DesignBar, type LayoutVariant, type Theme } from './components/DesignBar'
import { focusSearch } from './lib/focusSearch'
import { FocusLayout } from './layouts/Focus'
import { ThreeColumnLayout } from './layouts/ThreeColumn'

const DEFAULT_WIDTH = 1024

export default function App() {
  const [variant, setVariant] = useState<LayoutVariant>('three')
  const [theme, setTheme] = useState<Theme>('light')
  const [width, setWidth] = useState(DEFAULT_WIDTH)
  const [frameWidth, setFrameWidth] = useState(DEFAULT_WIDTH)
  const frameRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    document.documentElement.dataset.theme = theme
  }, [theme])

  useEffect(() => {
    const frame = frameRef.current
    if (!frame) return undefined
    const update = () => {
      const rect = frame.getBoundingClientRect()
      setFrameWidth(rect.width)
      const root = document.documentElement.style
      root.setProperty('--frame-left', `${rect.left}px`)
      root.setProperty('--frame-top', `${rect.top}px`)
      root.setProperty('--frame-width', `${rect.width}px`)
      root.setProperty('--frame-height', `${rect.height}px`)
    }
    update()
    const observer = new ResizeObserver(update)
    observer.observe(frame)
    window.addEventListener('resize', update)
    return () => {
      observer.disconnect()
      window.removeEventListener('resize', update)
    }
  }, [])

  useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      if (event.metaKey || event.ctrlKey) {
        if (event.key.toLowerCase() === 'k') {
          event.preventDefault()
          focusSearch()
        }
        return
      }
      const target = event.target as HTMLElement | null
      if (
        target &&
        (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable)
      ) {
        return
      }
      if (event.key === '1') setVariant('three')
      if (event.key === '2') setVariant('focus')
      if (event.key.toLowerCase() === 't') {
        setTheme((current) => (current === 'light' ? 'dark' : 'light'))
      }
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [])

  return (
    <Tooltip.Provider delay={380}>
      <div className={app.app}>
        <div className={app.stage}>
          <div className={app.stageBody}>
            <div ref={frameRef} className={app.frame} style={{ width }} data-variant={variant}>
              {variant === 'three' && <ThreeColumnLayout frameWidth={frameWidth} />}
              {variant === 'focus' && <FocusLayout />}
            </div>
          </div>
        </div>

        <DesignBar
          variant={variant}
          onVariant={setVariant}
          theme={theme}
          onTheme={setTheme}
          width={width}
          onWidth={setWidth}
          onReset={() => {
            setVariant('three')
            setTheme('light')
            setWidth(DEFAULT_WIDTH)
          }}
        />
      </div>
    </Tooltip.Provider>
  )
}
