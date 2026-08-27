<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <TheImageGallery
    v-if="imageFiles.length > 0"
    :images="imageFiles"
    :layout="layout"
    :baseUrl="baseUrl"
    :priority="priority"
  />
  <TheAudioPlayer v-else-if="audioFiles.length > 0" :files="audioFiles" :baseUrl="baseUrl" />
  <TheVideoPlayer v-else-if="videoFiles.length > 0" :files="videoFiles" :baseUrl="baseUrl" />
</template>

<script setup lang="ts">
import { computed, defineAsyncComponent } from 'vue'
import { getEchoFilesBy } from '@/utils/echo'
import TheAudioPlayer from './audio/TheAudioPlayer.vue'
import TheVideoPlayer from './video/TheVideoPlayer.vue'

const TheImageGallery = defineAsyncComponent(() => import('./image/TheImageGallery.vue'))

const props = withDefaults(
  defineProps<{
    echo: App.Api.Ech0.Echo | App.Api.Hub.Echo
    layout?: string
    baseUrl?: string
    priority?: boolean
  }>(),
  { priority: false },
)

const imageFiles = computed(() =>
  getEchoFilesBy(props.echo, { categories: ['image'], dedupeBy: 'id' }),
)
const audioFiles = computed(() =>
  getEchoFilesBy(props.echo, { categories: ['audio'], dedupeBy: 'id' }),
)
const videoFiles = computed(() =>
  getEchoFilesBy(props.echo, { categories: ['video'], dedupeBy: 'id' }),
)
</script>

<style scoped></style>
