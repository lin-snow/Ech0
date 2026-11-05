#!/usr/bin/env bun
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow
//
// Draws the Open Graph cover as a deterministic SVG. The right half is the Ech0
// mark turned into a character: its two eyes — the orange dot and the ink
// capsule — inside a cream-cored, orange-rimmed orb.
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

const ORB = { x: 1040, y: 315, radius: 300, coreX: 1000, coreY: 310 }
const EYES = {
  dot: { cx: 952, cy: 300, r: 27 },
  capsule: { x: 1036, y: 275, width: 104, height: 50, radius: 25, rotation: -14 },
}

const PALETTE = {
  paper: '#FFFFFF',
  paperWarm: '#FFFBF4',
  ink: '#141210',
  body: '#6f6154',
  clay: '#a3947f',
  accent: '#F54A00',
  core: '#FFFDF7',
  coreWarm: '#FFF4E2',
  rimLight: '#FFE0B9',
  rimMid: '#FCAE77',
  rimDeep: '#F26A30',
}

const capsuleCenter = {
  x: EYES.capsule.x + EYES.capsule.width / 2,
  y: EYES.capsule.y + EYES.capsule.height / 2,
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
  <radialGradient id="orb" gradientUnits="userSpaceOnUse" cx="${ORB.coreX}" cy="${ORB.coreY}" r="${ORB.radius}">
    <stop offset="0" stop-color="${PALETTE.core}"/>
    <stop offset="0.50" stop-color="${PALETTE.coreWarm}"/>
    <stop offset="0.72" stop-color="${PALETTE.rimLight}"/>
    <stop offset="0.86" stop-color="${PALETTE.rimMid}"/>
    <stop offset="0.95" stop-color="${PALETTE.rimDeep}"/>
    <stop offset="0.985" stop-color="${PALETTE.rimDeep}" stop-opacity="0.35"/>
    <stop offset="1" stop-color="${PALETTE.rimDeep}" stop-opacity="0"/>
  </radialGradient>
  <radialGradient id="glow" cx="0.5" cy="0.5" r="0.5">
    <stop offset="0" stop-color="${PALETTE.rimMid}" stop-opacity="0.30"/>
    <stop offset="0.6" stop-color="${PALETTE.rimMid}" stop-opacity="0.10"/>
    <stop offset="1" stop-color="${PALETTE.rimMid}" stop-opacity="0"/>
  </radialGradient>
</defs>

<rect width="${WIDTH}" height="${HEIGHT}" fill="url(#bg)"/>
<circle cx="${ORB.x}" cy="${ORB.y}" r="${ORB.radius + 140}" fill="url(#glow)"/>
<circle cx="${ORB.x}" cy="${ORB.y}" r="${ORB.radius}" fill="url(#orb)"/>

<circle cx="${EYES.dot.cx}" cy="${EYES.dot.cy}" r="${EYES.dot.r}" fill="${PALETTE.accent}"/>
<g transform="rotate(${EYES.capsule.rotation} ${capsuleCenter.x} ${capsuleCenter.y})">
  <rect x="${EYES.capsule.x}" y="${EYES.capsule.y}" width="${EYES.capsule.width}" height="${EYES.capsule.height}" rx="${EYES.capsule.radius}" fill="${PALETTE.ink}"/>
</g>

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
