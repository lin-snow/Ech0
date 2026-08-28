// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package version

import (
	"fmt"
	"time"
)

const (
	Version = "5.7.0"

	// License is the SPDX identifier of the project license.
	License = "AGPL-3.0-or-later"

	Author = "L1nSn0w"

	RepoURL = "https://github.com/lin-snow/Ech0"

	StartYear = 2025
)

var Commit = "unknown"

var BuildTime = ""

type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"build_time"`
	License   string `json:"license"`
	Author    string `json:"author"`
	RepoURL   string `json:"repo_url"`
}

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
func Copyright() string {
	end := time.Now().UTC().Year()
	if end <= StartYear {
		return fmt.Sprintf("Copyright (C) %d %s", StartYear, Author)
	}
	return fmt.Sprintf("Copyright (C) %d-%d %s", StartYear, end, Author)
}
