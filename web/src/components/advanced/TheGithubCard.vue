<template>
  <div
    class="max-w-sm flex justify-center items-center bg-[var(--card-color)] rounded-lg shadow-sm ring-1 ring-inset ring-[var(--ring-color-100)] p-2 gap-2"
  >
    <a :href="props.GithubURL" target="_blank">
      <div class="flex justify-between items-center">
        <div class="shrink-0 px-6">
          <img
            v-if="CardData?.owner?.avatar_url"
            :src="CardData?.owner?.avatar_url"
            alt="头像"
            class="w-14 h-14 rounded-full shadow"
          />
          <Githubproj v-else class="w-14 h-14" />
        </div>

        <div v-if="CardData" class="py-1">
          <span class="text-lg font-bold text-[var(--text-color-600)]">{{
            CardData?.name || repo
          }}</span>
          <p
            class="text-sm text-[var(--text-color-400)] font-mono line-clamp-2"
            :title="CardData?.description"
          >
            {{ CardData?.description }}
          </p>
          <div class="flex justify-start items-center h-auto text-[var(--text-color-500)]">
            <!-- star -->
            <Star class="w-4 h-4 mr-1" /> <span> {{ CardData?.stargazers_count }} </span>
            <!-- fork -->
            <Fork class="w-4 h-4 mx-1" /> <span> {{ CardData?.forks_count }} </span>
          </div>
        </div>
      </div>
    </a>
  </div>
</template>

<script setup lang="ts">
import Githubproj from '../icons/githubproj.vue'
import Star from '../icons/star.vue'
import Fork from '../icons/fork.vue'
import { fetchGetGithubRepo } from '@/service/api'
import { onMounted, ref } from 'vue'

const githubRepoCache = new Map<string, App.Api.Ech0.GithubCardData | null>()
const githubRepoInFlight = new Map<string, Promise<App.Api.Ech0.GithubCardData | null>>()

const props = defineProps<{
  GithubURL: string
}>()

// 处理GithubURL(提取owner和repo)
const githubUrlSegments = props.GithubURL.split('/').filter(Boolean)
const [ownerRaw, repoRaw] = githubUrlSegments.slice(-2)
const owner = ownerRaw ?? ''
const repo = repoRaw ?? ''
const CardData = ref<App.Api.Ech0.GithubCardData>()
const repoKey = `${owner}/${repo}`

const loadGithubRepo = async () => {
  if (!owner || !repo) {
    return
  }

  if (githubRepoCache.has(repoKey)) {
    const cachedData = githubRepoCache.get(repoKey)
    if (cachedData) {
      CardData.value = cachedData
    }
    return
  }

  if (!githubRepoInFlight.has(repoKey)) {
    const task = fetchGetGithubRepo({ owner, repo })
      .then((res) => res ?? null)
      .catch(() => null)
      .finally(() => {
        githubRepoInFlight.delete(repoKey)
      })
    githubRepoInFlight.set(repoKey, task)
  }

  const repoData = await githubRepoInFlight.get(repoKey)
  githubRepoCache.set(repoKey, repoData ?? null)
  if (repoData) {
    CardData.value = repoData
  }
}

onMounted(() => {
  void loadGithubRepo()
})
</script>

<style scoped></style>
