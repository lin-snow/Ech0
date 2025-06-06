package model

import (
	"github.com/golang-jwt/jwt/v5"
)

type MyClaims struct {
	Userid   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

const (
	MAX_USER_COUNT  = 5
	NO_USER_LOGINED = uint(0)
)
