// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

type ActivityHeatmapProps = {
  data: number[]
  label?: string
}

export function ActivityHeatmap({ data, label = '活跃度' }: ActivityHeatmapProps) {
  const weeks: number[][] = []
  for (let index = 0; index < data.length; index += 4) {
    weeks.push(data.slice(index, index + 4))
  }

  return (
    <div className="heatmap" role="img" aria-label={label}>
      {weeks.map((week, weekIndex) => (
        <div className="heatmap__week" key={`week-${weekIndex}`}>
          {week.map((level, dayIndex) => (
            <span
              key={`day-${weekIndex}-${dayIndex}`}
              className="heat-cell"
              data-level={Math.max(0, Math.min(4, level))}
              title={`${level} 条记录`}
            />
          ))}
        </div>
      ))}
    </div>
  )
}
