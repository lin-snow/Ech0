<template>
  <div class="flex justify-center items-center h-screen">
    <div class="h-1/2 max-w-sm sm:max-w-md md:max-w-lg">
      <h1
        class="text-6xl italic font-bold text-center text-[var(--text-color-next-300)] mb-4 font-serif"
      >
        Ech0
      </h1>
      <!-- 登录  -->
      <div v-if="AuthMode === 'login'">
        <!-- 模式切换 -->
        <div class="flex justify-between items-center">
          <h2 class="text-lg font-bold text-[var(--text-color-next-400)] mb-3">登录</h2>
          <div class="mb-3">
            <button
              @click="AuthMode = 'register'"
              class="text-[var(--text-color-next-500)] hover:text-[var(--text-color-next-700)] transition duration-200"
            >
              <div class="flex flex-row gap-0 items-center">
                注册
                <Arrow class="text-2xl" />
              </div>
            </button>
          </div>
        </div>
        <!-- 账号密码输入 -->
        <BaseInput v-model="username" type="text" placeholder="请输入用户名" class="mb-4" />
        <BaseInput v-model="password" type="password" placeholder="请输入密码" class="mb-4" />
        <div class="flex justify-between items-center">
          <BaseButton
            @click="router.push({ name: 'home' })"
            title="返回首页"
            :icon="Home"
            class="rounded-md w-9 h-9 flex-shrink-0"
          />
          <div class="w-full flex items-center justify-end gap-1">
            <!-- Passkey 登录（Resident Key / 无用户名） -->
            <BaseButton
              :icon="Passkey"
              v-if="passkeySupported"
              @click="handlePasskeyLogin"
              class="rounded-md w-9 h-9"
              title="使用 Passkey 登录"
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
              class="w-9 h-9 rounded-md"
              title="使用 OAuth2 登录"
            />
          </div>
          <!-- 账号密码登录 -->
          <BaseButton @click="handleLogin" class="w-12 h-9 rounded-md ml-1 flex-shrink-0">
            <span class="text-[var(--text-color-next-500)]">登录</span>
          </BaseButton>
        </div>
      </div>
      <!-- 注册 -->
      <div v-else-if="AuthMode === 'register'">
        <div class="flex justify-between items-center">
          <h2 class="text-lg font-bold text-[var(--text-color-next-400)] mb-3">注册</h2>
          <div class="mb-3">
            <button
              @click="AuthMode = 'login'"
              class="text-[var(--text-color-next-500)] hover:text-[var(--text-color-next-700)] transition duration-200"
            >
              <div class="flex flex-row gap-0 items-center">
                登录
                <Arrow class="text-2xl rotate-180" />
              </div>
            </button>
          </div>
        </div>
        <BaseInput v-model="username" type="text" placeholder="请输入用户名" class="mb-4" />
        <BaseInput v-model="password" type="password" placeholder="请输入密码" class="mb-4" />
        <div class="flex justify-between items-center px-0.5">
          <BaseButton
            @click="router.push({ name: 'home' })"
            title="返回首页"
            :icon="Home"
            class="rounded-md w-9 h-9"
          />
          <BaseButton @click="handleRegister" class="rounded-md">
            <span class="text-[var(--text-color-next-500)]">注册</span>
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
import { fetchGetOAuth2Status } from '@/service/api'
import { OAuth2Provider } from '@/enums/enums'
import { fetchPasskeyLoginBegin, fetchPasskeyLoginFinish } from '@/service/api'
import { theToast } from '@/utils/toast'
import { base64urlToUint8Array, uint8ArrayToBase64url } from '@/utils/other'

const AuthMode = ref<'login' | 'register'>('login') // login / register
const username = ref<string>('')
const password = ref<string>('')
const userStore = useUserStore()
const passkeySupported = !!(window.PublicKeyCredential && navigator.credentials)

const oauth2Status = ref<App.Api.Setting.OAuth2Status | null>(null)
const baseURL =
  import.meta.env.VITE_SERVICE_BASE_URL === '/'
    ? window.location.origin
    : import.meta.env.VITE_SERVICE_BASE_URL
const oauthURL = ref<string>(`${baseURL}/oauth/github/login`)

const gotoOAuth2URL = () => {
  window.location.href = oauthURL.value
}

const getOAuth2Status = async () => {
  const res = await fetchGetOAuth2Status()
  if (res.code === 1) {
    oauth2Status.value = res.data
    oauthURL.value = res.data.provider
      ? `${baseURL}/oauth/${res.data.provider}/login?redirect_uri=${window.location.origin}/auth`
      : `${baseURL}/auth`
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
  if (!raw || typeof raw !== 'object') throw new Error('服务端返回的 publicKey 不合法')
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
  if (!passkeySupported) return
  try {
    const begin = await fetchPasskeyLoginBegin()
    if (begin.code !== 1) return

    const options = normalizeRequestOptions(begin.data.publicKey)
    const got = await navigator.credentials.get({ publicKey: options })
    if (!got) throw new Error('获取凭证失败')
    const cred = got as PublicKeyCredential

    const finish = await fetchPasskeyLoginFinish(begin.data.nonce, credentialToJSON(cred))
    if (finish.code !== 1) return

    await userStore.loginWithToken(finish.data)
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : 'Passkey 登录失败'
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
  const token = url.searchParams.get('token')
  if (token) {
    console.log('检测到 token，尝试使用 token 登录', token)
    // 有 token，直接登录
    await userStore.loginWithToken(token)
    return
  }
  getOAuth2Status()
})
</script>
