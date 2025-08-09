package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"blockChainBrowser/client/scanner/config"
	"blockChainBrowser/client/scanner/internal/scanner"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	configFile string
	startBlock uint64
	chain      string
	interval   time.Duration
)

// rootCmd 根命令
var rootCmd = &cobra.Command{
	Use:   "block-scanner",
	Short: "Blockchain block scanner tool",
	Long: `A powerful blockchain block scanner that supports:
- Multiple blockchain networks (Bitcoin, Ethereum)
- Continuous block scanning with configurable intervals
- Automatic retry mechanisms and error handling
- Progress tracking and monitoring
- File output and server submission
- Graceful shutdown with Ctrl+C`,
	RunE: run,
}

// init 初始化命令
func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Configuration file path")
	rootCmd.PersistentFlags().Uint64VarP(&startBlock, "start", "s", 0, "Start block height (0 = from config or resume)")
	rootCmd.PersistentFlags().StringVarP(&chain, "chain", "n", "", "Specific chain to scan (btc, eth, or empty for all)")
	rootCmd.PersistentFlags().DurationVarP(&interval, "interval", "i", 0, "Scanning interval (0 = use config default)")
}

// run 运行主程序
func run(cmd *cobra.Command, args []string) error {
	// 加载配置
	if err := config.Load(configFile); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// 应用命令行参数
	if startBlock > 0 {
		config.AppConfig.Scan.StartBlockHeight = startBlock
	}
	// 如果指定了间隔，则使用命令行参数
	if interval > 0 {
		config.AppConfig.Scan.Interval = interval
	}
	// 如果指定了链，则使用命令行参数
	if chain != "" {
		logrus.Infof("Will scan specific chain: %s", chain)
	}

	// 设置日志格式
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)

	logrus.Info("Starting blockchain block scanner...")

	// 创建扫块器
	blockScanner := scanner.NewBlockScanner(config.AppConfig)

	// 如果配置了自动启动，则启动扫描器
	if config.AppConfig.Scan.AutoStart {
		logrus.Info("Auto-starting block scanner...")
		if err := blockScanner.Start(); err != nil {
			logrus.Errorf("Failed to start block scanner: %v", err)
			return err
		}
	}

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down block scanner...")

	// 停止扫描器
	blockScanner.Stop()

	// 等待一段时间确保所有操作完成
	time.Sleep(2 * time.Second)

	logrus.Info("Block scanner stopped")

	return nil
}

// main 主函数
func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
