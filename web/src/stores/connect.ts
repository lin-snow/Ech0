import { ref } from 'vue'
import { defineStore } from 'pinia'
import { fetchGetConnectList, fetchGetAllConnectInfo } from '@/service/api'

export const useConnectStore = defineStore('connectStore', () => {
  /**
   * State
   */
  const connects = ref<App.Api.Connect.Connected[]>([])
  const connectsInfo = ref<App.Api.Connect.Connect[]>([])

  /**
   * Actions
   */
  async function getConnect() {
    await fetchGetConnectList()
      .then((res) => {
        if (res.code === 1) {
          connects.value = res.data
        }
      })
      .catch((err) => {
        console.error(err)
      })
  }

  async function getConnectInfo() {
    await fetchGetAllConnectInfo().then((res) => {
      if (res.code === 1) {
        connectsInfo.value = res.data
      }
    })
  }

  return {
    connects,
    connectsInfo,
    getConnect,
    getConnectInfo,
  }
})
