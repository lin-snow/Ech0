package main

import "github.com/lin-snow/ech0/internal/server"

func main() {
	// 创建Server
	s := server.New()

	// 初始化Server
	server.Init(s)

	// 启动Server
	server.Start(s)
}
