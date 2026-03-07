package model

// PageQueryDto 用于分页查询的请求数据传输对象
//
// swagger:model PageQueryDto
type PageQueryDto struct {
	Page     int    `json:"page"     form:"page"`
	PageSize int    `json:"pageSize" form:"pageSize"`
	Search   string `json:"search"   form:"search"`
}

// FileDto is the unified response for file operations.
// The Key field is the single source of truth — URLs are resolved at runtime.
//
// swagger:model FileDto
type FileDto struct {
	ID          uint   `json:"id"`
	Key         string `json:"key"`
	URL         string `json:"url"`
	ContentType string `json:"content_type,omitempty"`
	Category    string `json:"category,omitempty"`
	Size        int64  `json:"size,omitempty"`
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
}

// FileDeleteDto is the request body for deleting a file.
//
// swagger:model FileDeleteDto
type FileDeleteDto struct {
	Key string `json:"key" binding:"required"`
}

// PresignDto 用于响应 S3 预签名 URL 的请求数据传输对象
//
// swagger:model PresignDto
type PresignDto struct {
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	Key         string `json:"key"`
	PresignURL  string `json:"presign_url"`
	FileURL     string `json:"file_url"`
}

// GetPresignURLDto 用于请求 S3 预签名 URL 的请求数据传输对象
//
// swagger:model GetPresignURLDto
type GetPresignURLDto struct {
	FileName    string `json:"file_name"    binding:"required"`
	ContentType string `json:"content_type"`
}

// GetWebsiteTitleDto 用于请求网站标题的请求数据传输对象
//
// swagger:model GetWebsiteTitleDto
type GetWebsiteTitleDto struct {
	WebSiteURL string `json:"website_url" form:"website_url" binding:"required"`
}
