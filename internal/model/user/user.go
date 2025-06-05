package model

const (
	USER_NOT_EXISTS_ID = 0
)

type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Username string `gorm:"size:255;not null;unique" json:"username"`
	Password string `gorm:"size:255;not null" json:"password"`
	IsAdmin  bool   `gorm:"bool" json:"is_admin"`
	Avatar   string `gorm:"size:255" json:"avatar"`
}
