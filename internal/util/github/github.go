// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package util

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/mod/semver"
)

func isHelmChartArtifactStyleTag(tag string) bool {
	t := strings.TrimSpace(tag)
	return len(t) >= 5 && strings.EqualFold(t[:5], "ech0-")
}

func canonicalStableSemverFromReleaseTag(tag string) string {
	t := strings.TrimSpace(tag)
	if t == "" {
		return ""
	}
	if !strings.HasPrefix(t, "v") {
		t = "v" + t
	}
	t = semver.Canonical(t)
	if t == "" {
		return ""
	}
	if semver.Prerelease(t) != "" {
		return ""
	}
	return t
}

var latestVersionCache struct {
	mu        sync.Mutex
	version   string
	expiresAt time.Time
}

const (
	listReleasesMaxPages = 10
	releasesPerPage      = 30
	githubAPIBase        = "https://api.github.com"
)

type githubRelease struct {
	TagName    string `json:"tag_name"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
}

func fetchReleasesPage(ctx context.Context, owner, repo string, page int) ([]githubRelease, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases?per_page=%d&page=%d", githubAPIBase, owner, repo, releasesPerPage, page)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("github releases request failed: status %d", resp.StatusCode)
	}

	var releases []githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}
	return releases, nil
}

func GetLatestVersion() (string, error) {
	now := time.Now().UTC()
	latestVersionCache.mu.Lock()
	if latestVersionCache.version != "" && now.Before(latestVersionCache.expiresAt) {
		v := latestVersionCache.version
		latestVersionCache.mu.Unlock()
		return v, nil
	}
	latestVersionCache.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	owner, repo := "lin-snow", "Ech0"

	var best string
pageLoop:
	for page := 1; page <= listReleasesMaxPages; page++ {
		releases, err := fetchReleasesPage(ctx, owner, repo, page)
		if err != nil {
			return "", fmt.Errorf("list releases failed: %w", err)
		}
		for _, rel := range releases {
			if rel.Draft {
				continue
			}
			tag := strings.TrimSpace(rel.TagName)
			if tag == "" {
				continue
			}
			if isHelmChartArtifactStyleTag(tag) {
				continue
			}
			if rel.Prerelease {
				continue
			}
			canon := canonicalStableSemverFromReleaseTag(tag)
			if canon == "" {
				continue
			}
			best = canon
			break pageLoop
		}
		if len(releases) < releasesPerPage {
			break
		}
	}

	if best == "" {
		return "", fmt.Errorf("no stable application release found (ech0-* chart tags are ignored)")
	}

	result := strings.TrimPrefix(best, "v")

	latestVersionCache.mu.Lock()
	latestVersionCache.version = result
	latestVersionCache.expiresAt = time.Now().UTC().Add(30 * time.Minute)
	latestVersionCache.mu.Unlock()

	return result, nil
}
