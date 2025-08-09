package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"blockChainBrowser/server/config"
	"blockChainBrowser/server/internal/database"
	"blockChainBrowser/server/internal/handlers"
	"blockChainBrowser/server/internal/repository"
	"blockChainBrowser/server/internal/routes"
	"blockChainBrowser/server/internal/services"

	"github.com/gin-gonic/gin"
)

// Server 服务器结构
type Server struct {
	router    *gin.Engine
	wsHandler *handlers.WebSocketHandler
	server    *http.Server
}

// New 创建服务器实例
func New() *Server {
	// 初始化数据库
	if err := database.Init(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// 创建仓储层
	blockRepo := repository.NewBlockRepository()
	txRepo := repository.NewTransactionRepository()
	addressRepo := repository.NewAddressRepository()
	assetRepo := repository.NewAssetRepository()
	coinConfigRepo := repository.NewCoinConfigRepository()
	baseConfigRepo := repository.NewBaseConfigRepository()
	// 创建服务层
	blockService := services.NewBlockService(blockRepo)
	txService := services.NewTransactionService(txRepo)
	addressService := services.NewAddressService(addressRepo)
	assetService := services.NewAssetService(assetRepo)
	coinConfigService := services.NewCoinConfigService(coinConfigRepo)
	baseConfigService := services.NewBaseConfigService(baseConfigRepo)

	// 创建处理器
	blockHandler := handlers.NewBlockHandler(blockService)
	txHandler := handlers.NewTransactionHandler(txService)
	wsHandler := handlers.NewWebSocketHandler()
	addressHandler := handlers.NewAddressHandler(addressService)
	assetHandler := handlers.NewAssetHandler(assetService)
	coinConfigHandler := handlers.NewCoinConfigHandler(coinConfigService)
	scannerHandler := handlers.NewScannerHandler(baseConfigService)

	// 启动WebSocket处理器
	wsHandler.Start()

	// 设置路由
	router := routes.SetupRoutes(blockHandler, txHandler, wsHandler, addressHandler, assetHandler, coinConfigHandler, scannerHandler)

	// 创建HTTP服务器
	addr := fmt.Sprintf("%s:%d", config.AppConfig.Server.Host, config.AppConfig.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  config.AppConfig.Server.ReadTimeout,
		WriteTimeout: config.AppConfig.Server.WriteTimeout,
	}

	return &Server{
		router:    router,
		wsHandler: wsHandler,
		server:    server,
	}
}

// Start 启动服务器
func (s *Server) Start() error {
	log.Printf("Starting server on %s", s.server.Addr)
	log.Printf("Server configuration: %+v", config.AppConfig.Server)

	return s.server.ListenAndServe()
}

// Shutdown 优雅关闭服务器
func (s *Server) Shutdown(timeout time.Duration) error {
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return s.server.Shutdown(ctx)
}

// GetRouter 获取路由实例（用于测试）
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}
