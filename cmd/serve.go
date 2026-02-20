package cmd

import (
	"github.com/lin-snow/ech0/internal/cli"
	"github.com/spf13/cobra"
)

// serveCmd 是启动 Web 和 SSH 服务的命令
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "启动 Web 和 SSH 服务",
	Run: func(cmd *cobra.Command, args []string) {
		cli.DoServeWithSSHAndBlock()
	},
}

// webCmd 是仅启动 Web 服务并阻塞的命令
var webCmd = &cobra.Command{
	Use:   "web",
	Short: "仅启动 Web 服务（阻塞）",
	Run: func(cmd *cobra.Command, args []string) {
		cli.DoServeWithBlock()
	},
}

// init 函数用于初始化根命令和子命令
func init() {
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(webCmd)
}
