<template>
  <PanelCard>
    <div class="backup-wrap">
      <div class="backup-header">
        <h1 class="backup-title">{{ t('backupSetting.title') }}</h1>
        <p class="backup-desc">{{ t('backupSetting.description') }}</p>
      </div>

      <div class="backup-action">
        <BaseButton
          @click="handleBackupExport"
          class="backup-export-btn"
          :tooltip="t('backupSetting.exportSnapshot')"
        >
          {{ t('backupSetting.exportSnapshot') }}
        </BaseButton>
      </div>
    </div>
  </PanelCard>
</template>

<script setup lang="ts">
import PanelCard from '@/layout/PanelCard.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import { theToast } from '@/utils/toast'
import { useUserStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'

const userStore = useUserStore()
const authStore = useAuthStore()
const { isLogin } = storeToRefs(userStore)
const { t } = useI18n()

const handleBackupExport = async () => {
  if (!isLogin.value) {
    theToast.info(String(t('backupSetting.loginRequired')), { duration: 3000 })
    return
  }

  try {
    theToast.info(String(t('backupSetting.exporting')), {
      duration: 4000,
    })

    const token = authStore.accessToken
    const baseURL =
      import.meta.env.VITE_SERVICE_BASE_URL === '/'
        ? window.location.origin
        : import.meta.env.VITE_SERVICE_BASE_URL
    const downloadUrl = `${baseURL}/api/backup/export?token=${token}`

    // 创建隐藏的 a 标签触发下载
    const link = document.createElement('a')
    link.href = downloadUrl
    link.style.display = 'none'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)

    theToast.success(String(t('backupSetting.exportStarted')))
  } catch (error) {
    theToast.error(String(t('backupSetting.exportFailed')))
    console.error(String(t('backupSetting.exportFailed')), error)
  }
}
</script>

<style scoped>
.backup-wrap {
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
}

.backup-header {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.backup-title {
  color: var(--color-text-primary);
  font-weight: 700;
  font-size: 1.05rem;
}

.backup-desc {
  color: var(--color-text-secondary);
  font-size: 0.9rem;
}

.backup-action {
  display: flex;
  align-items: center;
}

.backup-export-btn {
  border-radius: var(--radius-md);
  color: var(--color-text-primary) !important;
}
</style>
