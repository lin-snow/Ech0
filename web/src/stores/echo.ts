import { defineStore } from 'pinia'
import { ref, computed, watch } from 'vue'
import { fetchGetEchosByPage } from '@/service/api'

export const useEchoStore = defineStore('echoStore', () => {
  /**
   * state
   */
  const echoList = ref<App.Api.Ech0.Echo[]>([])
  const isLoading = ref<boolean>(true)
  const total = ref<number>(0)
  const pageSize = ref<number>(5)
  const page = ref<number>(0)
  const current = ref<number>(1)
  const searchValue = ref<string>('')
  const hasMore = computed(() => {
    return total.value > echoList.value.length
  })
  const searchingMode = computed(() => {
    return searchValue.value.length > 0
  })

  // 监听 searchingMode 的变化
  watch(searchingMode, (newValue, oldValue) => {
    // 如果从搜索模式切换到非搜索模式，重置当前页码和数据列表
    if (newValue === false && oldValue === true) {
      refreshEchos()
    }
  })

  /**
   * actions
   */
  async function getEchosByPage() {
    if (current.value <= page.value) return

    isLoading.value = true

    await fetchGetEchosByPage({
      page: current.value,
      pageSize: pageSize.value,
      search: searchValue.value,
    })
      .then((res) => {
        if (res.code === 1) {
          total.value = res.data.total
          echoList.value = [...echoList.value, ...res.data.items]
          page.value += 1
        }
      })
      .finally(() => {
        isLoading.value = false
      })
  }

  const refreshEchos = () => {
    current.value = 1
    page.value = 0
    echoList.value = []
    getEchosByPage()
  }

  const refreshForSearch = () => {
    current.value = 1
    page.value = 0
    echoList.value = []
  }

  return {
    echoList,
    isLoading,
    total,
    pageSize,
    page,
    current,
    searchValue,
    searchingMode,
    hasMore,
    getEchosByPage,
    refreshEchos,
    refreshForSearch,
  }
})
