<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <img :src="avatarSrc" :alt="alt" decoding="async" />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Avatar, Style, type StyleOptions } from '@dicebear/core'
import micah from '@dicebear/styles/micah.json'

// DiceBear 10 distributes styles as JSON definitions. `new Style()` validates
// the definition against its JSON Schema, so build it once at module scope and
// reuse it for every avatar instead of paying that cost on each render.
const micahStyle = new Style(micah)

type MicahOptions = StyleOptions<typeof micah>
type MicahOptionKey = keyof MicahOptions
type AvatarOptionItem = string | number | boolean
type AvatarOptionValue = AvatarOptionItem | AvatarOptionItem[] | null | undefined
type BaseAvatarOptions = Partial<Record<MicahOptionKey, AvatarOptionValue>>

const avatarCache = new Map<string, string>()

const props = withDefaults(
  defineProps<{
    seed?: string
    size?: number | string
    alt?: string
    src?: string
    options?: BaseAvatarOptions
  }>(),
  {
    seed: 'guest',
    size: 128,
    alt: 'avatar',
    src: '',
    options: () => ({}),
  },
)

const avatarSrc = computed(() => {
  const customSrc = props.src.trim()
  if (customSrc) return customSrc

  const seed = props.seed.trim() || 'guest'
  const sizeNum = Number(props.size)
  const size = Number.isFinite(sizeNum) && sizeNum > 0 ? Math.round(sizeNum) : 128
  const normalizedOptions: Partial<MicahOptions> = {}
  const sortedOptionEntries = (
    Object.entries(props.options) as [MicahOptionKey, AvatarOptionValue][]
  ).sort(([a], [b]) => String(a).localeCompare(String(b)))

  sortedOptionEntries.forEach(([key, value]) => {
    if (value === undefined || value === null) return
    ;(normalizedOptions as Record<MicahOptionKey, unknown>)[key] = value
  })

  const cacheKey = JSON.stringify({
    seed,
    size,
    options: normalizedOptions,
  })

  const cachedAvatar = avatarCache.get(cacheKey)
  if (cachedAvatar) return cachedAvatar

  const generatedAvatar = new Avatar(micahStyle, {
    seed,
    size,
    ...normalizedOptions,
  } as MicahOptions).toDataUri()

  avatarCache.set(cacheKey, generatedAvatar)
  return generatedAvatar
})
</script>
