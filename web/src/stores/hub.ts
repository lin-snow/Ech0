import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useFetch } from '@vueuse/core'
import { theToast } from '@/utils/toast'
import { useConnectStore } from './connect'

export const useHubStore = defineStore('hubStore', () => {
  /**
   * state
   */

  const connectStore = useConnectStore()

  // hub
  const hubList = ref<App.Api.Hub.HubList>([])
  const hubinfoList = ref<App.Api.Hub.HubInfoList>([])
  const hubInfoMap = ref<Map<string, App.Api.Hub.HubItemInfo>>(new Map())

  // echo
  const echoList = ref<App.Api.Hub.Echo[]>([]) // 存储Echo列表

  const isPreparing = ref<boolean>(true) // 是否正在准备数据
  const isLoading = ref<boolean>(false) // 是否正在加载数据
  const currentPage = ref<number>(1) // 延迟加载的页码，从0开始计数
  const pageSize = ref<number>(3) // 延迟加载的数量
  const hasMore = ref<boolean>(true) // 是否还有更多数据可加载

  /**
   * actions
   */

  // 1. 获取hubList
  const getHubList = async () => {
    isPreparing.value = true
    await connectStore.getConnect()

    hubList.value = connectStore.connects
  }

  // 2. 根据hubList 获取每个item的info
  const getHubInfoList = async () => {
    if (hubList.value.length === 0) {
      theToast.info('Hub列表为空，请到设置中添加Connect吧~')
      isPreparing.value = false
      return
    }

    // 处理 hubList 中的每个Hub（末尾的 / 去除）
    hubList.value = hubList.value.map((item) => {
      return typeof item === 'string'
        ? item.endsWith('/')
          ? item.slice(0, -1)
          : item
        : item.connect_url.endsWith('/')
          ? {
              ...item,
              connect_url: item.connect_url.slice(0, -1),
            }
          : item
    })

    // 创建带超时的请求函数
    const fetchWithTimeout = async (
      url: string,
      timeout: number = 5000,
    ): Promise<App.Api.Hub.HubItemInfo | null> => {
      return new Promise((resolve) => {
        let timeoutId: ReturnType<typeof setTimeout>
        let isResolved = false

        // 设置超时
        timeoutId = setTimeout(() => {
          if (!isResolved) {
            isResolved = true
            console.warn(`[Hub] 请求超时: ${url}`)
            resolve(null)
          }
        }, timeout)

        // 发起请求
        ;(async () => {
          try {
            const { error, data } = await useFetch<App.Api.Response<App.Api.Hub.HubItemInfo>>(
              `${url}/api/connect`,
            ).json()

            clearTimeout(timeoutId)
            if (!isResolved) {
              isResolved = true
              if (error.value || data.value?.code !== 1) {
                console.warn(`[Hub] 请求失败: ${url}`, error.value)
                resolve(null)
              } else {
                resolve(data.value?.data || null)
              }
            }
          } catch (err) {
            clearTimeout(timeoutId)
            if (!isResolved) {
              isResolved = true
              console.error(`[Hub] 请求异常: ${url}`, err)
              resolve(null)
            }
          }
        })()
      })
    }

    // 使用 Promise.allSettled 来并行获取每个Hub的info
    const promises = hubList.value.map(async (hub) => {
      const url = typeof hub === 'string' ? hub : hub.connect_url
      return await fetchWithTimeout(url, 5000) // 5秒超时
    })

    const results = await Promise.allSettled(promises)

    // 收集成功的结果，并从 hubList 中移除失败的实例
    const validHubs: typeof hubList.value = []
    const failedHubs: string[] = []

    results.forEach((result, index) => {
      const hub = hubList.value[index]
      if (!hub) return // 防止 undefined

      const hubUrl = typeof hub === 'string' ? hub : hub.connect_url

      if (result.status === 'fulfilled' && result.value) {
        // 成功获取信息
        hubinfoList.value.push(result.value)
        validHubs.push(hub)

        // 将Hub信息存入Map
        if (typeof hubUrl === 'string') {
          hubInfoMap.value.set(hubUrl, result.value)
        }
      } else {
        // 失败的实例，记录并排除
        if (typeof hubUrl === 'string') {
          failedHubs.push(hubUrl)
          console.warn(`[Hub] 实例不可用，已排除: ${hubUrl}`)
        }
      }
    })

    // 更新 hubList，只保留可用的实例
    hubList.value = validHubs

    // 提示用户
    if (failedHubs.length > 0) {
      theToast.warning(`${failedHubs.length} 个实例不可用，已自动排除`)
    }

    // 处理结果
    if (hubList.value.length === 0) {
      theToast.info('当前Hub暂无可连接的实例。')
      isPreparing.value = false
      return
    }

    isPreparing.value = false
    theToast.success(`成功连接 ${hubList.value.length} 个实例，开始加载 Echos`)
  }

  // 3. 根据 hubList 获取 list 中每个 item 的 echo
  const loadEchoListPage = async () => {
    if (!hasMore.value || isLoading.value || isPreparing.value) return

    isLoading.value = true
    try {
      const promises = hubList.value.map(async (item) => {
        const url = typeof item === 'string' ? item : item.connect_url
        const { error, data } = await useFetch<App.Api.Response<App.Api.Ech0.PaginationResult>>(
          url + '/api/echo/page',
        )
          .post({
            page: currentPage.value,
            pageSize: pageSize.value,
          })
          .json()

        if (error.value || data.value?.code !== 1) return []

        // 增加必要字段
        return (data.value?.data.items || []).map((echo: App.Api.Ech0.Echo) => ({
          ...echo,
          createdTs: new Date(echo.created_at).getTime(),
          server_name: hubInfoMap.value.get(url)?.server_name || 'Ech0',
          server_url: url,
          logo:
            hubInfoMap.value.get(url)?.logo !== ''
              ? hubInfoMap.value.get(url)?.logo
              : '/favicon.ico',
        }))
      })

      const results = await Promise.allSettled(promises)
      results.forEach((result, index) => {
        if (result.status === 'fulfilled' && Array.isArray(result.value)) {
          echoList.value.push(...result.value)
        } else {
          console.warn(`加载Hub ${hubList.value[index]} 的Echo数据失败:`)
        }
      })
      // 全局时间倒序排序
      echoList.value.sort((a, b) => b.createdTs - a.createdTs)

      // 检查是否还有更多数据
      hasMore.value = results.some((result) => {
        if (result.status === 'fulfilled' && Array.isArray(result.value)) {
          return result.value.length >= pageSize.value
        }
        return false
      })

      if (!hasMore.value && echoList.value.length > 0) {
        theToast.info('没有更多数据了🙃')
      }

      currentPage.value += 1
    } finally {
      isLoading.value = false
    }
  }

  return {
    echoList,
    hubList,
    hubInfoMap,
    hubinfoList,
    isLoading,
    isPreparing,
    currentPage,
    pageSize,
    hasMore,
    getHubList,
    getHubInfoList,
    loadEchoListPage,
  }
})
