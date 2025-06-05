<template>
  <button
    @click="scrollToTop"
    id="backToTop"
    class="cursor-pointer rounded-full shadow hover:shadow-md bg-white ring-1 ring-inset ring-gray-100"
  >
    <Arrowup class="w-full h-full" />
  </button>
</template>

<script setup lang="ts">
import Arrowup from '../icons/arrowup.vue'

const scrollToTop = () => {
  window.scrollTo({ top: 0, behavior: 'smooth' })
}
</script>
