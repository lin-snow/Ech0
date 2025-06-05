<template>
  <div
    class="max-w-sm flex justify-center items-center bg-white rounded-lg shadow-sm ring-1 ring-inset ring-gray-100 p-2 gap-2"
  >
    <a :href="websiteInfo.site" target="_blank">
      <div class="flex justify-between items-center">
        <div class="shrink-0 px-2">
          <Link class="w-5 h-5" />
        </div>
        <span class="text-lg font-bold text-gray-600 truncate max-w-30">{{
          websiteInfo.title
        }}</span>
      </div>
    </a>
  </div>
</template>

<script setup lang="ts">
import Link from '../icons/link.vue'

const props = defineProps<{
  website: string
}>()

const websiteInfo = JSON.parse(props.website) as { title: string; site: string }
</script>
