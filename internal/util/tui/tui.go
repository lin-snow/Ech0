// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package tui

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"charm.land/lipgloss/v2"
)

var LightDark = sync.OnceValue(func() lipgloss.LightDarkFunc {
	return lipgloss.LightDark(lipgloss.HasDarkBackground(os.Stdin, os.Stdout))
})

var (
	infoStyle = sync.OnceValue(func() lipgloss.Style {
		return lipgloss.NewStyle().
			PaddingLeft(2).
			Foreground(LightDark()(lipgloss.Color("236"), lipgloss.Color("252")))
	})

	titleStyle = sync.OnceValue(func() lipgloss.Style {
		return lipgloss.NewStyle().
			Bold(true).
			Foreground(LightDark()(lipgloss.Color("#4338ca"), lipgloss.Color("#FF7F7F")))
	})

	highlight = sync.OnceValue(func() lipgloss.Style {
		return lipgloss.NewStyle().
			Bold(false).
			Italic(true).
			Foreground(LightDark()(lipgloss.Color("#7c3aed"), lipgloss.Color("#53b7f5ff")))
	})

	boxStyle = lipgloss.NewStyle().
			Bold(true).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#fb5151ff")).
			Padding(1, 1).
			Margin(1, 1)
)

const (
	banner = `
    ______     __    ____ 
   / ____/____/ /_  / __ \
  / __/ / ___/ __ \/ / / /
 / /___/ /__/ / / / /_/ / 
/_____/\___/_/ /_/\____/  
`
)

func GetLogoBanner() string {
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
	gradientBanner := lipgloss.JoinVertical(lipgloss.Left, rendered...)

	full := lipgloss.JoinVertical(lipgloss.Left,
		gradientBanner,
	)

	return full
}

func PrintCLIBanner() {
	banner := GetLogoBanner()

	if _, err := fmt.Fprintln(os.Stdout, banner); err != nil {
		fmt.Fprintf(os.Stderr, "failed to print banner: %v\n", err)
	}
}

func PrintCLIInfo(title, msg string) {
	if _, err := fmt.Fprintln(os.Stdout, infoStyle().Render(titleStyle().Render(title)+": "+highlight().Render(msg))); err != nil {
		fmt.Fprintf(os.Stderr, "failed to print cli info: %v\n", err)
	}
}

type CLIInfoItem struct {
	Title string
	Msg   string
}

type CLIBoxHeader struct {
	Icon  string
	Title string
	Value string
}

const boxLabelGap = 2

func GetCLIPrintWithBox(header CLIBoxHeader, items ...CLIInfoItem) string {
	lines := make([]string, 0, len(items)+2)

	if head := renderBoxHeader(header); head != "" {
		lines = append(lines, head)
		if len(items) > 0 {
			lines = append(lines, "")
		}
	}

	labelWidth := 0
	for _, item := range items {
		if w := lipgloss.Width(item.Title); w > labelWidth {
			labelWidth = w
		}
	}

	indent := ""
	if header.Title != "" || header.Value != "" {
		indent = "  "
	}

	for _, item := range items {
		lines = append(lines, renderBoxItem(item, labelWidth, indent)...)
	}

	if len(lines) == 0 {
		return ""
	}
	return boxStyle.Render(strings.Join(lines, "\n"))
}

func renderBoxHeader(h CLIBoxHeader) string {
	if strings.TrimSpace(h.Title) == "" && strings.TrimSpace(h.Value) == "" {
		return ""
	}

	var b strings.Builder
	if h.Icon != "" {
		b.WriteString(h.Icon)
		b.WriteString("  ")
	}
	b.WriteString(titleStyle().Render(h.Title))
	if h.Value != "" {
		b.WriteString("  ")
		b.WriteString(highlight().Render(h.Value))
	}
	return infoStyle().Render(b.String())
}

func renderBoxItem(item CLIInfoItem, labelWidth int, indent string) []string {
	parts := strings.Split(item.Msg, "\n")

	if item.Title == "" {
		out := make([]string, 0, len(parts))
		for _, line := range parts {
			out = append(out, infoStyle().Render(indent+highlight().Render(line)))
		}
		return out
	}

	label := indent + titleStyle().Render(item.Title) +
		strings.Repeat(" ", labelWidth-lipgloss.Width(item.Title)+boxLabelGap)
	out := []string{infoStyle().Render(label + highlight().Render(parts[0]))}

	continuation := indent + strings.Repeat(" ", labelWidth+boxLabelGap)
	for _, line := range parts[1:] {
		out = append(out, infoStyle().Render(continuation+highlight().Render(line)))
	}
	return out
}

func PrintCLIWithBox(header CLIBoxHeader, items ...CLIInfoItem) {
	if _, err := fmt.Fprintln(os.Stdout, GetCLIPrintWithBox(header, items...)); err != nil {
		fmt.Fprintf(os.Stderr, "failed to print cli box: %v\n", err)
	}
}

func ClearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to clear screen: %v\n", err)
	}
}
