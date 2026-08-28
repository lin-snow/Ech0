<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <PanelCard>
    <div class="w-full">
      <div class="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
        <div>
          <h1 class="text-lg font-bold text-[var(--color-text-primary)]">Ech0 MCP</h1>
          <p class="mcp-lead mt-1">{{ t('mcpSetting.description') }}</p>
        </div>
        <BaseButton
          class="top-action-btn top-action-btn-primary shrink-0 whitespace-nowrap px-2.5 py-1 text-xs"
          @click="goToAccessTokens"
        >
          {{ t('mcpSetting.createToken') }}
        </BaseButton>
      </div>

      <div v-if="isLoading" class="mt-6 text-center text-sm text-[var(--color-text-muted)]">
        {{ t('commonUi.loading') }}
      </div>

      <div v-else-if="!manifest" class="mt-6 text-center text-sm text-[var(--color-text-muted)]">
        {{ t('mcpSetting.loadFailed') }}
      </div>

      <template v-else>
        <section class="mcp-block">
          <h2 class="mcp-block-title">{{ t('mcpSetting.endpointTitle') }}</h2>
          <p class="mcp-hint">{{ t('mcpSetting.endpointHint') }}</p>

          <div class="mcp-address">
            <code>{{ endpointUrl }}</code>
            <button
              type="button"
              class="mcp-copy"
              v-tooltip="t('mcpSetting.copyEndpoint')"
              @click="copyEndpoint"
            >
              <Clipboard class="h-full w-full" />
            </button>
          </div>

          <dl class="mcp-rows">
            <dt>{{ t('mcpSetting.transport') }}</dt>
            <dd>{{ transportLabel }}</dd>
            <dt>{{ t('mcpSetting.audience') }}</dt>
            <dd>{{ manifest.audience }}</dd>
            <dt>{{ t('mcpSetting.protocolVersions') }}</dt>
            <dd>{{ manifest.protocol_versions.join(' · ') }}</dd>
          </dl>
        </section>

        <section class="mcp-block">
          <h2 class="mcp-block-title">{{ t('mcpSetting.capabilitiesTitle') }}</h2>

          <div class="mcp-groups">
            <div v-for="group in capabilityGroups" :key="group.key" class="mcp-group">
              <p class="mcp-group-label">
                {{ group.label }}
                <span class="mcp-count">{{ group.items.length }}</span>
              </p>
              <div class="mcp-names">
                <button
                  v-for="item in group.items"
                  :key="item.id"
                  type="button"
                  class="mcp-name"
                  :class="{ 'mcp-name-danger': item.danger }"
                  @click="openDetail(item)"
                >
                  {{ item.id }}
                </button>
              </div>
            </div>
          </div>

          <p class="mcp-hint mcp-legend">{{ t('mcpSetting.dangerLegend') }}</p>
        </section>
      </template>
    </div>
  </PanelCard>

  <TransitionRoot appear :show="selectedItem !== null" as="template">
    <Dialog as="div" class="relative z-5000" @close="closeDetail">
      <TransitionChild
        as="template"
        enter="duration-200 ease-out"
        enter-from="opacity-0"
        enter-to="opacity-100"
        leave="duration-150 ease-in"
        leave-from="opacity-100"
        leave-to="opacity-0"
      >
        <div class="fixed inset-0 bg-black/30 backdrop-blur-sm" />
      </TransitionChild>

      <div class="fixed inset-0 overflow-y-auto">
        <div class="flex min-h-full items-center justify-center p-4">
          <TransitionChild
            as="template"
            enter="duration-200 ease-out"
            enter-from="opacity-0 scale-95"
            enter-to="opacity-100 scale-100"
            leave="duration-150 ease-in"
            leave-from="opacity-100 scale-100"
            leave-to="opacity-0 scale-95"
          >
            <DialogPanel
              class="w-full max-w-lg transform rounded-[var(--radius-lg)] bg-[var(--dialog-bg-color)] p-5 text-left align-middle shadow-[var(--shadow-md)] ring-1 ring-inset ring-[var(--color-border-subtle)] transition-all"
            >
              <div class="flex items-start justify-between gap-3">
                <div class="min-w-0">
                  <DialogTitle class="mcp-detail-name">{{ selectedItem?.id }}</DialogTitle>
                  <p class="mt-1 text-xs text-[var(--color-text-muted)]">
                    {{ selectedItem?.kind }}
                  </p>
                </div>
                <button
                  type="button"
                  class="-mt-1 -mr-1 shrink-0 cursor-pointer rounded-md p-1.5 text-[var(--color-text-muted)] transition-colors hover:bg-[var(--color-bg-muted)] hover:text-[var(--color-text-primary)]"
                  :aria-label="t('mcpSetting.detailClose')"
                  @click="closeDetail"
                >
                  <Close class="h-4 w-4" />
                </button>
              </div>

              <dl class="mcp-rows mcp-detail-rows">
                <dt>{{ t('mcpSetting.detailScopes') }}</dt>
                <dd>{{ selectedItem?.scopes.join(' · ') }}</dd>
                <template v-if="selectedItem?.danger">
                  <dt>{{ t('mcpSetting.detailRisk') }}</dt>
                  <dd class="text-[var(--color-danger)]">
                    {{ t('mcpSetting.detailRiskValue') }}
                  </dd>
                </template>
              </dl>

              <p class="mcp-detail-desc">{{ selectedItem?.description }}</p>
            </DialogPanel>
          </TransitionChild>
        </div>
      </div>
    </Dialog>
  </TransitionRoot>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import PanelCard from '@/layout/PanelCard.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import Clipboard from '@/components/icons/clipboard.vue'
