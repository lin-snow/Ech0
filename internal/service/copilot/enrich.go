// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"encoding/base64"
	"io"
	"strings"

	"github.com/lin-snow/ech0/internal/agent"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	embeddingModel "github.com/lin-snow/ech0/internal/model/embedding"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
)

const maxChatImages = 4

// maxSourceImageBytes is the largest stored image worth reading at all; it is
// downscaled well below this before it is sent (see fitForVision).
const maxSourceImageBytes = 20 << 20

func (s *CopilotService) enrichHits(
	ctx context.Context,
	results []embeddingModel.SearchResult,
	multimodal bool,
	budget *imageBudget,
) (map[string]string, []agent.ImagePart) {
	exts := make(map[string]string, len(results))
	var images []agent.ImagePart
	for i := range results {
		echo, err := s.echoService.GetEchoById(ctx, results[i].EchoID)
		if err != nil || echo == nil {
			continue
		}
		if txt := formatExtension(echo.Extension); txt != "" {
			exts[results[i].EchoID] = txt
		}
		results[i].Extension = echo.Extension

		var files []fileModel.File
		for _, ef := range echo.EchoFiles {
			cat := fileModel.CategoryOf(ef.File.ContentType)
			switch cat {
			case fileModel.CategoryImage, fileModel.CategoryVideo, fileModel.CategoryAudio:
				files = append(files, ef.File)
			}
			if cat == fileModel.CategoryImage && multimodal && len(images) < maxChatImages {
				if part, ok := s.loadImagePart(ctx, ef.File); ok && budget.take(len(part.Base64)+len(part.URL)) {
					images = append(images, part)
				}
			}
		}
		results[i].Files = files
	}
	return exts, images
}

func formatExtension(ext *echoModel.EchoExtension) string {
	if ext == nil {
		return ""
	}
	str := func(k string) string {
		if v, ok := ext.Payload[k].(string); ok {
			return strings.TrimSpace(v)
		}
		return ""
	}
	switch ext.Type {
	case echoModel.Extension_MUSIC:
		if u := str("url"); u != "" {
			return "[音乐分享] " + u
		}
	case echoModel.Extension_VIDEO:
		if id := str("videoId"); id != "" {
			return "[视频分享] 视频ID " + id
		}
	case echoModel.Extension_GITHUBPROJ:
		if u := str("repoUrl"); u != "" {
			return "[GitHub 项目] " + u
		}
	case echoModel.Extension_WEBSITE:
		title, site := str("title"), str("site")
		switch {
		case title != "" && site != "":
			return "[网站] " + title + " " + site
		case site != "":
			return "[网站] " + site
		case title != "":
			return "[网站] " + title
		}
	case echoModel.Extension_LOCATION:
		if place := str("placeholder"); place != "" {
			return "[位置] " + place
		}
	case echoModel.Extension_TWEET:
		u, user := str("url"), str("username")
		switch {
		case u != "" && user != "":
			return "[X 推文] @" + user + " " + u
		case u != "":
			return "[X 推文] " + u
		}
	}
	return ""
}

func (s *CopilotService) loadImagePart(ctx context.Context, f fileModel.File) (agent.ImagePart, bool) {
	if fileModel.CategoryOf(f.ContentType) != fileModel.CategoryImage {
		return agent.ImagePart{}, false
	}

	if f.IsRef() {
		if f.RefURL == nil || *f.RefURL == "" {
			return agent.ImagePart{}, false
		}
		mediaType, ok := mediaTypeOfURL(f.ContentType, *f.RefURL)
		if !ok {
			return agent.ImagePart{}, false
		}
		return agent.ImagePart{MediaType: mediaType, URL: *f.RefURL}, true
	}

	if f.Size > maxSourceImageBytes {
		return agent.ImagePart{}, false
	}
	opened, err := s.files.Open(ctx, f.ID, nil)
	if err != nil || opened.Reader == nil {
		return agent.ImagePart{}, false
	}
	defer func() { _ = opened.Reader.Close() }()
	data, err := io.ReadAll(io.LimitReader(opened.Reader, maxSourceImageBytes+1))
	if err != nil || len(data) == 0 || len(data) > maxSourceImageBytes {
		return agent.ImagePart{}, false
	}
	mediaType, data, ok := fitForVision(data)
	if !ok {
		return agent.ImagePart{}, false
	}
	return agent.ImagePart{MediaType: mediaType, Base64: base64.StdEncoding.EncodeToString(data)}, true
}
