package visitor

// DailyStat 保存 PV/UV 的日粒度快照。
type DailyStat struct {
	Date string `gorm:"type:char(10);primaryKey" json:"date"`
	PV   int64  `gorm:"default:0" json:"pv"`
	UV   int64  `gorm:"default:0" json:"uv"`
}
