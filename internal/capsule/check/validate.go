// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package check

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lin-snow/ech0/internal/capsule"
	"github.com/lin-snow/ech0/internal/storage"
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
)

const echoFileMode = 0o644

func validateManifest(
	r *Report,
	loaded *capsule.Loaded,
	site capsule.Site,
	referenced map[string]struct{},
) {
	p := capsule.ManifestPath

	if loaded.ManifestErr != nil {
		r.errorf(p, "", "%v", loaded.ManifestErr)
	}
	for _, u := range loaded.ManifestUnknown {
		r.warnf(p, "", "unknown field ignored: %s", u)
	}
	if loaded.Manifest == nil {
		return
	}

	switch v := loaded.Manifest.SchemaVersion; {
	case v <= 0:
		r.errorf(p, "schema_version", "schema_version is required")
	case v > capsule.SchemaVersion:
		r.errorf(p, "schema_version", "unsupported schema_version %d, this build supports up to %d", v, capsule.SchemaVersion)
	}

	if loaded.Manifest.Owner.Username == "" {
		r.errorf(p, "owner.username", "owner.username is required")
	}

	if site.CustomJS != "" {
		r.warnf(p, "site.custom_js", "custom_js is not empty: consuming this capsule means running third-party code")
	}
	if site.CustomCSS != "" {
		r.warnf(p, "site.custom_css", "custom_css is not empty: consuming this capsule means applying third-party styles")
	}
	if marker := instanceMarker(site.ServerLogo, site.ServerURL); marker != "" {
		r.warnf(p, "site.server_logo", "embeds source instance URL (%s), the link may break after migration", marker)
	}

	validateFiles(r, loaded, p, loaded.Manifest.Files, referenced)
}

func validateEchoes(r *Report, loaded *capsule.Loaded, opts Options, serverURL string) (ids map[string]struct{}, referenced map[string]struct{}, err error) {
	ids = make(map[string]struct{}, len(loaded.Echoes))
	referenced = make(map[string]struct{})
	firstSeen := make(map[string]string, len(loaded.Echoes))

	for i := range loaded.Echoes {
		e := &loaded.Echoes[i]
		if e.Err != nil {
			r.errorf(e.Path, "", "%v", e.Err)
			continue
		}
		for _, u := range e.Unknown {
			r.warnf(e.Path, "", "unknown field ignored: %s", u)
		}
		doc := e.Doc
		if doc == nil {
			continue
		}

		if doc.ID == "" && opts.Fix {
			if err := fixEchoID(r, loaded, e); err != nil {
				return nil, nil, err
			}
		}

		switch {
		case doc.ID == "":
			r.errorf(e.Path, "id", "id is required, run `ech0 check --fix` to generate one")
		case !uuidUtil.IsValid(doc.ID):
			r.errorf(e.Path, "id", "id %q is not a valid UUID", doc.ID)
		default:
			if first, dup := firstSeen[doc.ID]; dup {
				r.errorf(e.Path, "id", "duplicate id %s, already used by %s", doc.ID, first)
			} else {
				firstSeen[doc.ID] = e.Path
			}
			ids[doc.ID] = struct{}{}
		}

		if doc.CreatedAt == "" {
			r.errorf(e.Path, "created_at", "created_at is required")
		} else if _, perr := capsule.ParseTime(doc.CreatedAt); perr != nil {
			r.errorf(e.Path, "created_at", "%v", perr)
		}

		if doc.Layout != "" {
			if _, ok := capsule.ValidLayouts[doc.Layout]; !ok {
				r.warnf(e.Path, "layout", "unknown layout %q, consumers fall back to %q", doc.Layout, capsule.DefaultLayout)
			}
		}

		validateExtension(r, e.Path, doc.Extension, serverURL)
		validateFiles(r, loaded, e.Path, doc.Files, referenced)

		if marker := instanceMarker(doc.Content, serverURL); marker != "" {
			r.warnf(e.Path, "content", "embeds source instance URL (%s), the link may break after migration", marker)
		}
	}
	return ids, referenced, nil
}

func validateExtension(r *Report, echoPath string, ext *capsule.Extension, serverURL string) {
	if ext == nil {
		return
	}
	if ext.Type == "" {
		r.errorf(echoPath, "extension.type", "extension.type is required when extension is present")
	} else if _, ok := capsule.ValidExtensionTypes[ext.Type]; !ok {
		r.warnf(echoPath, "extension.type", "unknown extension type %q, consumers skip rendering it", ext.Type)
	}
	if ext.Payload == nil {
		r.errorf(echoPath, "extension.payload", "extension.payload is required when extension is present")
		return
	}
	for _, hit := range scanPayload(ext.Payload, serverURL) {
		r.warnf(echoPath, hit.field, "embeds source instance URL (%s), the link may break after migration", hit.marker)
	}
}

func validateFiles(r *Report, loaded *capsule.Loaded, echoPath string, files []capsule.FileRef, referenced map[string]struct{}) {
	for i := range files {
		f := files[i]
		field := fmt.Sprintf("files[%d]", i)

		switch {
		case f.Key != "" && f.URL != "":
			r.errorf(echoPath, field, "key and url are mutually exclusive")
		case f.Key == "" && f.URL == "":
			r.errorf(echoPath, field, "one of key or url is required")
		}

		if f.Category != "" {
			if _, ok := capsule.ValidCategories[f.Category]; !ok {
				r.warnf(echoPath, field+".category", "unknown category %q, consumers fall back to %q", f.Category, string(storage.CategoryFile))
			}
		}

		if f.Key == "" {
			continue
		}
		if err := capsule.ValidateKey(f.Key); err != nil {
			r.errorf(echoPath, field+".key", "%v", err)
			continue
		}

		media := capsule.MediaPath(f.Key)
		referenced[media] = struct{}{}

		size, ok := loaded.MediaPaths[media]
		if !ok {
			r.errorf(echoPath, field+".key", "capsule is not self-contained: %s is missing for key %q", media, f.Key)
			continue
		}
		if f.Size > 0 && f.Size != size {
			r.warnf(echoPath, field+".size", "declared size %d does not match %s (%d bytes on disk)", f.Size, media, size)
		}
	}
}

