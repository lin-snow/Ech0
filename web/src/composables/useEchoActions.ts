// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { fetchDeleteEcho, fetchGetEchoById } from '@/service/api'
import { useEchoStore, useEditorStore } from '@/stores'
import { theToast } from '@/utils/toast'
import { useBaseDialog } from '@/composables/useBaseDialog'

type Echo = App.Api.Ech0.Echo

export function useEchoActions() {
  const echoStore = useEchoStore()
  const editorStore = useEditorStore()
  const router = useRouter()
  const { openConfirm } = useBaseDialog()
  const { t } = useI18n()

  const startUpdate = async (echo: Echo) => {
    const res = await fetchGetEchoById(String(echo.id))
    echoStore.echoToUpdate = res.code === 1 && res.data ? res.data : echo
    editorStore.isUpdateMode = true
    await router.push({
      name: 'home',
      query: { ...router.currentRoute.value.query, tab: 'publish' },
    })
  }

  const updateEcho = async (echo: Echo) => {
    if (editorStore.isUpdateMode) {
      openConfirm({
        title: String(t('echoCard.editAnotherEchoTitle')),
        description: String(t('echoCard.editAnotherEchoDesc')),
        onConfirm: () => {
          void startUpdate(echo)
        },
      })
      return
    }

    await startUpdate(echo)
  }

  const deleteEcho = (echoId: string) => {
    openConfirm({
      title: String(t('echoCard.deleteConfirmTitle')),
      description: String(t('echoCard.deleteConfirmDesc')),
      onConfirm: () => {
        fetchDeleteEcho(echoId).then(() => {
          theToast.success(String(t('echoCard.deleteSuccess')))
          echoStore.removeEcho(echoId)
        })
      },
    })
  }

  return { updateEcho, deleteEcho }
}
