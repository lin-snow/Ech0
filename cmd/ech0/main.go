package main

import "github.com/lin-snow/ech0/internal/server"

func main() {
	// 创建Server
	s := server.New()

	// 初始化Server
	s.Init()

	// 启动Server
	s.Start()
}
