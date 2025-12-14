<template>
  <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24">
    <!-- Icon from Flowbite Icons by Themesberg - https://github.com/themesberg/flowbite-icons/blob/main/LICENSE -->
    <path 
      fill="none"
      :stroke="color"
      stroke-linecap="round"
      stroke-linejoin="round" 
      stroke-width="2"
      d="m16 4 3 3H5v3m3 10-3-3h14v-3m-9-2.5 2-1.5v4"
    />
  </svg>
</template>
  
<script lang="ts">
export default {
name: 'FlowbiteArrowsRepeatCountOutline',
props: {
    color: {
    type: String,
    default: '#888888',
    },
},
}
</script>
  