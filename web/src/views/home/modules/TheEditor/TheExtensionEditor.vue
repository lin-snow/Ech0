<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div>
    <div v-if="editorStore.currentExtensionType === ExtensionType.MUSIC">
      <h2 class="text-[var(--color-text-secondary)] font-bold mb-1">
        {{ t('editor.musicShare') }}
      </h2>
      <p class="text-[var(--color-text-muted)] text-sm">{{ t('editor.musicSupportHint') }}</p>
      <p class="text-[var(--color-text-muted)] text-sm mb-1">
        {{ t('editor.musicVipHint') }}
      </p>
      <BaseInput
        v-model="editorStore.extensionToAdd.extension"
        class="rounded-lg h-auto w-full"
        :placeholder="t('editor.musicUrlPlaceholder')"
      />
      <div
        v-if="
          editorStore.extensionToAdd.extension.length > 0 &&
          editorStore.extensionToAdd.extension_type === ExtensionType.MUSIC
        "
        class="mt-1 text-[var(--color-text-muted)] text-md"
      >
        {{ t('editor.parseResult') }}:
        <span
          v-if="parseMusicURL(editorStore.extensionToAdd.extension)"
          class="text-[var(--color-accent)]"
          >{{ t('editor.parseSuccess') }}</span
        >
        <span v-else class="text-[var(--color-danger)]">{{ t('editor.parseFailed') }}</span>
      </div>
    </div>
    <div v-if="editorStore.currentExtensionType === ExtensionType.VIDEO">
      <div class="text-[var(--color-text-secondary)] font-bold mb-1">
        {{ t('editor.videoShare') }}
      </div>
      <div class="text-[var(--color-text-muted)] mb-1">{{ t('editor.videoExtractHint') }}</div>
      <BaseInput
        v-model="editorStore.videoURL"
        class="rounded-lg h-auto w-full my-2"
        :placeholder="t('editor.videoUrlPlaceholder')"
      />
      <div class="text-[var(--color-text-secondary)] my-1">
        {{ t('editor.videoId') }}: {{ editorStore.extensionToAdd.extension }}
      </div>
    </div>
    <div v-if="editorStore.currentExtensionType === ExtensionType.GITHUBPROJ">
      <div class="text-[var(--color-text-secondary)] font-bold mb-1">
        {{ t('editor.githubShare') }}
      </div>
      <BaseInput
        v-model="editorStore.extensionToAdd.extension"
        class="rounded-lg h-auto w-full"
        :placeholder="t('editor.githubUrlPlaceholder')"
      />
    </div>
    <div v-if="editorStore.currentExtensionType === ExtensionType.TWEET">
      <div class="text-[var(--color-text-secondary)] font-bold mb-1">
        {{ t('editor.tweetShare') }}
      </div>
      <p class="text-[var(--color-text-muted)] text-sm mb-1">{{ t('editor.tweetHint') }}</p>
      <BaseInput
        v-model="editorStore.tweetToAdd.url"
        class="rounded-lg h-auto w-full"
        :placeholder="t('editor.tweetUrlPlaceholder')"
      />
      <div
        v-if="editorStore.tweetToAdd.url.trim()"
        class="mt-1 text-[var(--color-text-muted)] text-md"
      >
        {{ t('editor.parseResult') }}:
        <span v-if="parsedTweet" class="text-[var(--color-accent)]">
          @{{ parsedTweet.username }} · {{ parsedTweet.statusId }}
        </span>
        <span v-else class="text-[var(--color-danger)]">{{ t('editor.tweetUrlInvalid') }}</span>
      </div>
    </div>
    <div v-if="editorStore.currentExtensionType === ExtensionType.WEBSITE">
      <div class="text-[var(--color-text-secondary)] font-bold mb-1">
        {{ t('editor.websiteShare') }}
      </div>
      <BaseInput
        v-model="editorStore.websiteToAdd.title"
        class="rounded-lg h-auto w-full mb-2"
        :placeholder="t('editor.websiteTitlePlaceholder')"
      />
      <div class="flex items-center gap-2">
        <BaseInput
          v-model="editorStore.websiteToAdd.site"
          class="rounded-lg h-auto flex-1"
          :placeholder="t('editor.websiteUrlPlaceholder')"
        />
        <BaseButton
          class="rounded-md px-3 py-2 text-sm whitespace-nowrap"
          :disabled="isFetchingWebsiteTitle"
          @click="handleFetchWebsiteTitle"
        >
          {{ isFetchingWebsiteTitle ? t('editor.fetchingTitle') : t('editor.fetchTitle') }}
        </BaseButton>
      </div>
    </div>
    <div v-if="editorStore.currentExtensionType === ExtensionType.LOCATION">
      <div class="text-[var(--color-text-secondary)] font-bold mb-1">
        {{ t('editor.locationShare') }}
      </div>
      <p class="text-[var(--color-text-muted)] text-sm mb-2">
        {{ t('editor.locationHint') }}
      </p>

      <LocationPicker class="mb-2" :lat-lng="pickerLatLng" @change="handleMapChange" />

      <div class="flex flex-wrap items-center gap-2 mb-2">
        <BaseButton
          class="rounded-md px-3 py-2 text-sm whitespace-nowrap"
          :disabled="isGeolocating"
          @click="handleUseCurrentLocation"
        >
          {{ isGeolocating ? t('editor.locationGeolocating') : t('editor.locationUseCurrent') }}
        </BaseButton>
        <BaseButton
          class="rounded-md px-3 py-2 text-sm whitespace-nowrap"
          :disabled="!hasLocationSelected"
          @click="handleClearLocation"
        >
          {{ t('editor.locationClear') }}
        </BaseButton>
        <span v-if="isFetchingGeocoding" class="text-[var(--color-text-muted)] text-xs">
          {{ t('editor.locationLookingUp') }}
        </span>
      </div>

      <div class="grid grid-cols-2 gap-2 mb-2">
        <BaseInput
          v-model.number="latInput"
          type="number"
          :placeholder="t('editor.locationLatPlaceholder')"
          class="rounded-lg h-auto w-full"
        />
        <BaseInput
          v-model.number="lngInput"
          type="number"
          :placeholder="t('editor.locationLngPlaceholder')"
          class="rounded-lg h-auto w-full"
        />
      </div>

      <textarea
        v-model="editorStore.locationToAdd.placeholder"
        class="w-full rounded-lg border border-[var(--input-border-color)] bg-[var(--input-bg-color)]! text-[var(--input-text-color)] shadow-[var(--shadow-sm)] px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-[var(--input-focus-ring-color)] resize-y min-h-[5.5rem]"
        rows="4"
        :placeholder="t('editor.locationPlaceholderInput')"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import BaseInput from '@/components/common/BaseInput.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import LocationPicker from '@/components/advanced/extension/shared/LocationPicker.vue'
