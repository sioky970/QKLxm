package router

import (
	"exchange-go/config"
	_ "exchange-go/docs"
	"exchange-go/internal/api/handler"
	"exchange-go/internal/api/middleware"
	"exchange-go/internal/websocket"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Setup 初始化路由
func Setup() *gin.Engine {
	// 设置运行模式
	if config.GlobalConfig.App.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// 全局中间件
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.Cors())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Swagger文档
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// WebSocket路由
	r.GET("/ws", websocket.Handler)

	// 静态文件服务（用于充值截图等上传文件访问）
	r.Static("/uploads", "./uploads")

	// API 路由组
	api := r.Group("/api")
	{
		// 公开接口 (无需登录)
		public := api.Group("")
		{
			// 用户模块
			public.POST("/user/register", handler.UserRegister)
			public.POST("/user/login", handler.UserLogin)
			public.POST("/user/send_code", handler.SendVerifyCode)
			public.POST("/user/reset_password", handler.ResetPassword)

			// 公开配置模块
			public.GET("/config/register", handler.GetRegisterConfig)

			// 币种模块
			public.GET("/currency/list", handler.GetCurrencyList)

			// 行情模块
			public.GET("/quotation/new", handler.NewQuotation)
			public.POST("/market/market", handler.MarketData)
			public.GET("/kline", handler.GetKline)
			public.GET("/currency/quotation_new", handler.CurrencyQuotation)

			// 新闻模块
			public.GET("/news/list", handler.GetNewsList)
			public.GET("/news/detail", handler.GetNewsDetail)
			public.GET("/news/category", handler.GetNewsCategoryList)
			public.GET("/news/recommended", handler.GetRecommendedNews)
			public.GET("/news/hot", handler.GetHotNews)
		}

		// 需要登录的接口
		auth := api.Group("")
		auth.Use(middleware.JWTAuth())
		{
			// 用户模块
			auth.GET("/user/info", handler.UserInfo)
			auth.POST("/user/update", handler.UserUpdate)
			auth.POST("/user/change_password", handler.ChangePassword)
			auth.POST("/user/change_pay_password", handler.ChangePayPassword)
			auth.GET("/user/cash_info", handler.GetUserCashInfo)
			auth.POST("/user/cash_info", handler.SaveUserCashInfo)

			// 钱包模块
			auth.GET("/wallet/list", handler.WalletList)
			auth.GET("/wallet/info", handler.WalletInfo)
			auth.GET("/wallet/asset-overview", handler.GetAssetOverview)
			auth.GET("/wallet/all-assets", handler.GetAllAssetsWithBalance)
			auth.POST("/wallet/recharge", handler.WalletRecharge)
			auth.POST("/wallet/withdraw", handler.WalletWithdraw)
			auth.GET("/wallet/withdraw-config", handler.GetWithdrawConfig)
			auth.GET("/wallet/logs", handler.WalletLogs)
			auth.GET("/wallet/withdrawals", handler.GetWithdrawalList)
			auth.GET("/wallet/today-profit-loss", handler.GetTodayProfitLoss)
			auth.GET("/wallet/seven-days-profit-loss", handler.GetSevenDaysProfitLoss)

			// 钱包划转模块(新)
			auth.GET("/wallet/transfer/balance", handler.GetWalletTransferHandler().GetBalance)          // 获取总余额
			auth.GET("/wallet/transfer/spot", handler.GetWalletTransferHandler().GetSpotBalance)        // 获取现货钱包余额
			auth.GET("/wallet/transfer/contract", handler.GetWalletTransferHandler().GetContractBalance) // 获取合约钱包余额
			auth.POST("/wallet/transfer", handler.GetWalletTransferHandler().Transfer)                  // 钱包划转
			auth.GET("/wallet/transfer/records", handler.GetWalletTransferHandler().GetTransferRecords)  // 划转记录

			// 交易模块 (币币)
			auth.POST("/transaction/submit", handler.SpotSubmit)
			auth.POST("/transaction/cancel", handler.SpotCancel)
			auth.POST("/transaction/chase", handler.SpotChase)
			auth.GET("/transaction/list", handler.SpotList)
			auth.GET("/transaction/history", handler.SpotHistory)

			// 合约交易模块(旧API保持兼容)
			auth.POST("/lever/submit", handler.LeverSubmit)
			auth.POST("/lever/close", handler.LeverClose)
			auth.GET("/lever/position", handler.LeverPosition)
			auth.GET("/lever/history", handler.LeverHistory)

			// 永续合约交易模块(新API)
			auth.POST("/contract/open", handler.ContractOpenPosition)                      // 开仓(市价/限价)
			auth.POST("/contract/close", handler.ContractClosePosition)                    // 平仓
			auth.POST("/contract/chase", handler.ContractChaseOrder)                       // 追单
			auth.POST("/contract/tpsl", handler.ContractSetTPSL)                           // 设置止盈止损
			auth.POST("/contract/cancel/:order_id", handler.ContractCancelOrder)           // 撤单
			auth.GET("/contract/positions", handler.ContractGetPositions)                  // 持仓列表
			auth.GET("/contract/pending", handler.ContractGetPendingOrders)                // 挂单列表
			auth.GET("/contract/position/:position_id", handler.ContractGetPositionDetail) // 持仓详情

			// 交割合约交易模块
			auth.POST("/delivery/contract/open", handler.DeliveryOpen)           // 开仓(市价/限价)
			auth.POST("/delivery/contract/close", handler.DeliveryClose)         // 平仓
			auth.GET("/delivery/contract/positions", handler.DeliveryPositions)  // 持仓列表
			auth.GET("/delivery/account/balance", handler.DeliveryAccount)        // 账户余额
			r.GET("/delivery/contracts", handler.DeliveryContracts)               // 合约列表(无需登录)

			// 秒合约模块
			auth.GET("/micro/periods", handler.MicroPeriods)       // 获取周期配置(无需登录也可访问)
			auth.POST("/micro/submit", handler.MicroSubmit)        // 提交订单
			auth.GET("/micro/orders", handler.MicroList)           // 订单列表
			auth.GET("/micro/active", handler.MicroActiveOrder)    // 当前进行中的订单
			auth.GET("/micro/order/:id", handler.MicroOrderDetail) // 订单详情

			// 实名认证模块
			auth.POST("/kyc/submit", handler.SubmitKYC)
			auth.GET("/kyc/status", handler.GetKYCStatus)

			// 站内信模块
			auth.GET("/messages/list", handler.GetMessageList)
			auth.GET("/messages/detail", handler.GetMessageDetail)
			auth.POST("/messages/read", handler.MarkMessageAsRead)
			auth.POST("/messages/unread", handler.MarkMessageAsUnread)
			auth.POST("/messages/batch-read", handler.BatchMarkMessagesAsRead)
			auth.POST("/messages/read-all", handler.MarkAllMessagesAsRead)
			auth.POST("/messages/delete", handler.DeleteMessage)
			auth.POST("/messages/batch-delete", handler.BatchDeleteMessages)
			auth.GET("/messages/unread-count", handler.GetUnreadMessageCount)
			auth.GET("/messages/statistics", handler.GetMessageStatistics)

			// 统一订单模块（聚合现货、合约、秒合约）
			auth.GET("/orders/all", handler.GetUnifiedOrderList)       // 统一订单列表
			auth.GET("/orders/statistics", handler.GetOrderStatistics) // 订单统计
			auth.GET("/orders/detail", handler.GetOrderDetail)         // 订单详情

			// 充值模块
			auth.GET("/deposit/addresses", handler.GetDepositAddresses)   // 获取充值地址列表
			auth.POST("/deposit/create", handler.CreateDepositOrder)      // 创建充值订单
			auth.POST("/deposit/upload", handler.UploadDepositScreenshot) // 上传转账截图
			auth.GET("/deposit/orders", handler.GetDepositOrders)         // 用户充值记录
			auth.POST("/deposit/cancel", handler.CancelDepositOrder)      // 取消充值订单

			// 通用上传
			auth.POST("/upload", handler.UploadFile) // 通用文件上传
		}

		// 管理端路由
		SetupAdminRoutes(api)
	}

	return r
}
