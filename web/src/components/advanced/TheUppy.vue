<template>
  <!-- Uppy Dashboard 容器 -->
  <div id="uppy-dashboard" class="rounded-md overflow-hidden shadow-inner"></div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onBeforeUnmount } from 'vue'
import { getAuthToken } from '@/service/request/shared'
import { useUserStore, useEditorStore } from '@/stores'
import { theToast } from '@/utils/toast'
import { storeToRefs } from 'pinia'
import { FILE_CATEGORY, FILE_STORAGE_TYPE } from '@/constants/file'
import { getPresign, globalFileRegistry, updateFileMeta } from '@/lib/file'
import { isSafari } from '@/utils/other'
import { getImageSize } from '@/utils/image'
import { useI18n } from 'vue-i18n'

/* --------------- 与Uppy相关 ---------------- */
import Uppy from '@uppy/core'
import Dashboard from '@uppy/dashboard'
import Compressor from '@uppy/compressor'
import XHRUpload from '@uppy/xhr-upload'
import AwsS3 from '@uppy/aws-s3'
import '@uppy/core/css/style.min.css'
import '@uppy/dashboard/css/style.min.css'
import zh_CN from '@uppy/locales/lib/zh_CN'
import en_US from '@uppy/locales/lib/en_US'
import de_DE from '@uppy/locales/lib/de_DE'

let uppy: Uppy | null = null

const props = defineProps<{
  fileStorageType: App.Api.File.StorageType
  EnableCompressor: boolean
  fileCategory?: App.Api.File.Category
  allowedFileTypes?: string[]
}>()
// const emit = defineEmits(['uppyUploaded'])

const memorySource = ref<string>(props.fileStorageType) // 用于记住上传方式
const isUploading = ref<boolean>(false) // 是否正在上传
const files = ref<App.Api.Ech0.FileToAdd[]>([]) // 已上传的文件列表
const tempFiles = ref<Map<string, { id: string; url: string; key: string }>>(new Map()) // 用于S3临时存储文件回显地址的 Map(key: uppyFileId, value: {id, url, key})
const pendingUploadTasks = ref<Set<Promise<void>>>(new Set()) // 跟踪 upload-success 异步任务，避免 complete 抢跑

const userStore = useUserStore()
const editorStore = useEditorStore()
const { isLogin } = storeToRefs(userStore)
const envURL = import.meta.env.VITE_SERVICE_BASE_URL as string
const backendURL = envURL.endsWith('/') ? envURL.slice(0, -1) : envURL
const { t, locale } = useI18n()

const outputMimeType = isSafari() ? 'image/jpeg' : 'image/webp'
const currentCategory = props.fileCategory || FILE_CATEGORY.IMAGE
const currentAllowedTypes = props.allowedFileTypes?.length ? props.allowedFileTypes : ['image/*']

function getUppyLocale(currentLocale: string) {
  const normalized = String(currentLocale || '').toLowerCase()
  if (normalized.startsWith('de')) return de_DE
  if (normalized.startsWith('en')) return en_US
  return zh_CN
}

function inferFileExtFromType(contentType: string): string {
  const normalized = String(contentType || '').toLowerCase()
  if (normalized.includes('png')) return '.png'
  if (normalized.includes('webp')) return '.webp'
  if (normalized.includes('gif')) return '.gif'
  if (normalized.includes('bmp')) return '.bmp'
  if (normalized.includes('avif')) return '.avif'
  if (normalized.includes('jpeg') || normalized.includes('jpg')) return '.jpg'
  return '.bin'
}

function tryParseJSON(input: unknown): Record<string, unknown> | undefined {
  if (!input) return undefined
  if (typeof input === 'string') {
    try {
      const parsed = JSON.parse(input) as unknown
      return typeof parsed === 'object' && parsed !== null
        ? (parsed as Record<string, unknown>)
        : undefined
    } catch {
      return undefined
    }
  }
  return typeof input === 'object' ? (input as Record<string, unknown>) : undefined
}

