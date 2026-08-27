// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package storage

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"
)

func compactUserID(userID string) string {
	uid := strings.TrimSpace(userID)
	if uid == "" {
		return "anon"
	}
	uid = strings.ReplaceAll(uid, "-", "")
	if len(uid) > 8 {
		uid = uid[:8]
	}
	return strings.ToLower(uid)
}

type RandomKeyGenerator struct {
	RandSource  io.Reader
	SuffixBytes int
	Now         func() time.Time
}

func NewRandomKeyGenerator() *RandomKeyGenerator {
	return &RandomKeyGenerator{
		RandSource:  rand.Reader,
		SuffixBytes: 4,
		Now:         time.Now,
	}
}

func (g *RandomKeyGenerator) GenerateKey(_ Category, userID string, originalFilename string) (string, error) {
	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(originalFilename)))
	if ext == "" {
		ext = ".bin"
	}
	size := g.SuffixBytes
	if size <= 0 {
		size = 4
	}
	src := g.RandSource
	if src == nil {
		src = rand.Reader
	}
	randPart := make([]byte, size)
	if _, err := io.ReadFull(src, randPart); err != nil {
		return "", err
	}
	now := g.Now()
	uid := compactUserID(userID)
	return strings.ToLower(fmt.Sprintf("%s_%d_%s%s", uid, now.UTC().Unix(), hex.EncodeToString(randPart), ext)), nil
}

type StaticKeyGenerator struct {
	Name string
}

func (g *StaticKeyGenerator) GenerateKey(_ Category, _ string, _ string) (string, error) {
	name := strings.TrimSpace(g.Name)
	if name == "" {
		return "", fmt.Errorf("static key name is empty")
	}
	return name, nil
}
