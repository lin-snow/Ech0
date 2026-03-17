<template>
  <PanelCard>
    <!-- 系统设置 -->
    <div class="w-full">
      <div class="flex flex-row items-center justify-between mb-3">
        <h1 class="text-[var(--color-text-primary)] font-bold text-lg">
          {{ t('systemSetting.title') }}
        </h1>
        <div class="flex flex-row items-center justify-end">
          <BaseEditCapsule
            :editing="editMode"
            :apply-title="t('commonUi.apply')"
            :cancel-title="t('commonUi.cancel')"
            :edit-title="t('commonUi.edit')"
            @apply="handleUpdateSystemSetting"
            @toggle="editMode = !editMode"
          />
        </div>
      </div>
      <!-- 服务器&站点图标 -->
      <div class="flex justify-start items-center mb-4">
        <div class="w-28 sm:w-23">
          <img
            :src="systemLogoSrc"
            :alt="t('systemSetting.logoAlt')"
            class="w-12 h-12 rounded-full ml-2 mr-9 ring-1 ring-[var(--color-border-subtle)] shadow-[var(--shadow-sm)]"
          />
        </div>
        <div>
          <!-- 点击上传头像 -->
          <input
            id="file-input"
            class="hidden"
            type="file"
            accept="image/*"
            ref="fileInput"
            @change="handleUploadImage"
          />
          <BaseButton
            v-if="editMode"
            class="rounded-md text-center w-auto text-align-center h-8 md:ml-5"
            @click="handTriggerUpload"
          >
            {{ t('systemSetting.changeLogo') }}
          </BaseButton>
        </div>
      </div>

      <!-- 站点标题 -->
      <div
        class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 mb-1"
      >
        <h2 class="font-semibold min-w-26 w-max shrink-0 whitespace-nowrap">
          {{ t('systemSetting.siteTitle') }}:
        </h2>
        <span v-if="!editMode" class="flex-1 min-w-0 truncate" :title="SystemSetting.site_title">{{
          SystemSetting?.site_title.length === 0 ? t('commonUi.none') : SystemSetting.site_title
        }}</span>
        <BaseInput
          v-else
          v-model="SystemSetting.site_title"
          type="text"
          :placeholder="t('systemSetting.siteTitlePlaceholder')"
          class="w-full py-1!"
        />
      </div>
      <!-- 服务名称 -->
      <div
        class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 mb-1"
      >
        <h2 class="font-semibold min-w-26 w-max shrink-0 whitespace-nowrap">
          {{ t('systemSetting.serverName') }}:
        </h2>
        <span v-if="!editMode" class="flex-1 min-w-0 truncate" :title="SystemSetting.server_name">{{
          SystemSetting?.server_name.length === 0 ? t('commonUi.none') : SystemSetting.server_name
        }}</span>
        <BaseInput
          v-else
          v-model="SystemSetting.server_name"
          type="text"
          :placeholder="t('systemSetting.serverNamePlaceholder')"
          class="w-full py-1!"
        />
      </div>
      <!-- 服务地址 -->
      <div
        class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 mb-1"
      >
        <h2 class="font-semibold min-w-26 w-max shrink-0 whitespace-nowrap">
          {{ t('systemSetting.serverUrl') }}:
        </h2>
        <span v-if="!editMode" class="flex-1 min-w-0 truncate" :title="SystemSetting.server_url">{{
          SystemSetting?.server_name.length === 0 ? t('commonUi.none') : SystemSetting.server_url
        }}</span>
        <BaseInput
          v-else
          v-model="SystemSetting.server_url"
          type="text"
          :placeholder="t('systemSetting.serverUrlPlaceholder')"
          class="w-full py-1!"
        />
      </div>
      <!-- 自定义页脚内容 -->
      <div
        class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 mb-1"
      >
        <h2 class="font-semibold min-w-26 w-max shrink-0 whitespace-nowrap">
          {{ t('systemSetting.footerContent') }}:
        </h2>
        <span
          v-if="!editMode"
          class="flex-1 min-w-0 truncate inline-block align-middle"
          :title="SystemSetting.footer_content"
          style="vertical-align: middle"
        >
          {{
            SystemSetting.footer_content.length === 0
              ? t('commonUi.none')
              : SystemSetting.footer_content
          }}
        </span>
        <BaseInput
          v-else
          v-model="SystemSetting.footer_content"
          type="text"
          :placeholder="t('systemSetting.footerContentPlaceholder')"
          class="w-full py-1!"
        />
      </div>
      <!-- 自定义页脚链接 -->
      <div
        class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 mb-1"
      >
        <h2 class="font-semibold min-w-26 w-max shrink-0 whitespace-nowrap">
          {{ t('systemSetting.footerLink') }}:
        </h2>
        <span
          v-if="!editMode"
          class="flex-1 min-w-0 truncate inline-block align-middle"
          :title="SystemSetting.footer_link"
          style="vertical-align: middle"
        >
          {{
            SystemSetting.footer_link.length === 0 ? t('commonUi.none') : SystemSetting.footer_link
          }}
        </span>
        <BaseInput
          v-else
          v-model="SystemSetting.footer_link"
          type="text"
          :placeholder="t('systemSetting.footerLinkPlaceholder')"
          class="w-full py-1!"
        />
      </div>
      <!-- Meting API -->
      <div
        class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 mb-1"
      >
        <h2 class="font-semibold min-w-26 w-max shrink-0 whitespace-nowrap">
          {{ t('systemSetting.metingApi') }}:
        </h2>
        <span
          v-if="!editMode"
          class="flex-1 min-w-0 truncate inline-block align-middle"
          :title="SystemSetting.meting_api"
          style="vertical-align: middle"
        >
          {{
            SystemSetting.meting_api.length === 0 ? t('commonUi.none') : SystemSetting.meting_api
          }}
        </span>
        <BaseInput
          v-else
          v-model="SystemSetting.meting_api"
          type="text"
          :placeholder="t('systemSetting.metingApiPlaceholder')"
          class="w-full py-1!"
        />
      </div>
      <!-- 自定义 CSS -->
      <div class="flex flex-row justify-start text-[var(--color-text-secondary)] gap-2 mb-1">
        <h2 class="font-semibold min-w-26 w-max shrink-0 whitespace-nowrap">
          {{ t('systemSetting.customCss') }}:
        </h2>
        <span
          v-if="!editMode"
          class="flex-1 min-w-0 truncate inline-block align-middle"
          :title="SystemSetting.custom_css"
          style="vertical-align: middle"
          >{{ SystemSetting?.custom_css?.length === 0 ? t('commonUi.none') : '******' }}</span
        >
        <BaseTextArea
          v-else
          v-model="SystemSetting.custom_css"
          type="text"
          :placeholder="t('systemSetting.customCssPlaceholder')"
          class="w-full py-1!"
        />
      </div>
      <!-- 自定义 Script -->
      <div class="flex flex-row justify-start text-[var(--color-text-secondary)] gap-2 mb-1">
        <h2 class="font-semibold min-w-26 w-max shrink-0 whitespace-nowrap">
          {{ t('systemSetting.customJs') }}:
        </h2>
        <span
          v-if="!editMode"
          class="flex-1 min-w-0 truncate inline-block align-middle"
          :title="SystemSetting.custom_js"
          style="vertical-align: middle"
          >{{ SystemSetting?.custom_js?.length === 0 ? t('commonUi.none') : '******' }}</span
        >
        <BaseTextArea
          v-else
          v-model="SystemSetting.custom_js"
          type="text"
          :placeholder="t('systemSetting.customJsPlaceholder')"
          class="w-full py-1!"
        />
      </div>
      <!-- 默认语言 -->
      <div
        class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 mb-1"
      >
        <h2 class="font-semibold min-w-26 w-max shrink-0 whitespace-nowrap">
          {{ t('systemSetting.defaultLocale') }}:
        </h2>
        <span v-if="!editMode" class="flex-1 min-w-0 truncate">
          {{
            ({
              'en-US': t('commonUi.localeEnUS'),
              'de-DE': t('commonUi.localeDeDe'),
              'zh-CN': t('commonUi.localeZhCN'),
            }[SystemSetting.default_locale] || t('commonUi.localeZhCN'))
          }}
        </span>
        <BaseSelect
          v-else
          v-model="SystemSetting.default_locale"
          :options="localeOptions"
          class="w-fit h-8"
        />
      </div>
      <!-- 允许注册 -->
      <div class="flex flex-row items-center justify-start text-[var(--color-text-secondary)]">
        <h2 class="font-semibold min-w-26 w-max shrink-0 whitespace-nowrap">
          {{ t('systemSetting.allowRegister') }}:
        </h2>
        <BaseSwitch v-model="SystemSetting.allow_register" :disabled="!editMode" />
      </div>
    </div>
  </PanelCard>
