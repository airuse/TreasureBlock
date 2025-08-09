package routes

import (
	"blockChainBrowser/server/internal/handlers"

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
	scannerHandler *handlers.ScannerHandler,
) *gin.Engine {
	router := gin.Default()

	// 添加CORS中间件
	router.Use(corsMiddleware())

	// API版本分组
	v1 := router.Group("/api/v1")
	{
		// 区块相关路由
		blocks := v1.Group("/blocks")
		{
			blocks.GET("", blockHandler.ListBlocks)                      // 获取区块列表
			blocks.GET("/latest", blockHandler.GetLatestBlock)           // 获取最新区块
			blocks.GET("/hash/:hash", blockHandler.GetBlockByHash)       // 根据哈希获取区块
			blocks.GET("/height/:height", blockHandler.GetBlockByHeight) // 根据高度获取区块
			blocks.POST("/create", blockHandler.CreateBlock)             // 创建区块
		}

		// 交易相关路由
		transactions := v1.Group("/transactions")
		{
			transactions.GET("", txHandler.ListTransactions)                            // 获取交易列表
			transactions.GET("/hash/:hash", txHandler.GetTransactionByHash)             // 根据哈希获取交易
			transactions.GET("/address/:address", txHandler.GetTransactionsByAddress)   // 根据地址获取交易
			transactions.GET("/block/:blockHash", txHandler.GetTransactionsByBlockHash) // 根据区块哈希获取交易
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
			coinConfigs.POST("/:symbol", coinConfigHandler.CreateCoinConfig)
			coinConfigs.GET("/:symbol", coinConfigHandler.GetCoinConfigBySymbol)
		}
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
