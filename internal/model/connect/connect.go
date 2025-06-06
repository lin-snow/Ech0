package model

type Connect struct {
	ServerName  string `json:"server_name"`  // 服务器名称
	ServerURL   string `json:"server_url"`   // 服务器地址
	Logo        string `json:"logo"`         // 站点logo
	Ech0s       int    `json:"ech0s"`        // 留言数量
	SysUsername string `json:"sys_username"` // 系统管理员用户名
}

type Connected struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	ConnectURL string `json:"connect_url"` // 连接地址
}
