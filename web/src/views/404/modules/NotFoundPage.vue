<template>
  <div class="flex justify-center items-center h-screen relative">
    <div>
      <div class="text-7xl italic font-bold text-center text-gray-300 mb-4 flex-wrap">
        <span> 404 </span>
        <span class="text-nowrap"> 页面不存在 </span>
      </div>
      <BaseButton
        @click="$router.push({ name: 'home' })"
        title="返回首页"
        class="rounded-md absolute top-10 left-10"
      >
        <Arrow class="text-3xl rotate-180" />
      </BaseButton>
    </div>
  </div>
</template>

<script setup lang="ts">
import BaseButton from '@/components/common/BaseButton.vue'
import Arrow from '@/components/icons/arrow.vue'
</script>
