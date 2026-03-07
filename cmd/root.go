package cmd

import (
	"os"

	"github.com/lin-snow/ech0/internal/app"
	"github.com/lin-snow/ech0/internal/cli"
	"github.com/spf13/cobra"
)

// rootCmd 是 Ech0 的根命令
// 默认启动CLI With TUI
var rootCmd = &cobra.Command{
	Use:   "ech0",
	Short: "面向个人的新一代开源、自托管、专注思想流动的轻量级发布平台",
	Long:  `面向个人的新一代开源、自托管、专注思想流动的轻量级发布平台`,

	// 这个 Run 会在没有子命令时执行
	Run: func(cmd *cobra.Command, args []string) {
		cli.DoTui()
	},
}

// tuiCmd 是启动 Ech0 TUI 的命令
var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "启动 Ech0 TUI",
	Run: func(cmd *cobra.Command, args []string) {
		cli.DoTui()
	},
}

// versionCmd 是查看当前版本信息的命令
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "查看当前版本信息",
	Run: func(cmd *cobra.Command, args []string) {
		cli.DoVersion()
	},
}

// infoCmd 是查看当前信息的命令
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "查看当前信息",
	Run: func(cmd *cobra.Command, args []string) {
		cli.DoEch0Info()
	},
}

// helloCmd 是输出 Ech0 Logo 的命令
var helloCmd = &cobra.Command{
	Use:   "hello",
	Short: "输出 Ech0 Logo",
	Run: func(cmd *cobra.Command, args []string) {
		cli.DoHello()
	},
}

// init 函数用于初始化根命令和子命令
func init() {
	// 解决Windows下使用 Cobra 触发 mousetrap 提示
	cobra.MousetrapHelpText = ""
	rootCmd.AddCommand(tuiCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(helloCmd)
}

// Bootstrap 注入应用工厂，按需装配 Web 生命周期应用。
func Bootstrap(build func() (*app.App, func(), error)) {
	cli.SetAppFactory(build)
}

// Execute 是根命令的入口函数
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
