// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package server

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"charm.land/lipgloss/v2"
	tuiUtil "github.com/lin-snow/ech0/internal/util/tui"
	versionPkg "github.com/lin-snow/ech0/internal/version"
)

const (
	GreetingBanner = `
███████╗     ██████╗    ██╗  ██╗     ██████╗ 
██╔════╝    ██╔════╝    ██║  ██║    ██╔═████╗
█████╗      ██║         ███████║    ██║██╔██║
██╔══╝      ██║         ██╔══██║    ████╔╝██║
███████╗    ╚██████╗    ██║  ██║    ╚██████╔╝
╚══════╝     ╚═════╝    ╚═╝  ╚═╝     ╚═════╝ 
                                             
`
)

var (
	infoStyle = sync.OnceValue(func() lipgloss.Style {
		return lipgloss.NewStyle().
			PaddingLeft(2).
			Foreground(tuiUtil.LightDark()(lipgloss.Color("236"), lipgloss.Color("252")))
	})

	titleStyle = sync.OnceValue(func() lipgloss.Style {
		return lipgloss.NewStyle().
			Bold(true).
			Foreground(tuiUtil.LightDark()(lipgloss.Color("#4338ca"), lipgloss.Color("#f7b457ff")))
	})

	highlight = sync.OnceValue(func() lipgloss.Style {
		return lipgloss.NewStyle().
			Bold(false).
			Italic(true).
			Foreground(tuiUtil.LightDark()(lipgloss.Color("#7c3aed"), lipgloss.Color("#53b7f5ff")))
	})

	boxStyle = lipgloss.NewStyle().
			Bold(true).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#fb5151ff")).
			Padding(1, 1).
			Margin(1, 1)
)

func PrintGreetings(port string) {
	banner := gradientBanner(GreetingBanner)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		infoStyle().Render(
			"📦 "+titleStyle().Render("Version")+": "+highlight().Render(versionPkg.Version),
		),
		infoStyle().Render("🎈 "+titleStyle().Render("Port")+": "+highlight().Render(port)),
		infoStyle().Render("🧙 "+titleStyle().Render("Author")+": "+highlight().Render("L1nSn0w")),
		infoStyle().Render(
			"👉 "+titleStyle().Render("Website")+": "+highlight().Render("https://ech0.app/"),
		),
		infoStyle().Render(
			"👉 "+titleStyle().Render(
				"GitHub",
			)+": "+highlight().Render(
				"https://github.com/lin-snow/Ech0",
			),
		),
	)

	full := lipgloss.JoinVertical(lipgloss.Left,
		banner,
		boxStyle.Render(content),
	)

	if _, err := fmt.Fprintln(os.Stdout, full); err != nil {
		fmt.Fprintf(os.Stderr, "failed to print greetings: %v\n", err)
	}
}

func gradientBanner(banner string) string {
	lines := strings.Split(banner, "\n")
	var rendered []string

	colors := []string{
		"#FF7F7F",
		"#FFB347",
		"#FFEB9C",
		"#B8E6B8",
		"#87CEEB",
		"#DDA0DD",
		"#F0E68C",
	}

	for i, line := range lines {
		color := lipgloss.Color(colors[i%len(colors)])
		style := lipgloss.NewStyle().Foreground(color)
		rendered = append(rendered, style.Render(line))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rendered...)
}
