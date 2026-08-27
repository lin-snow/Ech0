<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="about-page">
    <div class="about-frame">
      <button
        type="button"
        class="about-back"
        :aria-label="t('commonNav.backHome')"
        @click="$router.push({ name: 'home' })"
      >
        <Arrow class="about-back__icon" />
        <span>{{ t('commonNav.backHome') }}</span>
      </button>

      <div class="about-colophon">
        <span
          ref="stickerRef"
          class="about-sticker"
          :class="{ 'is-dragging': dragging }"
          :style="{ transform: `translate(${drag.x}px, ${drag.y}px) rotate(${drag.x * 0.04}deg)` }"
          aria-hidden="true"
          @pointerdown="onPointerDown"
          @pointermove="onPointerMove"
          @pointerup="onPointerUp"
          @pointercancel="onPointerUp"
        >
          <img class="about-sticker__img" src="/Ech0.svg" alt="" draggable="false" />
        </span>

        <div class="about-colophon__body">
          <h1 class="about-colophon__name">Ech0</h1>
          <p class="about-colophon__build">
            v{{ version
            }}<template v-if="hasCommit">
              ·
              <a
                :href="commitURL"
                target="_blank"
                rel="noopener noreferrer"
                class="about-link about-colophon__commit"
                :aria-label="t('about.viewSourceAtCommit', { commit })"
                :title="t('about.viewSourceAtCommit', { commit })"
                >{{ commit }}</a
              ></template
            >
          </p>

          <p class="about-colophon__links">
            <a
              :href="repoURL"
              target="_blank"
              rel="noopener noreferrer"
              class="about-link"
              :aria-label="t('about.viewSource')"
              :title="t('about.viewSource')"
            >
              {{ t('about.fieldSource') }}<span class="about-link__ext" aria-hidden="true">↗</span>
            </a>
            <span class="about-colophon__sep" aria-hidden="true">·</span>
            <a
              :href="`${repoURL}/blob/main/LICENSE`"
              target="_blank"
              rel="noopener noreferrer"
              class="about-link"
            >
              {{ license }}<span class="about-link__ext" aria-hidden="true">↗</span>
            </a>
          </p>

          <p class="about-colophon__copyright">{{ copyright }}</p>
          <p class="about-colophon__powered">{{ t('about.poweredBy') }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSettingStore } from '@/stores'
import Arrow from '@/components/icons/arrow.vue'

const { t } = useI18n()
const settingStore = useSettingStore()

const stickerRef = ref<HTMLElement | null>(null)
const dragging = ref(false)
const drag = reactive({ x: 0, y: 0 })
let originX = 0
let originY = 0

const onPointerDown = (e: PointerEvent) => {
  if (e.pointerType === 'mouse' && e.button !== 0) return
  dragging.value = true
  originX = e.clientX
  originY = e.clientY
  drag.x = 0
  drag.y = 0
  stickerRef.value?.setPointerCapture(e.pointerId)
}

const onPointerMove = (e: PointerEvent) => {
  if (!dragging.value) return
  drag.x = e.clientX - originX
  drag.y = e.clientY - originY
}

const onPointerUp = (e: PointerEvent) => {
  if (!dragging.value) return
  dragging.value = false
  drag.x = 0
  drag.y = 0
  try {
    stickerRef.value?.releasePointerCapture(e.pointerId)
  } catch {}
}

const FALLBACK_REPO = 'https://github.com/lin-snow/Ech0'
const FALLBACK_AUTHOR = 'L1nSn0w'
const FALLBACK_LICENSE = 'AGPL-3.0-or-later'

const version = computed(() => settingStore.hello?.version || '--')
const commit = computed(() => settingStore.hello?.commit || '')
const hasCommit = computed(() => commit.value !== '' && commit.value !== 'unknown')
const author = computed(() => settingStore.hello?.author || FALLBACK_AUTHOR)
const license = computed(() => settingStore.hello?.license || FALLBACK_LICENSE)
const repoURL = computed(() => settingStore.hello?.repo_url || FALLBACK_REPO)

