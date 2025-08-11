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
	router     *gin.Engine
	wsHandler  *handlers.WebSocketHandler
	httpServer *http.Server
	tlsServer  *http.Server
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

	// 认证相关仓储层
	userRepo := repository.NewUserRepository(database.GetDB())
	apiKeyRepo := repository.NewAPIKeyRepository(database.GetDB())
	requestLogRepo := repository.NewRequestLogRepository(database.GetDB())

	// 创建服务层
	blockService := services.NewBlockService(blockRepo)
	txService := services.NewTransactionService(txRepo)
	addressService := services.NewAddressService(addressRepo)
	assetService := services.NewAssetService(assetRepo)
	coinConfigService := services.NewCoinConfigService(coinConfigRepo)
	baseConfigService := services.NewBaseConfigService(baseConfigRepo)

	// 认证服务
	authService := services.NewAuthService(
		userRepo,
		apiKeyRepo,
		requestLogRepo,
		config.AppConfig.Security.JWTSecret,
		config.AppConfig.Security.JWTExpiration,
	)

	// 创建处理器
	blockHandler := handlers.NewBlockHandler(blockService)
	txHandler := handlers.NewTransactionHandler(txService)
	wsHandler := handlers.NewWebSocketHandler()
	addressHandler := handlers.NewAddressHandler(addressService)
	assetHandler := handlers.NewAssetHandler(assetService)
	coinConfigHandler := handlers.NewCoinConfigHandler(coinConfigService)
	scannerHandler := handlers.NewScannerHandler(baseConfigService)
	authHandler := handlers.NewAuthHandler(authService)

	// 启动WebSocket处理器
	wsHandler.Start()

	// 设置路由
	router := routes.SetupRoutes(blockHandler, txHandler, wsHandler, addressHandler, assetHandler, coinConfigHandler, scannerHandler, authHandler, authService, apiKeyRepo, requestLogRepo)

	// 创建HTTP服务器
	httpAddr := fmt.Sprintf("%s:%d", config.AppConfig.Server.Host, config.AppConfig.Server.Port)
	httpServer := &http.Server{
		Addr:         httpAddr,
		Handler:      router,
		ReadTimeout:  config.AppConfig.Server.ReadTimeout,
		WriteTimeout: config.AppConfig.Server.WriteTimeout,
	}

	var tlsServer *http.Server
	if config.AppConfig.Server.TLSEnabled {
		tlsAddr := fmt.Sprintf("%s:%d", config.AppConfig.Server.Host, config.AppConfig.Server.TLSPort)
		tlsServer = &http.Server{
			Addr:         tlsAddr,
			Handler:      router,
			ReadTimeout:  config.AppConfig.Server.ReadTimeout,
			WriteTimeout: config.AppConfig.Server.WriteTimeout,
		}
	}

	return &Server{
		router:     router,
		wsHandler:  wsHandler,
		httpServer: httpServer,
		tlsServer:  tlsServer,
	}
}

// Start 启动服务器
func (s *Server) Start() error {
	log.Printf("Server configuration: %+v", config.AppConfig.Server)

	// 如果启用了TLS，同时启动HTTP和HTTPS服务器
	if s.tlsServer != nil {
		// 启动HTTPS服务器
		go func() {
			log.Printf("Starting HTTPS server on %s", s.tlsServer.Addr)
			if err := s.tlsServer.ListenAndServeTLS(
				config.AppConfig.Server.CertFile,
				config.AppConfig.Server.KeyFile,
			); err != nil && err != http.ErrServerClosed {
				log.Fatalf("HTTPS server failed: %v", err)
			}
		}()

		// 启动HTTP服务器（用于重定向到HTTPS）
		log.Printf("Starting HTTP server on %s (redirects to HTTPS)", s.httpServer.Addr)
		return s.httpServer.ListenAndServe()
	}

	// 只启动HTTP服务器
	log.Printf("Starting HTTP server on %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

// Shutdown 优雅关闭服务器
func (s *Server) Shutdown(timeout time.Duration) error {
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 关闭HTTP服务器
	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	// 关闭HTTPS服务器（如果存在）
	if s.tlsServer != nil {
		if err := s.tlsServer.Shutdown(ctx); err != nil {
			log.Printf("HTTPS server shutdown error: %v", err)
		}
	}

	return nil
}

// GetRouter 获取路由实例（用于测试）
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}
