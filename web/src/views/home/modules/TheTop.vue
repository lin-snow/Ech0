<template>
  <div class="mx-auto mb-1 sm:mb-0">
    <div class="w-full flex justify-between items-center">
      <!-- 搜索与过滤 -->
      <div class="flex justify-start items-center gap-2">
        <BaseInput
          v-if="!isFilteringMode"
          title="搜索"
          type="text"
          v-model="searchContent"
          placeholder="搜索..."
          class="w-42! h-10 bg-[var(--input-bg-color)]"
          @keyup.enter="$event.target.blur()"
          @blur="handleSearch"
        />
        <!-- 过滤条件 -->
        <Filter v-if="isFilteringMode" class="w-7 h-7" />
        <div
          v-if="isFilteringMode && filteredTag"
          @click="handleCancelFilter"
          class="w-34 text-nowrap flex items-center justify-between px-1 py-0.5 text-[var(--text-color-300)] border border-dashed border-[var(--border-color-400)] rounded-md hover:cursor-pointer hover:line-through hover:text-[var(--text-color-500)]"
        >
          <p class="text-nowrap truncate">{{ filteredTag.name }}</p>
          <Close class="inline w-4 h-4 ml-1" />
        </div>
      </div>

      <!-- 右侧图标组 -->
      <div class="flex justify-end items-center gap-1">
        <!-- RSS -->
        <div>
          <a href="/rss" title="RSS">
            <!-- icon -->
            <Rss class="w-8 h-8 text-[var(--text-color-400)]" />
          </a>
        </div>
        <!-- Ech0 Hub -->
        <div>
          <RouterLink to="/hub" title="Ech0 Hub">
            <!-- icon -->
            <HubIcon class="w-8 h-8 text-[var(--text-color-400)]" />
          </RouterLink>
        </div>
        <!-- Ech0 Widget -->
        <div class="block xl:hidden">
          <RouterLink to="/Widget" title="Ech0 Widget">
            <!-- icon -->
            <Widget class="w-8 h-8 text-[var(--text-color-400)]" />
          </RouterLink>
        </div>
        <!-- PanelPage -->
        <div>
          <RouterLink to="/panel" title="面板">
            <!-- icon -->
            <Panel class="w-8 h-8 text-[var(--text-color-400)]" />
          </RouterLink>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import BaseInput from '@/components/common/BaseInput.vue'
import Panel from '@/components/icons/panel.vue'
import Rss from '@/components/icons/rss.vue'
import HubIcon from '@/components/icons/hub.vue'
import { RouterLink } from 'vue-router'
import { useEchoStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { ref } from 'vue'
import Close from '@/components/icons/close.vue'
import Filter from '@/components/icons/filter.vue'
import Widget from '@/components/icons/widget.vue'
const echoStore = useEchoStore()
const { refreshForSearch, getEchosByPage } = echoStore
const { searchingMode, filteredTag, isFilteringMode } = storeToRefs(echoStore)

const searchContent = ref<string>('')

const handleSearch = () => {
  console.log('搜索内容:', searchContent.value)

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

const handleCancelFilter = () => {
  echoStore.isFilteringMode = false
  echoStore.filteredTag = null
  echoStore.refreshEchosForFilter()
}
</script>
