// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package version is the single source of truth for build / release metadata
// of the Ech0 binary: semantic version, git commit, build time, license,
// author, repository URL.
//
// Const values are source-controlled (bump in code on release).
// Var values are injected at build time via -ldflags "-X .../version.Commit=...".
// See Makefile / docker/build.Dockerfile / .github/workflows/release.yml.
package version

import (
	"fmt"
	"time"
)

const (
	// Version is the current semantic version. Bump on release.
	// release-please-action edits the literal below — do NOT remove the
	// trailing marker comment; see .release-please-config.json.
	Version = "4.7.0" // x-release-please-version

	// License is the SPDX identifier of the project license.
	License = "AGPL-3.0-or-later"

	// Author is the primary author / copyright holder.
	Author = "lin-snow"

	// RepoURL is the canonical source repository URL. AGPL-3.0 §13 requires
	// network users be able to obtain the corresponding source — this URL
	// (combined with Commit) is what the About page surfaces to satisfy that.
	RepoURL = "https://github.com/lin-snow/Ech0"

	// StartYear is the project inception year, used to render copyright ranges.
	// 首个公开 commit 是 2025-03-21；以此为版权起始年。
	StartYear = 2025
)

// Commit is the short git commit hash, injected at build time.
// Defaults to "unknown" so `go run` / unbranded local builds still compile.
var Commit = "unknown"

// BuildTime is the RFC3339 build timestamp, injected at build time.
// Empty when not injected.
var BuildTime = ""

// Info bundles all build / release metadata. Useful for handlers that want
// to return a single struct instead of hand-spelling every field.
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"build_time"`
	License   string `json:"license"`
	Author    string `json:"author"`
	RepoURL   string `json:"repo_url"`
}

// Get returns a snapshot of the current build / release metadata.
func Get() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		BuildTime: BuildTime,
		License:   License,
		Author:    Author,
		RepoURL:   RepoURL,
	}
}

// Copyright returns the human-readable copyright line, e.g.
// "Copyright (C) 2025-2026 lin-snow".
// The end year is the current UTC year, or StartYear if the clock is broken.
func Copyright() string {
	end := time.Now().UTC().Year()
	if end <= StartYear {
		return fmt.Sprintf("Copyright (C) %d %s", StartYear, Author)
	}
	return fmt.Sprintf("Copyright (C) %d-%d %s", StartYear, end, Author)
}
