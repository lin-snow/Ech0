import { defineStore } from 'pinia'
import { ref } from 'vue'

interface ZonePrintPayload {
  key: string
  text: string
}

const MAX_PRINT_LENGTH = 2000
const EXTENSION_LABEL_MAP: Record<string, string> = {
  GITHUBPROJ: 'GitHub',
  WEBSITE: 'Website',
  VIDEO: 'Video',
  MUSIC: 'Music',
}

const buildPrintableEchoText = (echo: App.Api.Ech0.Echo): string => {
  const content = echo.content?.trim() || ''
  const imageCount = Array.isArray(echo.images) ? echo.images.length : 0
  const hasExtension = Boolean(echo.extension)
  const hasMedia = imageCount > 0 || hasExtension

  if (!hasMedia) {
    return content
  }

  const lines: string[] = []
  lines.push(content)
  lines.push('')
  lines.push('---')
  lines.push('[METADATA]')
  lines.push(`EchoID: ${echo.id}`)

  const firstTag = echo.tags?.[0]?.name
  if (firstTag) {
    lines.push(`Tag: #${firstTag}`)
  }

  if (imageCount > 0) {
    lines.push(`Images: ${imageCount} file(s)`)
  }

  if (hasExtension) {
    const rawType = String(echo.extension_type || 'Unknown').toUpperCase()
    const extensionLabel = EXTENSION_LABEL_MAP[rawType] || String(echo.extension_type || 'Unknown')
    lines.push(`Extension: ${extensionLabel}`)
  }

  return lines.join('\n')
}

export const useZoneStore = defineStore('zoneStore', () => {
  const pendingPrint = ref<ZonePrintPayload | null>(null)

  const setPendingPrintEcho = (echo: App.Api.Ech0.Echo) => {
    const text = buildPrintableEchoText(echo)
    const normalized = text.trim()
    if (!normalized) return

    pendingPrint.value = {
      key: `echo:${echo.id}:${Date.now()}`,
      text: normalized.slice(0, MAX_PRINT_LENGTH),
    }
  }

  const consumePendingPrint = (): ZonePrintPayload | null => {
    const payload = pendingPrint.value
    pendingPrint.value = null
    return payload
  }

  const init = () => {}

  return {
    pendingPrint,
    setPendingPrintEcho,
    consumePendingPrint,
    init,
  }
})