function extractUploadPayload(response: unknown): Record<string, unknown> {
  const resp = (response || {}) as Record<string, unknown>
  const body = tryParseJSON(resp.body)
  const nestedBody = tryParseJSON((resp.response as Record<string, unknown> | undefined)?.body)
  const responseText = tryParseJSON(resp.responseText)
  const candidates = [body, nestedBody, responseText].filter(Boolean) as Record<string, unknown>[]
  if (candidates.length === 0) return {}
  const first = candidates[0]
  const data = first.data
  if (data && typeof data === 'object') return data as Record<string, unknown>
  return first
}

function getUppyFileSize(file: unknown): number | undefined {
  const uppyFile = (file || {}) as Record<string, unknown>
  const directSize = uppyFile.size
  if (typeof directSize === 'number' && Number.isFinite(directSize) && directSize >= 0) {
    return directSize
  }
  const data = uppyFile.data
  if (data instanceof File || data instanceof Blob) {
    return data.size
  }
  return undefined
}

function getUppyFileContentType(file: unknown): string | undefined {
  const uppyFile = (file || {}) as Record<string, unknown>
  const directType = String(uppyFile.type || '').trim()
  if (directType) return directType
  const data = uppyFile.data
  if (data instanceof File || data instanceof Blob) {
    const blobType = String(data.type || '').trim()
    if (blobType) return blobType
  }
  return undefined
}

async function getImageDimensionsFromUppyFile(
  file: unknown,
): Promise<{ width?: number; height?: number }> {
  const uppyFile = (file || {}) as Record<string, unknown>
  const data = uppyFile.data
  if (!(data instanceof Blob)) return {}

  const blobUrl = URL.createObjectURL(data)
  try {
    const size = await getImageSize(blobUrl)
    return { width: size.width, height: size.height }
  } catch {
    return {}
  } finally {
    URL.revokeObjectURL(blobUrl)
  }
}

// ✨ 监听粘贴事件
const handlePaste = async (e: ClipboardEvent) => {
  if (!e.clipboardData) return

  for (const item of e.clipboardData.items) {
    if (item.type.startsWith('image/')) {
      const file = item.getAsFile()
      if (file) {
        // 保留原始文件的 lastModified，避免同一图片因时间戳不同被当作不同文件
        const pasteFile = new File([file], file.name, {
          type: file.type,
          lastModified: file.lastModified,
        })

        // 检查是否可能是重复文件：同一 Uppy 实例中是否存在 name、type、size 都相同的文件
        const existingFile = uppy?.getFiles().find(
          (f) =>
            f.name === pasteFile.name &&
            f.type === pasteFile.type &&
            f.size === pasteFile.size,
        )
        if (existingFile) {
          // 检测到重复文件，静默跳过，继续处理下一个
          continue
        }

        uppy?.addFile({
          id: `pasted-image-${Date.now()}-${Math.random().toString(36).slice(2, 9)}`,
          name: pasteFile.name,
          type: pasteFile.type,
          data: pasteFile,
          source: 'PastedImage',
        })
        // 注意：不再手动调用 upload()
        // 因为 Uppy 配置了 autoProceed: true，会自动开始上传
      }
    }
  }
}

