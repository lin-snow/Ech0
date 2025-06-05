<template>
  <div class="w-full px-2">
    <div class="rounded-md shadow-sm ring-1 ring-gray-200 ring-inset bg-white p-4 mb-3">
      <!-- 设置 -->
      <div>
        <div class="flex flex-row items-center justify-between mb-3">
          <h1 class="text-gray-600 font-bold text-lg">系统设置</h1>
          <div class="flex flex-row items-center justify-end gap-2 w-14">
            <button v-if="editMode" @click="handleUpdateSystemSetting" title="编辑">
              <Saveupdate class="w-5 h-5 text-gray-400 hover:w-6 hover:h-6" />
            </button>
            <button @click="editMode = !editMode" title="编辑">
              <Edit v-if="!editMode" class="w-5 h-5 text-gray-400 hover:w-6 hover:h-6" />
              <Close v-else class="w-5 h-5 text-gray-400 hover:w-6 hover:h-6" />
            </button>
          </div>
        </div>
        <!-- 站点标题 -->
        <div class="flex flex-row items-center justify-start text-gray-500 gap-2 h-10">
          <h2 class="font-semibold w-30 flex-shrink-0">站点标题:</h2>
          <span v-if="!editMode">{{
            SystemSetting?.site_title.length === 0 ? '暂无' : SystemSetting.site_title
          }}</span>
          <BaseInput
            v-else
            v-model="SystemSetting.site_title"
            type="text"
            placeholder="请输入站点标题"
            class="w-full !py-1"
          />
        </div>
        <!-- 服务名称 -->
        <div class="flex flex-row items-center justify-start text-gray-500 gap-2 h-10">
          <h2 class="font-semibold w-30 flex-shrink-0">服务名称:</h2>
          <span v-if="!editMode">{{
            SystemSetting?.server_name.length === 0 ? '暂无' : SystemSetting.server_name
          }}</span>
          <BaseInput
            v-else
            v-model="SystemSetting.server_name"
            type="text"
            placeholder="请输入服务名称"
            class="w-full !py-1"
          />
        </div>
        <!-- 服务地址 -->
        <div class="flex flex-row items-center justify-start text-gray-500 gap-2 h-10">
          <h2 class="font-semibold w-30 flex-shrink-0">服务地址:</h2>
          <span v-if="!editMode">{{
            SystemSetting?.server_name.length === 0 ? '暂无' : SystemSetting.server_url
          }}</span>
          <BaseInput
            v-else
            v-model="SystemSetting.server_url"
            type="text"
            placeholder="请输入服务地址,带http(s)"
            class="w-full !py-1"
          />
        </div>
        <!-- ICP备案号 -->
        <div class="flex flex-row items-center justify-start text-gray-500 gap-2 h-10">
          <h2 class="font-semibold w-30 flex-shrink-0">ICP备案:</h2>
          <span
            v-if="!editMode"
            class="truncate max-w-40 inline-block align-middle"
            :title="SystemSetting.ICP_number"
            style="vertical-align: middle"
          >
            {{ SystemSetting.ICP_number.length === 0 ? '暂无' : SystemSetting.ICP_number }}
          </span>
          <BaseInput
            v-else
            v-model="SystemSetting.ICP_number"
            type="text"
            placeholder="请输入ICP备案号"
            class="w-full !py-1"
          />
        </div>
        <!-- Meting API -->
        <div class="flex flex-row items-center justify-start text-gray-500 gap-2 h-10">
          <h2 class="font-semibold w-30 flex-shrink-0">MetingAPI:</h2>
          <span
            v-if="!editMode"
            class="truncate max-w-40 inline-block align-middle"
            :title="SystemSetting.meting_api"
            style="vertical-align: middle"
          >
            {{ SystemSetting.meting_api.length === 0 ? '暂无' : SystemSetting.meting_api }}
          </span>
          <BaseInput
            v-else
            v-model="SystemSetting.meting_api"
            type="text"
            placeholder="请输入Meting API地址,带http(s)"
            class="w-full !py-1"
          />
        </div>
        <!-- 自定义 CSS -->
        <div class="flex flex-row items-center justify-start text-gray-500 gap-2 h-10">
          <h2 class="font-semibold w-30 flex-shrink-0">自定义 CSS:</h2>
          <span
            v-if="!editMode"
            class="truncate max-w-full inline-block align-middle"
            :title="SystemSetting.custom_css"
            style="vertical-align: middle"
            >{{ SystemSetting?.custom_css?.length === 0 ? '暂无' : '******' }}</span
          >
          <BaseInput
            v-else
            v-model="SystemSetting.custom_css"
            type="text"
            placeholder="请输入自定义 CSS"
            class="w-full !py-1"
          />
        </div>
        <!-- 自定义 Script -->
        <div class="flex flex-row items-center justify-start text-gray-500 gap-2 h-10">
          <h2 class="font-semibold w-30 flex-shrink-0">自定义 JS:</h2>
          <span
            v-if="!editMode"
            class="truncate max-w-full inline-block align-middle"
            :title="SystemSetting.custom_js"
            style="vertical-align: middle"
            >{{ SystemSetting?.custom_js?.length === 0 ? '暂无' : '******' }}</span
          >
          <BaseInput
            v-else
            v-model="SystemSetting.custom_js"
            type="text"
            placeholder="请输入自定义 Script"
            class="w-full !py-1"
          />
        </div>
        <!-- 允许注册 -->
        <div class="flex flex-row items-center justify-start text-gray-500 h-10">
          <h2 class="font-semibold w-30 flex-shrink-0">允许注册:</h2>
          <BaseSwitch v-model="SystemSetting.allow_register" :disabled="!editMode" />
        </div>
      </div>
    </div>

    <div class="rounded-md shadow-sm ring-1 ring-gray-200 ring-inset bg-white p-4">
      <!-- Ech0 Connect设置 -->
      <div>
        <div class="flex flex-row items-center justify-between mb-3">
          <h1 class="text-gray-600 font-bold text-lg">Ech0 Connect</h1>
          <div class="flex flex-row items-center justify-end gap-2 w-14">
            <button @click="connectsEdit = !connectsEdit" title="编辑">
              <Edit v-if="!connectsEdit" class="w-5 h-5 text-gray-400 hover:w-6 hover:h-6" />
              <Close v-else class="w-5 h-5 text-gray-400 hover:w-6 hover:h-6" />
            </button>
          </div>
        </div>
        <!-- 添加 Connect -->
        <div v-if="connectsEdit" class="flex flex-row items-center justify-between h-10">
          <BaseInput
            v-model="connectUrl"
            type="text"
            placeholder="请输入Connect地址（带https/http）"
            class="w-full h-7"
          />
          <BaseButton
            :icon="Publish"
            @click="handleAddConnect"
            class="w-7 h-7 ml-2 rounded-md"
            title="连接"
          />
        </div>
        <!-- Connect列表 -->
        <div
          v-if="connects.length === 0 && !connectsEdit"
          class="flex flex-col items-center justify-center mt-2"
        >
          <span class="text-gray-400">暂无连接...</span>
        </div>
        <div v-else class="mt-2">
          <div
            v-for="(connect, index) in connects"
            :key="index"
            class="flex flex-row items-center justify-between text-gray-500 gap-3 h-10"
          >
            <div class="flex items-center gap-2 flex-1 min-w-0">
              <h2 class="font-semibold w-30 flex-shrink-0">Connect {{ index + 1 }}:</h2>
              <span
                class="truncate max-w-full"
                :title="connect.connect_url"
                style="display: inline-block"
                >{{ connect.connect_url }}</span
              >
            </div>
            <BaseButton
              :icon="Disconnect"
              :disabled="!connectsEdit"
              @click="handleDisconnect(connect.id)"
              class="w-7 h-7 rounded-md flex-shrink-0"
              title="断开连接"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSwitch from '@/components/common/BaseSwitch.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import Edit from '@/components/icons/edit.vue'
