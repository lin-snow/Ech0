<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <Teleport to="body">
    <transition
      enter-active-class="transition duration-150 ease-out"
      enter-from-class="opacity-0 scale-95"
      enter-to-class="opacity-100 scale-100"
      leave-active-class="transition duration-100 ease-in"
      leave-from-class="opacity-100 scale-100"
      leave-to-class="opacity-0 scale-95"
    >
      <div
        v-if="menu"
        ref="panelEl"
        :style="{ top: `${menu.top}px`, right: `${menu.right}px` }"
        class="fixed z-5000 w-36 origin-top-right rounded-[var(--radius-md)] bg-[var(--color-bg-muted)] ring-1 ring-[var(--color-border-subtle)] shadow-[var(--shadow-md)] p-1"
        role="menu"
      >
        <div
          v-if="menu.echo.private"
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
  </Teleport>
</template>

<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import EditEcho from '@/components/icons/editecho.vue'
import Lock from '@/components/icons/lock.vue'
import Roll from '@/components/icons/roll.vue'
import { useEchoActions } from '@/composables/useEchoActions'
import { useEchoMenu } from '@/composables/useEchoMenu'

const { menu, closeMenu } = useEchoMenu()
const { updateEcho, deleteEcho } = useEchoActions()
const { t } = useI18n()
const route = useRoute()
const panelEl = ref<HTMLElement | null>(null)

watch(
  () => route.fullPath,
  () => closeMenu(),
)

const onOutside = (event: PointerEvent) => {
  const target = event.target as Node | null
  if (!target) return
  if (panelEl.value?.contains(target)) return
  closeMenu()
}

const onKeydown = (event: KeyboardEvent) => {
  if (event.key === 'Escape') closeMenu()
}

const onResize = () => closeMenu()

const bind = () => {
  document.addEventListener('pointerdown', onOutside)
  document.addEventListener('keydown', onKeydown)
  window.addEventListener('resize', onResize)
}

const unbind = () => {
  document.removeEventListener('pointerdown', onOutside)
  document.removeEventListener('keydown', onKeydown)
  window.removeEventListener('resize', onResize)
}

watch(menu, (open) => {
  unbind()
  if (open) bind()
})

onBeforeUnmount(() => {
  unbind()
  closeMenu()
})

const onUpdate = () => {
  const echo = menu.value?.echo
  closeMenu()
  if (echo) void updateEcho(echo)
}

const onDelete = () => {
  const echoId = menu.value?.echo.id
  closeMenu()
  if (echoId) deleteEcho(echoId)
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
