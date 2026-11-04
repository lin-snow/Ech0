<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <main class="min-h-screen w-full px-4 flex items-center justify-center">
    <div class="w-full max-w-[18rem]">
      <TheInitIntro />
      <TheInitForm
        :email="form.email"
        :username="form.username"
        :password="form.password"
        :site-title="form.siteTitle"
        :submitting="submitting"
        @update:email="form.email = $event"
        @update:username="form.username = $event"
        @update:password="form.password = $event"
        @update:site-title="form.siteTitle = $event"
        @submit="onSubmit"
      />
    </div>
  </main>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useInitStore, useSettingStore, useUserStore } from '@/stores'
import { theToast } from '@/utils/toast'
import TheInitForm from './TheInitForm.vue'
import TheInitIntro from './TheInitIntro.vue'
import { useI18n } from 'vue-i18n'

const router = useRouter()
const initStore = useInitStore()
const userStore = useUserStore()
const settingStore = useSettingStore()
const { isLogin } = storeToRefs(userStore)
const { t } = useI18n()
const submitting = ref(false)
const form = reactive({
  email: '',
  username: '',
  password: '',
  siteTitle: '',
})

onMounted(async () => {
  await initStore.getStatus().catch(() => undefined)
  if (initStore.initialized || initStore.ownerExists) {
    await router.replace({ name: 'auth' })
  }
})

/**
 * The Huma init endpoint is framework-neutral and cannot mint a session, so the
 * client signs in with the credentials it already holds. Success lands the new
 * owner on the editor instead of a second login form.
 */
const enterAfterInit = async () => {
  await userStore.login({ username: form.username.trim(), password: form.password })
  if (isLogin.value) {
    // The SPA read the settings before this owner existed, so the freshly
    // captured site name would still look unset downstream.
    await settingStore.init()
    await router.replace({ name: 'home', query: { tab: 'publish' } })
    return
  }
  theToast.info(String(t('init.autoLoginFailed')))
  await router.replace({ name: 'auth' })
}

const onSubmit = async () => {
  if (!form.email || !form.username || !form.password || submitting.value) {
    return
  }

  submitting.value = true
  try {
    const siteTitle = form.siteTitle.trim()
    const res = await initStore.initOwner({
      email: form.email.trim(),
      username: form.username.trim(),
      password: form.password,
      ...(siteTitle ? { site_title: siteTitle } : {}),
    })
    if (res.code === 1) {
      theToast.success(res.msg || String(t('init.initDone')))
      await enterAfterInit()
      return
    }
    if (initStore.initialized || initStore.ownerExists) {
      theToast.success(String(t('init.alreadyInitializedRedirect')))
      await router.replace({ name: 'auth' })
      return
    }
    theToast.error(res.msg || String(t('init.initFailed')))
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped></style>