const copyright = computed(
  () =>
    settingStore.hello?.copyright || `Copyright (C) ${new Date().getFullYear()} ${author.value}`,
)

const commitURL = computed(() => `${repoURL.value}/tree/${commit.value}`)
</script>

<style scoped>
.about-page {
  display: flex;
  justify-content: center;
  width: 100%;
  min-height: 100vh;
  padding: 4rem 1.5rem;
}

.about-frame {
  display: flex;
  flex-direction: column;
  width: 100%;
  max-width: 28rem;
}

.about-back {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  align-self: flex-start;
  padding: 0;
  border: 0;
  background: none;
  color: var(--color-text-muted);
  font-size: 0.8125rem;
  cursor: pointer;
  transition: color 0.15s ease;
}

.about-back:hover {
  color: var(--color-text-primary);
}

.about-back__icon {
  width: 0.95rem;
  height: 0.95rem;
  transform: rotate(180deg);
  transition: transform 0.15s ease;
}

.about-back:hover .about-back__icon {
  transform: rotate(180deg) translateX(2px);
}

.about-colophon {
  display: flex;
  align-items: flex-start;
  gap: 0.85rem;
  margin-top: 4rem;
}

.about-colophon__body {
  min-width: 0;
}

.about-sticker {
  flex: 0 0 auto;
  margin-top: 0.15rem;
  touch-action: none;
  cursor: grab;
  transition: transform 0.55s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.about-sticker.is-dragging {
  z-index: 2;
  cursor: grabbing;
  transition: none;
}

.about-sticker__img {
  display: block;
  width: 2.75rem;
  height: 2.75rem;
  border-radius: 0.55rem;
  box-shadow: var(--shadow-sm);
  pointer-events: none;
  user-select: none;
  transition:
    box-shadow 0.2s ease,
    transform 0.2s ease;
}

.about-sticker.is-dragging .about-sticker__img {
  box-shadow: var(--shadow-md);
  transform: scale(1.06);
}

@media (prefers-reduced-motion: reduce) {
  .about-sticker {
    transition-duration: 0.2s;
    transition-timing-function: ease;
  }
}

.about-colophon__name {
  margin: 0;
  font-family: var(--font-family-display);
  font-size: 2rem;
  font-weight: 700;
  letter-spacing: -0.01em;
  line-height: 1.1;
  color: var(--color-text-primary);
}

.about-colophon__build {
  margin: 0.4rem 0 0;
  font-family: var(--font-family-mono);
  font-size: 0.8125rem;
  letter-spacing: 0.01em;
  color: var(--color-text-muted);
}

.about-colophon__commit {
  text-decoration: underline;
  text-decoration-color: color-mix(in oklab, var(--color-text-muted) 50%, transparent);
  text-underline-offset: 0.2em;
}

.about-colophon__commit:hover {
  text-decoration-color: var(--color-accent);
}

.about-colophon__links {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
  margin: 1.75rem 0 0;
  font-size: 0.875rem;
}

.about-colophon__sep {
  color: var(--color-text-muted);
  opacity: 0.6;
}

.about-link {
  display: inline-flex;
  align-items: center;
  color: var(--color-text-secondary);
  text-decoration: none;
  transition: color 0.15s ease;
}

.about-link:hover {
  color: var(--color-accent);
}

.about-link__ext {
  margin-left: 0.15em;
  font-size: 0.85em;
  color: var(--color-text-muted);
}

.about-link:hover .about-link__ext {
  color: var(--color-accent);
}

.about-colophon__copyright {
  margin: 1.75rem 0 0;
  font-size: 0.75rem;
  line-height: 1.6;
  color: var(--color-text-muted);
}

.about-colophon__powered {
  margin: 0.3rem 0 0;
  font-size: 0.75rem;
  color: var(--color-text-muted);
}

@media (width <= 480px) {
  .about-page {
    padding: 3rem 1.25rem;
  }

  .about-colophon {
    margin-top: 3rem;
  }
}
</style>
