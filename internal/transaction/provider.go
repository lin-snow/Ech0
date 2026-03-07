package transaction

import "github.com/google/wire"

var TransactorSet = wire.NewSet(
	NewGormTransactor,
	wire.Bind(new(Transactor), new(*GormTransactor)),
)

var ProviderSet = TransactorSet
