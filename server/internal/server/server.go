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
	"blockChainBrowser/server/internal/utils"

	"github.com/gin-gonic/gin"
)

// Server 服务器结构
type Server struct {
	router                     *gin.Engine
	wsHandler                  *handlers.WebSocketHandler
	httpServer                 *http.Server
	tlsServer                  *http.Server
	transactionStatusScheduler *services.TransactionStatusScheduler
	feeScheduler               *services.FeeScheduler
	dataCleanupScheduler       *services.DataCleanupScheduler
}

// New 创建服务器实例
func New() *Server {
	// 初始化数据库
	if err := database.Init(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// 创建仓库
	blockRepo := repository.NewBlockRepository()
	txRepo := repository.NewTransactionRepository()
	addressRepo := repository.NewAddressRepository()
	assetRepo := repository.NewAssetRepository()
	baseConfigRepo := repository.NewBaseConfigRepository()
	coinConfigRepo := repository.NewCoinConfigRepository()
	contractRepo := repository.NewContractRepository(database.GetDB())
	userRepo := repository.NewUserRepository(database.GetDB())
	apiKeyRepo := repository.NewAPIKeyRepository(database.GetDB())
	requestLogRepo := repository.NewRequestLogRepository(database.GetDB())
	userAddressRepo := repository.NewUserAddressRepository(database.GetDB())
	parserConfigRepo := repository.NewParserConfigRepository(database.GetDB())
	statsRepo := repository.NewStatsRepository()
	earningsRepo := repository.NewEarningsRepository(database.GetDB())
	userBalanceRepo := repository.NewUserBalanceRepository(database.GetDB())
	btcUTXORepo := repository.NewBTCUTXORepository(database.GetDB())

	// RPC 管理器与合约调用服务
	rpcManager := utils.NewRPCClientManager()
	contractCallService := services.NewContractCallService(rpcManager)

	// 创建解析配置服务
	parserConfigService := services.NewParserConfigService(parserConfigRepo)

	// 创建服务
	blockService := services.NewBlockService(blockRepo)
	txService := services.NewTransactionService(txRepo, coinConfigRepo, btcUTXORepo)
	addressService := services.NewAddressService(addressRepo)
	assetService := services.NewAssetService(assetRepo)
	baseConfigService := services.NewBaseConfigService(baseConfigRepo)
	coinConfigService := services.NewCoinConfigService(coinConfigRepo)
	contractService := services.NewContractService(contractRepo, coinConfigRepo)
	btcUTXOService := services.NewBTCUTXOService(btcUTXORepo)
	// 权限服务依赖
	roleRepo := repository.NewRoleRepository(database.GetDB())
	permissionRepo := repository.NewPermissionRepository(database.GetDB())
	permissionService := services.NewPermissionService(userRepo, roleRepo, permissionRepo)
	authService := services.NewAuthService(
		userRepo,
		apiKeyRepo,
		requestLogRepo,
		permissionService,
		config.AppConfig.Security.JWTSecret,
		config.AppConfig.Security.JWTExpiration,
	)
	userAddressService := services.NewUserAddressService(userAddressRepo, blockRepo, contractRepo, contractCallService, btcUTXOService, baseConfigRepo)
	statsService := services.NewStatsService(statsRepo)

	// 创建收益相关服务
	earningsService := services.NewEarningsService(earningsRepo, userBalanceRepo)

	// 创建交易凭证相关服务
	transactionReceiptRepo := repository.NewTransactionReceiptRepository(database.GetDB())
	transactionReceiptService := services.NewTransactionReceiptService(transactionReceiptRepo)

	// 创建区块验证服务
	blockVerificationService := services.NewBlockVerificationService(
		blockRepo,
		txRepo,
		transactionReceiptRepo,
		coinConfigRepo,
	)
	contractParseResultRepo := repository.NewContractParseResultRepository()
	contractParseResultService := services.NewContractParseService(contractParseResultRepo, transactionReceiptRepo, txRepo, blockRepo, parserConfigRepo, coinConfigRepo, userAddressRepo)

	// 创建处理器
	txHandler := handlers.NewTransactionHandler(txService, transactionReceiptService, parserConfigRepo, blockVerificationService, contractParseResultService, coinConfigService, userAddressService)
	wsHandler := handlers.NewWebSocketHandler()
	blockHandler := handlers.NewBlockHandler(blockService, wsHandler)
	addressHandler := handlers.NewAddressHandler(addressService)
	assetHandler := handlers.NewAssetHandler(assetService)
	coinConfigHandler := handlers.NewCoinConfigHandler(coinConfigService, parserConfigService)
	contractHandler := handlers.NewContractHandler(contractService)
	scannerHandler := handlers.NewScannerHandler(baseConfigService)
	authHandler := handlers.NewAuthHandler(authService)
	userAddressHandler := handlers.NewUserAddressHandler(userAddressService)
	baseConfigHandler := handlers.NewBaseConfigHandler(baseConfigService)
	parserConfigHandler := handlers.NewParserConfigHandler(parserConfigService)
	homeHandler := handlers.NewHomeHandler(blockService, txService, statsService)
	earningsHandler := handlers.NewEarningsHandler(earningsService)
	userTransactionHandler := handlers.NewUserTransactionHandler()
	contractParseResultHandler := handlers.NewContractParseResultHandler(contractParseResultService)
	blockVerificationHandler := handlers.NewBlockVerificationHandler(blockVerificationService, earningsService, contractParseResultService, btcUTXOService, txService)

	// 启动WebSocket处理器
	wsHandler.Start()

	// 创建并启动交易状态调度器
	transactionStatusScheduler := services.NewTransactionStatusScheduler()
	transactionStatusScheduler.SetWebSocketHandler(wsHandler)
	go transactionStatusScheduler.Start(context.Background())

	// 创建并启动费率调度器
	feeScheduler := services.NewFeeScheduler()
	feeScheduler.SetWebSocketHandler(wsHandler)
	go feeScheduler.Start(context.Background())

	// 创建并启动数据清理调度器
	dataCleanupScheduler := services.NewDataCleanupScheduler(
		database.GetDB(),
		userAddressRepo,
		blockRepo,
		txRepo,
		transactionReceiptRepo,
		contractParseResultRepo,
	)
	go dataCleanupScheduler.Start(context.Background())

	// 创建Gas费率处理器
	gasHandler := handlers.NewGasHandler(feeScheduler)

	// 设置路由
	router := routes.SetupRoutes(
		blockHandler,
		txHandler,
		wsHandler,
		addressHandler,
		assetHandler,
		coinConfigHandler,
		contractHandler,
		parserConfigHandler,
		scannerHandler,
		authHandler,
		userAddressHandler,
		userTransactionHandler,
		baseConfigHandler,
		homeHandler,
		earningsHandler,
		gasHandler,
		authService,
		apiKeyRepo,
		requestLogRepo,
		earningsService,
		config.AppConfig.Server.TLSEnabled, // 传递TLS配置
		contractParseResultHandler,
		blockVerificationHandler,
	)

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
		router:                     router,
		wsHandler:                  wsHandler,
		httpServer:                 httpServer,
		tlsServer:                  tlsServer,
		transactionStatusScheduler: transactionStatusScheduler,
		feeScheduler:               feeScheduler,
		dataCleanupScheduler:       dataCleanupScheduler,
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

	// 停止交易状态调度器
	if s.transactionStatusScheduler != nil {
		s.transactionStatusScheduler.Stop()
	}

	// 停止数据清理调度器
	if s.dataCleanupScheduler != nil {
		s.dataCleanupScheduler.Stop()
	}

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
