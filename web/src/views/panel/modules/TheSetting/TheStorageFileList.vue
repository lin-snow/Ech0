<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<script setup lang="ts">
import PanelCard from '@/layout/PanelCard.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import { fetchDownloadFileById, fetchListFiles } from '@/service/api'
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { theToast } from '@/utils/toast'
import { formatBytes } from '@/utils/file'
import DownloadIcon from '@/components/icons/download.vue'
import ViewIcon from '@/components/icons/view.vue'

const PAGE_SIZE = 20
const SEARCH_DEBOUNCE = 300

const { t } = useI18n()

const files = ref<App.Api.File.FileDto[]>([])
const total = ref(0)
const page = ref(1)
const loading = ref(false)
const error = ref('')
const downloadingId = ref('')
const previewingId = ref('')

const filters = reactive({
  search: '',
  kind: '',
  access: '',
})

const pages = computed(() => Math.max(1, Math.ceil(total.value / PAGE_SIZE)))

const KIND_OPTIONS = [
  { labelKey: 'storageFileList.kindOwned', value: 'owned' },
  { labelKey: 'storageFileList.kindBlob', value: 'blob' },
  { labelKey: 'storageFileList.kindRef', value: 'ref' },
] as const satisfies readonly { labelKey: string; value: App.Api.File.Kind }[]

const ACCESS_OPTIONS = [
  { labelKey: 'storageFileList.accessPublic', value: 'public' },
  { labelKey: 'storageFileList.accessOwner', value: 'owner' },
  { labelKey: 'storageFileList.accessPrivate', value: 'private' },
] as const satisfies readonly { labelKey: string; value: App.Api.File.Access }[]

const kindOptions = computed(() => [
  { label: String(t('storageFileList.allKinds')), value: '' },
  ...KIND_OPTIONS.map((item) => ({ label: String(t(item.labelKey)), value: item.value })),
])

const accessOptions = computed(() => [
  { label: String(t('storageFileList.allAccess')), value: '' },
  ...ACCESS_OPTIONS.map((item) => ({ label: String(t(item.labelKey)), value: item.value })),
])

const categoryLabel = (category: App.Api.File.Category) => {
  switch (category) {
    case 'image':
      return String(t('storageFileList.categoryImage'))
    case 'audio':
      return String(t('storageFileList.categoryAudio'))
    case 'video':
      return String(t('storageFileList.categoryVideo'))
    case 'pdf':
      return String(t('storageFileList.categoryPdf'))
    case 'markdown':
      return String(t('storageFileList.categoryMarkdown'))
    default:
      return String(t('storageFileList.categoryFile'))
  }
}

let searchTimer: ReturnType<typeof setTimeout> | undefined

