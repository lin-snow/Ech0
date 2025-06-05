package util

import (
	model "github.com/lin-snow/ech0/internal/model/common"
	"go.uber.org/zap"
)

var Logger *zap.Logger

func InitLogger() {
	var err error
	Logger, err = zap.NewProduction()
	if err != nil {
		panic(model.INIT_LOGGER_PANIC + ": " + err.Error())
	}
}
