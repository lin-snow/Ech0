package model

type UserInfoDto struct {
	Username string `json:"username"`
	Password string `json:"password"`
	IsAdmin  bool   `json:"is_admin"`
	Avatar   string `json:"avatar"`
}
