<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div>
    <PanelCard class="mb-3">
      <div class="w-full">
        <div class="flex flex-row items-center justify-between mb-3">
          <h1 class="text-[var(--color-text-primary)] font-bold text-lg">
            {{ t('oauth2Setting.title') }}
          </h1>
          <div class="flex flex-row items-center justify-end">
            <BaseEditCapsule
              :editing="oauth2EditMode"
              :apply-title="t('commonUi.apply')"
              :cancel-title="t('commonUi.cancel')"
              :edit-title="t('commonUi.edit')"
              @apply="handleUpdateOAuth2Setting"
              @toggle="oauth2EditMode = !oauth2EditMode"
            />
          </div>
        </div>

        <div
          v-if="OAuth2Setting.enable"
          class="mb-3 border border-dashed border-[var(--color-border-strong)] rounded-md p-3"
        >
          <h2 class="text-[var(--color-text-primary)] font-semibold mb-2">
            {{ t('oauth2Setting.healthCheck') }}
          </h2>
          <div class="flex flex-col sm:flex-row sm:flex-wrap gap-2 text-sm">
            <div class="flex items-center gap-2">
              <span class="text-[var(--color-text-secondary)]"
                >{{ t('oauth2Setting.oauthReady') }}:</span
              >
              <span
                class="px-2 py-0.5 rounded-md"
                :class="
                  oauthRuntimeStatus?.oauth_ready
                    ? 'bg-green-500/15 text-green-500'
                    : 'bg-yellow-500/15 text-yellow-500'
                "
              >
                {{
                  oauthRuntimeStatus?.oauth_ready
                    ? t('oauth2Setting.ready')
                    : t('oauth2Setting.notReady')
                }}
              </span>
            </div>
          </div>
          <p
            v-if="missingBoundaryItems.length > 0"
            class="mt-2 text-xs text-[var(--color-text-muted)] break-all"
          >
            {{ t('oauth2Setting.missingItems') }}: {{ missingBoundaryItems.join('、') }}
          </p>
          <div class="mt-2">
            <BaseButton
              class="rounded-md h-8 text-xs"
              @click="handleAutoFillBoundary"
              :disabled="missingBoundaryItems.length === 0"
            >
              {{ t('oauth2Setting.autofill') }}
            </BaseButton>
          </div>
        </div>

        <div
          class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] h-10"
        >
          <h2 class="font-semibold min-w-30 w-max shrink-0 whitespace-nowrap">
            {{ t('oauth2Setting.enableOAuth2') }}:
          </h2>
          <BaseSwitch v-model="OAuth2Setting.enable" :disabled="!oauth2EditMode" />
        </div>

        <div
          class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
        >
          <h2 class="font-semibold min-w-30 w-max shrink-0 whitespace-nowrap">
            {{ t('oauth2Setting.template') }}:
          </h2>
          <BaseSelect
            v-model="OAuth2Setting.provider"
            :options="OAuth2ProviderOptions"
            :disabled="!oauth2EditMode"
            class="w-34 h-8"
          />
        </div>

        <div
          class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
        >
          <h2 class="font-semibold min-w-30 w-max shrink-0 whitespace-nowrap">Client ID:</h2>
          <span
            v-if="!oauth2EditMode"
            class="flex-1 min-w-0 truncate inline-block align-middle"
            v-tooltip="OAuth2Setting.client_id"
            style="vertical-align: middle"
          >
            {{
              OAuth2Setting.client_id.length === 0 ? t('commonUi.none') : OAuth2Setting.client_id
            }}
          </span>
          <BaseInput
            v-else
            v-model="OAuth2Setting.client_id"
            type="text"
            :placeholder="t('oauth2Setting.clientIdPlaceholder')"
            class="w-full py-1!"
          />
        </div>

        <div
          class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
        >
          <h2 class="font-semibold min-w-30 w-max shrink-0 whitespace-nowrap">Client Secret:</h2>
          <span
            v-if="!oauth2EditMode"
            class="flex-1 min-w-0 truncate inline-block align-middle"
            v-tooltip="OAuth2Setting.client_secret"
            style="vertical-align: middle"
          >
            {{
              OAuth2Setting.client_secret.length === 0
                ? t('commonUi.none')
                : OAuth2Setting.client_secret
            }}
          </span>
          <BaseInput
            v-else
            v-model="OAuth2Setting.client_secret"
            type="text"
            :placeholder="t('oauth2Setting.clientSecretPlaceholder')"
            class="w-full py-1!"
          />
        </div>

        <div
          class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
        >
          <h2 class="font-semibold min-w-30 w-max shrink-0 whitespace-nowrap">Callback URL:</h2>
          <span
            v-if="!oauth2EditMode"
            class="flex-1 min-w-0 truncate inline-block align-middle"
            v-tooltip="OAuth2Setting.redirect_uri"
            style="vertical-align: middle"
          >
            {{ redirect_uri.length === 0 ? t('commonUi.none') : redirect_uri }}
          </span>
          <BaseInput
            v-else
            v-model="redirect_uri"
            type="text"
            :placeholder="t('oauth2Setting.callbackPlaceholder')"
            class="w-full py-1!"
          />
        </div>

        <div
          class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
        >
          <h2 class="font-semibold min-w-30 w-max shrink-0 whitespace-nowrap">Auth URL:</h2>
          <span
            v-if="!oauth2EditMode"
            class="flex-1 min-w-0 truncate inline-block align-middle"
            v-tooltip="OAuth2Setting.auth_url"
            style="vertical-align: middle"
          >
            {{ OAuth2Setting.auth_url.length === 0 ? t('commonUi.none') : OAuth2Setting.auth_url }}
          </span>
          <BaseInput
            v-else
            v-model="OAuth2Setting.auth_url"
            type="text"
            :placeholder="t('oauth2Setting.authUrlPlaceholder')"
            class="w-full py-1!"
          />
        </div>

        <div
          class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
        >
          <h2 class="font-semibold min-w-30 w-max shrink-0 whitespace-nowrap">Token URL:</h2>
          <span
            v-if="!oauth2EditMode"
            class="flex-1 min-w-0 truncate inline-block align-middle"
            v-tooltip="OAuth2Setting.token_url"
            style="vertical-align: middle"
          >
            {{
              OAuth2Setting.token_url.length === 0 ? t('commonUi.none') : OAuth2Setting.token_url
            }}
          </span>
          <BaseInput
            v-else
            v-model="OAuth2Setting.token_url"
            type="text"
            :placeholder="t('oauth2Setting.tokenUrlPlaceholder')"
            class="w-full py-1!"
          />
        </div>

        <div
          class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
        >
          <h2 class="font-semibold min-w-30 w-max shrink-0 whitespace-nowrap">UserInfo URL:</h2>
          <span
            v-if="!oauth2EditMode"
            class="flex-1 min-w-0 truncate inline-block align-middle"
            v-tooltip="OAuth2Setting.user_info_url"
            style="vertical-align: middle"
          >
            {{
              OAuth2Setting.user_info_url.length === 0
                ? t('commonUi.none')
                : OAuth2Setting.user_info_url
            }}
          </span>
          <BaseInput
            v-else
            v-model="OAuth2Setting.user_info_url"
            type="text"
            :placeholder="t('oauth2Setting.userInfoUrlPlaceholder')"
            class="w-full py-1!"
          />
        </div>

        <div
          class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
        >
          <h2 class="font-semibold min-w-30 w-max shrink-0 whitespace-nowrap">Scopes:</h2>
          <span
            v-if="!oauth2EditMode"
            class="flex-1 min-w-0 truncate inline-block align-middle"
            v-tooltip="OAuth2Setting.scopes.join(', ')"
            style="vertical-align: middle"
          >
            {{
              OAuth2Setting.scopes.length === 0
                ? t('commonUi.none')
                : OAuth2Setting.scopes.join(', ')
            }}
          </span>
          <BaseInput
            v-else
            v-model="scopeString"
            type="text"
            :placeholder="t('oauth2Setting.scopesPlaceholder')"
            class="w-full py-1!"
            @blur="OAuth2Setting.scopes = scopeString.split(',').map((s) => s.trim())"
          />
        </div>

        <div
          v-if="OAuth2Setting.enable"
          class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] h-10"
        >
          <h2 class="font-semibold min-w-30 w-max shrink-0 whitespace-nowrap">
            {{ t('oauth2Setting.enableOidc') }}:
          </h2>
          <BaseSwitch v-model="OAuth2Setting.is_oidc" :disabled="!oauth2EditMode" />
        </div>

        <div
          v-if="OAuth2Setting.is_oidc"
          class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
        >
          <h2 class="font-semibold min-w-30 w-max shrink-0 whitespace-nowrap">Issuer:</h2>
          <span
            v-if="!oauth2EditMode"
            class="flex-1 min-w-0 truncate inline-block align-middle"
            v-tooltip="OAuth2Setting.issuer"
            style="vertical-align: middle"
          >
            {{ OAuth2Setting.issuer.length === 0 ? t('commonUi.none') : OAuth2Setting.issuer }}
          </span>
          <BaseInput
            v-else
            v-model="OAuth2Setting.issuer"
            type="text"
            :placeholder="t('oauth2Setting.issuerPlaceholder')"
            class="w-full py-1!"
          />
        </div>

        <div
          v-if="OAuth2Setting.is_oidc"
          class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
        >
          <h2 class="font-semibold min-w-30 w-max shrink-0 whitespace-nowrap">JWKS URL:</h2>
          <span
            v-if="!oauth2EditMode"
            class="flex-1 min-w-0 truncate inline-block align-middle"
            v-tooltip="OAuth2Setting.jwks_url"
            style="vertical-align: middle"
          >
            {{ OAuth2Setting.jwks_url.length === 0 ? t('commonUi.none') : OAuth2Setting.jwks_url }}
          </span>
          <BaseInput
            v-else
            v-model="OAuth2Setting.jwks_url"
            type="text"
            :placeholder="t('oauth2Setting.jwksPlaceholder')"
            class="w-full py-1!"
          />
        </div>

        <div class="mt-3 border border-dashed border-[var(--color-border-strong)] rounded-md p-3">
          <h3 class="text-[var(--color-text-primary)] font-semibold mb-2">
            {{ t('oauth2Setting.securityBoundary') }}
          </h3>
          <div
            class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
          >
            <h2 class="font-semibold min-w-40 w-max shrink-0 whitespace-nowrap">
              Redirect Allowlist:
            </h2>
            <span v-if="!oauth2EditMode" class="flex-1 min-w-0 truncate inline-block align-middle">
              {{
                OAuth2Setting.auth_redirect_allowed_return_urls.length === 0
                  ? t('commonUi.none')
                  : OAuth2Setting.auth_redirect_allowed_return_urls.join(', ')
              }}
            </span>
            <BaseInput
              v-else
              v-model="redirectAllowlistString"
              type="text"
              :placeholder="t('oauth2Setting.multiUrlPlaceholder')"
              class="w-full py-1!"
            />
          </div>
          <div
            class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
          >
            <h2 class="font-semibold min-w-40 w-max shrink-0 whitespace-nowrap">CORS Origins:</h2>
            <span v-if="!oauth2EditMode" class="flex-1 min-w-0 truncate inline-block align-middle">
              {{
                OAuth2Setting.cors_allowed_origins.length === 0
                  ? t('commonUi.none')
                  : OAuth2Setting.cors_allowed_origins.join(', ')
              }}
            </span>
            <BaseInput
              v-else
              v-model="corsOriginsString"
              type="text"
              :placeholder="t('oauth2Setting.multiOriginPlaceholder')"
              class="w-full py-1!"
            />
          </div>
        </div>
      </div>
    </PanelCard>

    <PanelCard v-if="OAuth2Setting.enable && OAuth2Setting.provider" class="mb-3">
      <div class="w-full border border-dashed border-[var(--color-border-strong)] rounded-md p-3">
        <div>
          <h1 class="text-[var(--color-text-primary)] font-semibold text-lg">
            {{ t('oauth2Setting.accountBind') }}
          </h1>
          <p class="text-[var(--color-text-muted)] text-sm mt-1">
            {{ t('oauth2Setting.bindNotice') }}
          </p>
          <div
            v-if="oauthInfo && isBound"
            class="mt-2 border border-dashed border-[var(--color-border-strong)] rounded-md p-3 flex items-center justify-center"
          >
            <p class="text-[var(--color-text-secondary)] font-bold flex items-center">
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
                      : t('oauth2Setting.customAccount', {
                          type: oauthInfo?.auth_type === 'oidc' ? 'OIDC' : 'OAuth2',
                        })
              }}</span>
              {{ t('oauth2Setting.bound') }}
            </p>
          </div>
          <BaseButton v-else class="rounded-md mt-3" @click="handleBindOAuth2()">
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
                    ? t('oauth2Setting.bindGithub')
                    : OAuth2Setting.provider === OAuth2Provider.GOOGLE
                      ? t('oauth2Setting.bindGoogle')
                      : OAuth2Setting.provider === OAuth2Provider.QQ
                        ? t('oauth2Setting.bindQQ')
                        : t('oauth2Setting.bindCustom', {
                            type: OAuth2Setting.is_oidc ? 'OIDC' : 'OAuth2',
                          })
                }}
              </span>
            </div>
          </BaseButton>
        </div>
      </div>
    </PanelCard>
  </div>