import Disconnect from '@/components/icons/disconnect.vue'
import Close from '@/components/icons/close.vue'
import Publish from '@/components/icons/publish.vue'
import Saveupdate from '@/components/icons/saveupdate.vue'
import { ref, onMounted } from 'vue'
import { fetchUpdateSettings, fetchAddConnect, fetchDeleteConnect } from '@/service/api'
import { theToast } from '@/utils/toast'
import { useSettingStore } from '@/stores/settting'
import { useConnectStore } from '@/stores/connect'
import { storeToRefs } from 'pinia'

const settingStore = useSettingStore()
const { getSystemSetting } = settingStore
const { SystemSetting } = storeToRefs(settingStore)

const connectStore = useConnectStore()
const { getConnect } = connectStore
const { connects } = storeToRefs(connectStore)

const editMode = ref<boolean>(false)
const connectsEdit = ref<boolean>(false)
const connectUrl = ref<string>('')

const handleUpdateSystemSetting = async () => {
  await fetchUpdateSettings(settingStore.SystemSetting)
    .then((res) => {
      if (res.code === 1) {
        theToast.success(res.msg)
      }
    })
    .finally(() => {
      editMode.value = false
      // 重新获取设置
      getSystemSetting()
    })
}

const handleAddConnect = async () => {
  if (connectUrl.value.length === 0) {
    theToast.error('请输入Connect地址')
    return
  }
  await fetchAddConnect(connectUrl.value).then((res) => {
    if (res.code === 1) {
      theToast.success(res.msg)
      connectUrl.value = ''
      getConnect()
    }
  })
}

const handleDisconnect = async (connect_id: number) => {
  // 弹出确认框
  if (confirm('确定要断开连接吗？')) {
    await fetchDeleteConnect(connect_id).then((res) => {
      if (res.code === 1) {
        theToast.success(res.msg)
        getConnect()
      }
    })
  }
}

onMounted(async () => {
  await getSystemSetting()
  await getConnect()
})
</script>
