<template>
  <div class="flex justify-center items-center h-screen">
    <div class="h-1/2 w-[min(86vw,12rem)] sm:w-[min(82vw,18rem)] md:w-[15rem]">
      <h1
        class="text-6xl italic font-bold text-center text-[var(--color-text-muted)] mb-4 font-serif"
      >
        Ech0
      </h1>
      <!-- 登录  -->
      <div v-if="AuthMode === 'login'">
        <!-- 模式切换 -->
        <div class="flex items-center justify-between gap-3 mb-3">
          <h2 class="text-lg font-bold text-[var(--color-text-muted)] leading-tight">
            {{ t('authPage.login') }}
          </h2>
          <button
            @click="AuthMode = 'register'"
            class="text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)] transition duration-200 whitespace-nowrap flex-shrink-0"
          >
            <div class="flex flex-row gap-1 items-center leading-tight">
              <span>{{ t('authPage.register') }}</span>
              <Arrow class="text-xl" />
            </div>
          </button>
        </div>
        <!-- 账号密码输入 -->
        <BaseInput
          v-model="username"
          type="text"
          :placeholder="t('authPage.usernamePlaceholder')"
          class="mb-4"
        />
        <BaseInput
          v-model="password"
          type="password"
          :placeholder="t('authPage.passwordPlaceholder')"
          class="mb-4"
        />
        <div class="flex justify-between items-center">
          <BaseButton
            @click="router.push({ name: 'home' })"
            :tooltip="t('authPage.backHome')"
            :icon="Home"
            class="rounded-md w-9 h-9 flex-shrink-0"
          />
          <div class="w-full flex items-center justify-end gap-1">
            <!-- Passkey 登录（Resident Key / 无用户名） -->
            <BaseButton
              :icon="Passkey"
              v-if="passkeySupported"
              @click="handlePasskeyLogin"
              :disabled="!!passkeyStatus && !passkeyStatus.passkey_ready"
              class="rounded-md w-9 h-9"
              :tooltip="t('authPage.passkeyLoginTitle')"
            />
            <!-- OAuth2 登录 -->
            <BaseButton
              v-if="oauth2Status && oauth2Status.enabled"
              :icon="
                oauth2Status.provider === OAuth2Provider.GITHUB
                  ? Github
                  : oauth2Status.provider === OAuth2Provider.GOOGLE
                    ? Google
                    : oauth2Status.provider === OAuth2Provider.QQ
                      ? QQ
                      : Customoauth
              "
              @click="gotoOAuth2URL"
              :disabled="!oauth2Status.oauth_ready"
              class="w-9 h-9 rounded-md"
              :tooltip="t('authPage.oauth2LoginTitle')"
            />
          </div>
          <!-- 账号密码登录 -->
          <BaseButton @click="handleLogin" class="min-w-fit px-3 h-9 rounded-md ml-1 flex-shrink-0">
            <span class="text-[var(--color-text-secondary)]">{{ t('authPage.login') }}</span>
          </BaseButton>
        </div>
      </div>
      <!-- 注册 -->
      <div v-else-if="AuthMode === 'register'">
        <div class="flex items-center justify-between gap-3 mb-3">
          <h2 class="text-lg font-bold text-[var(--color-text-muted)] leading-tight">
            {{ t('authPage.register') }}
          </h2>
          <button
            @click="AuthMode = 'login'"
            class="text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)] transition duration-200 whitespace-nowrap flex-shrink-0"
          >
            <div class="flex flex-row gap-1 items-center leading-tight">
              <span>{{ t('authPage.login') }}</span>
              <Arrow class="text-xl rotate-180" />
            </div>
          </button>
        </div>
        <BaseInput
          v-model="username"
          type="text"
          :placeholder="t('authPage.usernamePlaceholder')"
          class="mb-4"
        />
        <BaseInput
          v-model="password"
          type="password"
          :placeholder="t('authPage.passwordPlaceholder')"
          class="mb-4"
        />
        <div class="flex justify-between items-center px-0.5">
          <BaseButton
            @click="router.push({ name: 'home' })"
            :tooltip="t('authPage.backHome')"
            :icon="Home"
            class="rounded-md w-9 h-9"
          />
          <BaseButton @click="handleRegister" class="rounded-md min-w-fit px-3">
            <span class="text-[var(--color-text-secondary)]">{{ t('authPage.register') }}</span>
          </BaseButton>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import { useUserStore } from '@/stores'
import Arrow from '@/components/icons/arrow.vue'
import Passkey from '@/components/icons/passkey.vue'
import Home from '@/components/icons/home.vue'
import Github from '@/components/icons/github.vue'
import Google from '@/components/icons/google.vue'
import QQ from '@/components/icons/qq.vue'
import Customoauth from '@/components/icons/customoauth.vue'
import { fetchGetOAuth2Status, fetchGetPasskeyStatus } from '@/service/api'
import { OAuth2Provider } from '@/enums/enums'
import { fetchPasskeyLoginBegin, fetchPasskeyLoginFinish } from '@/service/api'
import { theToast } from '@/utils/toast'
import { base64urlToUint8Array, uint8ArrayToBase64url } from '@/utils/other'
import { useI18n } from 'vue-i18n'