import Close from '@/components/icons/close.vue'
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from '@headlessui/vue'
import { fetchGetMCPManifest } from '@/service/api'
import { theToast } from '@/utils/toast'

const TRANSPORT_LABELS: Record<string, string> = {
  'streamable-http': 'Streamable HTTP',
}

const { t } = useI18n()
const router = useRouter()

const manifest = ref<App.Api.MCP.Manifest | null>(null)
const isLoading = ref<boolean>(true)

const endpointUrl = computed<string>(
  () => `${window.location.origin}${manifest.value?.path ?? '/mcp'}`,
)

const transportLabel = computed<string>(() => {
  const transport = manifest.value?.transport ?? ''
  return TRANSPORT_LABELS[transport] ?? transport
})

const groupLabels = computed<Record<string, string>>(() => ({
  echo: String(t('mcpSetting.groupEcho')),
  profile: String(t('mcpSetting.groupProfile')),
  comment: String(t('mcpSetting.groupComment')),
  file: String(t('mcpSetting.groupFile')),
  connect: String(t('mcpSetting.groupConnect')),
  admin: String(t('mcpSetting.groupAdmin')),
}))

type CapabilityItem = {
  id: string
  kind: string
  scopes: string[]
  description: string
  danger: boolean
}

const capabilityGroups = computed<{ key: string; label: string; items: CapabilityItem[] }[]>(() => {
  const groups = new Map<string, CapabilityItem[]>()
  for (const tool of manifest.value?.tools ?? []) {
    const key = tool.scopes[0]?.split(':')[0] || 'other'
    const item: CapabilityItem = {
      id: tool.name,
      kind: String(t('mcpSetting.kindTool')),
      scopes: tool.scopes,
      description: tool.description,
      danger: tool.destructive,
    }
    const bucket = groups.get(key)
    if (bucket) bucket.push(item)
    else groups.set(key, [item])
  }

  const resources: CapabilityItem[] = [
    ...(manifest.value?.resources ?? []).map((resource) => ({
      id: resource.uri,
      kind: String(t('mcpSetting.kindResource')),
      scopes: resource.scopes,
      description: resource.description || resource.name,
      danger: false,
    })),
    ...(manifest.value?.resource_templates ?? []).map((template) => ({
      id: template.uri_template,
      kind: String(t('mcpSetting.kindTemplate')),
      scopes: template.scopes,
      description: template.description || template.name,
      danger: false,
    })),
  ]

  const rows = [...groups].map(([key, items]) => ({
    key,
    label: groupLabels.value[key] ?? key,
    items,
  }))
  if (resources.length > 0) {
    rows.push({ key: 'resources', label: String(t('mcpSetting.groupResources')), items: resources })
  }
  return rows
})

const selectedItem = ref<CapabilityItem | null>(null)
const openDetail = (item: CapabilityItem) => {
  selectedItem.value = item
}
const closeDetail = () => {
  selectedItem.value = null
}

const copyEndpoint = async () => {
  try {
    await navigator.clipboard.writeText(endpointUrl.value)
    theToast.success(String(t('mcpSetting.copySuccess')))
  } catch {
    theToast.error(String(t('mcpSetting.copyFailed')))
  }
}

const goToAccessTokens = () => {
  router.push('/panel/setting')
}

onMounted(async () => {
  try {
    const res = await fetchGetMCPManifest()
    if (res.code === 1 && res.data) {
      manifest.value = res.data
    }
  } finally {
    isLoading.value = false
  }
})
</script>

