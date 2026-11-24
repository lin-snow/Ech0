#!/usr/bin/env bash
# SPDX-License-Identifier: AGPL-3.0-or-later
# Copyright (C) 2025-2026 lin-snow
#
# Fails when the working-tree OpenAPI spec removes or narrows what clients
# already speak. Regenerated specs are checked in, so `just openapi-check`
# already proves the spec matches the code; this step proves the spec did not
# regress the wire contract.
#
# Baseline, in order: internal/openapi/baseline.yaml when present, otherwise the
# newest `v*` tag. A baseline file exists only for an intended break (e.g. the
# envelope removal): `just api-baseline` accepts the current spec as the new
# promise, and the file is deleted once the breaking release is tagged, because
# from then on the tag itself carries the new contract.
# One-shot escape hatch: ECH0_API_BREAKING_ALLOW=1.

set -euo pipefail

SPEC="internal/openapi/openapi.yaml"
BASELINE="internal/openapi/baseline.yaml"
OASDIFF_VERSION="v1.11.7"

if [ "${1:-}" = "--write" ]; then
  cp "$SPEC" "$BASELINE"
  echo "api-breaking: baseline written to $BASELINE"
  echo "api-breaking: delete it once the breaking release is tagged"
  exit 0
fi

tmp=""
trap 'rm -f "$tmp"' EXIT

if [ -f "$BASELINE" ]; then
  base="$BASELINE"
  base_label="$BASELINE"
else
  base_ref="${ECH0_API_BASE_REF:-$(git describe --tags --abbrev=0 --match 'v*' 2>/dev/null || true)}"
  if [ -z "$base_ref" ] || ! git rev-parse --verify --quiet "$base_ref" >/dev/null; then
    echo "api-breaking: no baseline, skipping"
    exit 0
  fi
  tmp="$(mktemp)"
  git show "$base_ref:$SPEC" >"$tmp"
  base="$tmp"
  base_label="$base_ref"
fi

if cmp -s "$base" "$SPEC"; then
  echo "api-breaking: spec unchanged against $base_label"
  exit 0
fi

if [ "${ECH0_API_BREAKING_ALLOW:-0}" = "1" ]; then
  echo "api-breaking: breaking changes allowed by ECH0_API_BREAKING_ALLOW=1 against $base_label"
  exit 0
fi

go run "github.com/oasdiff/oasdiff@${OASDIFF_VERSION}" breaking "$base" "$SPEC" --fail-on ERR --color never
