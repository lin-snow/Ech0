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

    // 使用 Promise.allSettled 来并行获取每个Hub的info
    const promises = hubList.value.map(async (hub) => {
      const { error, data } = await useFetch<App.Api.Response<App.Api.Hub.HubItemInfo>>(
        `${typeof hub === 'string' ? hub : hub.connect_url}/api/connect`,
      ).json()

      if (error.value || data.value?.code !== 1) {
        return null
      }

      return data.value?.data || null
    })

    await Promise.allSettled(promises).then((results) => {
      results.forEach((result, index) => {
        if (result.status === 'fulfilled' && result.value) {
          hubinfoList.value.push(result.value)
          const hubKey =
            typeof hubList.value?.[index] === 'string'
              ? hubList.value?.[index]
              : hubList.value?.[index]?.connect_url

          // 将Hub信息存入Map（确保 hubKey 为 string）
          if (typeof hubKey === 'string') {
            hubInfoMap.value.set(hubKey, result.value)
          }
        } else {
          theToast.warning(`获取Hub信息失败: ${hubList.value[index]}`)
        }
      })
    })

    // 处理结果
    if (hubinfoList.value.length === 0) {
      theToast.info('当前Hub暂无可连接的实例。')
      return
    }

    isPreparing.value = false
    theToast.success('开始加载 Echos')
  }

  // 3. 根据 hubList 获取 list 中每个 item 的 echo
  const loadEchoListPage = async () => {
    if (!hasMore.value || isLoading.value || isPreparing.value) return

    // 数据标准化函数：将旧版本的 images 字段转换为 media 字段
    const normalizeEchoData = (echo: any, serverUrl: string): App.Api.Hub.Echo => {
      // 如果没有 media 字段或 media 为空，但有 images 字段，则进行转换
      if ((!echo.media || echo.media.length === 0) && echo.images && Array.isArray(echo.images)) {
        // 将 images 转换为 media 格式
        echo.media = echo.images.map((image: any) => ({
          id: image.id,
          message_id: image.message_id,
          media_url: image.image_url || image.media_url, // 兼容两种字段名
          media_type: 'image' as const, // 旧版本只支持图片
          media_source: image.image_source || image.media_source, // 兼容两种字段名
          object_key: image.object_key,
          width: image.width,
          height: image.height,
        }))

        // 开发环境日志
        if (import.meta.env.DEV) {
          console.log('[兼容性转换] 检测到旧版本数据格式，已转换 images → media', {
            echoId: echo.id,
            serverUrl: serverUrl,
            imagesCount: echo.images.length,
          })
        }
      }

      // 如果既没有 media 也没有 images，设置为空数组
      if (!echo.media || !Array.isArray(echo.media)) {
        echo.media = []
        if (import.meta.env.DEV && echo.images === undefined) {
          console.warn('[兼容性转换] Echo 数据缺少 media 和 images 字段', {
            echoId: echo.id,
            serverUrl: serverUrl,
          })
        }
      }

      return echo
    }

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

        // 增加必要字段并进行数据标准化
        return (data.value?.data.items || []).map((echo: App.Api.Ech0.Echo) => {
          // 先进行数据标准化（images → media 转换）
          const normalizedEcho = normalizeEchoData(echo, url)

          // 然后添加 Hub 相关字段
          return {
            ...normalizedEcho,
            createdTs: new Date(normalizedEcho.created_at).getTime(),
            server_name: hubInfoMap.value.get(url)?.server_name || 'Ech0',
            server_url: url,
            // 设置echo.logo为站点Logo（来自/api/connect接口）
            logo:
              hubInfoMap.value.get(url)?.logo && hubInfoMap.value.get(url)?.logo !== ''
                ? hubInfoMap.value.get(url)?.logo
                : '/favicon.ico',
          }
        })
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
