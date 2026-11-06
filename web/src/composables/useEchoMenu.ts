// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { ref } from 'vue'

// There is only one action menu at a time. Keeping the id outside the card
// instances lets the virtualizer keep the open row alive while it scrolls.
const activeMenuId = ref<string | null>(null)

export function useEchoMenu() {
  return { activeMenuId }
}
