// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package storage

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// PublicFS serves the local storage root over HTTP without authentication,
// restricted to the media category directories written by the file schema.
//
// Everything else in the data directory stays unreachable: directories are
// refused (no listings, so stored artifacts cannot be enumerated), and only
// keys under a media prefix are opened. Snapshots, capsules and migration
// working copies live outside the storage root (see the migrator artifact
// slots) and are served exclusively by the admin-checked download route.
type PublicFS struct {
	root     http.Dir
	prefixes []string
}

func NewPublicFS(root string) http.FileSystem {
	return &PublicFS{root: http.Dir(root), prefixes: PublicPathPrefixes()}
}

func (p *PublicFS) Open(name string) (http.File, error) {
	rel, ok := publicKey(name, p.prefixes)
	if !ok {
		return nil, fs.ErrNotExist
	}

	file, err := p.root.Open("/" + rel)
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	if info.IsDir() {
		// Returning a directory handle would make http.FileServer emit the
		// listing (or a 500 when it refuses); 404 instead.
		_ = file.Close()
		return nil, fs.ErrNotExist
	}
	return file, nil
}

func publicKey(name string, prefixes []string) (string, bool) {
	// http.FileSystem names are absolute URL paths (FileServer strips the
	// request prefix); anything else is not a request we serve.
	if !strings.HasPrefix(name, "/") {
		return "", false
	}
	rel := strings.TrimPrefix(path.Clean(name), "/")
	if rel == "" || rel == "." {
		return "", false
	}
	for _, prefix := range prefixes {
		if rel == strings.TrimSuffix(prefix, "/") || strings.HasPrefix(rel, prefix) {
			return rel, true
		}
	}
	return "", false
}
