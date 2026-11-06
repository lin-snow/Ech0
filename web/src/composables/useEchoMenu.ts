// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { shallowRef } from 'vue'

type Echo = App.Api.Ech0.Echo

export type EchoMenuState = {
  echo: Echo
  top: number
  right: number
}

const MENU_GAP = 8

// One action menu for the whole app. The menu is deliberately independent of
// the card's DOM: a virtual list may recycle the row at any time, but the menu
// still has the Echo and can finish the update/delete.
const menu = shallowRef<EchoMenuState | null>(null)
let openedAt = 0

export function useEchoMenu() {
  const openMenu = (echo: Echo, trigger: HTMLElement) => {
    const rect = trigger.getBoundingClientRect()
    menu.value = {
      echo,
      top: Math.round(rect.bottom + MENU_GAP),
      right: Math.max(MENU_GAP, Math.round(window.innerWidth - rect.right)),
    }
    openedAt = performance.now()
  }

  const closeMenu = () => {
    menu.value = null
  }

  const isMenuOpen = (id: string) => menu.value?.echo.id === id
  const canCloseOnScroll = () => performance.now() - openedAt > 200

  return { menu, openMenu, closeMenu, isMenuOpen, canCloseOnScroll }
}