func validateComments(r *Report, loaded *capsule.Loaded, echoIDs map[string]struct{}) {
	p := capsule.CommentsPath
	if !loaded.HasComments {
		return
	}
	if loaded.CommentsErr != nil {
		r.errorf(p, "", "%v", loaded.CommentsErr)
		return
	}
	for _, u := range loaded.CommentsUnknown {
		r.warnf(p, "", "unknown field ignored: %s", u)
	}

	raw, err := capsule.RawComments(loaded.CommentsRaw)
	if err != nil {
		r.errorf(p, "", "parse %s: %v", p, err)
	}
	for i, item := range raw {
		for _, forbidden := range capsule.ForbiddenCommentFields {
			if _, ok := item[forbidden]; ok {
				r.errorf(p, fmt.Sprintf("comments[%d].%s", i, forbidden),
					"forbidden field %q must not appear in a capsule", forbidden)
			}
		}
	}

	if loaded.Comments == nil {
		return
	}
	firstSeen := make(map[string]int, len(loaded.Comments.Comments))
	for i := range loaded.Comments.Comments {
		c := loaded.Comments.Comments[i]
		at := func(name string) string { return fmt.Sprintf("comments[%d].%s", i, name) }

		if c.ID == "" {
			r.errorf(p, at("id"), "id is required")
		} else if first, dup := firstSeen[c.ID]; dup {
			r.errorf(p, at("id"), "duplicate id %s, already used by comments[%d]", c.ID, first)
		} else {
			firstSeen[c.ID] = i
		}

		if c.EchoID == "" {
			r.errorf(p, at("echo_id"), "echo_id is required")
		} else if _, ok := echoIDs[c.EchoID]; !ok {
			r.warnf(p, at("echo_id"), "orphan comment: echo %s is not in this capsule", c.EchoID)
		}

		if c.Nickname == "" {
			r.errorf(p, at("nickname"), "nickname is required")
		}
		if c.Content == "" {
			r.errorf(p, at("content"), "content is required")
		}
		if c.CreatedAt == "" {
			r.errorf(p, at("created_at"), "created_at is required")
		} else if _, perr := capsule.ParseTime(c.CreatedAt); perr != nil {
			r.errorf(p, at("created_at"), "%v", perr)
		}

		if c.Status != "" && c.Status != capsule.DefaultCommentStatus {
			r.warnf(p, at("status"), "status %q is not %q: a capsule should only carry approved comments",
				c.Status, capsule.DefaultCommentStatus)
		}
	}
}

func validateMedia(r *Report, loaded *capsule.Loaded, referenced map[string]struct{}, site capsule.Site) {
	logo := logoMedia(site.ServerLogo, loaded.MediaPaths)
	for _, p := range sortedMediaPaths(loaded.MediaPaths) {
		if _, ok := referenced[p]; ok {
			continue
		}
		if p == logo {
			continue
		}
		r.warnf(p, "", "dangling media: not referenced by any echo file or site.server_logo")
	}
}

func validatePaths(r *Report, loaded *capsule.Loaded) {
	for i := range loaded.Echoes {
		reportTraversal(r, loaded.Echoes[i].Path)
	}
	for _, p := range sortedMediaPaths(loaded.MediaPaths) {
		reportTraversal(r, p)
	}
	for _, p := range loaded.UnknownPaths {
		reportTraversal(r, p)
		r.warnf(p, "", "unknown path ignored: not defined by the capsule spec")
	}
}

func reportTraversal(r *Report, p string) {
	if hasTraversal(p) {
		r.errorf(p, "", "path traversal (\"..\") is forbidden inside a capsule")
	}
}

func hasTraversal(p string) bool {
	for _, seg := range strings.Split(p, "/") {
		if seg == ".." {
			return true
		}
	}
	return false
}

func fixEchoID(r *Report, loaded *capsule.Loaded, e *capsule.LoadedEcho) error {
	id := uuidUtil.NewV7()
	e.Doc.ID = id

	data, err := capsule.EncodeEcho(e.Doc)
	if err != nil {
		return fmt.Errorf("capsule check: re-encode %s: %w", e.Path, err)
	}
	target := filepath.Join(loaded.Source.Path, filepath.FromSlash(e.Path))
	if err := os.WriteFile(target, data, echoFileMode); err != nil {
		return fmt.Errorf("capsule check: write back %s: %w", e.Path, err)
	}

	r.Fixed = append(r.Fixed, fmt.Sprintf("%s: generated id %s", e.Path, id))
	return nil
}

func logoMedia(logo string, media map[string]int64) string {
	if logo == "" || len(media) == 0 {
		return ""
	}
	base := logo
	if u, err := url.Parse(logo); err == nil && u.Path != "" {
		base = u.Path
	}
	base = path.Base(base)
	if base == "" || base == "." || base == "/" {
		return ""
	}
	for p := range media {
		if path.Base(p) == base {
			return p
		}
	}
	return ""
}

func sortedMediaPaths(media map[string]int64) []string {
	paths := make([]string, 0, len(media))
	for p := range media {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	return paths
}
