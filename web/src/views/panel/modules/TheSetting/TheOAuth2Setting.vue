<template>
  <div>
    <PanelCard class="mb-3">
      <!-- OAuth2 设置 -->
      <div class="w-full">
        <div class="flex flex-row items-center justify-between mb-3">
          <h1 class="text-gray-600 font-bold text-lg">OAuth2设置</h1>
          <div class="flex flex-row items-center justify-end gap-2 w-14">
            <button v-if="oauth2EditMode" @click="handleUpdateOAuth2Setting" title="编辑">
              <Saveupdate class="w-5 h-5 text-gray-400 hover:w-6 hover:h-6" />
            </button>
            <button @click="oauth2EditMode = !oauth2EditMode" title="编辑">
              <Edit v-if="!oauth2EditMode" class="w-5 h-5 text-gray-400 hover:w-6 hover:h-6" />
              <Close v-else class="w-5 h-5 text-gray-400 hover:w-6 hover:h-6" />
            </button>
          </div>
        </div>

        <!-- 开启OAuth2 -->
        <div class="flex flex-row items-center justify-start text-stone-500 h-10">
          <h2 class="font-semibold w-30 shrink-0">启用OAuth2:</h2>
          <BaseSwitch v-model="OAuth2Setting.enable" :disabled="!oauth2EditMode" />
        </div>

        <!-- OAuth2 Provider -->
        <div class="flex flex-row items-center justify-start text-stone-500 gap-2 h-10">
          <h2 class="font-semibold w-30 shrink-0">OAuth2 模板:</h2>
          <BaseSelect
            v-model="OAuth2Setting.provider"
            :options="OAuth2ProviderOptions"
            :disabled="!oauth2EditMode"
            class="w-34 h-8"
          />
        </div>

        <!-- Client ID -->
        <div class="flex flex-row items-center justify-start text-stone-500 gap-2 h-10">
          <h2 class="font-semibold w-30 shrink-0">Client ID:</h2>
          <span
            v-if="!oauth2EditMode"
            class="truncate max-w-40 inline-block align-middle"
            :title="OAuth2Setting.client_id"
            style="vertical-align: middle"
          >
            {{ OAuth2Setting.client_id.length === 0 ? '暂无' : OAuth2Setting.client_id }}
          </span>
          <BaseInput
            v-else
            v-model="OAuth2Setting.client_id"
            type="text"
            placeholder="请输入Client ID"
            class="w-full py-1!"
          />
        </div>

        <!-- Client Secret -->
        <div class="flex flex-row items-center justify-start text-stone-500 gap-2 h-10">
          <h2 class="font-semibold w-30 shrink-0">Client Secret:</h2>
          <span
            v-if="!oauth2EditMode"
            class="truncate max-w-40 inline-block align-middle"
            :title="OAuth2Setting.client_secret"
            style="vertical-align: middle"
          >
            {{ OAuth2Setting.client_secret.length === 0 ? '暂无' : OAuth2Setting.client_secret }}
          </span>
          <BaseInput
            v-else
            v-model="OAuth2Setting.client_secret"
            type="text"
            placeholder="请输入Client Secret"
            class="w-full py-1!"
          />
        </div>

        <!-- Callback URL -->
        <div class="flex flex-row items-center justify-start text-stone-500 gap-2 h-10">
          <h2 class="font-semibold w-30 shrink-0">Callback URL:</h2>
          <span
            v-if="!oauth2EditMode"
            class="truncate max-w-40 inline-block align-middle"
            :title="OAuth2Setting.redirect_uri"
            style="vertical-align: middle"
          >
            {{ redirect_uri.length === 0 ? '暂无' : redirect_uri }}
          </span>
          <BaseInput
            v-else
            v-model="redirect_uri"
            type="text"
            placeholder="请输入回调地址"
            class="w-full py-1!"
          />
        </div>

        <!-- Auth URL -->
        <div class="flex flex-row items-center justify-start text-stone-500 gap-2 h-10">
          <h2 class="font-semibold w-30 shrink-0">Auth URL:</h2>
          <span
            v-if="!oauth2EditMode"
            class="truncate max-w-40 inline-block align-middle"
            :title="OAuth2Setting.auth_url"
            style="vertical-align: middle"
          >
            {{ OAuth2Setting.auth_url.length === 0 ? '暂无' : OAuth2Setting.auth_url }}
          </span>
          <BaseInput
            v-else
            v-model="OAuth2Setting.auth_url"
            type="text"
            placeholder="请输入授权地址"
            class="w-full py-1!"
          />
        </div>

        <!-- Token URL -->
        <div class="flex flex-row items-center justify-start text-stone-500 gap-2 h-10">
          <h2 class="font-semibold w-30 shrink-0">Token URL:</h2>
          <span
            v-if="!oauth2EditMode"
            class="truncate max-w-40 inline-block align-middle"
            :title="OAuth2Setting.token_url"
            style="vertical-align: middle"
          >
            {{ OAuth2Setting.token_url.length === 0 ? '暂无' : OAuth2Setting.token_url }}
          </span>
          <BaseInput
            v-else
            v-model="OAuth2Setting.token_url"
            type="text"
            placeholder="请输入Token地址"
            class="w-full py-1!"
          />
        </div>

        <!-- User Info URL -->
        <div class="flex flex-row items-center justify-start text-stone-500 gap-2 h-10">
          <h2 class="font-semibold w-30 shrink-0">UserInfo URL:</h2>
          <span
            v-if="!oauth2EditMode"
            class="truncate max-w-40 inline-block align-middle"
            :title="OAuth2Setting.user_info_url"
            style="vertical-align: middle"
          >
            {{ OAuth2Setting.user_info_url.length === 0 ? '暂无' : OAuth2Setting.user_info_url }}
          </span>
          <BaseInput
            v-else
            v-model="OAuth2Setting.user_info_url"
            type="text"
            placeholder="请输入用户信息地址"
            class="w-full py-1!"
          />
        </div>

        <!-- Scopes -->
        <div class="flex flex-row items-center justify-start text-stone-500 gap-2 h-10">
          <h2 class="font-semibold w-30 shrink-0">Scopes:</h2>
          <span
            v-if="!oauth2EditMode"
            class="truncate max-w-40 inline-block align-middle"
            :title="OAuth2Setting.scopes.join(', ')"
            style="vertical-align: middle"
          >
            {{ OAuth2Setting.scopes.length === 0 ? '暂无' : OAuth2Setting.scopes.join(', ') }}
          </span>
          <BaseInput
            v-else
            v-model="scopeString"
            type="text"
            placeholder="请输入Scopes，多个用逗号分隔"
            class="w-full py-1!"
            @blur="OAuth2Setting.scopes = scopeString.split(',').map((s) => s.trim())"
          />
        </div>
      </div>
    </PanelCard>

    <div
      v-if="OAuth2Setting.enable && OAuth2Setting.provider"
      class="rounded-md border border-dashed border-gray-400 p-4 mb-3"
    >
      <!-- OAuth2 账号绑定 -->
      <div class="w-full">
        <div class="mb-3">
          <h1 class="text-gray-600 font-bold text-lg">账号绑定</h1>
          <p class="text-stone-400 text-sm">注意：需先配置OAuth2信息</p>
          <div
            v-if="
              oauthInfo && oauthInfo.oauth_id.length && oauthInfo.provider && oauthInfo.user_id != 0
            "
            class="mt-2 border border-dashed border-stone-400 rounded-md p-3 flex items-center justify-center"
          >
            <p class="text-stone-500 font-bold flex items-center">
              <component
                :is="
                  oauthInfo.provider === OAuth2Provider.GITHUB
                    ? Github
                    : oauthInfo.provider === OAuth2Provider.GOOGLE
                      ? Google
                      : oauthInfo.provider === OAuth2Provider.QQ
                        ? QQ
                        : Custom
                "
                class="w-5 h-5 mr-2"
              />
              <span>{{
                oauthInfo.provider === OAuth2Provider.GITHUB
                  ? 'GitHub'
                  : oauthInfo.provider === OAuth2Provider.GOOGLE
                    ? 'Google'
                    : oauthInfo.provider === OAuth2Provider.QQ
                      ? 'QQ'
                      : '自定义 OAuth2 账号'
              }}</span>
              账号已绑定
            </p>
          </div>
          <BaseButton
            v-else
            class="rounded-md mt-2 bg-stone-50! bg-op-80"
            @click="handleBindOAuth2()"
          >
            <div class="flex items-center justify-between">
              <component
                :is="
                  OAuth2Setting.provider === OAuth2Provider.GITHUB
                    ? Github
                    : OAuth2Setting.provider === OAuth2Provider.GOOGLE
                      ? Google
                      : OAuth2Setting.provider === OAuth2Provider.QQ
                        ? QQ
                        : Custom
                "
                class="w-5 h-5 mr-2"
              />
              <span class="flex-1 text-left">
                {{
                  OAuth2Setting.provider === OAuth2Provider.GITHUB
                    ? '绑定 GitHub 账号'
                    : OAuth2Setting.provider === OAuth2Provider.GOOGLE
                      ? '绑定 Google 账号'
                      : OAuth2Setting.provider === OAuth2Provider.QQ
                        ? '绑定 QQ 账号'
                        : '绑定自定义 OAuth2 账号'
                }}
              </span>
            </div>
          </BaseButton>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import PanelCard from '@/layout/PanelCard.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSwitch from '@/components/common/BaseSwitch.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import Edit from '@/components/icons/edit.vue'