const AuthMode = ref<'login' | 'register'>('login') // login / register
const username = ref<string>('')
const password = ref<string>('')
const userStore = useUserStore()
const { t } = useI18n()
const passkeySupported = !!(window.PublicKeyCredential && navigator.credentials)

const oauth2Status = ref<App.Api.Setting.OAuth2Status | null>(null)
const passkeyStatus = ref<App.Api.Setting.PasskeyStatus | null>(null)
const baseURL =
  import.meta.env.VITE_SERVICE_BASE_URL === '/'
    ? window.location.origin
    : import.meta.env.VITE_SERVICE_BASE_URL
const oauthURL = ref<string>(`${baseURL}/oauth/github/login`)

const gotoOAuth2URL = () => {
  if (!oauth2Status.value?.oauth_ready) {
    theToast.warning(String(t('authPage.oauth2NotReady')))
    return
  }
  if (!oauthURL.value) {
    theToast.error(String(t('authPage.oauth2UrlUnavailable')))
    return
  }
  window.location.href = oauthURL.value
}

const getOAuth2Status = async () => {
  const res = await fetchGetOAuth2Status()
  if (res.code === 1) {
    oauth2Status.value = res.data
    oauthURL.value = res.data.provider
      ? `${baseURL}/oauth/${res.data.provider}/login?redirect_uri=${window.location.origin}/auth`
      : ''
  }
}

const getPasskeyStatus = async () => {
  const res = await fetchGetPasskeyStatus()
  if (res.code === 1) {
    passkeyStatus.value = res.data
  }
}

const router = useRouter()

const handleLogin = async () => {
  // console.log('登录', username.value, password.value)
  await userStore.login({
    username: username.value,
    password: password.value,
  })
}

type RequestOptionsJSON = Omit<
  PublicKeyCredentialRequestOptions,
  'challenge' | 'allowCredentials'
> & {
  challenge: string
  allowCredentials?: Array<{
    type: PublicKeyCredentialType
    id: string
    transports?: AuthenticatorTransport[]
  }>
}

function normalizeRequestOptions(raw: unknown): PublicKeyCredentialRequestOptions {
  if (!raw || typeof raw !== 'object') throw new Error(String(t('authPage.invalidPublicKey')))
  const o = raw as RequestOptionsJSON
  const { challenge, allowCredentials, ...rest } = o

  const allow = Array.isArray(allowCredentials)
    ? allowCredentials.map((c) => ({
        ...c,
        id: base64urlToUint8Array(c.id) as BufferSource,
      }))
    : undefined

  return {
    ...rest,
    challenge: base64urlToUint8Array(challenge) as BufferSource,
    ...(allow ? { allowCredentials: allow } : {}),
  } as PublicKeyCredentialRequestOptions
}

function credentialToJSON(cred: PublicKeyCredential) {
  const obj: Record<string, unknown> = {
    id: cred.id,
    rawId: uint8ArrayToBase64url(cred.rawId),
    type: cred.type,
    clientExtensionResults: cred.getClientExtensionResults?.() ?? {},
  }

  const response: Record<string, unknown> = {}
  response.clientDataJSON = uint8ArrayToBase64url(cred.response.clientDataJSON)

  if ('authenticatorData' in cred.response) {
    const r = cred.response as AuthenticatorAssertionResponse
    response.authenticatorData = uint8ArrayToBase64url(r.authenticatorData)
    response.signature = uint8ArrayToBase64url(r.signature)
    if (r.userHandle && r.userHandle.byteLength > 0) {
      response.userHandle = uint8ArrayToBase64url(r.userHandle)
    }
  }

  obj.response = response
  return obj
}

const handlePasskeyLogin = async () => {
  if (passkeyStatus.value && !passkeyStatus.value.passkey_ready) {
    theToast.warning(String(t('authPage.passkeyNotReady')))
    return
  }
  if (!passkeySupported) return
  try {
    const begin = await fetchPasskeyLoginBegin()
    if (begin.code !== 1) return

    const options = normalizeRequestOptions(begin.data.publicKey)
    const got = await navigator.credentials.get({ publicKey: options })
    if (!got) throw new Error(String(t('authPage.getCredentialFailed')))
    const cred = got as PublicKeyCredential

    const finish = await fetchPasskeyLoginFinish(begin.data.nonce, credentialToJSON(cred))
    if (finish.code !== 1 || !finish.data?.access_token) return

    await userStore.loginWithTokenPair(finish.data)
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : String(t('authPage.passkeyLoginFailed'))
    theToast.error(msg)
  }
}

const handleRegister = async () => {
  // console.log('注册', username.value, password.value)
  if (
    await userStore.signup({
      username: username.value,
      password: password.value,
    })
  ) {
    // 注册成功，切换到登录模式
    AuthMode.value = 'login'
  }
}

onMounted(async () => {
  const url = new URL(window.location.href)
  const code = url.searchParams.get('code')
  if (code) {
    await userStore.loginWithCode(code)
    return
  }
  await Promise.all([getOAuth2Status(), getPasskeyStatus()])
})
</script>