// 初始化 Uppy 实例
const initUppy = () => {
  // 创建 Uppy 实例
  uppy = new Uppy({
    restrictions: {
      maxNumberOfFiles: 6,
      allowedFileTypes: currentAllowedTypes,
    },
    autoProceed: true,
  })

  // 使用 Dashboard 插件
  uppy.use(Dashboard, {
    inline: true,
    target: '#uppy-dashboard',
    hideProgressDetails: false,
    hideUploadButton: false,
    hideCancelButton: false,
    hideRetryButton: false,
    hidePauseResumeButton: false,
    proudlyDisplayPoweredByUppy: false,
    height: 200,
    locale: getUppyLocale(locale.value),
    note: String(t('uppy.dashboardNote')),
  })

  // 是否启用智能压缩
  if (props.EnableCompressor) {
    uppy.use(Compressor, {
      mimeType: outputMimeType,
      convertTypes: ['image/jpeg', 'image/png', 'image/webp'],
    })
  }

  // 根据 props.fileStorageType 动态切换上传插件
  if (memorySource.value == FILE_STORAGE_TYPE.LOCAL) {
    uppy.setMeta({
      category: currentCategory,
      storage_type: FILE_STORAGE_TYPE.LOCAL,
    })
    uppy.use(XHRUpload, {
      endpoint: `${backendURL}/api/files/upload`, // 本地上传接口
      fieldName: 'file',
      formData: true,
      headers: {
        Authorization: `${getAuthToken()}`,
      },
    })
  } else if (memorySource.value == FILE_STORAGE_TYPE.OBJECT) {
    uppy.use(AwsS3, {
      endpoint: '', // 走自定义的签名接口
      shouldUseMultipart: false, // 禁用分块上传
      // 每来一个文件都调用一次该函数，获取签名参数
      async getUploadParameters(file) {
        // console.log("Uploading to S3:", file)
        const contentType = String(file.type || 'application/octet-stream')
        const rawName = String(file.name || '').trim()
        const fileName = rawName || `upload_${Date.now()}${inferFileExtFromType(contentType)}`
        const data = await getPresign({
          fileName,
          contentType,
          storageType: FILE_STORAGE_TYPE.OBJECT,
        })
        const uppyFileId = String((file as unknown as Record<string, unknown>)?.id || '')
        if (uppyFileId) {
          tempFiles.value.set(uppyFileId, {
            id: String(data.id || ''),
            url: data.file_url,
            key: data.key || '',
          })
        }
        // 兜底保留一份按 file_name 的索引，兼容潜在的插件行为差异。
        tempFiles.value.set(data.file_name, {
          id: String(data.id || ''),
          url: data.file_url,
          key: data.key || '',
        })
        return {
          method: 'PUT',
          url: data.presign_url, // 预签名 URL
          headers: {
            // 必须跟签名时的 Content-Type 完全一致
            'Content-Type': file.type,
          },
          // PUT 上传没有 fields
          fields: {},
        }
      },
    })
  }

  // 监听粘贴事件
  document.addEventListener('paste', handlePaste)

  // 添加文件时
  uppy.on('files-added', () => {
    if (!isLogin.value) {
      theToast.error(String(t('uppy.loginRequired')))
      return
    }
    isUploading.value = true
    editorStore.fileUploading = true
  })
  // 上传开始前，检查是否登录
  uppy.on('upload', () => {
    if (!isLogin.value) {
      theToast.error(String(t('uppy.loginRequired')))
      return
    }
    theToast.info(String(t('uppy.uploadingImageWait')), { duration: 500 })
    isUploading.value = true
    editorStore.fileUploading = true
  })
  // 单个文件上传失败后，显示错误信息
  uppy.on('upload-error', (file, error, response) => {
    if (props.fileStorageType === FILE_STORAGE_TYPE.LOCAL) {
      type ResponseBody = {
        code: number
        msg: string
        // @ts-nocheck
        /* eslint-disable */
        data: any
      }

      let errorMsg = String(t('uppy.uploadImageError'))
      // @ts-nocheck
      /* eslint-disable */
      const resp = response as any // 忽略 TS 类型限制
      if (resp?.response) {
        let resObj: ResponseBody

        if (typeof resp.response === 'string') {
          resObj = JSON.parse(resp.response) as ResponseBody
        } else {
          resObj = resp.response as ResponseBody
        }

        if (resObj?.msg) {
          errorMsg = resObj.msg
        }
      }
      theToast.error(errorMsg)
    } else {
      const msg = String((error as Error | undefined)?.message || '').trim()
      if (msg) theToast.error(msg)
    }
    isUploading.value = false
    editorStore.fileUploading = false
  })
  // 单个文件上传成功后，保存文件 URL 到 files 列表
  uppy.on('upload-success', (file, response) => {
    const task = (async () => {
      theToast.success(String(t('uppy.uploadSuccess')))

      // 分两种情况: Local 或者 S3
      if (memorySource.value === FILE_STORAGE_TYPE.LOCAL) {
        const payload = extractUploadPayload(response) as App.Api.File.FileDto &
          Record<string, unknown>

        const fileId = String(payload.id || payload.file_id || payload.ID || '')
        const fileKey = String(payload.key || payload.object_key || '')
        const fileUrl = String(
          payload.url ||
            payload.file_url ||
            payload.access_url ||
            (response as Record<string, unknown>)?.uploadURL ||
            '',
        )
        const size =
          typeof payload.size === 'number' && Number.isFinite(payload.size)
            ? payload.size
            : getUppyFileSize(file)
        const width = typeof payload.width === 'number' ? payload.width : undefined
        const height = typeof payload.height === 'number' ? payload.height : undefined
        if (!fileId || !fileUrl) {
          theToast.error(String(t('uppy.missingFileIdentifier')))
          return
        }
        const item: App.Api.Ech0.FileToAdd = {
          id: fileId,
          url: fileUrl,
          storage_type: FILE_STORAGE_TYPE.LOCAL,
          key: fileKey,
          size,
          width: width,
          height: height,
        }
        files.value.push(item)
        globalFileRegistry.upsert({
          id: fileId,
          key: fileKey,
          url: fileUrl,
          storageType: FILE_STORAGE_TYPE.LOCAL,
          size,
          width,
          height,
        })
      } else if (memorySource.value === FILE_STORAGE_TYPE.OBJECT) {
        const uppyFileId = String((file as unknown as Record<string, unknown>)?.id || '')
        const uploadedFile =
          tempFiles.value.get(uppyFileId) || tempFiles.value.get(String(file?.name || '')) || ''
        if (!uploadedFile) return
        if (!uploadedFile.id) {
          theToast.error(String(t('uppy.missingFileId')))
          return
        }

        const rawSize = getUppyFileSize(file)
        const contentType = getUppyFileContentType(file)
        let width: number | undefined
        let height: number | undefined
        if (currentCategory === FILE_CATEGORY.IMAGE) {
          const dimensions = await getImageDimensionsFromUppyFile(file)
          width = dimensions.width
          height = dimensions.height
        }

        let resolvedSize = rawSize
        let resolvedWidth = width
        let resolvedHeight = height
        let resolvedContentType = contentType
        try {
          const updated = await updateFileMeta({
            id: uploadedFile.id,
            size: rawSize ?? 0,
            width,
            height,
            contentType,
          })
          resolvedSize = updated.size
          resolvedWidth = updated.width
          resolvedHeight = updated.height
          resolvedContentType = updated.contentType || contentType
        } catch (err) {
          const msg = err instanceof Error ? err.message : String(t('uppy.fileMetaFillbackFailed'))
          theToast.error(msg)
        }

        const item: App.Api.Ech0.FileToAdd = {
          id: uploadedFile.id,
          url: uploadedFile.url,
          storage_type: FILE_STORAGE_TYPE.OBJECT,
          key: uploadedFile.key,
          content_type: resolvedContentType,
          size: resolvedSize,
          width: resolvedWidth,
          height: resolvedHeight,
        }
        files.value.push(item)
        globalFileRegistry.upsert({
          id: uploadedFile.id,
          key: uploadedFile.key,
          url: uploadedFile.url,
          storageType: FILE_STORAGE_TYPE.OBJECT,
          contentType: resolvedContentType,
          size: resolvedSize,
          width: resolvedWidth,
          height: resolvedHeight,
        })
      }
    })()
    pendingUploadTasks.value.add(task)
    task.finally(() => {
      pendingUploadTasks.value.delete(task)
    })
  })
  // 全部文件上传完成后，发射事件到父组件
  uppy.on('complete', async (result) => {
    // 等待所有 upload-success 异步流程（含 meta 回填）完成，避免 complete 先执行。
    if (pendingUploadTasks.value.size > 0) {
      await Promise.allSettled(Array.from(pendingUploadTasks.value))
    }

    const filesToAddResult = [...files.value]
    if (result?.successful?.length && filesToAddResult.length === 0) {
      theToast.error(String(t('uppy.uploadSuccessNoFileId')))
    }
    // 保持“上传中”直到写回编辑器状态完成，避免用户立即发布导致 echo_files 为空。
    Promise.resolve(
      filesToAddResult.length > 0 ? editorStore.handleUppyUploaded(filesToAddResult) : undefined,
    ).finally(() => {
      isUploading.value = false
      editorStore.fileUploading = false
      files.value = []
      tempFiles.value.clear()
      pendingUploadTasks.value.clear()
    })
  })
}

