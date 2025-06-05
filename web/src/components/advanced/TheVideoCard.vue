<template>
  <div
    class="max-w-sm flex justify-center items-center bg-white rounded-lg shadow-sm ring-1 ring-inset ring-gray-100 p-2 gap-2"
  >
    <!-- 使用移动端外链播放器（解决跳转） -->
    <iframe
      :src="`https://www.bilibili.com/blackboard/html5mobileplayer.html?bvid=${props.bvnumber}&as_wide=1&high_quality=1&danmaku=0`"
      scrolling="no"
      border="0"
      frameborder="no"
      framespacing="0"
      allowfullscreen="true"
      loading="lazy"
      class="w-full aspect-video rounded"
    ></iframe>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  bvnumber: string
}>()
</script>

<style scoped></style>
