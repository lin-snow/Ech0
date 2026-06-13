<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div v-if="echo.extension" class="extension-renderer">
    <MusicPlayerCard v-if="echo.extension.type === ExtensionType.MUSIC" :echo="echo" />
    <VideoCard
      v-else-if="echo.extension.type === ExtensionType.VIDEO"
      :video-id="echo.extension.payload.videoId"
    />
    <GithubCard
      v-else-if="
        echo.extension.type === ExtensionType.GITHUBPROJ && echo.extension.payload?.repoUrl
      "
      :github-url="echo.extension.payload.repoUrl"
    />
    <WebsiteCard
      v-else-if="echo.extension.type === ExtensionType.WEBSITE"
      :website="echo.extension.payload"
    />
    <LocationCard
      v-else-if="echo.extension.type === ExtensionType.LOCATION"
      :location="echo.extension.payload"
    />
    <TweetCard
      v-else-if="echo.extension.type === ExtensionType.TWEET && echo.extension.payload?.url"
      :payload="echo.extension.payload"
    />
  </div>
</template>

<script setup lang="ts">
import { defineAsyncComponent } from 'vue'
import { ExtensionType } from '@/enums/enums'
const MusicPlayerCard = defineAsyncComponent(() => import('./cards/MusicPlayerCard.vue'))
const VideoCard = defineAsyncComponent(() => import('./cards/VideoCard.vue'))
const GithubCard = defineAsyncComponent(() => import('./cards/GithubCard.vue'))
const WebsiteCard = defineAsyncComponent(() => import('./cards/WebsiteCard.vue'))
const LocationCard = defineAsyncComponent(() => import('./cards/LocationCard.vue'))
const TweetCard = defineAsyncComponent(() => import('./cards/TweetCard.vue'))

defineProps<{
  echo: {
    extension?: App.Api.Ech0.EchoExtension | null
  }
}>()
</script>

<style scoped>
.extension-renderer {
  width: 100%;
}
</style>
