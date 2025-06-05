package util

import (
	model "github.com/lin-snow/ech0/internal/model/common"
	util "github.com/lin-snow/ech0/internal/util/log"
	"go.uber.org/zap"
)

func HandleError(se *model.ServerError) *model.ServerError {
	if se.Err != nil {
		util.Logger.Error(
			se.Msg,
			zap.Error(se.Err),
		)
	}

	return se
}

func HandlePanicError(se *model.ServerError) {
	if se.Err != nil {
		util.Logger.Panic(
			se.Msg,
			zap.Error(se.Err),
		)
	}

	panic(se.Msg)
}
