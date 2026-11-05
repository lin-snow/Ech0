#!/usr/bin/env bun
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow
//
// Draws the Open Graph cover as deterministic SVG: a warm paper field with a
// square-pixel halftone that ripples out from the mark. Rasterise with sips:
//
//   bun docs/assets/og-cover/generate.mjs
//   sips -s format png docs/assets/og-cover/og-cover.svg --out /tmp/og@2x.png
//   sips -z 630 1200 /tmp/og@2x.png --out /tmp/og.png
//   sips -s format jpeg -s formatOptions 86 /tmp/og.png --out web/public/og-cover.jpg

import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

const WIDTH = 1200
const HEIGHT = 630
const RENDER_SCALE = 2
const ORIGIN = { x: 892, y: 314 }
const GRID = 16

const PALETTE = {
  paper: '#FFFBF3',
  paperMid: '#F8F2E7',
  paperDeep: '#F2E8D8',
  shade: '#E8DAC3',
  ink: '#1c1713',
  umber: '#6f4a2f',
  clay: '#9a7a5e',
  accent: '#F54A00',
}

const number = (value) => Math.round(value * 100) / 100

const smoothstep = (edge0, edge1, x) => {
  const t = Math.min(1, Math.max(0, (x - edge0) / (edge1 - edge0)))
  return t * t * (3 - 2 * t)
}

// The halftone is one field, not a texture image: square size and opacity are
// driven by the distance to the mark, so the rings read as an outgoing echo.
const buildHalftone = () => {
  const squares = []
  const accents = []

  for (let y = GRID / 2; y <= HEIGHT; y += GRID) {
    for (let x = GRID / 2; x <= WIDTH; x += GRID) {
      const distance = Math.hypot(x - ORIGIN.x, y - ORIGIN.y)
      const wave = Math.pow(0.5 + 0.5 * Math.sin(distance / 18 - 0.4), 1.5)
      const falloff = 1 - smoothstep(0, 560, distance)
      const leftMargin = smoothstep(540, 700, x)
      const clearing = smoothstep(92, 168, distance)
      const value = wave * falloff * leftMargin * clearing

      if (value < 0.03) continue

      const size = number(2.5 + value * 10.5)
      const half = size / 2
      const x0 = number(x - half)
      const y0 = number(y - half)
      const opacity = number(0.04 + value * 0.28)

      squares.push(
        `<rect x="${x0}" y="${y0}" width="${size}" height="${size}" fill="${PALETTE.umber}" fill-opacity="${opacity}"/>`,
      )

      const scattered = ((x / GRID) * 3 + y / GRID) % 17 < 1
      if (scattered && value > 0.42) {
        const accentSize = number(size * 0.6)
        const accentHalf = accentSize / 2
        accents.push(
          `<rect x="${number(x - accentHalf)}" y="${number(y - accentHalf)}" width="${accentSize}" height="${accentSize}" fill="${PALETTE.accent}" fill-opacity="${number(value * 0.42)}"/>`,
        )
      }
    }
  }

  return `<g>${squares.join('')}</g><g>${accents.join('')}</g>`
}

// The product mark, cropped to its ink bounds so it reads at small sizes.
const buildMark = (height) => {
  const scale = height / 260
  const x = ORIGIN.x - (228 * scale) / 2
  const y = ORIGIN.y - (260 * scale) / 2

  return `<g transform="translate(${number(x)} ${number(y)}) scale(${number(scale)}) translate(-100 -68)">
    <circle cx="132" cy="102" r="14" fill="${PALETTE.accent}"/>
    <path d="M261.448 114.512C257.029 114.513 253.447 110.931 253.448 106.512L253.45 98.7982C253.451 94.3806 257.033 90.799 261.451 90.798L301.505 90.7884C305.924 90.7873 309.505 94.3692 309.504 98.7882L309.502 106.502C309.501 110.919 305.919 114.501 301.502 114.502L261.448 114.512Z" fill="${PALETTE.ink}"/>
    <path d="M199.965 160.5L215 181.845L215 234.5L199.965 209L199.965 160.5Z" fill="${PALETTE.ink}"/>
    <rect x="150" y="291" width="124" height="20" fill="${PALETTE.ink}"/>
  </g>`
}

const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="${WIDTH * RENDER_SCALE}" height="${HEIGHT * RENDER_SCALE}" viewBox="0 0 ${WIDTH} ${HEIGHT}">
<defs>
  <linearGradient id="paper" x1="0" y1="0" x2="1" y2="1">
    <stop offset="0" stop-color="${PALETTE.paper}"/>
    <stop offset="0.52" stop-color="${PALETTE.paperMid}"/>
    <stop offset="1" stop-color="${PALETTE.paperDeep}"/>
  </linearGradient>
  <radialGradient id="light" cx="0.16" cy="0.08" r="0.72">
    <stop offset="0" stop-color="#FFFDF8" stop-opacity="0.95"/>
    <stop offset="1" stop-color="#FFFDF8" stop-opacity="0"/>
  </radialGradient>
  <radialGradient id="shade" cx="0.9" cy="0.95" r="0.6">
    <stop offset="0" stop-color="${PALETTE.shade}" stop-opacity="0.55"/>
    <stop offset="1" stop-color="${PALETTE.shade}" stop-opacity="0"/>
  </radialGradient>
  <radialGradient id="clearing" cx="0.5" cy="0.5" r="0.5">
    <stop offset="0" stop-color="#FCF7EE" stop-opacity="0.96"/>
    <stop offset="0.62" stop-color="#FCF7EE" stop-opacity="0.72"/>
    <stop offset="1" stop-color="#FCF7EE" stop-opacity="0"/>
  </radialGradient>
</defs>

<rect width="${WIDTH}" height="${HEIGHT}" fill="url(#paper)"/>
<rect width="${WIDTH}" height="${HEIGHT}" fill="url(#light)"/>
<rect width="${WIDTH}" height="${HEIGHT}" fill="url(#shade)"/>
${buildHalftone()}
<circle cx="${ORIGIN.x}" cy="${ORIGIN.y}" r="190" fill="url(#clearing)"/>
${buildMark(104)}

<text x="88" y="150" font-family="'Avenir Next','SF Pro Text',Helvetica,Arial,sans-serif" font-size="12" letter-spacing="4.2" fill="${PALETTE.clay}">SELF-HOSTED MICROBLOG</text>
<text x="84" y="318" font-family="'Iowan Old Style','Palatino Linotype',Palatino,Georgia,serif" font-size="138" letter-spacing="2" fill="#5f3f27">Ech0</text>
<rect x="88" y="352" width="382" height="1" fill="${PALETTE.umber}" fill-opacity="0.25"/>
<text x="88" y="400" font-family="'Avenir Next','SF Pro Text',Helvetica,Arial,sans-serif" font-size="24" fill="#7a5a3f">A quiet home for your thoughts,</text>
<text x="88" y="434" font-family="'Avenir Next','SF Pro Text',Helvetica,Arial,sans-serif" font-size="24" fill="#7a5a3f">notes and links.</text>
<text x="88" y="584" font-family="'Avenir Next','SF Pro Text',Helvetica,Arial,sans-serif" font-size="12" letter-spacing="3.4" fill="#a98a6c">ECH0.APP</text>
</svg>`

const target = join(dirname(fileURLToPath(import.meta.url)), 'og-cover.svg')
await Bun.write(target, svg)
console.log(`wrote ${target} (${(svg.length / 1024).toFixed(1)} KiB)`)
