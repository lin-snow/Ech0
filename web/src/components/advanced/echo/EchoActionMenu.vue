<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div ref="rootEl" class="relative flex items-center justify-center mr-2">
    <button
      ref="triggerEl"
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

    <!-- Teleported on purpose: every virtual row is its own stacking context
         (transform) and the timeline scroller clips its overflow, so an in-row
         popover is painted underneath the next row and cut off at the viewport
         edge. Fixed coordinates are measured from the trigger when opening. -->
    <Teleport to="body">
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
          ref="panelEl"
          :style="panelStyle"
          class="fixed z-5000 w-36 rounded-[var(--radius-md)] bg-[var(--color-bg-muted)] ring-1 ring-[var(--color-border-subtle)] shadow-[var(--shadow-md)] p-1"
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
    </Teleport>
  </div>
</template>

<script lang="ts">
import { ref as vueRef } from 'vue'

// One open menu across every timeline row: opening another row closes this one,
// and virtual rows recycling away cannot leave a menu stranded.
const openMenuId = vueRef<string | null>(null)
</script>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import EditEcho from '@/components/icons/editecho.vue'
import Lock from '@/components/icons/lock.vue'
import More from '@/components/icons/more.vue'
import Roll from '@/components/icons/roll.vue'
import { fetchDeleteEcho, fetchGetEchoById } from '@/service/api'
import { useEchoStore, useEditorStore } from '@/stores'
import { useBaseDialog } from '@/composables/useBaseDialog'
import { theToast } from '@/utils/toast'

type Echo = App.Api.Ech0.Echo

const props = defineProps<{
  echo: Echo
}>()

const echoStore = useEchoStore()
const editorStore = useEditorStore()
const router = useRouter()
const route = useRoute()
const { openConfirm } = useBaseDialog()
const { t } = useI18n()

const rootEl = ref<HTMLElement | null>(null)
const triggerEl = ref<HTMLElement | null>(null)
const panelEl = ref<HTMLElement | null>(null)
const open = computed(() => openMenuId.value === props.echo.id)

const PANEL_WIDTH = 144
const PANEL_GAP = 8
const VIEWPORT_INSET = 8

// The panel has not mounted yet when the first style is written; a per-shape
// estimate keeps the opening frame in the right place, then the measured height
// refines it (and decides the flip).
const estimatedPanelHeight = () => (props.echo.private ? 92 : 68)

const panelStyle = ref({ top: '0px', right: '0px' })

const placePanel = (fallbackHeight: number) => {
  const trigger = triggerEl.value
  if (!trigger) return
  const rect = trigger.getBoundingClientRect()
  const width = panelEl.value?.offsetWidth || PANEL_WIDTH
  const height = panelEl.value?.offsetHeight || fallbackHeight
  // Right-align with the trigger, then clamp inside the viewport.
  const right = Math.min(
    Math.max(VIEWPORT_INSET, window.innerWidth - rect.right),
    Math.max(VIEWPORT_INSET, window.innerWidth - width - VIEWPORT_INSET),
  )
  const below = rect.bottom + PANEL_GAP
  const top =
    below + height > window.innerHeight - VIEWPORT_INSET
      ? Math.max(VIEWPORT_INSET, rect.top - PANEL_GAP - height)
      : below
  panelStyle.value = { top: `${Math.round(top)}px`, right: `${Math.round(right)}px` }
}

const close = () => {
  if (open.value) openMenuId.value = null
}

const onOutside = (event: PointerEvent) => {
  if (!open.value) return
  const target = event.target as Node | null
  if (!target) return
  if (rootEl.value?.contains(target) || panelEl.value?.contains(target)) return
  close()
}

const onKeydown = (event: KeyboardEvent) => {
  if (event.key === 'Escape') close()
}

const bind = () => {
  document.addEventListener('pointerdown', onOutside)
  document.addEventListener('keydown', onKeydown)
  window.addEventListener('resize', close)
  // A fixed panel stops tracking its trigger on scroll; closing is the only
  // honest option, and scrolling while a menu is open is a dismiss gesture.
  window.addEventListener('scroll', close, { capture: true, passive: true })
}

const unbind = () => {
  document.removeEventListener('pointerdown', onOutside)
  document.removeEventListener('keydown', onKeydown)
  window.removeEventListener('resize', close)
  window.removeEventListener('scroll', close, { capture: true })
}

const toggle = () => {
  openMenuId.value = open.value ? null : props.echo.id
}

watch(
  open,
  (value) => {
    unbind()
    if (!value) return
    placePanel(estimatedPanelHeight())
    void nextTick(() => {
      if (open.value) placePanel(estimatedPanelHeight())
    })
    bind()
  },
  { immediate: true },
)

watch(
  () => route.fullPath,
  () => close(),
)

onBeforeUnmount(() => {
  unbind()
  close()
})

const startUpdate = async (echo: Echo) => {
  const res = await fetchGetEchoById(String(echo.id))
  echoStore.echoToUpdate = res.code === 1 && res.data ? res.data : echo
  editorStore.isUpdateMode = true
  await router.push({
    name: 'home',
    query: { ...router.currentRoute.value.query, tab: 'publish' },
  })
}

const onUpdate = () => {
  close()
  if (!editorStore.isUpdateMode) {
    void startUpdate(props.echo)
    return
  }
  openConfirm({
    title: String(t('echoCard.editAnotherEchoTitle')),
    description: String(t('echoCard.editAnotherEchoDesc')),
    onConfirm: () => {
      void startUpdate(props.echo)
    },
  })
}

const onDelete = () => {
  const echoId = props.echo.id
  close()
  openConfirm({
    title: String(t('echoCard.deleteConfirmTitle')),
    description: String(t('echoCard.deleteConfirmDesc')),
    onConfirm: () => {
      fetchDeleteEcho(echoId).then(() => {
        theToast.success(String(t('echoCard.deleteSuccess')))
        echoStore.removeEcho(echoId)
      })
    },
  })
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
