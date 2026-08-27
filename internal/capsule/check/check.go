// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package check

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/lin-snow/ech0/internal/capsule"
)

type Options struct {
	Fix bool
}

func Validate(ctx context.Context, loaded *capsule.Loaded, opts Options) (*Report, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if loaded == nil {
		return nil, errors.New("capsule check: nil capsule")
	}
	if opts.Fix {
		if err := ensureWritable(loaded); err != nil {
			return nil, err
		}
	}

	r := &Report{}

	site := capsule.Site{}
	if loaded.Manifest != nil {
		site = loaded.Manifest.Site
	}

	echoIDs, referenced, err := validateEchoes(r, loaded, opts, site.ServerURL)
	if err != nil {
		return nil, err
	}
	validateManifest(r, loaded, site, referenced)
	validateComments(r, loaded, echoIDs)
	validateMedia(r, loaded, referenced, site)
	validatePaths(r, loaded)

	sortIssues(r.Issues)
	return r, nil
}

func Run(ctx context.Context, src *capsule.Source, opts Options) (*capsule.Loaded, *Report, error) {
	loaded, err := capsule.Load(ctx, src)
	if err != nil {
		return nil, nil, err
	}
	report, err := Validate(ctx, loaded, opts)
	if err != nil {
		return loaded, nil, err
	}
	return loaded, report, nil
}

func ensureWritable(loaded *capsule.Loaded) error {
	if loaded.Source == nil || loaded.Source.Path == "" {
		return errors.New("capsule check: --fix requires a capsule directory on disk")
	}
	info, err := os.Stat(loaded.Source.Path)
	if err != nil {
		return fmt.Errorf("capsule check: --fix cannot access %q: %w", loaded.Source.Path, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("capsule check: --fix is not supported for archive capsules (%s)", loaded.Source.Path)
	}
	return nil
}
