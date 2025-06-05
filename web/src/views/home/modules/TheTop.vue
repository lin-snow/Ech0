<template>
  <div class="mx-auto flex justify-between items-center mb-2">
    <div class="w-full flex justify-between items-center">
      <BaseInput
        title="搜索"
        type="text"
        v-model="searchContent"
        placeholder="搜索..."
        class="w-2/3 h-10"
        @keyup.enter="$event.target.blur()"
        @blur="handleSearch"
      />
      <!-- RSS -->
      <div class="relative mr-1">
        <a href="/rss" title="RSS">
          <!-- icon -->
          <Rss class="w-8 h-8 text-gray-400" />
        </a>
      </div>
      <!-- ConnectPage -->
      <div class="relative mr-2 xl:hidden">
        <RouterLink to="/connect" title="连接">
          <!-- icon -->
          <Other class="w-8 h-8 text-gray-400" />
        </RouterLink>
        <!-- <span class="absolute -top-1 -right-1 block w-2 h-2 bg-green-500 rounded-full ring-1 ring-white"></span> -->
      </div>
      <!-- PanelPage -->
      <div class="relative mr-1">
        <RouterLink to="/panel" title="面板">
          <!-- icon -->
          <Panel class="w-8 h-8 text-gray-400" />
        </RouterLink>
        <!-- <span class="absolute -top-1 -right-1 block w-2 h-2 bg-green-500 rounded-full ring-1 ring-white"></span> -->
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import BaseInput from '@/components/common/BaseInput.vue'
import Other from '@/components/icons/other.vue'
import Panel from '@/components/icons/panel.vue'
import Rss from '@/components/icons/rss.vue'
import { RouterLink } from 'vue-router'
import { useEchoStore } from '@/stores/echo'
import { storeToRefs } from 'pinia'
import { ref } from 'vue'
const echoStore = useEchoStore()
const { refreshForSearch, getEchosByPage } = echoStore
const { searchingMode } = storeToRefs(echoStore)

const searchContent = ref<string>('')

const handleSearch = () => {
  // 设置搜索内容
  echoStore.searchValue = searchContent.value

  // 判断是否是搜索模式
  if (searchingMode.value) {
    // 初始化搜索
    refreshForSearch()
    // 开始搜索
    getEchosByPage()
  }
}

// const notDeveloper = () => {
//   theToast.info('该功能正在开发中，请耐心等待！')
// }
</script>
