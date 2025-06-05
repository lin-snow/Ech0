package model

import (
	"github.com/golang-jwt/jwt/v5"
)

type MyClaims struct {
	Userid   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}