</template>

<script setup lang="ts">
import PanelCard from '@/layout/PanelCard.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSwitch from '@/components/common/BaseSwitch.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseEditCapsule from '@/components/common/BaseEditCapsule.vue'
import { ref, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSettingStore } from '@/stores'
import { theToast } from '@/utils/toast'
import { OAuth2Provider } from '@/enums/enums'
import {
  fetchUpdateOAuth2Settings,
  fetchBindOAuth2,
  fetchGetOAuthInfo,
  fetchGetOAuth2Status,
} from '@/service/api'
import Github from '@/components/icons/github.vue'
import Google from '@/components/icons/google.vue'
import QQ from '@/components/icons/qq.vue'
import Custom from '@/components/icons/customoauth.vue'
import { storeToRefs } from 'pinia'

const settingStore = useSettingStore()
const { t } = useI18n()
const { getOAuth2Setting } = settingStore
const { OAuth2Setting } = storeToRefs(settingStore)

const oauth2EditMode = ref(false)

const OAuth2ProviderOptions = [
  { label: 'GitHub', value: OAuth2Provider.GITHUB },
  { label: 'Google', value: OAuth2Provider.GOOGLE },
  { label: 'QQ', value: OAuth2Provider.QQ },
  { label: String(t('oauth2Setting.customTemplate')), value: OAuth2Provider.CUSTOM },
]

