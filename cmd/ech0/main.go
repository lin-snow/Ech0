package main

import (
	"github.com/lin-snow/ech0/cmd"
	"github.com/lin-snow/ech0/internal/bootstrap"
	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/internal/di"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
)

func main() {
	bootstrap.Bootstrap()
	logUtil.InitLogger()
	config.Config()

	cmd.Bootstrap(di.BuildWebApp)
	cmd.Execute()
}
