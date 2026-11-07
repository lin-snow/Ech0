<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div ref="rootEl" class="relative flex items-center justify-center mr-2">
    <button
      type="button"
      :aria-label="t('echoCard.moreActions')"
      :aria-expanded="open"
      aria-haspopup="menu"
      class="w-7 h-7 flex items-center justify-center bg-[var(--color-bg-surface)] ring-1 ring-[var(--color-border-subtle)] ring-inset rounded-full shadow-sm transition-shadow duration-150 hover:shadow-md focus:outline-none"
      @mousedown.prevent
      @click.stop="toggle"
    >
      <More class="w-5 h-5" />
    </button>

    <transition
      enter-active-class="transition-opacity duration-100 ease-out"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition-opacity duration-75 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="open"
        class="absolute right-0 top-full z-50 mt-2 w-36 rounded-[var(--radius-md)] bg-[var(--color-bg-muted)] ring-1 ring-[var(--color-border-subtle)] shadow-[var(--shadow-md)] p-1"
        role="menu"
      >
        <div
          v-if="echo.private"
          class="flex items-center gap-1.5 px-2 pt-1 pb-1.5 text-[var(--color-text-muted)]"
        >
          <Lock class="w-3 h-3" />
          <span class="text-[10px] font-semibold tracking-[0.08em] uppercase">
            {{ t('echoCard.privateStatus') }}
          </span>
        </div>

        <button type="button" class="menu-row" role="menuitem" @click="onUpdate">
          <EditEcho class="w-3.5 h-3.5 shrink-0" />
          <span>{{ t('echoCard.update') }}</span>
        </button>
        <button type="button" class="menu-row menu-row--danger" role="menuitem" @click="onDelete">
          <Roll class="w-3.5 h-3.5 shrink-0" />
          <span>{{ t('echoCard.delete') }}</span>
        </button>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import EditEcho from '@/components/icons/editecho.vue'
import Lock from '@/components/icons/lock.vue'
import More from '@/components/icons/more.vue'
import Roll from '@/components/icons/roll.vue'
import { useEchoActions } from '@/composables/useEchoActions'

type Echo = App.Api.Ech0.Echo

const props = defineProps<{
  echo: Echo
}>()

const { updateEcho, deleteEcho } = useEchoActions()
const { t } = useI18n()
const route = useRoute()

const rootEl = ref<HTMLElement | null>(null)
const open = ref(false)

const close = () => {
  open.value = false
}

const onOutside = (event: PointerEvent) => {
  const target = event.target as Node | null
  if (target && rootEl.value?.contains(target)) return
  close()
}

const onKeydown = (event: KeyboardEvent) => {
  if (event.key === 'Escape') close()
}

const bind = () => {
  document.addEventListener('pointerdown', onOutside)
  document.addEventListener('keydown', onKeydown)
}

const unbind = () => {
  document.removeEventListener('pointerdown', onOutside)
  document.removeEventListener('keydown', onKeydown)
}

const toggle = () => {
  open.value = !open.value
}

watch(open, (value) => {
  unbind()
  if (value) bind()
})

watch(
  () => route.fullPath,
  () => close(),
)

onBeforeUnmount(unbind)

const onUpdate = () => {
  close()
  void updateEcho(props.echo)
}

const onDelete = () => {
  const echoId = props.echo.id
  close()
  deleteEcho(echoId)
}
</script>

<style scoped>
.menu-row {
  display: flex;
  align-items: center;
  width: 100%;
  gap: 8px;
  height: 28px;
  padding: 0 8px;
  border-radius: var(--radius-sm);
  color: var(--color-text-secondary);
  font-size: 12px;
  font-weight: 500;
  text-align: left;
  transition:
    color 150ms ease,
    background-color 150ms ease;
}

.menu-row:hover {
  color: var(--color-text-primary);
  background: var(--color-bg-surface);
}

.menu-row:focus-visible {
  outline: none;
  background: var(--color-bg-surface);
  box-shadow: inset 0 0 0 1.5px var(--color-border-strong);
}

.menu-row--danger:hover {
  color: var(--color-danger);
}
</style>
