package routes

import (
	"time"

	"blockChainBrowser/server/internal/database"
	"blockChainBrowser/server/internal/handlers"
	"blockChainBrowser/server/internal/middleware"
	"blockChainBrowser/server/internal/repository"
	"blockChainBrowser/server/internal/services"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置路由
func SetupRoutes(
	blockHandler *handlers.BlockHandler,
	txHandler *handlers.TransactionHandler,
	wsHandler *handlers.WebSocketHandler,
	addressHandler *handlers.AddressHandler,
	assetHandler *handlers.AssetHandler,
	coinConfigHandler *handlers.CoinConfigHandler,
	contractHandler *handlers.ContractHandler,
	parserConfigHandler *handlers.ParserConfigHandler,
	scannerHandler *handlers.ScannerHandler,
	authHandler *handlers.AuthHandler,
	userAddressHandler *handlers.UserAddressHandler,
	baseConfigHandler *handlers.BaseConfigHandler,
	homeHandler *handlers.HomeHandler,
	authService services.AuthService,
	apiKeyRepo repository.APIKeyRepository,
	requestLogRepo repository.RequestLogRepository,
	tlsEnabled bool, // 添加TLS配置参数
) *gin.Engine {
	router := gin.Default()

	// 添加安全相关中间件
	router.Use(middleware.SecurityHeadersMiddleware())
	router.Use(middleware.RequestSizeLimitMiddleware(10 * 1024 * 1024)) // 10MB限制
	router.Use(middleware.HTTPSRedirectMiddleware(tlsEnabled))          // 根据配置启用HTTPS重定向
	router.Use(corsMiddleware())

	// 创建暴力破解防护器
	bruteForceProtector := middleware.NewBruteForceProtector()
	router.Use(bruteForceProtector.LoginAttemptMiddleware())

	// 创建认证和限流中间件实例
	jwtAuthMiddleware := middleware.JWTAuthMiddleware(authService)
	rateLimitMiddleware := middleware.RateLimitMiddleware(apiKeyRepo, requestLogRepo)
	requestLogMiddleware := middleware.RequestLogMiddleware(requestLogRepo)

	// 创建区块验证服务
	blockVerificationService := services.NewBlockVerificationService(
		repository.NewBlockRepository(),
		repository.NewTransactionRepository(),
		repository.NewTransactionReceiptRepository(database.GetDB()),
		repository.NewCoinConfigRepository(),
	)

	// 公开API路由（不需要认证）
	api := router.Group("/api")
	{
		// 认证相关路由
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)                       // 用户注册
			auth.POST("/login", authHandler.Login)                             // 用户登录
			auth.POST("/token", authHandler.GetAccessToken)                    // 获取访问令牌
			auth.POST("/refresh", jwtAuthMiddleware, authHandler.RefreshToken) // 刷新令牌（需要登录令牌）
		}

		// 权限相关路由（公开）
		api.GET("/permissions", baseConfigHandler.GetPermissionTypes) // 获取权限类型列表

		// 基础配置相关路由（公开）
		api.GET("/base-configs/group/:group", baseConfigHandler.GetConfigsByGroup) // 根据分组获取配置
		api.GET("/base-configs/type/:type", baseConfigHandler.GetConfigsByType)    // 根据类型获取配置
	}

	// 需要用户认证的路由（使用登录令牌）
	userAPI := api.Group("/user")
	userAPI.Use(jwtAuthMiddleware) // 使用JWT认证中间件
	{
		userAPI.GET("/profile", authHandler.GetProfile)              // 获取用户资料
		userAPI.POST("/change-password", authHandler.ChangePassword) // 修改密码

		// API密钥管理
		apiKeys := userAPI.Group("/api-keys")
		{
			apiKeys.POST("", authHandler.CreateAPIKey)           // 创建API密钥
			apiKeys.GET("", authHandler.GetAPIKeys)              // 获取API密钥列表
			apiKeys.PUT("/:id", authHandler.UpdateAPIKey)        // 更新API密钥
			apiKeys.DELETE("/:id", authHandler.DeleteAPIKey)     // 删除API密钥
			apiKeys.GET("/:id/stats", authHandler.GetUsageStats) // 获取使用统计
		}

		// 用户地址管理
		addresses := userAPI.Group("/addresses")
		{
			addresses.POST("", userAddressHandler.CreateAddress)       // 创建地址
			addresses.GET("", userAddressHandler.GetUserAddresses)     // 获取地址列表
			addresses.GET("/:id", userAddressHandler.GetAddressByID)   // 获取地址详情
			addresses.PUT("/:id", userAddressHandler.UpdateAddress)    // 更新地址
			addresses.DELETE("/:id", userAddressHandler.DeleteAddress) // 删除地址
		}
	}

	// 创建公开API限流器（每小时100次请求）
	publicRateLimiter := middleware.NewPublicRateLimiter(100, time.Hour)

	// 不需要JWT认证的API路由，仅限查看区块等信息（带限流保护）
	noAuthAPI := api.Group("/no-auth")
	noAuthAPI.Use(publicRateLimiter.PublicRateLimitMiddleware()) // 添加限流中间件
	{
		noAuthAPI.GET("/blocks", blockHandler.ListBlocksPublic)                                                // 限制最多20个区块
		noAuthAPI.GET("/blocks/height/:height", blockHandler.GetBlockByHeightPublic)                           // 根据高度获取区块详情（公开接口）
		noAuthAPI.GET("/blocks/search", blockHandler.SearchBlocksPublic)                                       // 搜索区块（公开接口）
		noAuthAPI.GET("/transactions", txHandler.ListTransactionsPublic)                                       // 限制最多1000条交易
		noAuthAPI.GET("/transactions/block-height/:blockHeight", txHandler.GetTransactionsByBlockHeightPublic) // 根据区块高度获取交易（公开接口）
	}

	// 区块验证相关路由
	blockVerificationHandler := handlers.NewBlockVerificationHandler(blockVerificationService)

	// 需要AccessToken认证的API路由（区块链数据API）
	v1 := api.Group("/v1")
	v1.Use(jwtAuthMiddleware)    // JWT认证
	v1.Use(rateLimitMiddleware)  // 限流
	v1.Use(requestLogMiddleware) // 请求日志
	{
		// 首页相关路由
		home := v1.Group("/home")
		{
			home.GET("/stats", homeHandler.GetHomeStats) // 获取首页统计数据
		}

		// 区块相关路由
		blocks := v1.Group("/blocks")
		blocks.Use(jwtAuthMiddleware) // Add auth middleware here
		{
			blocks.GET("", blockHandler.ListBlocks)                      // 获取区块列表
			blocks.GET("/latest", blockHandler.GetLatestBlock)           // 获取最新区块
			blocks.GET("/hash/:hash", blockHandler.GetBlockByHash)       // 根据哈希获取区块
			blocks.GET("/height/:height", blockHandler.GetBlockByHeight) // 根据高度获取区块
			blocks.GET("/search", blockHandler.SearchBlocks)             // 搜索区块（认证用户）
			blocks.POST("/create", blockHandler.CreateBlock)             // 创建区块
			blocks.POST("/update", blockHandler.UpdateBlock)             // 更新区块

			// 区块验证相关路由 - 需要认证
			blocks.GET("/verification/last-verified", blockVerificationHandler.GetLastVerifiedBlockHeight)
			blocks.POST("/:blockID/verify", blockVerificationHandler.VerifyBlock)
		}

		// 交易相关路由
		transactions := v1.Group("/transactions")
		{
			transactions.GET("", txHandler.ListTransactions)                                       // 获取交易列表
			transactions.GET("/hash/:hash", txHandler.GetTransactionByHash)                        // 根据哈希获取交易
			transactions.GET("/address/:address", txHandler.GetTransactionsByAddress)              // 根据地址获取交易
			transactions.GET("/block-hash/:blockHash", txHandler.GetTransactionsByBlockHash)       // 根据区块哈希获取交易
			transactions.GET("/block-height/:blockHeight", txHandler.GetTransactionsByBlockHeight) // 根据区块高度获取交易
			transactions.POST("/create", txHandler.CreateTransaction)                              // 创建交易记录
			transactions.POST("/create/batch", txHandler.CreateTransactionsBatch)                  // 批量创建交易记录
			transactions.GET("/receipt/:hash", txHandler.GetTransactionReceiptByHash)              // 根据哈希获取交易凭证
		}

		// 地址相关路由
		addresses := v1.Group("/addresses")
		{
			addresses.GET("/:address/transactions", txHandler.GetTransactionsByAddress) // 根据地址获取交易
		}

		// 资产相关路由
		assets := v1.Group("/assets")
		{
			assets.POST("/:address", assetHandler.CreateAsset)
		}

		// 币种配置相关路由
		coinConfigs := v1.Group("/coin-configs")
		{
			coinConfigs.POST("", coinConfigHandler.CreateCoinConfig)                                        // 创建币种配置
			coinConfigs.GET("", coinConfigHandler.ListCoinConfigs)                                          // 分页获取币种配置列表
			coinConfigs.GET("/all", coinConfigHandler.GetAllCoinConfigs)                                    // 获取所有币种配置
			coinConfigs.GET("/scanner", coinConfigHandler.GetCoinConfigsForScanner)                         // 扫块程序专用接口（兼容性）
			coinConfigs.GET("/maintenance/:contractAddress", coinConfigHandler.GetCoinConfigForMaintenance) // 获取币种配置信息（维护用，包含解析配置）
			coinConfigs.GET("/contract/:contractAddress", coinConfigHandler.GetCoinConfigByContractAddress) // 根据合约地址获取币种配置
			coinConfigs.GET("/chain/:chain", coinConfigHandler.GetCoinConfigsByChain)                       // 根据链名称获取币种配置
			coinConfigs.GET("/chain/:chain/stablecoins", coinConfigHandler.GetStablecoins)                  // 获取指定链的稳定币
			coinConfigs.GET("/chain/:chain/verified", coinConfigHandler.GetVerifiedTokens)                  // 获取指定链的已验证代币
			coinConfigs.GET("/id/:id", coinConfigHandler.GetCoinConfigByID)                                 // 根据ID获取币种配置
			coinConfigs.GET("/symbol/:symbol", coinConfigHandler.GetCoinConfigBySymbol)                     // 根据符号获取币种配置
			coinConfigs.PUT("/:id", coinConfigHandler.UpdateCoinConfig)                                     // 更新币种配置
			coinConfigs.DELETE("/:id", coinConfigHandler.DeleteCoinConfig)                                  // 删除币种配置
		}

		// 合约相关路由
		contracts := v1.Group("/contracts")
		{
			contracts.POST("", contractHandler.CreateOrUpdateContract)                      // 创建或更新合约
			contracts.GET("/chain/:chainName", contractHandler.GetContractsByChain)         // 根据链名称获取合约列表
			contracts.GET("/type/:type", contractHandler.GetContractsByType)                // 根据合约类型获取合约列表
			contracts.GET("/erc20", contractHandler.GetERC20Tokens)                         // 获取所有ERC-20代币合约
			contracts.GET("", contractHandler.GetAllContracts)                              // 获取所有合约
			contracts.GET("/address/:address", contractHandler.GetContractByAddress)        // 根据地址获取合约
			contracts.PUT("/:address/status/:status", contractHandler.UpdateContractStatus) // 更新合约状态
			contracts.PUT("/:address/verify", contractHandler.VerifyContract)               // 验证合约
			contracts.DELETE("/:address", contractHandler.DeleteContract)                   // 删除合约
		}

		// 解析配置相关路由
		parserConfigs := v1.Group("/parser-configs")
		{
			parserConfigs.POST("", parserConfigHandler.CreateParserConfig)                                                       // 创建解析配置
			parserConfigs.GET("", parserConfigHandler.ListParserConfigs)                                                         // 分页获取解析配置列表
			parserConfigs.GET("/:id", parserConfigHandler.GetParserConfigByID)                                                   // 根据ID获取解析配置
			parserConfigs.GET("/contract/:contractAddress", parserConfigHandler.GetParserConfigsByContract)                      // 根据合约地址获取解析配置
			parserConfigs.GET("/contract/:contractAddress/signature/:signature", parserConfigHandler.GetParserConfigBySignature) // 根据合约地址和签名获取解析配置
			parserConfigs.GET("/contract/:contractAddress/info", parserConfigHandler.GetContractParserInfo)                      // 获取合约完整解析信息（三表联查）
			parserConfigs.POST("/contract/:contractAddress/parse", parserConfigHandler.ParseInputData)                           // 解析交易输入数据
			parserConfigs.PUT("/:id", parserConfigHandler.UpdateParserConfig)                                                    // 更新解析配置
			parserConfigs.DELETE("/:id", parserConfigHandler.DeleteParserConfig)                                                 // 删除解析配置
		}

		// 扫描器相关路由
		scanner := v1.Group("/scanner")
		{
			scanner.GET("/getconfig", scannerHandler.GetScannerConfig)
		}
	}

	// WebSocket路由
	router.GET("/ws", wsHandler.HandleWebSocket)

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"success": true,
			"message": "Blockchain Browser API is running",
		})
	})

	return router
}

// corsMiddleware CORS中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
