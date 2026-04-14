package auth

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewAuthService,
	wire.Bind(new(Service), new(*AuthService)),
)
