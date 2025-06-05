<template>
  <div class="px-4 pb-4 py-2 mt-4 mb-10 mx-auto flex justify-center items-center h-screen">
    <div
      class="h-3/5 sm:h-1/2 max-w-sm sm:max-w-md md:max-w-lg px-2 sm:px-4 md:px-6 my-4 sm:my-5 md:my-6"
    >
      <h1 class="text-5xl italic font-bold text-center text-gray-300 mb-5">Ech0s Connect</h1>
      <div class="mb-5">
        <!-- 返回首页 -->
        <div class="mb-2">
          <BaseButton
            @click="$router.push('/')"
            class="text-gray-600 rounded-md !shadow-none !border-none !ring-0 !bg-transparent group"
            title="返回首页"
          >
            <Arrow
              class="w-9 h-9 rotate-180 transition-transform duration-200 group-hover:-translate-x-1"
            />
          </BaseButton>
        </div>

        <TheConnects />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import BaseButton from '@/components/common/BaseButton.vue'
import Arrow from '@/components/icons/arrow.vue'
import TheConnects from './TheConnects.vue'
</script>
