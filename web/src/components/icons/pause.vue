<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24">
    <!-- Icon from Solar by 480 Design - https://creativecommons.org/licenses/by/4.0/ -->
    <path
      :fill="color"
      d="M8.25 3A2.25 2.25 0 0 0 6 5.25v13.5A2.25 2.25 0 0 0 8.25 21h.5A2.25 2.25 0 0 0 11 18.75V5.25A2.25 2.25 0 0 0 8.75 3zm7 0A2.25 2.25 0 0 0 13 5.25v13.5A2.25 2.25 0 0 0 15.25 21h.5A2.25 2.25 0 0 0 18 18.75V5.25A2.25 2.25 0 0 0 15.75 3z"
    />
  </svg>
</template>

<script lang="ts">
export default {
  name: 'SolarPauseBold',
  props: {
    color: {
      type: String,
      default: '#888888',
    },
  },
}
</script>
