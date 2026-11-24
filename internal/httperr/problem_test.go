// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package httperr

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/lin-snow/ech0/internal/apperr"
)

func TestCatalogProjectsEverySentinel(t *testing.T) {
	if len(apperr.Classes()) == 0 {
		t.Fatal("apperr catalog is empty")
	}
	seen := map[string]struct{}{}
	for _, class := range apperr.Classes() {
		text := class.Err.Error()
		if _, ok := seen[text]; ok {
			t.Fatalf("duplicate detail %q", text)
		}
		seen[text] = struct{}{}

		problem := From(fmt.Errorf("while saving: %w", class.Err))
		if problem.Status != class.Status {
			t.Fatalf("%q status = %d, catalog %d", text, problem.Status, class.Status)
		}
		if problem.Type != class.Type {
			t.Fatalf("%q type = %q, catalog %q", text, problem.Type, class.Type)
		}
		if class.Status >= http.StatusInternalServerError {
			if problem.Detail != detailInternal {
				t.Fatalf("%q detail = %q, want redacted", text, problem.Detail)
			}
			continue
		}
		if problem.Detail != text {
			t.Fatalf("%q detail = %q", text, problem.Detail)
		}
	}
}

func TestUnknownErrorIsRedacted(t *testing.T) {
	problem := From(errors.New("dial tcp 10.0.0.1:5432: i/o timeout"))
	if problem.Status != http.StatusInternalServerError || problem.Detail != detailInternal {
		t.Fatalf("problem = %+v", problem)
	}
}

func TestEdgeCauseOverridesStatus(t *testing.T) {
	problem := From(EdgeCause(http.StatusTooManyRequests, apperr.ErrInvalidRequestBody))
	if problem.Status != http.StatusTooManyRequests {
		t.Fatalf("status = %d", problem.Status)
	}
	if problem.Detail != apperr.ErrInvalidRequestBody.Error() {
		t.Fatalf("detail = %q", problem.Detail)
	}
}
