<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<script setup lang="ts">
import PanelCard from '@/layout/PanelCard.vue'
import { FILE_STORAGE_TYPE } from '@/constants/file'
import { fetchDownloadFileById, fetchDownloadFileByPath, fetchFileTree } from '@/service/api'
import { nextTick, reactive, ref, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { theToast } from '@/utils/toast'
import DownloadIcon from '@/components/icons/download.vue'
import ViewIcon from '@/components/icons/view.vue'
import { useSettingStore } from '@/stores'
import { storeToRefs } from 'pinia'

const EASE_OUT = 'cubic-bezier(0.33, 1, 0.68, 1)'
const EASE_IN_OUT = 'cubic-bezier(0.65, 0, 0.35, 1)'

const runningAnims = new WeakMap<HTMLElement, Animation>()

const cancelAnim = (el: HTMLElement) => {
  runningAnims.get(el)?.cancel()
  runningAnims.delete(el)
}

const playAnim = (
  el: HTMLElement,
  keyframes: Keyframe[],
  options: KeyframeAnimationOptions,
  onFinish?: () => void,
) => {
  cancelAnim(el)
  const anim = el.animate(keyframes, options)
  runningAnims.set(el, anim)
  anim.onfinish = () => {
    if (runningAnims.get(el) === anim) runningAnims.delete(el)
    onFinish?.()
  }
  anim.oncancel = () => {
    if (runningAnims.get(el) === anim) runningAnims.delete(el)
  }
  return anim
}

type RootStorageType = typeof FILE_STORAGE_TYPE.LOCAL | typeof FILE_STORAGE_TYPE.OBJECT

type TreeNode = App.Api.File.FileTreeNode & {
  children: TreeNode[]
  expanded: boolean
  loaded: boolean
  loading: boolean
  error: string
}

type TreeRow = {
  node: TreeNode
  depth: number
}

type FileSectionState = {
  title: string
  expanded: boolean
  loaded: boolean
  loading: boolean
  error: string
  nodes: TreeNode[]
}

type LoadChildrenOptions = {
  force?: boolean
  preserveOnError?: boolean
}

const { t } = useI18n()

const sections = reactive<Record<RootStorageType, FileSectionState>>({
  [FILE_STORAGE_TYPE.LOCAL]: {
    title: '',
    expanded: false,
    loaded: false,
    loading: false,
    error: '',
    nodes: [],
  },
  [FILE_STORAGE_TYPE.OBJECT]: {
    title: '',
    expanded: false,
    loaded: false,
    loading: false,
    error: '',
    nodes: [],
  },
})

const downloadingId = ref('')
const previewingId = ref('')
const refreshingRoot = ref<RootStorageType | ''>('')
const selectedNodeKey = ref('')
const settingStore = useSettingStore()
const { S3Setting } = storeToRefs(settingStore)
const rootStorageTypes = computed<RootStorageType[]>(() =>
  S3Setting.value.enable
    ? [FILE_STORAGE_TYPE.LOCAL, FILE_STORAGE_TYPE.OBJECT]
    : [FILE_STORAGE_TYPE.LOCAL],
)
const nodeCache = reactive<Record<string, App.Api.File.FileTreeNode[]>>({})
const rootContentRefs = reactive<Partial<Record<RootStorageType, HTMLElement | null>>>({})

const toNode = (raw: App.Api.File.FileTreeNode): TreeNode => ({
  ...raw,
  children: [],
  expanded: false,
  loaded: false,
  loading: false,
  error: '',
})

const buildNodeKey = (storageType: RootStorageType, path: string) => `${storageType}:${path}`
const buildCacheKey = (storageType: RootStorageType, prefix: string) =>
  `${storageType}:${prefix.trim() === '' ? '/' : prefix}`

const cloneRawNodes = (items: App.Api.File.FileTreeNode[]) => items.map((item) => ({ ...item }))

const getActiveRoot = (): RootStorageType | '' => {
  return rootStorageTypes.value.find((type) => sections[type].expanded) || ''
}

const clearCacheByRoot = (storageType: RootStorageType) => {
  const prefix = `${storageType}:`
  for (const key of Object.keys(nodeCache)) {
    if (key.startsWith(prefix)) {
      delete nodeCache[key]
    }
  }
}

const loadChildren = async (
  storageType: RootStorageType,
  prefix: string,
  parentNode?: TreeNode,
  options: LoadChildrenOptions = {},
) => {
  const { force = false, preserveOnError = true } = options
  const cleanPrefix = prefix.trim()
  const cacheKey = buildCacheKey(storageType, cleanPrefix)
  const section = sections[storageType]

  if (!force && nodeCache[cacheKey]) {
    const cachedNodes = nodeCache[cacheKey].map(toNode)
    if (parentNode) {
      parentNode.children = cachedNodes
      parentNode.loaded = true
      parentNode.error = ''
    } else {
      section.nodes = cachedNodes
      section.loaded = true
      section.error = ''
    }
    return
  }

  const backupChildren = parentNode ? [...parentNode.children] : [...section.nodes]
  if (parentNode) {
    parentNode.loading = true
    parentNode.error = ''
  } else {
    section.loading = true
    section.error = ''
  }

  const res = await fetchFileTree({
    storage_type: storageType,
    prefix: cleanPrefix,
  })

  if (res.code === 1) {
    const rawItems = res.data?.items || []
    nodeCache[cacheKey] = cloneRawNodes(rawItems)
    const nodes = rawItems.map(toNode)
    if (parentNode) {
      parentNode.children = nodes
      parentNode.loaded = true
      parentNode.error = ''
    } else {
      section.nodes = nodes
      section.loaded = true
      section.error = ''
    }
  } else if (parentNode) {
    parentNode.error = res.msg || String(t('storageFileList.loadFailed'))
    if (preserveOnError && backupChildren.length > 0) {
      parentNode.children = backupChildren
      parentNode.loaded = true
    }
  } else {
    section.error = res.msg || String(t('storageFileList.loadFailed'))
    if (preserveOnError && backupChildren.length > 0) {
      section.nodes = backupChildren
      section.loaded = true
    }
  }

  if (parentNode) {
    parentNode.loading = false
  } else {
    section.loading = false
  }
}

const toggleRoot = async (storageType: RootStorageType) => {
  const section = sections[storageType]
  if (section.loading || refreshingRoot.value !== '') return
  const nextExpanded = !section.expanded
  for (const rootType of rootStorageTypes.value) {
    if (rootType !== storageType) {
      sections[rootType].expanded = false
    }
  }
  section.expanded = nextExpanded
  if (section.expanded && !section.loaded && !section.loading) {
    await loadChildren(storageType, '')
  }
}

const toggleFolder = async (storageType: RootStorageType, node: TreeNode) => {
  if (node.loading || refreshingRoot.value !== '') return
  const contentEl = rootContentRefs[storageType]
  const startHeight = contentEl?.scrollHeight ?? 0
  node.expanded = !node.expanded
  if (node.expanded && !node.loaded) {
    await loadChildren(storageType, node.path, node)
  }
  if (!contentEl || !sections[storageType].expanded) return
  await nextTick()
  const endHeight = contentEl.scrollHeight
  if (startHeight === 0 || startHeight === endHeight) return
  contentEl.style.overflow = 'hidden'
  playAnim(
    contentEl,
    [{ height: `${startHeight}px` }, { height: `${endHeight}px` }],
    { duration: 260, easing: EASE_OUT },
    () => {
      contentEl.style.height = 'auto'
      contentEl.style.overflow = ''
    },
  )
}

const handleNodeClick = async (storageType: RootStorageType, node: TreeNode) => {
  if (node.node_type === 'folder') {
    await toggleFolder(storageType, node)
    return
  }
  selectedNodeKey.value = buildNodeKey(storageType, node.path)
}

const flattenRows = (nodes: TreeNode[], depth = 0): TreeRow[] => {
  const rows: TreeRow[] = []
  for (const node of nodes) {
    rows.push({ node, depth })
    if (node.node_type === 'folder' && node.expanded && node.loaded && node.children.length > 0) {
      rows.push(...flattenRows(node.children, depth + 1))
    }
  }
  return rows
}

const getVisibleRows = (storageType: RootStorageType) => flattenRows(sections[storageType].nodes)

const isNodeSelected = (storageType: RootStorageType, node: TreeNode) => {
  return selectedNodeKey.value === buildNodeKey(storageType, node.path)
}

const rootMark = (storageType: RootStorageType) => {
  return storageType === FILE_STORAGE_TYPE.LOCAL ? 'L' : 'O'
}

const rootTitle = (storageType: RootStorageType) => {
  return storageType === FILE_STORAGE_TYPE.LOCAL
    ? String(t('storageFileList.localStorage'))
    : String(t('storageFileList.objectStorage'))
}

const refreshCurrentRoot = async () => {
  const activeRoot = getActiveRoot()
  if (!activeRoot) return
  if (refreshingRoot.value !== '' || sections[activeRoot].loading) return
  refreshingRoot.value = activeRoot
  clearCacheByRoot(activeRoot)
  sections[activeRoot].loaded = false
  sections[activeRoot].error = ''
  await loadChildren(activeRoot, '', undefined, { force: true, preserveOnError: true })
  refreshingRoot.value = ''
}

const actionKeyOf = (storageType: RootStorageType, node: TreeNode) =>
  node.file_id || `path:${storageType}:${node.path}`

const triggerDownload = async (storageType: RootStorageType, node: TreeNode) => {
  try {
    const actionKey = actionKeyOf(storageType, node)
    downloadingId.value = actionKey
    const blob = node.file_id
      ? await fetchDownloadFileById(node.file_id)
      : await fetchDownloadFileByPath({
          storage_type: storageType,
          path: node.path,
          name: node.name,
          content_type: node.content_type,
        })
    if (await isErrorLikeBlob(blob)) {
      theToast.error(String(t('storageFileList.downloadFailedResponse')))
      return
    }
    const objectUrl = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = objectUrl
    link.download = node.name || 'file'
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

const triggerPreview = async (storageType: RootStorageType, node: TreeNode) => {
  try {
    const actionKey = actionKeyOf(storageType, node)
    previewingId.value = actionKey
    const blob = node.file_id
      ? await fetchDownloadFileById(node.file_id)
      : await fetchDownloadFileByPath({
          storage_type: storageType,
          path: node.path,
          name: node.name,
          content_type: node.content_type,
        })
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

const onBeforeEnter = (el: Element) => {
  const element = el as HTMLElement
  cancelAnim(element)
  element.style.overflow = 'hidden'
  element.style.height = '0'
  element.style.opacity = '0'
}

const onEnter = (el: Element, done: () => void) => {
  const element = el as HTMLElement
  playAnim(
    element,
    [
      { height: '0px', opacity: 0 },
      { height: `${element.scrollHeight}px`, opacity: 1 },
    ],
    { duration: 340, easing: EASE_OUT },
    () => {
      element.style.height = 'auto'
      element.style.opacity = '1'
      done()
    },
  )
}

const onAfterEnter = (el: Element) => {
  const element = el as HTMLElement
  element.style.height = 'auto'
  element.style.overflow = ''
}

const onBeforeLeave = (el: Element) => {
  const element = el as HTMLElement
  cancelAnim(element)
  element.style.overflow = 'hidden'
  element.style.height = `${element.scrollHeight}px`
  element.style.opacity = '1'
}

const onLeave = (el: Element, done: () => void) => {
  const element = el as HTMLElement
  playAnim(
    element,
    [
      { height: `${element.scrollHeight}px`, opacity: 1 },
      { height: '0px', opacity: 0 },
    ],
    { duration: 240, easing: EASE_IN_OUT },
    () => {
      element.style.height = '0px'
      element.style.opacity = '0'
      done()
    },
  )
}

const onAfterLeave = (el: Element) => {
  const element = el as HTMLElement
  element.style.height = ''
  element.style.opacity = ''
  element.style.overflow = ''
}

const setRootContentRef = (storageType: RootStorageType, el: unknown) => {
  rootContentRefs[storageType] = el instanceof HTMLElement ? el : null
}

watch(
  () => S3Setting.value.enable,
  (enabled) => {
    if (!enabled) {
      sections[FILE_STORAGE_TYPE.OBJECT].expanded = false
      if (selectedNodeKey.value.startsWith(`${FILE_STORAGE_TYPE.OBJECT}:`)) {
        selectedNodeKey.value = ''
      }
    }
  },
)

onMounted(() => {
  settingStore.getS3Setting()
})
</script>

<template>
  <PanelCard>
    <div class="storage-file-list">
      <div class="header">
        <h1 class="title">{{ t('storageFileList.title') }}</h1>
        <button
          class="refresh-btn"
          :disabled="!getActiveRoot() || refreshingRoot !== ''"
          @click="refreshCurrentRoot"
        >
          {{
            refreshingRoot !== '' ? t('storageFileList.refreshing') : t('storageFileList.refresh')
          }}
        </button>
      </div>
      <div class="explorer-panel">
        <div class="tree-list">
          <template v-for="storageType in rootStorageTypes" :key="storageType">
            <div class="tree-row root-row" @click="toggleRoot(storageType)">
              <div class="tree-left">
                <span class="node-icon">{{ sections[storageType].expanded ? '▾' : '▸' }}</span>
                <span class="root-mark">{{ rootMark(storageType) }}</span>
                <span class="node-name">{{ rootTitle(storageType) }}</span>
              </div>
              <div class="tree-right">
                <span class="count">{{
                  sections[storageType].expanded
                    ? t('storageFileList.expanded')
                    : t('storageFileList.collapsed')
                }}</span>
              </div>
            </div>

            <Transition
              :css="false"
              @before-enter="onBeforeEnter"
              @enter="onEnter"
              @after-enter="onAfterEnter"
              @before-leave="onBeforeLeave"
              @leave="onLeave"
              @after-leave="onAfterLeave"
            >
              <div
                v-show="sections[storageType].expanded"
                :ref="(el) => setRootContentRef(storageType, el)"
                class="root-content"
                :class="{ 'is-collapsed': !sections[storageType].expanded }"
              >
                <div v-if="sections[storageType].loading" class="status-text nested-status">
                  {{ t('storageFileList.loading') }}
                </div>
                <div
                  v-else-if="sections[storageType].error"
                  class="status-text status-error nested-status"
                >
                  {{ sections[storageType].error }}
                  <button class="retry-btn" @click.stop="loadChildren(storageType, '')">
                    {{ t('storageFileList.retry') }}
                  </button>
                </div>
                <div
                  v-else-if="
                    sections[storageType].loaded && sections[storageType].nodes.length === 0
                  "
                  class="status-text nested-status"
                >
                  {{ t('storageFileList.emptyFiles') }}
                </div>
                <template v-else>
                  <div>
                    <div
                      v-for="row in getVisibleRows(storageType)"
                      :key="`${storageType}:${row.node.path}`"
                      class="tree-row"
                      :class="{
                        'is-file': row.node.node_type === 'file',
                        'is-selected': isNodeSelected(storageType, row.node),
                      }"
                      @click="handleNodeClick(storageType, row.node)"
                    >
                      <div
                        class="tree-left"
                        :style="{ paddingLeft: `${row.depth * 1.15 + 1.55}rem` }"
                      >
                        <span class="node-icon">
                          {{
                            row.node.node_type === 'folder' ? (row.node.expanded ? '▾' : '▸') : '•'
                          }}
                        </span>
                        <span class="node-name" v-tooltip="row.node.name">{{ row.node.name }}</span>
                      </div>
                      <div class="tree-right">
                        <span
                          v-if="row.node.node_type === 'folder' && row.node.loading"
                          class="node-status"
                        >
                          {{ t('storageFileList.loading') }}
                        </span>
                        <button
                          v-else-if="row.node.node_type === 'folder' && row.node.error"
                          class="retry-btn"
                          @click.stop="loadChildren(storageType, row.node.path, row.node)"
                        >
                          {{ t('storageFileList.retry') }}
                        </button>
                        <div
                          v-else-if="row.node.node_type === 'file'"
                          class="file-actions"
                          :class="{
                            'is-busy':
                              downloadingId === actionKeyOf(storageType, row.node) ||
                              previewingId === actionKeyOf(storageType, row.node),
                          }"
                        >
                          <button
                            class="download-btn icon-btn"
                            :disabled="
                              downloadingId === actionKeyOf(storageType, row.node) ||
                              previewingId === actionKeyOf(storageType, row.node)
                            "
                            v-tooltip="
                              previewingId === actionKeyOf(storageType, row.node)
                                ? t('storageFileList.previewing')
                                : t('storageFileList.preview')
                            "
                            @click.stop="triggerPreview(storageType, row.node)"
                          >
                            <span v-if="previewingId === actionKeyOf(storageType, row.node)"
                              >◌</span
                            >
                            <ViewIcon v-else class="preview-icon" />
                          </button>
                          <button
                            class="download-btn icon-btn"
                            :disabled="
                              downloadingId === actionKeyOf(storageType, row.node) ||
                              previewingId === actionKeyOf(storageType, row.node)
                            "
                            v-tooltip="
                              downloadingId === actionKeyOf(storageType, row.node)
                                ? t('storageFileList.downloading')
                                : t('storageFileList.download')
                            "
                            @click.stop="triggerDownload(storageType, row.node)"
                          >
                            <span v-if="downloadingId === actionKeyOf(storageType, row.node)"
                              >◌</span
                            >
                            <DownloadIcon v-else class="download-icon" />
                          </button>
                        </div>
                      </div>
                    </div>
                  </div>
                </template>
              </div>
            </Transition>
          </template>
        </div>
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

.retry-btn,
.download-btn {
  border: 1px solid var(--color-border-subtle);
  border-radius: var(--radius-sm);
  padding: 0.2rem 0.65rem;
  color: var(--color-text-secondary);
  background: var(--color-bg-surface);
  cursor: pointer;
}

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

.refresh-btn:disabled {
  cursor: not-allowed;
  opacity: 0.7;
}

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
  padding: 0.45rem 0.5rem 0.45rem 1.8rem;
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

.root-row {
  cursor: pointer;
  border-bottom: 1px solid var(--color-border-subtle);
  background: var(--storage-root-row-bg);
}

.tree-row:hover {
  background: var(--storage-tree-row-hover-bg);
}

.tree-row.is-file {
  cursor: pointer;
}

.tree-row.is-selected {
  background: var(--storage-tree-row-selected-bg);
}

.tree-left {
  min-width: 0;
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding-right: 0.6rem;
}

.node-icon {
  width: 0.95rem;
  color: var(--color-text-muted);
  text-align: center;
}

.root-mark {
  width: 1.15rem;
  height: 1.15rem;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 0.72rem;
  font-weight: 700;
  color: var(--color-accent);
  background: var(--storage-root-mark-bg);
  border: 1px solid var(--storage-root-mark-border);
}

.node-name {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.tree-right {
  min-width: 7rem;
  display: inline-flex;
  justify-content: flex-end;
  align-items: center;
  gap: 0.35rem;
  padding-right: 0.4rem;
}

.node-status {
  color: var(--color-text-muted);
  font-size: 0.82rem;
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

.tree-row.is-file:hover .file-actions,
.tree-row.is-selected .file-actions,
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

.root-content {
  overflow: hidden;
}

.root-content.is-collapsed {
  pointer-events: none;
}

.tree-list > :last-child {
  border-bottom: none;
}
</style>
