// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package humares

import (
	"errors"
	"net/http"
	"sync"

	"github.com/danielgtaylor/huma/v2"
	"github.com/lin-snow/ech0/internal/httperr"
)

var installErrorModelOnce sync.Once

// installErrorModel makes every error Huma raises itself (routing misses,
// decoding failures, validation errors) come out as the same RFC 9457 document
// the domain errors use.
func installErrorModel() {
	installErrorModelOnce.Do(func() {
		huma.NewErrorWithContext = func(_ huma.Context, status int, msg string, errs ...error) huma.StatusError {
			return frameworkProblem(status, msg, errs...)
		}
		huma.NewError = func(status int, msg string, errs ...error) huma.StatusError {
			return frameworkProblem(status, msg, errs...)
		}
	})
}

// frameworkProblem renders errors Huma raises itself. Field-level details ride
// in `errors[]`; the document-level detail is Huma's own English text.
func frameworkProblem(status int, msg string, errs ...error) *huma.ErrorModel {
	// Huma reports every unannotated handler error as 500. A handler that
	// returned a domain error meant 4xx, so re-project it instead of flattening
	// it into INTERNAL; framework-raised 4xx (decode/validation) keep their
	// shape.
	if status >= http.StatusInternalServerError && len(errs) == 1 {
		if _, isDetail := errors.AsType[*huma.ErrorDetail](errs[0]); !isDetail {
			return httperr.From(errs[0])
		}
	}

	problem := httperr.From(httperr.Edge(status, msg))
	for _, err := range errs {
		detail, ok := errors.AsType[*huma.ErrorDetail](err)
		if !ok {
			detail = &huma.ErrorDetail{Message: err.Error()}
		}
		problem.Errors = append(problem.Errors, detail)
	}
	return problem
}
