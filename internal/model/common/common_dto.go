package model

// 分页相关
type PageQueryDto struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Search   string `json:"search"`
}

type ImageDto struct {
	// 图片的 URL 地址
	URL    string `json:"url" binding:"required"`
	SOURCE string `json:"source" binding:"required"`
}
