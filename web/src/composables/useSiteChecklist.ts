// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { computed, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useEchoStore, useSettingStore, useUserStore } from '@/stores'
import { CHECKLIST_DISMISSED_KEY, shouldShowSiteChecklist } from './useSiteOnboarding'
import { localStg } from '@/utils/storage'

// Module scope: the timeline and the checklist itself must agree on visibility,
// otherwise the empty state and the checklist can render at the same time.
const dismissed = ref(localStg.getItem<boolean>(CHECKLIST_DISMISSED_KEY) === true)

export function useSiteChecklist() {
  const echoStore = useEchoStore()
  const settingStore = useSettingStore()
  const userStore = useUserStore()
  const { SystemSetting } = storeToRefs(settingStore)
  const { isLogin } = storeToRefs(userStore)

  const visible = computed(() =>
    shouldShowSiteChecklist({
      isOwner: isLogin.value,
      loading: echoStore.isLoading,
      filtering: echoStore.hasActiveQuery,
      dismissed: dismissed.value,
      echoTotal: echoStore.total,
      siteTitle: SystemSetting.value.site_title ?? '',
      serverURL: SystemSetting.value.server_url ?? '',
    }),
  )

  const dismiss = () => {
    dismissed.value = true
    localStg.setItem(CHECKLIST_DISMISSED_KEY, true)
  }

  return { visible, dismiss }
}
