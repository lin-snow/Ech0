// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

export function focusSearch() {
  const input = document.querySelector<HTMLInputElement>('[data-search-input]')
  if (!input) return
  input.focus()
  input.setSelectionRange(input.value.length, input.value.length)
}

export function focusComposer() {
  const editors = Array.from(document.querySelectorAll<HTMLTextAreaElement>('[data-composer]'))
  const editor = editors.find((element) => element.offsetParent !== null) ?? editors[0]
  if (!editor) return
  editor.focus()
  editor.scrollIntoView({ block: 'center', behavior: 'smooth' })
}
