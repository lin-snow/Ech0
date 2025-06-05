package jwt

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lin-snow/ech0/internal/config"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	"log"
	"time"
)

// CreateClaims 创建Claims
func CreateClaims(user userModel.User) jwt.Claims {

	claims := authModel.MyClaims{
		Userid:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    config.Config.Auth.Jwt.Issuer,
			Subject:   user.Username,
			Audience:  jwt.ClaimStrings{config.Config.Auth.Jwt.Audience},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(config.Config.Auth.Jwt.Expires) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	return claims
}

// GenerateToken 生成JWT Token
func GenerateToken(claim jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	return token.SignedString(config.JWT_SECRET)
}

// ParseToken 解析JWT Token
func ParseToken(tokenString string) (*authModel.MyClaims, error) {
	claims := &authModel.MyClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return config.JWT_SECRET, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*authModel.MyClaims); ok {
		return claims, nil
	}

	log.Println("unknown claims type, cannot proceed")
	return nil, errors.New("unknown claims type, cannot proceed")
}