</template>

<script setup lang="ts">
import PanelCard from '@/layout/PanelCard.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSwitch from '@/components/common/BaseSwitch.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseTextArea from '@/components/common/BaseTextArea.vue'
import BaseEditCapsule from '@/components/common/BaseEditCapsule.vue'
import { computed, ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { fetchUpdateSettings } from '@/service/api'
import { FILE_CATEGORY, FILE_STORAGE_TYPE } from '@/constants/file'
import { theToast } from '@/utils/toast'
import { useSettingStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { resolveAvatarUrl } from '@/service/request/shared'
import { useFileQueue } from '@/lib/file'

const settingStore = useSettingStore()
const { t } = useI18n()
const { getSystemSetting } = settingStore
const { SystemSetting } = storeToRefs(settingStore)

const editMode = ref<boolean>(false)
const systemLogoSrc = computed(() => resolveAvatarUrl(SystemSetting.value?.server_logo))
const localeOptions = computed(() => [
  { label: String(t('commonUi.localeZhCN')), value: 'zh-CN' },
  { label: String(t('commonUi.localeEnUS')), value: 'en-US' },
  { label: String(t('commonUi.localeDeDe')), value: 'de-DE' },
])
const { enqueueUpload, waitForTask, clearFinishedUploads } = useFileQueue()

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

const fileInput = ref<HTMLInputElement | null>(null)
const handTriggerUpload = () => {
  if (fileInput.value) {
    fileInput.value.click()
  }
}
const handleUploadImage = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  try {
    const taskId = enqueueUpload({
      file,
      storageType: FILE_STORAGE_TYPE.LOCAL,
      category: FILE_CATEGORY.IMAGE,
    })
    const task = await theToast.promise(waitForTask(taskId), {
      loading: String(t('systemSetting.logoUploading')),
      success: String(t('systemSetting.logoUploadSuccess')),
      error: String(t('systemSetting.uploadFailed')),
    })

    if (task.result?.url) {
      SystemSetting.value.server_logo = task.result.url
    } else {
      SystemSetting.value.server_logo = '/Ech0.svg'
    }
  } catch (err) {
    console.error('上传异常', err)
    // 注意：这里只有抛出异常时才会进入，正常 res.code ≠ 1 是不会进来的
  } finally {
    clearFinishedUploads()
    target.value = ''
  }
}

onMounted(() => {
  getSystemSetting()
})
</script>

<style scoped></style>
