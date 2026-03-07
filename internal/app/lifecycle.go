package app

import "context"

// Lifecycle 定义应用生命周期单元。
type Lifecycle interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// Namer 为可选能力，用于生成更友好的生命周期错误信息。
type Namer interface {
	Name() string
}
