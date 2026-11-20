#!/bin/sh
set -eu

# sqlite-vec needs sqlite3.h. Use the matching header of the pinned SQLite
# amalgamation linked into the binary by go-sqlite3.
headers=$(mktemp -d)
trap 'rm -f "$headers/sqlite3.h" "$headers/sqlite3ext.h"; rmdir "$headers"' EXIT
go mod download github.com/mattn/go-sqlite3
sqlite_source=$(go list -m -f '{{.Dir}}' github.com/mattn/go-sqlite3)
cp "$sqlite_source/sqlite3-binding.h" "$headers/sqlite3.h"
cp "$sqlite_source/sqlite3ext.h" "$headers/sqlite3ext.h"

export VITE_ECH0_OBJECT_UPLOADS=true
npm exec --yes --package=pnpm@12.16.1 -- pnpm -C web install --frozen-lockfile
npm exec --yes --package=pnpm@12.16.1 -- pnpm -C web build
CGO_ENABLED=1 CGO_CFLAGS="-I$headers" go build -trimpath -ldflags="-s -w" -o "${VERCEL_OUTPUT_FILE:-server}" ./cmd/api
