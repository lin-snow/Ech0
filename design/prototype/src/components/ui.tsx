// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { Tooltip } from '@base-ui/react/tooltip'
import type { ButtonHTMLAttributes, CSSProperties, ReactElement, ReactNode } from 'react'
import { Icon, type IconName } from '../lib/icons'
import styles from './ui.module.css'

export function cx(...parts: Array<string | false | null | undefined>): string {
  return parts.filter(Boolean).join(' ')
}

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: 'primary' | 'default' | 'ghost'
  size?: 'sm' | 'md'
  full?: boolean
  icon?: IconName
}

export function Button({
  variant = 'default',
  size = 'md',
  full = false,
  icon,
  className,
  children,
  ...rest
}: ButtonProps) {
  return (
    <button
      type="button"
      className={cx(
        styles.button,
        styles[`button--${variant}`],
        size === 'sm' && styles['button--sm'],
        full && styles['button--full'],
        className,
      )}
      {...rest}
    >
      {icon && <Icon name={icon} size={size === 'sm' ? 12 : 14} />}
      {children}
    </button>
  )
}

type IconButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  icon: IconName
  label: string
  size?: number
}

export function IconButton({
  icon,
  label,
  size = 15,
  className,
  ...rest
}: IconButtonProps) {
  return (
    <Tip label={label}>
      <button
        type="button"
        aria-label={label}
        className={cx(styles.iconButton, className)}
        {...rest}
      >
        <Icon name={icon} size={size} />
      </button>
    </Tip>
  )
}

export function Tip({
  label,
  side = 'top',
  children,
}: {
  label: ReactNode
  side?: 'top' | 'right' | 'bottom' | 'left'
  children: ReactNode
}) {
  return (
    <Tooltip.Root>
      <Tooltip.Trigger render={children as ReactElement} />
      <Tooltip.Portal>
        <Tooltip.Positioner side={side} sideOffset={6} className={styles.tipPositioner}>
          <Tooltip.Popup className={styles.tipPopup}>{label}</Tooltip.Popup>
        </Tooltip.Positioner>
      </Tooltip.Portal>
    </Tooltip.Root>
  )
}

export function Kbd({ children }: { children: ReactNode }) {
  return <kbd className={styles.kbd}>{children}</kbd>
}

export function Mono({
  children,
  className,
  style,
}: {
  children: ReactNode
  className?: string
  style?: CSSProperties
}) {
  return (
    <span className={cx(styles.mono, className)} style={style}>
      {children}
    </span>
  )
}

export function Avatar({ src, size = 28, alt = '' }: { src: string; size?: number; alt?: string }) {
  return (
    <img
      className={styles.avatar}
      src={src}
      alt={alt}
      style={{ width: size, height: size }}
    />
  )
}
