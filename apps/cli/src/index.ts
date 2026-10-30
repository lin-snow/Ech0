#!/usr/bin/env bun
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { createRootCommand } from "./commands/root.ts";

await createRootCommand().parseAsync(process.argv);
