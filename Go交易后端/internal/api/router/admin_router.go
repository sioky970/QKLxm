package router

import (
	"github.com/gin-gonic/gin"

	adminHandler "exchange-go/internal/api/handler/admin"
	"exchange-go/internal/api/middleware"
)

// SetupAdminRoutes 设置管理端路由
func SetupAdminRoutes(r *gin.RouterGroup) {
	// 管理端公开路由 (无需认证)
	adminPublic := r.Group("/admin")
	{
		adminPublic.POST("/login", adminHandler.AdminLogin)
	}

	// 管理端路由组 - 需要管理员权限
	admin := r.Group("/admin")
	admin.Use(middleware.AdminJWTAuth())
	{
		// 认证相关
		admin.POST("/logout", adminHandler.AdminLogout)
		admin.GET("/info", adminHandler.AdminInfo)
		admin.POST("/change-password", adminHandler.AdminChangePassword)

		// 用户管理
		user := admin.Group("/user")
		{
			user.POST("/list", adminHandler.GetUserList)
			user.GET("/:id", adminHandler.GetUserDetail)
			user.PUT("/:id", adminHandler.UpdateUser)
			user.DELETE("/:id", adminHandler.DeleteUser)
			user.POST("/freeze", adminHandler.FreezeUser)
			user.POST("/activate", adminHandler.ActivateUser)
			user.POST("/reset-password", adminHandler.ResetPassword)
			user.POST("/adjust-balance", adminHandler.AdjustBalance)
			user.POST("/batch-risk", adminHandler.BatchSetRisk)
			user.GET("/:id/wallets", adminHandler.GetUserWallets)
			user.POST("/search", adminHandler.SearchUsers)
			user.GET("/export", adminHandler.ExportUsers)
			
			// 钱包余额管理（新版）
			user.GET("/:id/wallet-balance", adminHandler.GetWalletBalanceHandler().GetUserWalletBalance)
			user.POST("/wallet-adjust", adminHandler.GetWalletBalanceHandler().AdjustBalance)
			user.GET("/wallet-adjust/records", adminHandler.GetWalletBalanceHandler().GetAdjustmentRecords)
		}

		// 币种管理
		currency := admin.Group("/currency")
		{
			currency.POST("/list", adminHandler.GetCurrencyList)
			currency.GET("/:id", adminHandler.GetCurrencyDetail)
			currency.POST("/create", adminHandler.CreateCurrency)
			currency.PUT("/:id", adminHandler.UpdateCurrency)
			currency.DELETE("/:id", adminHandler.DeleteCurrency)
			currency.POST("/toggle-display", adminHandler.ToggleCurrencyDisplay)
			currency.POST("/update-rate", adminHandler.UpdateCurrencyRate)
			currency.POST("/update-risk", adminHandler.UpdateCurrencyRisk)
			currency.POST("/refresh-market", adminHandler.RefreshCurrencyMarket)
		}

		// 钱包管理与提现审核
		wallet := admin.Group("/wallet")
		{
			wallet.POST("/list", adminHandler.GetWalletList)
			wallet.GET("/:id", adminHandler.GetWalletDetail)
			wallet.POST("/update-balance", adminHandler.UpdateWalletBalance)
			wallet.POST("/freeze", adminHandler.FreezeWallet)
			wallet.POST("/activate", adminHandler.ActivateWallet)

			// 资金钱包管理
			fundWallet := wallet.Group("/fund")
			{
				fundWallet.GET("/list", adminHandler.GetAdminFundWalletHandler().GetFundWalletList)                  // 资金钱包列表
				fundWallet.GET("/:id", adminHandler.GetAdminFundWalletHandler().GetFundWalletDetail)               // 资金钱包详情
				fundWallet.GET("/transfers", adminHandler.GetAdminFundWalletHandler().GetFundWalletTransfers)     // 划转记录列表
				fundWallet.GET("/statistics", adminHandler.GetAdminFundWalletHandler().GetFundWalletStatistics)   // 资金钱包统计
				fundWallet.POST("/adjust", adminHandler.GetAdminFundWalletHandler().AdminAdjustFundWallet)        // 调整资金钱包余额
			}
		}

		// 提现审核 (对齐前端 API 路径)
		withdrawal := admin.Group("/withdrawal")
		{
			withdrawal.POST("/list", adminHandler.GetWithdrawalsList)
			withdrawal.GET("/:id", adminHandler.GetWithdrawalDetail)
			withdrawal.POST("/approve", adminHandler.ApproveWithdrawal)
			withdrawal.POST("/reject", adminHandler.RejectWithdrawal)
		}

		// 充值管理
		charge := admin.Group("/charge")
		{
			charge.POST("/list", adminHandler.GetChargeList)
			charge.POST("/approve", adminHandler.ApproveCharge)
			charge.POST("/reject", adminHandler.RejectCharge)
		}

		// 交易管理
		transaction := admin.Group("/transaction")
		{
			transaction.POST("/spot/list", adminHandler.GetSpotTransactionList)
			transaction.POST("/lever/list", adminHandler.GetLeverTransactionList)
			transaction.POST("/micro/list", adminHandler.GetMicroTransactionList)
			transaction.POST("/cancel", adminHandler.CancelOrder)
			/* 法币、C2C、商家管理已删除
			transaction.POST("/legal/list", adminHandler.GetLegalOrderList)
			transaction.POST("/legal/confirm-pay", adminHandler.ConfirmLegalPay)
			transaction.POST("/legal/confirm-receive", adminHandler.ConfirmLegalReceive)
			transaction.POST("/legal/cancel", adminHandler.CancelLegalOrder)
			transaction.POST("/c2c/list", adminHandler.GetC2COrderList)
			transaction.POST("/c2c/back", adminHandler.BackC2CSend)
			transaction.DELETE("/c2c/send/:id", adminHandler.DeleteC2CSend)
			*/
			transaction.POST("/account-log/list", adminHandler.GetAccountLogList)
			/*
				transaction.POST("/seller/list", adminHandler.GetSellerListAdvanced)
				transaction.POST("/seller/approve", adminHandler.ApproveSeller)
				transaction.POST("/seller/reject", adminHandler.RejectSeller)
			*/
			transaction.GET("/micro/config", adminHandler.GetMicroConfig)
			transaction.GET("/lever/multiple", adminHandler.GetLeverMultipleList)
		}

		// 秒合约管理
		micro := admin.Group("/micro")
		{
			micro.GET("/number/list", adminHandler.GetMicroNumberList)
			micro.POST("/number/create", adminHandler.CreateMicroNumber)
			micro.DELETE("/number/:id", adminHandler.DeleteMicroNumber)
			micro.GET("/seconds/list", adminHandler.GetMicroSecondsList)
			micro.POST("/seconds/create", adminHandler.CreateMicroSeconds)
			micro.PUT("/seconds/:id", adminHandler.UpdateMicroSeconds)
			micro.POST("/seconds/status", adminHandler.ToggleMicroSecondsStatus)
			micro.DELETE("/seconds/:id", adminHandler.DeleteMicroSeconds)
			micro.POST("/order/list", adminHandler.GetMicroOrderList)
			micro.GET("/order/:id", adminHandler.GetMicroOrderDetail)
			micro.PUT("/order/:id", adminHandler.UpdateMicroOrder)
			micro.POST("/order/batch-risk", adminHandler.BatchMicroRisk)
		}

		// 杠杆交易管理
		lever := admin.Group("/lever")
		{
			lever.POST("/list", adminHandler.GetLeverOrderList)
			lever.GET("/:id", adminHandler.GetLeverOrderDetail)
			lever.GET("/multiple/list", adminHandler.GetLeverMultipleListFull)
			lever.POST("/multiple/create", adminHandler.CreateLeverMultiple)
			lever.DELETE("/multiple/:id", adminHandler.DeleteLeverMultiple)
			lever.POST("/close", adminHandler.CloseLeverOrder)
			lever.GET("/export", adminHandler.ExportLeverOrdersCSV)
		}

		/* 法币交易、C2C交易、商家管理已删除
		// 法币交易管理
		legal := admin.Group("/legal")
		{
			legal.POST("/order/list", adminHandler.GetLegalOrderList)
			legal.GET("/statistics", adminHandler.GetLegalStatistics)
			legal.POST("/cancel", adminHandler.CancelLegalOrder)
			legal.POST("/confirm-pay", adminHandler.ConfirmLegalPay)
			legal.POST("/confirm-receive", adminHandler.ConfirmLegalReceive)
			legal.POST("/send/list", adminHandler.GetLegalSendList)
			legal.GET("/export", adminHandler.ExportLegalOrders)
		}

		// C2C交易管理
		c2c := admin.Group("/c2c")
		{
			c2c.POST("/order/list", adminHandler.GetC2COrderList)
			c2c.GET("/statistics", adminHandler.GetC2CStatistics)
			c2c.POST("/send/back", adminHandler.BackC2CSend)
			c2c.DELETE("/send/:id", adminHandler.DeleteC2CSend)
			c2c.GET("/export", adminHandler.ExportC2COrders)
		}

		// 商家管理
		seller := admin.Group("/seller")
		{
			seller.POST("/list", adminHandler.GetSellerListAdvanced)
			seller.POST("/create", adminHandler.CreateSeller)
			seller.POST("/review", adminHandler.ReviewSeller)
			seller.DELETE("/:id", adminHandler.DeleteSeller)
		}
		*/

		// 账户流水
		account := admin.Group("/account")
		{
			account.POST("/log/list", adminHandler.GetAccountLogListAdvanced)
			account.GET("/log/:id", adminHandler.GetAccountLogDetail)
			account.GET("/log/types", adminHandler.GetAccountLogTypes)
			account.GET("/profits", adminHandler.GetProfitStatistics)
			account.POST("/user/:user_id/logs", adminHandler.GetUserAccountLogs)
			account.GET("/log/export", adminHandler.ExportAccountLogs)
		}

		// 统计功能
		statistics := admin.Group("/statistics")
		{
			statistics.GET("/dashboard", adminHandler.GetDashboard)
			statistics.GET("/user", adminHandler.GetUserStatistics)
			statistics.GET("/trade", adminHandler.GetTradeStatistics)
			statistics.GET("/finance", adminHandler.GetFinanceStatistics)
			statistics.GET("/top-traders", adminHandler.GetTopTraders)
			statistics.GET("/currency", adminHandler.GetCurrencyStatistics)
		}

		// 风控管理
		risk := admin.Group("/risk")
		{
			risk.GET("/config", adminHandler.GetRiskConfig)
			risk.POST("/config", adminHandler.UpdateRiskConfig)
			risk.POST("/users", adminHandler.GetUserRiskList)
			risk.POST("/users/set", adminHandler.SetUserRisk)
			risk.GET("/matches", adminHandler.GetCurrencyMatchRiskList)
			risk.POST("/matches/set", adminHandler.SetMatchRisk)
			risk.POST("/orders", adminHandler.GetMicroOrderRiskList)
			risk.POST("/orders/set", adminHandler.SetMicroOrderRisk)
		}

		// 实名认证
		kyc := admin.Group("/kyc")
		{
			kyc.POST("/list", adminHandler.GetKYCList)
			kyc.GET("/:id", adminHandler.GetKYCDetail)
			kyc.POST("/review", adminHandler.ReviewKYC)
			kyc.POST("/approve", adminHandler.ApproveKYC)
			kyc.POST("/reject", adminHandler.RejectKYC)
			kyc.DELETE("/:id", adminHandler.DeleteKYC)
		}

		// 系统配置
		config := admin.Group("/config")
		{
			config.POST("/list", adminHandler.GetSettingList)
			config.GET("/:key", adminHandler.GetSetting)
			config.POST("/update", adminHandler.UpdateSetting)
			config.POST("/batch-update", adminHandler.BatchUpdateSettings)
		}

		// 管理员管理
		adminMgmt := admin.Group("/admin")
		{
			adminMgmt.POST("/list", adminHandler.GetAdminList)
			adminMgmt.GET("/:id", adminHandler.GetAdminDetail)
			adminMgmt.POST("/create", adminHandler.CreateAdmin)
			adminMgmt.PUT("/update", adminHandler.UpdateAdmin)
			adminMgmt.DELETE("/:id", adminHandler.DeleteAdmin)
		}

		// 角色管理
		role := admin.Group("/role")
		{
			role.GET("/list", adminHandler.GetRoleList)
			role.GET("/:id", adminHandler.GetRoleDetail)
			role.POST("/create", adminHandler.CreateRole)
			role.PUT("/update", adminHandler.UpdateRole)
			role.DELETE("/:id", adminHandler.DeleteRole)
			role.POST("/assign-permissions", adminHandler.AssignRolePermissions)
		}

		// 等级管理
		level := admin.Group("/level")
		{
			level.GET("/list", adminHandler.GetLevelList)
			level.POST("/create", adminHandler.CreateLevel)
			level.PUT("/update", adminHandler.UpdateLevel)
			level.DELETE("/:id", adminHandler.DeleteLevel)
		}

		// 新闻管理
		news := admin.Group("/news")
		{
			news.POST("/list", adminHandler.GetNewsList)
			news.GET("/:id", adminHandler.GetNewsDetail)
			news.POST("/create", adminHandler.CreateNews)
			news.PUT("/:id", adminHandler.UpdateNews)
			news.DELETE("/:id", adminHandler.DeleteNews)
			news.POST("/publish", adminHandler.PublishNews)
			news.POST("/draft", adminHandler.DraftNews)
		}

		// 新闻分类管理
		newsCategory := admin.Group("/news-category")
		{
			newsCategory.GET("/list", adminHandler.GetNewsCategoryList)
			newsCategory.POST("/create", adminHandler.CreateNewsCategory)
			newsCategory.PUT("/:id", adminHandler.UpdateNewsCategory)
			newsCategory.DELETE("/:id", adminHandler.DeleteNewsCategory)
		}

		// 站内信管理
		message := admin.Group("/messages")
		{
			message.POST("/list", adminHandler.GetMessageList)
			message.GET("/:id", adminHandler.GetMessageDetail)
			message.POST("/send", adminHandler.CreateMessage)
			message.PUT("/:id", adminHandler.UpdateMessage)
			message.DELETE("/:id", adminHandler.DeleteMessage)
			message.POST("/withdraw", adminHandler.WithdrawMessage)
			message.POST("/batch-send", adminHandler.BatchSendMessage)
			message.POST("/send-all", adminHandler.SendToAllUsers)
			message.POST("/send-by-level", adminHandler.SendByUserLevel)
			message.GET("/:id/receivers", adminHandler.GetReceiverList)
			message.GET("/:id/statistics", adminHandler.GetMessageReceiverStatistics)
			message.GET("/statistics", adminHandler.GetMessageStatistics)
		}

		// 消息模板管理
		messageTemplate := admin.Group("/message-templates")
		{
			messageTemplate.GET("/list", adminHandler.GetTemplateList)
			messageTemplate.POST("/create", adminHandler.CreateTemplate)
			messageTemplate.PUT("/:id", adminHandler.UpdateTemplate)
			messageTemplate.DELETE("/:id", adminHandler.DeleteTemplate)
		}

		// 充值管理
		deposit := admin.Group("/deposit")
		{
			// 充值地址配置
			deposit.GET("/address/list", adminHandler.GetDepositAddressList)
			deposit.POST("/address/create", adminHandler.CreateDepositAddress)
			deposit.PUT("/address/:id", adminHandler.UpdateDepositAddress)
			deposit.DELETE("/address/:id", adminHandler.DeleteDepositAddress)
			// 充值订单审核
			deposit.POST("/order/list", adminHandler.GetDepositOrderList)
			deposit.GET("/order/:id", adminHandler.GetDepositOrderDetail)
			deposit.POST("/approve", adminHandler.ApproveDepositOrder)
			deposit.POST("/reject", adminHandler.RejectDepositOrder)
		}
	}
}
