#!/usr/bin/env node
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow
// Idempotent batch tool: prepend SPDX-License-Identifier + Copyright headers to
// already containing "SPDX-License-Identifier" are skipped.

import { readdir, readFile, writeFile, stat } from 'node:fs/promises'
import { join, relative, dirname } from 'node:path'
import { fileURLToPath } from 'node:url'

const REPO_ROOT = join(dirname(fileURLToPath(import.meta.url)), '..')

const SPDX_ID = 'SPDX-License-Identifier: AGPL-3.0-or-later'
const COPYRIGHT = 'Copyright (C) 2025-2026 lin-snow'

const SKIP_DIRS = new Set([
  '.git',
  '.github',
  '.claude',
  '.pnpm-store',
  '.worktrees',
  '.vscode',
  '.idea',
  '.cache',
  'node_modules',
  'data',
  'backup',
  'tmp',
  'bin',
  'release',
  'dist',
  '.react-router',
])

const SKIP_FILES = new Set([
  'internal/di/wire_gen.go',
  'web/src/components/icons',
])

const SKIP_PATH_FRAGMENTS = [
  'template/dist/',
  'web/dist/',
  'web/src/components/icons/',
]

const EXTS = new Set(['.go', '.ts', '.vue'])

const args = process.argv.slice(2)
const MODE = args.includes('--check') ? 'check' : args.includes('--dry-run') ? 'dry' : 'write'

async function* walk(dir) {
  const entries = await readdir(dir, { withFileTypes: true })
  for (const e of entries) {
    const full = join(dir, e.name)
    if (e.isDirectory()) {
      if (SKIP_DIRS.has(e.name)) continue
      yield* walk(full)
    } else if (e.isFile()) {
      const rel = relative(REPO_ROOT, full)
      if (SKIP_PATH_FRAGMENTS.some((f) => rel.includes(f))) continue
      if (SKIP_FILES.has(rel)) continue
      const dot = full.lastIndexOf('.')
      if (dot === -1) continue
      const ext = full.slice(dot)
      if (!EXTS.has(ext)) continue
      yield { full, rel, ext }
    }
  }
}

function insertGo(src) {
  // Respect //go:build constraints — they must remain the first non-blank line.
  const lines = src.split('\n')
  let insertAt = 0
  if (lines[0]?.startsWith('//go:build') || lines[0]?.startsWith('// +build')) {
    let i = 0
    while (i < lines.length && lines[i].trim() !== '') i++
    insertAt = i + 1
  }
  const header = ['// ' + SPDX_ID, '// ' + COPYRIGHT, '']
  return [...lines.slice(0, insertAt), ...header, ...lines.slice(insertAt)].join('\n')
}

function insertTs(src) {
  if (src.startsWith('#!')) {
    const nl = src.indexOf('\n')
    const head = src.slice(0, nl + 1)
    const tail = src.slice(nl + 1)
    return head + `// ${SPDX_ID}\n// ${COPYRIGHT}\n\n` + tail
  }
  return `// ${SPDX_ID}\n// ${COPYRIGHT}\n\n` + src
}

function insertVue(src) {
  return `<!-- ${SPDX_ID} -->\n<!-- ${COPYRIGHT} -->\n` + src
}

async function processFile({ full, rel, ext }) {
  const src = await readFile(full, 'utf8')
  if (src.includes('SPDX-License-Identifier')) return { rel, status: 'skip' }
  let next
  if (ext === '.go') next = insertGo(src)
  else if (ext === '.ts') next = insertTs(src)
  else if (ext === '.vue') next = insertVue(src)
  else return { rel, status: 'skip' }
  if (MODE === 'check') return { rel, status: 'missing' }
  if (MODE === 'dry') return { rel, status: 'would-write' }
  await writeFile(full, next, 'utf8')
  return { rel, status: 'wrote' }
}

async function main() {
  const stats = { wrote: 0, skip: 0, missing: 0, 'would-write': 0 }
  const missing = []
  for await (const f of walk(REPO_ROOT)) {
    const r = await processFile(f)
    stats[r.status] = (stats[r.status] || 0) + 1
    if (r.status === 'missing') missing.push(r.rel)
  }
  if (MODE === 'check') {
    console.log(`Files already covered: ${stats.skip}`)
    console.log(`Files missing header: ${stats.missing}`)
    if (stats.missing > 0) {
      console.log('\nMissing:')
      for (const m of missing.slice(0, 50)) console.log('  ' + m)
      if (missing.length > 50) console.log(`  ... and ${missing.length - 50} more`)
      process.exit(1)
    }
    return
  }
  if (MODE === 'dry') {
    console.log(`Would write headers to ${stats['would-write']} files; ${stats.skip} already covered.`)
    return
  }
  console.log(`Wrote headers to ${stats.wrote} files; ${stats.skip} already covered.`)
}

main().catch((err) => {
  console.error(err)
  process.exit(2)
})
