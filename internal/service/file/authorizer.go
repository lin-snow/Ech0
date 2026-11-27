// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"fmt"
	"sync"

	fileModel "github.com/lin-snow/ech0/internal/model/file"
	"github.com/lin-snow/ech0/pkg/role"
	"github.com/lin-snow/ech0/pkg/viewer"
)

// AuthorizerRegistry maps an owner kind ("echo", …) to the domain that owns
// its visibility rules. The file domain never imports a business domain; the
// business domain registers an Authorizer at boot, which is the only edge
// between the two.
type AuthorizerRegistry struct {
	mu     sync.RWMutex
	byKind map[string]Authorizer
}

func NewAuthorizerRegistry() *AuthorizerRegistry {
	return &AuthorizerRegistry{byKind: map[string]Authorizer{}}
}

// Register binds a domain's authorizer. Registering the same kind twice is a
// programming error: two answers to one question is exactly the drift this
// design removes, so it panics at boot rather than picking one.
func (r *AuthorizerRegistry) Register(a Authorizer) {
	if a == nil {
		panic("file: nil authorizer")
	}
	kind := a.Kind()
	if kind == "" {
		panic("file: authorizer with empty kind")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, dup := r.byKind[kind]; dup {
		panic(fmt.Sprintf("file: duplicate authorizer for %q", kind))
	}
	r.byKind[kind] = a
}

// For returns the authorizer of a kind, or nil when the domain never
// registered one. A missing authorizer denies: an owner kind nobody vouches
// for must not become public by omission.
func (r *AuthorizerRegistry) For(kind string) Authorizer {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.byKind[kind]
}

// decide is the whole authorization rule of a file, in one place:
//
//	public  → anyone holding the id
//	private → its uploader, and administrators
//	owner   → any link whose domain allows the principal; otherwise the
//	          uploader/administrator fallback, so an ownerless draft still has
//	          exactly one reader instead of zero
func (s *FileService) decide(ctx context.Context, f *fileModel.File, p viewer.Context) (bool, error) {
	switch fileModel.Access(f.Access) {
	case fileModel.AccessPublic:
		return true, nil
	case fileModel.AccessPrivate:
		return s.byUploader(f, p), nil
	default:
		links, err := s.repo.LinksOfFile(ctx, f.ID)
		if err != nil {
			return false, err
		}
		for _, link := range links {
			authorizer := s.authorizers.For(link.OwnerKind)
			if authorizer == nil {
				continue
			}
			allowed, err := authorizer.Authorize(ctx, link.OwnerID, p)
			if err != nil {
				return false, err
			}
			if allowed {
				return true, nil
			}
		}
		return s.byUploader(f, p), nil
	}
}

// byUploader is the "no domain says yes" answer: the person who put the bytes
// here, and administrators, may read them. It also covers a file with no link
// at all, which has no other reader.
func (s *FileService) byUploader(f *fileModel.File, p viewer.Context) bool {
	if viewer.IsAnonymous(p) {
		return false
	}
	return p.UserID() == f.UploaderID || p.Role().AtLeast(role.Admin)
}
