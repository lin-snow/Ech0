package transaction

import (
	"context"
)

type contextKey string

const TxKey contextKey = "tx"

// Transactor 定义事务执行器接口
type Transactor interface {
	// Run 执行一个事务
	Run(ctx context.Context, fn func(ctx context.Context) error) error
}