// 监听 props.fileStorageType 变化
watch(
  () => locale.value,
  (newLocale, oldLocale) => {
    if (newLocale === oldLocale) return
    if (isUploading.value) return

    uppy?.destroy()
    uppy = null
    files.value = []
    tempFiles.value.clear()
    pendingUploadTasks.value.clear()
    initUppy()
  },
)

watch(
  () => props.fileStorageType,
  (newSource, oldSource) => {
    if (newSource !== oldSource) {
      if (!isUploading.value) {
        memorySource.value = newSource
        // 销毁旧的 Uppy 实例
        uppy?.destroy()
        uppy?.clear()
        files.value = [] // 清空已上传文件列表
        // 初始化新的 Uppy 实例
        initUppy()
      } else {
        theToast.error(String(t('uppy.uploadingTaskSwitchBlocked')))
      }
    }
  },
)

// 监听 props.EnableCompressor 变化
watch(
  () => props.EnableCompressor,
  (newVal, oldVal) => {
    if (newVal === oldVal) return
    if (isUploading.value) {
      theToast.error(String(t('uppy.uploadingSwitchCompressBlocked')))
      return
    }

    uppy?.destroy()
    uppy = null
    files.value = []
    tempFiles.value.clear()

    initUppy()
  },
)

onMounted(() => {
  initUppy()
})