import Close from '@/components/icons/close.vue'
import Saveupdate from '@/components/icons/saveupdate.vue'
import { ref, onMounted, watch } from 'vue'
import { useSettingStore } from '@/stores/setting'
import { theToast } from '@/utils/toast'
import { OAuth2Provider } from '@/enums/enums'
import { fetchUpdateOAuth2Settings, fetchBindOAuth2, fetchGetOAuthInfo } from '@/service/api'
import Github from '@/components/icons/github.vue'
import Google from '@/components/icons/google.vue'
import QQ from '@/components/icons/qq.vue'
import Custom from '@/components/icons/customoauth.vue'
import { storeToRefs } from 'pinia'

const settingStore = useSettingStore()
const { getOAuth2Setting } = settingStore
const { OAuth2Setting } = storeToRefs(settingStore)

const oauth2EditMode = ref(false)

const OAuth2ProviderOptions = [
  { label: 'GitHub', value: OAuth2Provider.GITHUB },
  { label: 'Google', value: OAuth2Provider.GOOGLE },
  { label: 'QQ', value: OAuth2Provider.QQ },
  { label: 'Custom', value: OAuth2Provider.CUSTOM },
]

const redirect_uri = ref(`${window.location.origin}/oauth/${OAuth2Setting.value.provider}/callback`)
const scopeString = ref('read:user')

