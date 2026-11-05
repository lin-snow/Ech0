#!/usr/bin/env bun
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow
//
// Draws the Open Graph cover as a deterministic SVG: warm paper, a soft
// gradient orb with capsule eyes, and a bold editorial left column. The orb is
// the friendly face of the brand; a faint periwinkle blush keeps it from
// reading as flat orange.
//
//   bun docs/assets/og-cover/generate.mjs
//   sips -s format png docs/assets/og-cover/og-cover.svg --out /tmp/og@2x.png
//   sips -z 630 1200 /tmp/og@2x.png --out /tmp/og.png
//   sips -s format jpeg -s formatOptions 88 /tmp/og.png --out web/public/og-cover.jpg

import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

const WIDTH = 1200
const HEIGHT = 630
const RENDER_SCALE = 2

const ORB = { x: 1072, y: 312, radius: 330 }
const EYES = [
  { x: 966, y: 212, width: 50, height: 110, rotation: -14 },
  { x: 1076, y: 194, width: 50, height: 110, rotation: -18 },
]

const PALETTE = {
  paper: '#FFFFFF',
  paperWarm: '#FFFBF4',
  ink: '#141210',
  body: '#6f6154',
  clay: '#a3947f',
  accent: '#F54A00',
  orbCore: '#FFE3C2',
  orbMid: '#FFB57C',
  orbDeep: '#FB7F45',
  orbEdge: '#F1571B',
  blush: '#9DB2FF',
}

const eye = ({ x, y, width, height, rotation }) => {
  const cx = x + width / 2
  const cy = y + height / 2
  return `<g transform="rotate(${rotation} ${cx} ${cy})"><rect x="${x}" y="${y}" width="${width}" height="${height}" rx="${width / 2}" fill="url(#eye)"/></g>`
}

const mark = `<g transform="translate(88 84)">
  <svg width="44" height="44" viewBox="100 68 228 260" xmlns="http://www.w3.org/2000/svg">
    <circle cx="132" cy="102" r="14" fill="${PALETTE.accent}"/>
    <path d="M261.448 114.512C257.029 114.513 253.447 110.931 253.448 106.512L253.45 98.7982C253.451 94.3806 257.033 90.799 261.451 90.798L301.505 90.7884C305.924 90.7873 309.505 94.3692 309.504 98.7882L309.502 106.502C309.501 110.919 305.919 114.501 301.502 114.502L261.448 114.512Z" fill="${PALETTE.ink}"/>
    <path d="M199.965 160.5L215 181.845L215 234.5L199.965 209L199.965 160.5Z" fill="${PALETTE.ink}"/>
    <rect x="150" y="291" width="124" height="20" fill="${PALETTE.ink}"/>
  </svg>
</g>`

const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="${WIDTH * RENDER_SCALE}" height="${HEIGHT * RENDER_SCALE}" viewBox="0 0 ${WIDTH} ${HEIGHT}">
<defs>
  <radialGradient id="bg" cx="0.18" cy="0.1" r="0.9">
    <stop offset="0" stop-color="${PALETTE.paper}"/>
    <stop offset="1" stop-color="${PALETTE.paperWarm}"/>
  </radialGradient>
  <radialGradient id="orb" cx="0.5" cy="0.5" r="0.5">
    <stop offset="0" stop-color="${PALETTE.orbCore}"/>
    <stop offset="0.40" stop-color="${PALETTE.orbMid}"/>
    <stop offset="0.68" stop-color="${PALETTE.orbDeep}"/>
    <stop offset="0.86" stop-color="${PALETTE.orbEdge}" stop-opacity="0.92"/>
    <stop offset="0.96" stop-color="${PALETTE.orbEdge}" stop-opacity="0.35"/>
    <stop offset="1" stop-color="${PALETTE.orbEdge}" stop-opacity="0"/>
  </radialGradient>
  <radialGradient id="blush" cx="0.78" cy="0.16" r="0.46">
    <stop offset="0" stop-color="${PALETTE.blush}" stop-opacity="0.55"/>
    <stop offset="0.55" stop-color="${PALETTE.blush}" stop-opacity="0.18"/>
    <stop offset="1" stop-color="${PALETTE.blush}" stop-opacity="0"/>
  </radialGradient>
  <radialGradient id="glow" cx="0.5" cy="0.5" r="0.5">
    <stop offset="0" stop-color="${PALETTE.orbMid}" stop-opacity="0.35"/>
    <stop offset="0.6" stop-color="${PALETTE.orbMid}" stop-opacity="0.12"/>
    <stop offset="1" stop-color="${PALETTE.orbMid}" stop-opacity="0"/>
  </radialGradient>
  <linearGradient id="eye" x1="0" y1="0" x2="0" y2="1">
    <stop offset="0" stop-color="#FFFFFF"/>
    <stop offset="1" stop-color="#FFF6EA"/>
  </linearGradient>
</defs>

<rect width="${WIDTH}" height="${HEIGHT}" fill="url(#bg)"/>
<circle cx="${ORB.x}" cy="${ORB.y}" r="${ORB.radius + 140}" fill="url(#glow)"/>
<circle cx="${ORB.x}" cy="${ORB.y}" r="${ORB.radius}" fill="url(#orb)"/>
<circle cx="${ORB.x}" cy="${ORB.y}" r="${ORB.radius}" fill="url(#blush)"/>

${EYES.map(eye).join('\n')}

${mark}
<text x="148" y="117" font-family="'Avenir Next','SF Pro Display',Helvetica,Arial,sans-serif" font-size="32" font-weight="700" letter-spacing="-0.6" fill="${PALETTE.ink}">Ech0</text>

<text x="88" y="322" font-family="'Avenir Next','SF Pro Display',Helvetica,Arial,sans-serif" font-size="70" font-weight="700" letter-spacing="-1.8" fill="${PALETTE.ink}">Your thoughts.</text>
<text x="88" y="396" font-family="'Avenir Next','SF Pro Display',Helvetica,Arial,sans-serif" font-size="70" font-weight="700" letter-spacing="-1.8" fill="${PALETTE.ink}">Your server<tspan fill="${PALETTE.accent}">.</tspan></text>
<text x="88" y="450" font-family="'Avenir Next','SF Pro Text',Helvetica,Arial,sans-serif" font-size="21" fill="${PALETTE.body}">Self-hosted microblog · single Go binary</text>
<text x="88" y="566" font-family="'Avenir Next','SF Pro Text',Helvetica,Arial,sans-serif" font-size="15" letter-spacing="2.4" fill="${PALETTE.clay}">ECH0.APP</text>
</svg>`

const target = join(dirname(fileURLToPath(import.meta.url)), 'og-cover.svg')
await Bun.write(target, svg)
console.log(`wrote ${target} (${(svg.length / 1024).toFixed(1)} KiB)`)
