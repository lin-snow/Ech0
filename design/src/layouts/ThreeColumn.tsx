// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { ActivityTabs, MobileHeader, SiteRail } from '../components/Chrome'
import { Composer } from '../components/Composer'
import { Feed } from '../components/Feed'
import { focusComposer } from '../lib/focusSearch'
import { useComposer } from '../lib/useComposer'
import { useSketchFeed } from '../lib/useSketchFeed'
import layout from './layout.module.css'

export function ThreeColumnLayout({ frameWidth }: { frameWidth: number }) {
  const feed = useSketchFeed()
  const composer = useComposer(feed)
  const wide = frameWidth > 880

  return (
    <div className={layout.three}>
      <div className={layout.left}>
        <SiteRail
          composer={wide ? <Composer composer={composer} compact /> : undefined}
          feed={feed}
        />
      </div>

      <section className={layout.center}>
        <MobileHeader onCompose={focusComposer} />
        <div className={layout.scroll}>
          <Feed
            feed={feed}
            variant="columns"
            heading={
              wide ? undefined : (
                <div className={layout.activityBlock}>
                  <ActivityTabs />
                </div>
              )
            }
            composer={wide ? undefined : composer}
            showToolbar={!wide}
          />
        </div>
      </section>
    </div>
  )
}
