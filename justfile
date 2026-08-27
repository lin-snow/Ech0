
set shell := ["bash", "-cu"]

mod web
mod site
mod hub
mod docker

VERSION       := `git describe --tags --always 2>/dev/null || echo unknown`
BUILD_TIME    := `date -u +%Y-%m-%dT%H:%M:%SZ`
GIT_COMMIT    := `git rev-parse --short HEAD 2>/dev/null || echo unknown`

VERSION_PKG   := "github.com/lin-snow/ech0/internal/version"
LDFLAGS       := "-X " + VERSION_PKG + ".Commit=" + GIT_COMMIT + " -X " + VERSION_PKG + ".BuildTime=" + BUILD_TIME

MOCKERY_VERSION := env_var_or_default("MOCKERY_VERSION", "v3.7.4")

COVER_EXCLUDE := 'internal/test/mocks/|/wire_gen\.go:'

default:
    @just --list


run:
    ECH0_SERVER_MODE=debug go run -ldflags "{{LDFLAGS}}" ./cmd/ech0 serve

build:
    go build -ldflags "{{LDFLAGS}}" -o ./bin/ech0 ./cmd/ech0

dev:
    #!/usr/bin/env bash
    set -euo pipefail
    AIR_BIN="$(command -v air 2>/dev/null || echo "$(go env GOPATH)/bin/air")"
    if [ ! -x "$AIR_BIN" ]; then
        echo "air not found, installing..."
        just air-install
        AIR_BIN="$(go env GOPATH)/bin/air"
    fi
    ECH0_SERVER_MODE=debug "$AIR_BIN" -c .air.toml

air-install:
    go install github.com/air-verse/air@latest


lint:
    golangci-lint run

fmt:
    golangci-lint fmt

test:
    go test ./...

test-race:
    CGO_ENABLED=1 go test -race ./...

test-cover:
    CGO_ENABLED=1 go test -coverprofile=coverage.out -covermode=atomic ./...
    @grep -v -E '{{COVER_EXCLUDE}}' coverage.out > coverage.calibrated.out
    @printf 'RAW        (incl. generated): '; go tool cover -func=coverage.out            | tail -1 | awk '{print $NF}'
    @printf 'CALIBRATED (excl. generated): '; go tool cover -func=coverage.calibrated.out | tail -1 | awk '{print $NF}'


mocks:
    go run github.com/vektra/mockery/v3@{{MOCKERY_VERSION}}

mocks-check: mocks
    git diff --exit-code -- internal/test/mocks

wire:
    go generate ./internal/di

wire-check: wire
    git diff --exit-code -- internal/di/wire_gen.go

openapi:
    go run ./cmd/openapi-gen

openapi-check: openapi
    git diff --exit-code -- internal/openapi/openapi.yaml


spdx:
    node scripts/add-spdx-headers.mjs

spdx-check:
    node scripts/add-spdx-headers.mjs --check


check:
    bash scripts/check.sh


bump NEW_VERSION:
    #!/usr/bin/env bash
    set -euo pipefail
    SEMVER='^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$'
    if ! echo "{{NEW_VERSION}}" | grep -Eq "$SEMVER"; then
        echo "✘ '{{NEW_VERSION}}' is not valid semver (expected X.Y.Z[-prerelease])"
        exit 1
    fi
    if [ -n "$(git status --porcelain)" ]; then
        echo "✘ Working tree dirty — commit or stash first so the bump commit is clean."
        git status --short
        exit 1
    fi
    OLD_VERSION="$(grep -E '^[[:space:]]*Version[[:space:]]*=[[:space:]]*"' internal/version/version.go \
                    | head -n1 \
                    | sed -E 's/.*"([^"]+)".*/\1/')"
    if [ -z "$OLD_VERSION" ]; then
        echo "✘ Could not extract current Version from internal/version/version.go"
        exit 1
    fi
    if [ "$OLD_VERSION" = "{{NEW_VERSION}}" ]; then
        echo "✘ Version is already $OLD_VERSION — nothing to bump."
        exit 1
    fi
    echo "→ bumping $OLD_VERSION → {{NEW_VERSION}}"
    sed -i.bak -E "s/^([[:space:]]*Version[[:space:]]*=[[:space:]]*\")[^\"]+(\")/\\1{{NEW_VERSION}}\\2/" internal/version/version.go
    rm -f internal/version/version.go.bak
    echo "→ verifying go build still succeeds..."
    go build ./... >/dev/null || { echo "✘ go build failed after bump — reverting"; git checkout -- internal/version/version.go; exit 1; }
    echo ""
    echo "✓ Version bumped. Diff:"
    git --no-pager diff -- internal/version/version.go
    echo ""
    echo "Next steps (review the diff above, then run):"
    echo ""
    echo "  # 1. Update CHANGELOG.md: rename [Unreleased] → [{{NEW_VERSION}}] - $(date -u +%Y-%m-%d), open a new empty [Unreleased]"
    echo "  # 2. Commit + tag:"
    echo "       git commit -am 'chore(release): v{{NEW_VERSION}}'"
    echo "       git tag -a v{{NEW_VERSION}} -m 'Release v{{NEW_VERSION}}'"
    echo "  # 3. Push to trigger release workflow:"
    echo "       git push origin main"
    echo "       git push origin v{{NEW_VERSION}}"