async function loadFiles() {
  loading.value = true
  error.value = ''
  try {
    const res = await fetchListFiles({
      page: page.value,
      pageSize: PAGE_SIZE,
      search: filters.search.trim() || undefined,
      kind: (filters.kind || undefined) as App.Api.File.Kind | undefined,
      access: (filters.access || undefined) as App.Api.File.Access | undefined,
    })
    if (!res) {
      error.value = String(t('storageFileList.loadFailed'))
      files.value = []
      total.value = 0
      return
    }
    files.value = res.items || []
    total.value = res.total || 0
    if (page.value > pages.value) {
      page.value = pages.value
      await loadFiles()
    }
  } catch {
    error.value = String(t('storageFileList.loadFailed'))
    files.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

function resetAndLoad() {
  page.value = 1
  void loadFiles()
}

function goTo(target: number) {
  const next = Math.min(Math.max(target, 1), pages.value)
  if (next === page.value) return
  page.value = next
  void loadFiles()
}

function refresh() {
  void loadFiles()
}

const actionKeyOf = (file: App.Api.File.FileDto) => file.id

const triggerDownload = async (file: App.Api.File.FileDto) => {
  try {
    downloadingId.value = actionKeyOf(file)
    const blob = await fetchDownloadFileById(file.id)
    if (await isErrorLikeBlob(blob)) {
      theToast.error(String(t('storageFileList.downloadFailedResponse')))
      return
    }
    const objectUrl = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = objectUrl
    link.download = file.name || 'file'
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(objectUrl)
  } catch {
    theToast.error(String(t('storageFileList.downloadFailed')))
  } finally {
    downloadingId.value = ''
  }
}

const triggerPreview = async (file: App.Api.File.FileDto) => {
  try {
    previewingId.value = actionKeyOf(file)
    const blob = await fetchDownloadFileById(file.id)
    if (await isErrorLikeBlob(blob)) {
      theToast.error(String(t('storageFileList.previewFailedResponse')))
      return
    }
    const objectUrl = window.URL.createObjectURL(blob)
    window.open(objectUrl, '_blank', 'noopener,noreferrer')
    window.setTimeout(() => {
      window.URL.revokeObjectURL(objectUrl)
    }, 60_000)
  } catch {
    theToast.error(String(t('storageFileList.previewFailed')))
  } finally {
    previewingId.value = ''
  }
}

const isErrorLikeBlob = async (blob: Blob) => {
  const mime = (blob.type || '').toLowerCase()
  if (!mime.includes('text') && !mime.includes('json') && !mime.includes('xml')) {
    return false
  }
  const text = (await blob.text()).trim().toLowerCase()
  return text.includes('文件不存在') || text.includes('error') || text.includes('not found')
}

watch(
  () => [filters.search, filters.kind, filters.access],
  () => {
    if (searchTimer) clearTimeout(searchTimer)
    searchTimer = setTimeout(resetAndLoad, SEARCH_DEBOUNCE)
  },
)

onMounted(() => {
  void loadFiles()
})
</script>

<template>
  <PanelCard>
    <div class="storage-file-list">
      <div class="header">
        <h1 class="title">{{ t('storageFileList.title') }}</h1>
        <button class="refresh-btn" :disabled="loading" @click="refresh">
          {{ loading ? t('storageFileList.refreshing') : t('storageFileList.refresh') }}
        </button>
      </div>

      <div class="filters">
        <BaseInput
          v-model="filters.search"
          class="search-input"
          :placeholder="t('storageFileList.searchPlaceholder')"
        />
        <BaseSelect v-model="filters.kind" class="filter-select" :options="kindOptions" />
        <BaseSelect v-model="filters.access" class="filter-select" :options="accessOptions" />
      </div>

      <div class="explorer-panel">
        <div v-if="loading && files.length === 0" class="status-text nested-status">
          {{ t('storageFileList.loading') }}
        </div>
        <div v-else-if="error" class="status-text status-error nested-status">
          {{ error }}
          <button class="retry-btn" @click="refresh">{{ t('storageFileList.retry') }}</button>
        </div>
        <div v-else-if="files.length === 0" class="status-text nested-status">
          {{ t('storageFileList.emptyFiles') }}
        </div>
        <div v-else class="tree-list">
          <div v-for="file in files" :key="file.id" class="tree-row is-file">
            <div class="tree-left">
              <div class="file-meta">
                <span class="node-name" v-tooltip="file.name">{{ file.name }}</span>
                <span class="file-sub">
                  {{ categoryLabel(file.category) }} · {{ formatBytes(file.size || 0) }}
                </span>
              </div>
            </div>
            <div class="tree-right">
              <span class="file-access">{{
                t(
                  ACCESS_OPTIONS.find((item) => item.value === file.access)?.labelKey ||
                    'storageFileList.accessPrivate',
                )
              }}</span>
              <div
                class="file-actions"
                :class="{
                  'is-busy':
                    downloadingId === actionKeyOf(file) || previewingId === actionKeyOf(file),
                }"
              >
                <button
                  class="download-btn icon-btn"
                  :disabled="
                    downloadingId === actionKeyOf(file) || previewingId === actionKeyOf(file)
                  "
                  v-tooltip="
                    previewingId === actionKeyOf(file)
                      ? t('storageFileList.previewing')
                      : t('storageFileList.preview')
                  "
                  @click.stop="triggerPreview(file)"
                >
                  <span v-if="previewingId === actionKeyOf(file)">◌</span>
                  <ViewIcon v-else class="preview-icon" />
                </button>
                <button
                  class="download-btn icon-btn"
                  :disabled="
                    downloadingId === actionKeyOf(file) || previewingId === actionKeyOf(file)
                  "
                  v-tooltip="
                    downloadingId === actionKeyOf(file)
                      ? t('storageFileList.downloading')
                      : t('storageFileList.download')
                  "
                  @click.stop="triggerDownload(file)"
                >
                  <span v-if="downloadingId === actionKeyOf(file)">◌</span>
                  <DownloadIcon v-else class="download-icon" />
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-if="total > 0" class="pager">
        <span class="count">{{ t('storageFileList.totalFiles', { total }) }}</span>
        <button class="refresh-btn" :disabled="page <= 1 || loading" @click="goTo(page - 1)">
          {{ t('storageFileList.prevPage') }}
        </button>
        <span class="count">{{ t('storageFileList.pageInfo', { page, pages }) }}</span>
        <button class="refresh-btn" :disabled="page >= pages || loading" @click="goTo(page + 1)">
          {{ t('storageFileList.nextPage') }}
        </button>
      </div>
    </div>
  </PanelCard>
</template>

<style scoped>
.storage-file-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.title {
  color: var(--color-text-primary);
  font-weight: 700;
  font-size: 1.1rem;
}

.filters {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.search-input {
  flex: 1;
  min-width: 10rem;
}

.filter-select {
  min-width: 8rem;
}

.retry-btn,
.download-btn,
.refresh-btn {
  border: 1px solid var(--color-border-subtle);
  border-radius: var(--radius-sm);
  padding: 0.2rem 0.65rem;
  color: var(--color-text-secondary);
  background: var(--color-bg-surface);
  cursor: pointer;
}

.refresh-btn:hover,
.retry-btn:hover,
.download-btn:hover:not(:disabled) {
  border-color: var(--color-border-strong);
}

.refresh-btn:disabled,
.download-btn:disabled {
  cursor: not-allowed;
  opacity: 0.7;
}

.count {
  color: var(--color-text-muted);
  font-size: 0.85rem;
}

.explorer-panel {
  border: 1px solid var(--color-border-subtle);
  border-radius: var(--radius-md);
  overflow: hidden;
}

.status-text {
  color: var(--color-text-muted);
  font-size: 0.9rem;
}

.nested-status {
  padding: 0.75rem;
}

.status-error {
  color: var(--color-error, #ef4444);
}

.tree-list {
  max-height: 24rem;
  overflow: auto;
  padding: 0;
}

.tree-row {
  min-height: 2.4rem;
  border-bottom: 1px solid var(--storage-tree-row-border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  padding: 0.08rem 0;
  color: var(--color-text-secondary);
  transition:
    background-color 0.18s ease,
    border-color 0.18s ease;
}

.tree-row:hover {
  background: var(--storage-tree-row-hover-bg);
}

.tree-left {
  min-width: 0;
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding-left: 0.75rem;
  padding-right: 0.6rem;
}

.file-meta {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 0.1rem;
}

.node-name {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.file-sub {
  color: var(--color-text-muted);
  font-size: 0.78rem;
}

.tree-right {
  min-width: 9rem;
  display: inline-flex;
  justify-content: flex-end;
  align-items: center;
  gap: 0.5rem;
  padding-right: 0.4rem;
}

.file-access {
  color: var(--color-text-muted);
  font-size: 0.8rem;
}

.icon-btn {
  min-width: 2.1rem;
  height: 1.9rem;
  padding: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.file-actions {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  opacity: 0;
  pointer-events: none;
  transform: translateX(4px);
  transition:
    opacity 0.16s ease,
    transform 0.16s ease;
}

.tree-row:hover .file-actions,
.file-actions.is-busy {
  opacity: 1;
  pointer-events: auto;
  transform: translateX(0);
}

.download-icon {
  font-size: 1rem;
}

.preview-icon {
  font-size: 0.96rem;
}

.pager {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.5rem;
}

.tree-list > :last-child {
  border-bottom: none;
}
</style>