onBeforeUnmount(() => {
  document.removeEventListener('paste', handlePaste)
})
</script>

<style scoped>
:deep(.uppy-Root) {
  border: 0;
  background-color: transparent;
}

:deep(.uppy-Dashboard-inner) {
  border: 1px solid var(--uppy-dropzone-border-color);
  border-radius: var(--radius-md);
  overflow: hidden;
  background-clip: padding-box;
}

:deep(.uppy-Dashboard-innerWrap) {
  background-color: var(--uppy-bg-color);
  background-clip: padding-box;
}

:deep(.uppy-Dashboard-AddFiles) {
  background-color: var(--uppy-dropzone-bg-color);
  border-color: var(--uppy-dropzone-border-color);
  box-shadow: var(--uppy-shadow-inset);
}

:deep(.uppy-Dashboard-AddFiles-title) {
  color: var(--uppy-dropzone-title-color);
}

:deep(.uppy-Dashboard-browse) {
  color: var(--uppy-link-color);
}

:deep(.uppy-Dashboard-browse:hover) {
  color: var(--uppy-link-hover-color);
}
:deep(.uppy-StatusBar) {
  color: var(--uppy-text-color);
  background-color: var(--uppy-panel-bg-color);
}

:deep(.uppy-DashboardContent-bar) {
  color: var(--uppy-text-color);
  background-color: var(--uppy-panel-bg-color);
}

:deep(.uppy-StatusBar-statusPrimary) {
  color: var(--uppy-text-color);
}

:deep(.uppy-DashboardContent-back) {
  color: var(--uppy-link-color);
}

:deep(.uppy-DashboardContent-addMore) {
  color: var(--uppy-link-color);
}

:deep(.uppy-DashboardContent-back:hover),
:deep(.uppy-DashboardContent-addMore:hover) {
  color: var(--uppy-link-hover-color);
}

:deep(.uppy-Dashboard-note),
:deep(.uppy-StatusBar-statusSecondary) {
  color: var(--uppy-muted-text-color);
}
</style>
