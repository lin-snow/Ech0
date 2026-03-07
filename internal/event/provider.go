package event

import (
	"sync"

	"github.com/google/wire"
)

type RegisteredRegistrar struct {
	Registrar *EventRegistrar
}

func ProvideEventBusProvider() func() IEventBus {
	var once sync.Once
	return func() IEventBus {
		once.Do(InitEventBus)
		return GetEventBus()
	}
}

func ProvideRegisteredRegistrar(registrar *EventRegistrar) (*RegisteredRegistrar, func(), error) {
	if err := registrar.Register(); err != nil {
		return nil, nil, err
	}
	return &RegisteredRegistrar{Registrar: registrar}, func() {
		_ = registrar.Stop()
	}, nil
}

var ProviderSet = wire.NewSet(ProvideEventBusProvider)