<style scoped>
.mcp-lead {
  font-size: 0.8125rem;
  line-height: 1.35rem;
  color: var(--color-text-muted);
}

.mcp-block {
  margin-top: 1.4rem;
}

.mcp-block + .mcp-block {
  padding-top: 1.4rem;
  border-top: 1px solid var(--color-border-subtle);
}

.mcp-block-title {
  font-size: 0.8125rem;
  font-weight: 600;
  letter-spacing: 0.02em;
  color: var(--color-text-primary);
}

.mcp-hint {
  margin-top: 0.25rem;
  font-size: 0.78rem;
  line-height: 1.25rem;
  color: var(--color-text-muted);
  overflow-wrap: anywhere;
}

.mcp-address {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-top: 0.7rem;
  border-radius: 0.5rem;
  padding: 0.45rem 0.5rem 0.45rem 0.7rem;
  background: var(--color-bg-muted);
}

.mcp-address code {
  flex: 1;
  min-width: 0;
  font-size: 0.8125rem;
  color: var(--color-text-primary);
  overflow-wrap: anywhere;
}

.mcp-copy {
  display: inline-flex;
  height: 1.65rem;
  width: 1.65rem;
  flex: none;
  align-items: center;
  justify-content: center;
  border-radius: var(--btn-radius);
  padding: 0.3rem;
  color: var(--color-text-muted);
  background: transparent;
  transition:
    color 0.2s,
    background-color 0.2s;
}

.mcp-copy:hover {
  color: var(--color-text-primary);
  background: var(--color-bg-surface);
}

.mcp-rows {
  display: grid;
  grid-template-columns: max-content minmax(0, 1fr);
  gap: 0.5rem 1.5rem;
  margin-top: 0.9rem;
  font-size: 0.8125rem;
  line-height: 1.35rem;
}

.mcp-rows dt {
  color: var(--color-text-muted);
  white-space: nowrap;
}

@media (width <= 640px) {
  .mcp-rows {
    grid-template-columns: minmax(0, 1fr);
    gap: 0.1rem;
  }

  .mcp-rows dt:not(:first-child) {
    margin-top: 0.55rem;
  }
}

.mcp-rows dd {
  min-width: 0;
  color: var(--color-text-primary);

  overflow-wrap: normal;
}

.mcp-groups {
  display: grid;

  grid-template-columns: minmax(0, 1fr);
  margin-top: 0.9rem;
}

.mcp-group + .mcp-group {
  margin-top: 0.7rem;
  padding-top: 0.7rem;
  border-top: 1px dashed var(--color-border-subtle);
}

.mcp-group-label {
  font-size: 0.75rem;
  color: var(--color-text-muted);
}

.mcp-count {
  margin-left: 0.3rem;
  opacity: 0.55;
}

.mcp-names {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.1rem 1.15rem;
  min-width: 0;
  margin-top: 0.3rem;
}

.mcp-name {
  flex: none;
  min-width: 0;
  max-width: 100%;
  padding: 0;
  overflow: hidden;
  font-family: var(--font-mono, monospace);
  font-size: 0.78rem;
  line-height: 1.5rem;
  color: inherit;

  white-space: nowrap;
  text-overflow: ellipsis;
  background: transparent;
  cursor: pointer;
  transition: color 0.15s;
}

@media (width <= 480px) {
  .mcp-name {
    font-size: 0.72rem;
  }
}

.mcp-name:hover,
.mcp-name:focus-visible {
  color: var(--color-text-primary);
  text-decoration: underline;
  text-underline-offset: 0.2rem;
}

.mcp-name-danger,
.mcp-name-danger:hover,
.mcp-name-danger:focus-visible {
  color: var(--color-danger);
}

.mcp-legend {
  margin-top: 0.9rem;
}

.mcp-detail-name {
  font-family: var(--font-mono, monospace);
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--color-text-primary);
  overflow-wrap: anywhere;
}

.mcp-detail-rows {
  margin-top: 1rem;
}

.mcp-detail-desc {
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid var(--color-border-subtle);
  font-size: 0.8125rem;
  line-height: 1.4rem;
  color: var(--color-text-secondary);
  overflow-wrap: anywhere;
}

.top-action-btn {
  border: 1px solid var(--color-border-subtle) !important;
  background: var(--color-bg-surface) !important;
  color: var(--color-text-secondary) !important;
}

.top-action-btn:hover {
  border-color: var(--color-border-strong) !important;
  background: var(--color-bg-muted) !important;
}

.top-action-btn-primary {
  border-color: var(--color-border-strong) !important;
}
</style>
