// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

const banner = `
███████╗     ██████╗    ██╗  ██╗     ██████╗
██╔════╝    ██╔════╝    ██║  ██║    ██╔═████╗
█████╗      ██║         ███████║    ██║██╔██║
██╔══╝      ██║         ██╔══██║    ████╔╝██║
███████╗    ╚██████╗    ██║  ██║    ╚██████╔╝
╚══════╝     ╚═════╝    ╚═╝  ╚═╝     ╚═════╝

` as const

const gradientColors = [
  '#f38ba8',
  '#fab387',
  '#f9e2af',
  '#a6e3a1',
  '#94e2d5',
  '#89b4fa',
  '#cba6f7',
  '#f5c2e7',
  '#eba0ac',
] as const

function printWelcome(): void {
  const lines = banner.trim().split('\n')
  console.log()
  for (const [index, line] of lines.entries()) {
    const color = gradientColors[index % gradientColors.length]
    console.log(`%c${line}`, `color: ${color}`)
  }
  console.log()
}

printWelcome()

export { printWelcome }
