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

// RandomKeyGenerator creates keys like 1_1700000000_abc123.png.
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

func (g *RandomKeyGenerator) GenerateKey(_ Category, userID uint, originalFilename string) (string, error) {
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
	return strings.ToLower(fmt.Sprintf("%d_%d_%s%s", userID, now.UTC().Unix(), hex.EncodeToString(randPart), ext)), nil
}

// StaticKeyGenerator produces a fixed key, useful for singleton files
// like background music.
type StaticKeyGenerator struct {
	Name string
}

func (g *StaticKeyGenerator) GenerateKey(_ Category, _ uint, _ string) (string, error) {
	name := strings.TrimSpace(g.Name)
	if name == "" {
		return "", fmt.Errorf("static key name is empty")
	}
	return name, nil
}
