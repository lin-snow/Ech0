<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <svg class="done-glyph" viewBox="0 0 24 24" width="1em" height="1em" aria-hidden="true">
    <path
      class="done-glyph__path"
      d="M5 12.5 10 17.5 19 7"
      pathLength="1"
      fill="none"
      stroke="currentColor"
      stroke-width="2.4"
      stroke-linecap="round"
      stroke-linejoin="round"
    />
  </svg>
</template>

<style scoped>
/* A check that draws itself once: enough to mark a settled trace without
   borrowing the old multi-state sprite. */
.done-glyph {
  display: inline-block;
  width: 1em;
  height: 1em;
  overflow: visible;
  vertical-align: -0.16em;
}

.done-glyph__path {
  stroke-dasharray: 1;
  stroke-dashoffset: 1;
  animation: done-glyph-draw 0.55s cubic-bezier(0.22, 1, 0.36, 1) 0.08s forwards;
}

@keyframes done-glyph-draw {
  to {
    stroke-dashoffset: 0;
  }
}

@media (prefers-reduced-motion: reduce) {
  .done-glyph__path {
    animation: none;
    stroke-dashoffset: 0;
  }
}
</style>
