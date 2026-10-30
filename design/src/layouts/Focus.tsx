// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { FocusTop } from '../components/Chrome'
import { Feed } from '../components/Feed'
import { useComposer } from '../lib/useComposer'
import { useSketchFeed } from '../lib/useSketchFeed'
import layout from './layout.module.css'

export function FocusLayout() {
  const feed = useSketchFeed()
  const composer = useComposer(feed)

  return (
    <div className={layout.focus}>
      <FocusTop />
      <div className={layout.focusScroll}>
        <div className={layout.focusContent}>
          <Feed feed={feed} variant="focus" composer={composer} />
        </div>
      </div>
    </div>
  )
}