import { useReverseGeocoding } from '@/components/advanced/extension/shared/useReverseGeocoding'
import { ExtensionType } from '@/enums/enums'
import { parseMusicURL, extractAndCleanMusicURL } from '@/utils/other'
import { parseTweetUrl } from '@/utils/tweet'
import { useEditorStore } from '@/stores'
import { computed, ref, watch } from 'vue'
import { fetchGetWebsiteTitle } from '@/service/api'
import { theToast } from '@/utils/toast'
import { useI18n } from 'vue-i18n'

const editorStore = useEditorStore()
const isFetchingWebsiteTitle = ref(false)
const { t } = useI18n()

const parsedTweet = computed(() => parseTweetUrl(editorStore.tweetToAdd.url))

const isGeolocating = ref(false)

const latInput = computed<number | null>({
  get: () => editorStore.locationToAdd.latitude,
  set: (v) => {
    editorStore.locationToAdd.latitude = typeof v === 'number' && Number.isFinite(v) ? v : null
  },
})

const lngInput = computed<number | null>({
  get: () => editorStore.locationToAdd.longitude,
  set: (v) => {
    editorStore.locationToAdd.longitude = typeof v === 'number' && Number.isFinite(v) ? v : null
  },
})

const pickerLatLng = computed(() => {
  const { latitude, longitude } = editorStore.locationToAdd
  if (
    typeof latitude !== 'number' ||
    typeof longitude !== 'number' ||
    !Number.isFinite(latitude) ||
    !Number.isFinite(longitude)
  ) {
    return null
  }
  return { lat: latitude, lng: longitude }
})

const latRef = computed(() => editorStore.locationToAdd.latitude)
const lngRef = computed(() => editorStore.locationToAdd.longitude)
const { displayName, isFetching: isFetchingGeocoding } = useReverseGeocoding(latRef, lngRef)

watch(displayName, (name) => {
  if (!name) return
  if (editorStore.currentExtensionType !== ExtensionType.LOCATION) return
  editorStore.locationToAdd.placeholder = name
})

function handleMapChange(p: { lat: number; lng: number }) {
  editorStore.locationToAdd.latitude = p.lat
  editorStore.locationToAdd.longitude = p.lng
}

const hasLocationSelected = computed(
  () =>
    editorStore.locationToAdd.latitude !== null ||
    editorStore.locationToAdd.longitude !== null ||
    !!editorStore.locationToAdd.placeholder.trim(),
)

function handleClearLocation() {
  editorStore.locationToAdd.latitude = null
  editorStore.locationToAdd.longitude = null
  editorStore.locationToAdd.placeholder = ''
}

function handleUseCurrentLocation() {
  if (!('geolocation' in navigator)) {
    theToast.error(String(t('editor.locationGeolocationUnsupported')))
    return
  }
  isGeolocating.value = true
  navigator.geolocation.getCurrentPosition(
    (pos) => {
      editorStore.locationToAdd.latitude = pos.coords.latitude
      editorStore.locationToAdd.longitude = pos.coords.longitude
      isGeolocating.value = false
    },
    () => {
      isGeolocating.value = false
      theToast.error(String(t('editor.locationGeolocationDenied')))
    },
    { enableHighAccuracy: true, timeout: 8000 },
  )
}

const handleFetchWebsiteTitle = async () => {
  const websiteURL = (editorStore.websiteToAdd.site || '').trim()
  if (!websiteURL) {
    theToast.warning(String(t('editor.websiteInputRequired')))
    return
  }

  isFetchingWebsiteTitle.value = true
  try {
    const res = await fetchGetWebsiteTitle(websiteURL)
    if (res.code === 1) {
      editorStore.websiteToAdd.title = res.data
      theToast.success(String(t('editor.fetchTitleSuccess')))
    } else {
      theToast.error(res.msg || String(t('editor.fetchTitleFailed')))
    }
  } catch (error) {
    console.error('Failed to fetch website title', error)
    theToast.error(String(t('editor.fetchTitleFailed')))
  } finally {
    isFetchingWebsiteTitle.value = false
  }
}

watch(
  () => editorStore.extensionToAdd.extension,
  (newValue: string) => {
    if (editorStore.currentExtensionType !== ExtensionType.MUSIC || !newValue) {
      return
    }

    const value = newValue.trim()

    if (!/https?:\/\//i.test(value)) {
      return
    }

    const cleanUrl = extractAndCleanMusicURL(value)

    if (cleanUrl && cleanUrl !== value) {
      editorStore.extensionToAdd.extension = cleanUrl
    }
  },
)
</script>

<style scoped></style>
