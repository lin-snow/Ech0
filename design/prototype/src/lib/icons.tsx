// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import type { ReactNode, SVGProps } from 'react'

export type IconName =
  | 'home'
  | 'tags'
  | 'globe'
  | 'sliders'
  | 'search'
  | 'close'
  | 'chevron-down'
  | 'plus'
  | 'refresh'
  | 'lock'
  | 'more'
  | 'grid'
  | 'rss'
  | 'github'

const paths: Record<IconName, ReactNode> = {
  home: (
    <>
      <path d="M3.5 10.4 12 3.4l8.5 7" />
      <path d="M5.6 9.4V20h12.8V9.4" />
      <path d="M10 20v-5.2h4V20" />
    </>
  ),
  tags: (
    <>
      <path d="M3.5 3.5h7.2l9.8 9.8-7.4 7.4-9.6-9.6V3.5Z" />
      <circle cx="7.9" cy="7.9" r="1.35" />
    </>
  ),
  globe: (
    <>
      <circle cx="12" cy="12" r="8.6" />
      <path d="M3.4 12h17.2" />
      <path d="M12 3.4c2.4 2.4 3.6 5.4 3.6 8.6S14.4 18.2 12 20.6c-2.4-2.4-3.6-5.4-3.6-8.6S9.6 5.8 12 3.4Z" />
    </>
  ),
  sliders: (
    <>
      <path d="M3 7h11.2M18.8 7H21M3 17h3.2M11 17h10" />
      <circle cx="16.4" cy="7" r="2.2" />
      <circle cx="8.6" cy="17" r="2.2" />
    </>
  ),
  search: (
    <>
      <circle cx="11" cy="11" r="6.8" />
      <path d="m20.2 20.2-4.4-4.4" />
    </>
  ),
  close: <path d="m6 6 12 12M18 6 6 18" />,
  'chevron-down': <path d="m6.4 9.4 5.6 5.6 5.6-5.6" />,
  plus: <path d="M12 4.8v14.4M4.8 12h14.4" />,
  refresh: (
    <>
      <path d="M20 5.6v5.2h-5.2" />
      <path d="M19.2 10.8a7.4 7.4 0 1 0-1.2 6.2" />
    </>
  ),
  lock: (
    <>
      <rect x="4.8" y="10.4" width="14.4" height="9.6" rx="1.6" />
      <path d="M8.2 10.4V7.8a3.8 3.8 0 0 1 7.6 0v2.6" />
    </>
  ),
  more: (
    <>
      <circle cx="5.2" cy="12" r="1.15" fill="currentColor" stroke="none" />
      <circle cx="12" cy="12" r="1.15" fill="currentColor" stroke="none" />
      <circle cx="18.8" cy="12" r="1.15" fill="currentColor" stroke="none" />
    </>
  ),
  grid: (
    <>
      <rect x="3.6" y="3.6" width="6.8" height="6.8" rx="1.2" />
      <rect x="13.6" y="3.6" width="6.8" height="6.8" rx="1.2" />
      <rect x="3.6" y="13.6" width="6.8" height="6.8" rx="1.2" />
      <rect x="13.6" y="13.6" width="6.8" height="6.8" rx="1.2" />
    </>
  ),
  rss: (
    <>
      <path d="M4.5 11.5a8 8 0 0 1 8 8" />
      <path d="M4.5 5.5a14 14 0 0 1 14 14" />
      <circle cx="6" cy="18" r="1.5" fill="currentColor" stroke="none" />
    </>
  ),
  github: (
    <path
      fill="currentColor"
      stroke="none"
      d="M12 2C6.48 2 2 6.58 2 12.25c0 4.53 2.87 8.37 6.84 9.73.5.1.68-.22.68-.49 0-.24-.01-.88-.01-1.73-2.78.62-3.37-1.37-3.37-1.37-.45-1.18-1.11-1.49-1.11-1.49-.91-.63.07-.62.07-.62 1 .07 1.53 1.06 1.53 1.06.89 1.56 2.34 1.11 2.91.85.09-.66.35-1.11.63-1.36-2.22-.26-4.56-1.14-4.56-5.07 0-1.12.39-2.03 1.03-2.75-.1-.26-.45-1.3.1-2.71 0 0 .84-.28 2.75 1.05a9.4 9.4 0 0 1 5 0c1.91-1.33 2.75-1.05 2.75-1.05.55 1.41.2 2.45.1 2.71.64.72 1.03 1.63 1.03 2.75 0 3.94-2.34 4.81-4.57 5.06.36.32.68.94.68 1.9 0 1.37-.01 2.47-.01 2.81 0 .27.18.6.69.49A10.02 10.02 0 0 0 22 12.25C22 6.58 17.52 2 12 2Z"
    />
  ),
}

type IconProps = SVGProps<SVGSVGElement> & {
  name: IconName
  size?: number
}

export function Icon({ name, size = 16, ...rest }: IconProps) {
  return (
    <svg
      viewBox="0 0 24 24"
      width={size}
      height={size}
      fill="none"
      stroke="currentColor"
      strokeWidth={1.6}
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
      focusable="false"
      {...rest}
    >
      {paths[name]}
    </svg>
  )
}

export function Wordmark({ compact = false }: { compact?: boolean }) {
  return (
    <span
      style={{
        fontFamily: 'var(--font-mono)',
        fontSize: compact ? 14 : 15,
        fontWeight: 600,
        letterSpacing: '-0.02em',
        lineHeight: 1,
      }}
    >
      ech0
    </span>
  )
}
