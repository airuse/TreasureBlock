package main

import (
	"log"

	"blockChainBrowser/server/config"
	"blockChainBrowser/server/internal/server"
)

func main() {
	// 加载配置
	if err := config.Load(); err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// 创建并启动服务器
	srv := server.New()
	if err := srv.Start(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
