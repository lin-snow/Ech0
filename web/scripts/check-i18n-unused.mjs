import { readdirSync, readFileSync, statSync } from 'node:fs'
import { join, extname } from 'node:path'

const root = process.cwd()
const sourceDir = join(root, 'src')
const zhPath = join(root, 'src/locales/messages/zh-CN.json')

const SOURCE_EXTENSIONS = new Set(['.vue', '.ts', '.tsx', '.js', '.jsx', '.mjs', '.cjs'])
const SKIP_DIRS = new Set(['node_modules', 'dist', 'coverage', '.git'])
const DYNAMIC_KEY_FIELDS = ['labelKey', 'titleKey', 'textKey', 'i18nKey', 'tooltipKey']
const ALLOW_UNUSED_PREFIXES = [
]
const ALLOW_UNUSED_KEYS = new Set([
])

const flatten = (obj, prefix = '', result = new Set()) => {
  for (const [key, value] of Object.entries(obj)) {
    const next = prefix ? `${prefix}.${key}` : key
    if (value && typeof value === 'object' && !Array.isArray(value)) {
      flatten(value, next, result)
    } else {
      result.add(next)
    }
  }
  return result
}

const walkFiles = (dir, files = []) => {
  for (const entry of readdirSync(dir)) {
    if (SKIP_DIRS.has(entry)) continue
    const abs = join(dir, entry)
    const stat = statSync(abs)
    if (stat.isDirectory()) {
      walkFiles(abs, files)
      continue
    }
    if (SOURCE_EXTENSIONS.has(extname(entry))) {
      files.push(abs)
    }
  }
  return files
}

const collectUsedKeys = (content, keyUniverse) => {
  const used = new Set()

  const tCallRegex = /(?:\b\$?t|\.t)\(\s*(['"`])([^'"`]+)\1/g
  for (const match of content.matchAll(tCallRegex)) {
    const key = match[2]
    if (keyUniverse.has(key)) used.add(key)
  }

  const fieldRegex = new RegExp(
    `\\b(?:${DYNAMIC_KEY_FIELDS.join('|')})\\b\\s*:\\s*(['"\`])([^'"\\\`]+)\\1`,
    'g',
  )
  for (const match of content.matchAll(fieldRegex)) {
    const key = match[2]
    if (keyUniverse.has(key)) used.add(key)
  }

  return used
}

const keyAllowed = (key) => {
  if (ALLOW_UNUSED_KEYS.has(key)) return true
  return ALLOW_UNUSED_PREFIXES.some((prefix) => key.startsWith(prefix))
}

const zh = JSON.parse(readFileSync(zhPath, 'utf8'))
const allKeys = flatten(zh)
const files = walkFiles(sourceDir)
const usedKeys = new Set()

for (const file of files) {
  const content = readFileSync(file, 'utf8')
  const found = collectUsedKeys(content, allKeys)
  for (const key of found) usedKeys.add(key)
}

const unused = [...allKeys].filter((key) => !usedKeys.has(key) && !keyAllowed(key)).sort()

if (unused.length > 0) {
  console.error(`Unused i18n keys found: ${unused.length}`)
  unused.forEach((key) => console.error(`- ${key}`))
  process.exit(1)
}

console.log('No unused i18n keys found.')
