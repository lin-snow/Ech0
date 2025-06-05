package util

import (
	model "github.com/lin-snow/ech0/internal/model/common"
	util "github.com/lin-snow/ech0/internal/util/log"
	"go.uber.org/zap"
)

func HandleError(se *model.ServerError) string {
	if se.Err != nil {
		if se.Msg == "" || len(se.Msg) == 0 {
			se.Msg = se.Err.Error()
		}
		util.Logger.Error(
			se.Msg,
			zap.Error(se.Err),
		)
	}

	return se.Msg
}

func HandlePanicError(se *model.ServerError) {
	if se.Err != nil {
		if se.Msg == "" || len(se.Msg) == 0 {
			se.Msg = se.Err.Error()
		}
		util.Logger.Panic(
			se.Msg,
			zap.Error(se.Err),
		)
	}

	panic(se.Msg)
}
