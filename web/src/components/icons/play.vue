<template>
  <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24">
    <!-- Icon from Solar by 480 Design - https://creativecommons.org/licenses/by/4.0/ -->
    <path
      :fill="color"
      d="M21.409 9.353a2.998 2.998 0 0 1 0 5.294L8.597 21.614C6.534 22.737 4 21.277 4 18.968V5.033c0-2.31 2.534-3.769 4.597-2.648z"
    />
  </svg>
</template>

<script lang="ts">
export default {
  name: 'SolarPlayBold',
  props: {
    color: {
      type: String,
      default: '#888888',
    },
  },
}
</script>