const handleUpdateOAuth2Setting = async () => {
  // 修改Scopes
  OAuth2Setting.value.scopes = scopeString.value.split(',').map((s) => s.trim())
  // 修改回调地址为当前域名加上固定路径
  OAuth2Setting.value.redirect_uri =
    redirect_uri.value || `${window.location.origin}/oauth/${OAuth2Setting.value.provider}/callback`

  // 提交更新
  await fetchUpdateOAuth2Settings(OAuth2Setting.value)
    .then((res) => {
      if (res.code === 1) {
        theToast.success(res.msg)
      }
    })
    .finally(() => {
      oauth2EditMode.value = false
      // 重新获取OAuth2设置
      getOAuth2Setting()
      // 重新获取OAuth信息
      if (OAuth2Setting.value.provider) {
        fetchGetOAuthInfo(OAuth2Setting.value.provider).then((res) => {
          if (res.code === 1) {
            oauthInfo.value = res.data
          }
        })
      }
    })
}

const handleBindOAuth2 = async () => {
  const res = await fetchBindOAuth2(OAuth2Setting.value.provider, `${window.location.origin}/panel`)
  if (res.code !== 1) {
    theToast.error(res.msg)
  } else {
    // 成功，跳转到授权URL
    window.location.href = res.data
  }
}

const oauthInfo = ref<App.Api.Setting.OAuthInfo | null>(null)

// 监听 OAuth2Setting.provider 变化，更新必填设置模板
watch(
  () => OAuth2Setting.value.provider,
  (newProvider) => {
    const template = getProviderTemplate(newProvider)
    Object.assign(OAuth2Setting.value, template)
  },
)

function getProviderTemplate(provider: string) {
  if (provider === String(OAuth2Provider.GITHUB)) {
    scopeString.value = 'read:user'
    redirect_uri.value = `${window.location.origin}/oauth/github/callback`
    return {
      redirect_uri: `${window.location.origin}/oauth/github/callback`,
      auth_url: 'https://github.com/login/oauth/authorize',
      token_url: 'https://github.com/login/oauth/access_token',
      user_info_url: 'https://api.github.com/user',
      scopes: ['read:user'],
    }
  } else if (provider === String(OAuth2Provider.GOOGLE)) {
    scopeString.value = 'openid'
    redirect_uri.value = `${window.location.origin}/oauth/google/callback`
    return {
      redirect_uri: `${window.location.origin}/oauth/google/callback`,
      auth_url: 'https://accounts.google.com/o/oauth2/v2/auth',
      token_url: 'https://oauth2.googleapis.com/token',
      user_info_url: 'https://openidconnect.googleapis.com/v1/userinfo',
      scopes: ['openid'], // 只要OAuth ID
    }
  } else if (provider === String(OAuth2Provider.QQ)) {
    scopeString.value = 'get_user_info'
    redirect_uri.value = `${window.location.origin}/oauth/qq/callback`
    return {
      redirect_uri: `${window.location.origin}/oauth/qq/callback`,
      auth_url: 'https://graph.qq.com/oauth2.0/authorize',
      token_url: 'https://graph.qq.com/oauth2.0/token',
      user_info_url: 'https://graph.qq.com/user/get_user_info',
      scopes: ['get_user_info'],
    }
  } else if (provider === String(OAuth2Provider.CUSTOM)) {
    scopeString.value = ''
    redirect_uri.value = `${window.location.origin}/oauth/custom/callback`
    return {
      redirect_uri: `${window.location.origin}/oauth/custom/callback`,
      auth_url: '',
      token_url: '',
      user_info_url: '',
      scopes: [],
    }
  }
  return {}
}

onMounted(async () => {
  const setting = await getOAuth2Setting()
  if (setting?.provider) {
    const res = await fetchGetOAuthInfo(setting.provider)
    if (res.code === 1) {
      oauthInfo.value = res.data
    }
  }
})
</script>

<style scoped></style>
