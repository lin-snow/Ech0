// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

/** Markdown is imported as text (`with { type: "text" }`) so builds embed it. */
declare module "*.md" {
  const text: string;
  export default text;
}
