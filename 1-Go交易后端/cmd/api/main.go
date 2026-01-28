package main

// @title Exchange API
// @version 1.0
// @description OKCoinsgp数字货币交易所API接口文档
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Token格式："Bearer {token}"

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"exchange-go/config"
	"exchange-go/internal/api/router"
	"exchange-go/internal/pkg/cache"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/huobi"
	"exchange-go/internal/pkg/logger"
	"exchange-go/internal/scheduler"
	"exchange-go/internal/service"
	"exchange-go/internal/websocket"

	"github.com/gin-gonic/gin"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "c", "config/config.yaml", "配置文件路径")
}

func main() {
	flag.Parse()

	// 初始化配置
	if err := config.Init(configPath); err != nil {
		fmt.Printf("初始化配置失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ 配置加载成功")

	// 初始化日志
	if err := logger.Init(&config.GlobalConfig.Log); err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()
	fmt.Println("✓ 日志系统初始化成功")

	// 初始化数据库
	if err := database.Init(&config.GlobalConfig.Database); err != nil {
		logger.Errorf("初始化数据库失败: %v", err)
		os.Exit(1)
	}
	defer database.Close()
	fmt.Println("✓ 数据库连接成功")

	// 初始化Redis（可选，失败不阻止服务启动）
	if err := cache.Init(&config.GlobalConfig.Redis); err != nil {
		logger.Warnf("初始化Redis失败: %v (部分缓存功能将不可用)", err)
		fmt.Println("⚠ Redis连接失败，部分缓存功能将不可用")
	} else {
		defer cache.Close()
		fmt.Println("✓ Redis连接成功")
	}

	// 启动火币WebSocket数据源
	huobiManager := huobi.GetManager()
	if err := huobiManager.Start(); err != nil {
		logger.Warnf("启动火币数据源失败: %v (行情数据将使用备用方案)", err)
	} else {
		defer huobiManager.Stop()
		fmt.Println("✓ 火币WebSocket数据源启动成功")
	}

	// 启动历史K线调度器（获取多周期K线数据）
	historyScheduler := huobi.GetHistoryKlineScheduler()
	if err := historyScheduler.Start(); err != nil {
		logger.Warnf("启动历史K线调度器失败: %v", err)
	} else {
		defer historyScheduler.Stop()
		fmt.Println("✓ 历史K线调度器启动成功")
	}

	// 初始化测试数据（开发模式下）
	if config.GlobalConfig.App.Mode == "debug" {
		initSvc := service.NewInitService()
		if err := initSvc.InitTestData(); err != nil {
			logger.Warnf("初始化测试数据失败: %v", err)
		} else {
			fmt.Println("✓ 测试数据初始化成功")
		}
	}

	// 初始化路由
	if err := service.EnsureDefaultCurrencies(); err != nil {
		logger.Warnf("default currencies init failed: %v", err)
	} else {
		fmt.Println("✅ default currencies initialized")
	}

	r := router.Setup()

	// 启动秒合约自动结算任务
	microScheduler := scheduler.NewMicroOrderScheduler()
	go microScheduler.Start()
	defer microScheduler.Stop()
	fmt.Println("✓ 秒合约自动结算任务启动成功")

	// 启动永续合约交易定时任务(爆仓检测、止盈止损检测、限价单检测)
	contractScheduler := scheduler.NewContractScheduler()
	go contractScheduler.Start()
	defer contractScheduler.Stop()
	fmt.Println("✓ 永续合约交易定时任务启动成功")

	// 启动现货订单价格监控任务（限价单自动成交）
	go service.GetMatchingEngine().Start()
	defer service.GetMatchingEngine().Stop()
	fmt.Println("✓ 现货订单价格监控任务启动成功")

	// 初始化WebSocket服务
	wsAddr := fmt.Sprintf(":%d", config.GlobalConfig.WebSocket.Port)
	wsServer := &http.Server{
		Addr:    wsAddr,
		Handler: setupWebSocketRoutes(),
	}

	// 启动服务
	addr := fmt.Sprintf(":%d", config.GlobalConfig.App.Port)
	fmt.Printf("\n🚀 服务启动成功，监听端口: %s\n", addr)
	fmt.Printf("   模式: %s\n", config.GlobalConfig.App.Mode)
	fmt.Printf("   健康检查: http://localhost%s/health\n\n", addr)

	// 在新协程中启动API服务
	go func() {
		if err := r.Run(addr); err != nil {
			logger.Fatalf("API服务启动失败: %v", err)
		}
	}()

	// 在新协程中启动WebSocket服务
	go func() {
		fmt.Printf("🌐 WebSocket服务启动，监听端口: %s\n", wsAddr)
		if err := wsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("WebSocket服务启动失败: %v", err)
		}
	}()

	// 等待中断信号优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 关闭WebSocket服务
	if err := wsServer.Close(); err != nil {
		logger.Errorf("关闭WebSocket服务失败: %v", err)
	}

	logger.Info("正在关闭服务...")
	fmt.Println("\n👋 服务已关闭")
}

// setupWebSocketRoutes 设置WebSocket路由
func setupWebSocketRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		// 使用gin适配器
		c, _ := gin.CreateTestContext(w)
		c.Request = r
		websocket.Handler(c)
	})
	return mux
}
