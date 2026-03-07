package task

import "github.com/google/wire"

var ProviderSet = wire.NewSet(NewTasker)
