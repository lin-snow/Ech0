package middleware

import (
	"github.com/google/wire"
	authService "github.com/lin-snow/ech0/internal/service/auth"
)

// Deps 聚合中间件所需的外部依赖，由 Wire 注入。
type Deps struct {
	TokenRevoker authService.TokenRevoker
}

func NewDeps(revoker authService.TokenRevoker) *Deps {
	return &Deps{TokenRevoker: revoker}
}

var ProviderSet = wire.NewSet(NewDeps)