const redirect_uri = ref<string>(OAuth2Setting.value.redirect_uri)
if (!redirect_uri.value) {
  redirect_uri.value = `${window.location.origin}/oauth/${OAuth2Setting.value.provider}/callback`
}
const scopeString = ref('read:user')
const redirectAllowlistString = ref('')
const corsOriginsString = ref('')

const parseList = (input: string) =>
  input
    .split(',')
    .map((s) => s.trim())
    .filter((s) => s.length > 0)

const handleUpdateOAuth2Setting = async () => {
  OAuth2Setting.value.scopes = scopeString.value.split(',').map((s) => s.trim())
  OAuth2Setting.value.redirect_uri =
    redirect_uri.value || `${window.location.origin}/oauth/${OAuth2Setting.value.provider}/callback`
  OAuth2Setting.value.auth_redirect_allowed_return_urls = parseList(redirectAllowlistString.value)
  OAuth2Setting.value.cors_allowed_origins = parseList(corsOriginsString.value)

  if (OAuth2Setting.value.auth_redirect_allowed_return_urls.some((u) => !/^https?:\/\//.test(u))) {
    theToast.error(String(t('oauth2Setting.redirectAllowlistInvalid')))
    return
  }
  if (OAuth2Setting.value.cors_allowed_origins.some((u) => !/^https?:\/\//.test(u))) {
    theToast.error(String(t('oauth2Setting.corsOriginsInvalid')))
    return
  }

  await fetchUpdateOAuth2Settings(OAuth2Setting.value)
    .then((res) => {
      if (res.code === 1) {
        theToast.success(res.msg)
      }
    })
    .finally(async () => {
      oauth2EditMode.value = false
      await getOAuth2Setting()
      await refreshHealthCheck()
      if (OAuth2Setting.value.provider) {
        const infoRes = await fetchGetOAuthInfo(OAuth2Setting.value.provider)
        if (infoRes.code === 1) {
          oauthInfo.value = infoRes.data
        }
      }
    })
}

const handleBindOAuth2 = async () => {
  const res = await fetchBindOAuth2(OAuth2Setting.value.provider, `${window.location.origin}/panel`)
  if (res.code !== 1) {
    theToast.error(res.msg)
  } else {
    window.location.href = res.data
  }
}

const oauthInfo = ref<App.Api.Setting.OAuthInfo | null>(null)
const isBound = ref<boolean>(false)
const oauthRuntimeStatus = ref<App.Api.Setting.OAuth2Status | null>(null)
const missingBoundaryItems = ref<string[]>([])

const refreshHealthCheck = async () => {
  const statusRes = await fetchGetOAuth2Status()
  if (statusRes.code === 1) {
    oauthRuntimeStatus.value = statusRes.data
  }
  const missing: string[] = []
  if ((OAuth2Setting.value.auth_redirect_allowed_return_urls || []).length === 0) {
    missing.push('Redirect Allowlist')
  }
  if ((OAuth2Setting.value.cors_allowed_origins || []).length === 0) {
    missing.push('CORS Origins')
  }
  missingBoundaryItems.value = missing
}

const handleAutoFillBoundary = () => {
  const currentOrigin = window.location.origin
  if (!OAuth2Setting.value.auth_redirect_allowed_return_urls?.length) {
    OAuth2Setting.value.auth_redirect_allowed_return_urls = [`${currentOrigin}/auth`]
  }
  if (!OAuth2Setting.value.cors_allowed_origins?.length) {
    OAuth2Setting.value.cors_allowed_origins = [currentOrigin]
  }

  redirectAllowlistString.value = OAuth2Setting.value.auth_redirect_allowed_return_urls.join(', ')
  corsOriginsString.value = OAuth2Setting.value.cors_allowed_origins.join(', ')
  oauth2EditMode.value = true
  void refreshHealthCheck()
  theToast.success(String(t('oauth2Setting.autofillDone')))
}

watch(
  () => OAuth2Setting.value.provider,
  (newProvider) => {
    const template = getProviderTemplate(newProvider)
    Object.assign(OAuth2Setting.value, template)
  },
)

watch(
  () => OAuth2Setting.value,
  (v) => {
    redirectAllowlistString.value = (v.auth_redirect_allowed_return_urls || []).join(', ')
    corsOriginsString.value = (v.cors_allowed_origins || []).join(', ')
  },
  { immediate: true, deep: true },
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
      scopes: ['openid'],
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
    scopeString.value = 'openid'
    redirect_uri.value = `${window.location.origin}/oauth/custom/callback`
    return {
      redirect_uri: `${window.location.origin}/oauth/custom/callback`,
      auth_url: '',
      token_url: '',
      user_info_url: '',
      scopes: ['openid'],
    }
  }
  return {}
}

onMounted(async () => {
  await getOAuth2Setting()
  await refreshHealthCheck()
  if (OAuth2Setting.value.provider) {
    const res = await fetchGetOAuthInfo(OAuth2Setting.value.provider)
    if (res.code === 1) {
      oauthInfo.value = res.data

      if (
        oauthInfo.value.auth_type === 'oidc' &&
        oauthInfo.value.issuer &&
        oauthInfo.value.oauth_id &&
        String(oauthInfo.value.user_id || '') !== '0'
      ) {
        isBound.value = true
      } else if (
        (oauthInfo.value.auth_type === '' || oauthInfo.value.auth_type === 'oauth2') &&
        oauthInfo.value.provider &&
        oauthInfo.value.oauth_id &&
        String(oauthInfo.value.user_id || '') !== '0'
      ) {
        isBound.value = true
      } else {
        isBound.value = false
      }
    }
  }
})
</script>

<style scoped></style>
