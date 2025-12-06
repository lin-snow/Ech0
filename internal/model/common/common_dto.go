package model

// PageQueryDto 用于分页查询的请求数据传输对象
//
// swagger:model PageQueryDto
type PageQueryDto struct {
	Page     int    `json:"page"     form:"page"`     // 页码，从1开始
	PageSize int    `json:"pageSize" form:"pageSize"` // 每页大小
	Search   string `json:"search"   form:"search"`   // 用于搜索的关键字
}

// ImageDto 用于图片相关的请求数据传输对象
//
// swagger:model ImageDto
type ImageDto struct {
	// 图片的 URL 地址
	URL       string `json:"url"        binding:"required"`
	SOURCE    string `json:"source"     binding:"required"`
	ObjectKey string `json:"object_key"` // 对象存储的 Key, 用于删除 S3/R2 上的图片
	Width     int    `json:"width"`      // 图片宽度
	Height    int    `json:"height"`     // 图片高度
}

// PresignDto 用于响应 S3 预签名 URL 的请求数据传输对象
//
// swagger:model PresignDto
type PresignDto struct {
	FileName    string `json:"file_name"` // 原始文件名
	ContentType string `json:"content_type"`
	ObjectKey   string `json:"object_key"`  // 预签名的对象存储 Key
	PresignURL  string `json:"presign_url"` // 预签名 URL
	FileURL     string `json:"file_url"`    // 文件访问 URL,用于回显
}

// GetPresignURLDto 用于请求 S3 预签名 URL 的请求数据传输对象
//
// swagger:model GetPresignURLDto
type GetPresignURLDto struct {
	FileName    string `json:"file_name"    binding:"required"` // 原始文件名
	ContentType string `json:"content_type"`                    // 文件的 MIME 类型
}

// GetWebsiteTitleDto 用于请求网站标题的请求数据传输对象
//
// swagger:model GetWebsiteTitleDto
type GetWebsiteTitleDto struct {
	WebSiteURL string `json:"website_url" form:"website_url" binding:"required"` // 网站URL
}

// ModelDto 用于3D模型相关的请求数据传输对象
//
// swagger:model ModelDto
type ModelDto struct {
	// 模型的 URL 地址
	URL string `json:"url" binding:"required"`
}
