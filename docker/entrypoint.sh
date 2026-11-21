#!/bin/sh
# SPDX-License-Identifier: AGPL-3.0-or-later
# Copyright (C) 2025-2026 lin-snow

# The image starts as root only to hand /app/data to uid/gid 10001: images
# before 5.10 wrote that directory as root, and a non-root server cannot take
# ownership of it from inside the container. The server itself never runs as
# root — this script drops privileges for good before executing it.
#
# `docker run --user 10001:10001` (and the Helm chart) skip the handoff
# entirely: nothing to fix, nothing to drop.

set -eu

UID_GID="10001:10001"
DATA_DIR="/app/data"

if [ "$(id -u)" = "0" ]; then
	if [ -d "${DATA_DIR}" ] && [ "$(stat -c %u "${DATA_DIR}")" != "10001" ]; then
		# A failed chown (read-only mount, foreign ownership) is not fatal here:
		# the server reports an unwritable data directory with the exact fix.
		chown -R "${UID_GID}" "${DATA_DIR}" ||
			echo "ech0: could not take ownership of ${DATA_DIR}" >&2
	fi
	exec su-exec "${UID_GID}" /app/ech0 "$@"
fi

exec /app/ech0 "$@"
