<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <PanelCard>
    <div class="flex flex-row items-center justify-between mb-3">
      <h1 class="text-[var(--color-text-primary)] font-bold text-lg">
        {{ t('userManager.title') }}
      </h1>
      <div class="flex flex-row items-center justify-end gap-2 w-14"></div>
    </div>

    <div v-if="loading" class="flex justify-center py-4 text-[var(--color-text-muted)]">
      {{ t('userManager.loading') }}
    </div>

    <div v-else>
      <div v-if="allusers.length === 0" class="flex flex-col items-center justify-center mt-2">
        <span class="text-[var(--color-text-muted)]">{{ t('userManager.empty') }}</span>
      </div>

      <div
        v-else
        class="mt-2 x-scrollbar overflow-x-auto border border-[var(--color-border-subtle)] rounded-lg"
      >
        <table class="w-full min-w-[640px] table-fixed text-sm">
          <thead>
            <tr class="bg-[var(--color-bg-muted)]/70 text-left text-[var(--color-text-muted)]">
              <th class="w-[40px] px-2 py-2 whitespace-nowrap">#</th>
              <th class="w-[120px] px-2 py-2 whitespace-nowrap">
                {{ t('userManager.username') }}
              </th>
              <th class="px-2 py-2 whitespace-nowrap">
                {{ t('userManager.email') }}
              </th>
              <th class="w-[94px] px-2 py-2 text-center whitespace-nowrap">
                {{ t('userManager.isAdmin') }}
              </th>
              <th class="w-[86px] px-2 py-2 text-right whitespace-nowrap">
                {{ t('commonUi.actions') }}
              </th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(user, index) in allusers"
              :key="user.id"
              class="border-t border-[var(--color-border-subtle)] text-[var(--color-text-secondary)]"
            >
              <td class="px-2 py-2 text-[var(--color-text-primary)]">{{ index + 1 }}</td>
              <td class="px-2 py-2 text-[var(--color-text-primary)] font-semibold truncate">
                {{ user.username }}
              </td>
              <td class="px-2 py-2 text-[var(--color-text-muted)] truncate">
                {{ user.email || '—' }}
              </td>
              <td class="px-2 py-2 text-center">
                <BaseSwitch v-model="user.is_admin" @click="handleUpdateUserPermission(user.id)" />
              </td>
              <td class="px-2 py-2 text-right">
                <BaseButton
                  class="h-8 w-8 !p-1.5"
                  :icon="Deluser"
                  @click="handleDeleteUser(user.id)"
                  :tooltip="t('userManager.deleteUser')"
                />
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </PanelCard>
</template>

<script setup lang="ts">
import PanelCard from '@/layout/PanelCard.vue'
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseSwitch from '@/components/common/BaseSwitch.vue'
import Deluser from '@/components/icons/deluser.vue'
import { theToast } from '@/utils/toast'
import { useBaseDialog } from '@/composables/useBaseDialog'
const { openConfirm } = useBaseDialog()
const { t } = useI18n()

const loading = ref<boolean>(true)

import { fetchGetAllUsers, fetchUpdateUserPermission, fetchDeleteUser } from '@/service/api'

const allusers = ref<App.Api.User.User[]>([])

const handleDeleteUser = async (userId: string) => {
  openConfirm({
    title: String(t('userManager.deleteConfirmTitle')),
    description: String(t('userManager.deleteConfirmDesc')),
    onConfirm: () => {
      fetchDeleteUser(userId).then((res) => {
        if (res.code === 1) {
          getAllUsers()
        }
      })
    },
  })
}

const handleUpdateUserPermission = async (userId: string) => {
  fetchUpdateUserPermission(userId)
    .then((res) => {
      if (res.code === 1) {
        theToast.success(res.msg)
      }
    })
    .finally(() => {
      getAllUsers()
    })
}

const getAllUsers = async () => {
  loading.value = true
  try {
    const res = await fetchGetAllUsers()
    if (res.code === 1) {
      allusers.value = res.data
    }
    loading.value = false
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  getAllUsers()
})
</script>

<style scoped></style>
