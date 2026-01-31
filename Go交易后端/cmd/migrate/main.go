package main

import (
	"flag"
	"fmt"
	"os"

	"exchange-go/config"
	"exchange-go/internal/pkg/database"
	"exchange-go/internal/pkg/logger"
)

// 版本信息
const (
	Version = "1.0.0"
	Banner  = `
╔═══════════════════════════════════════════════════════════╗
║           Exchange Database Migration Tool                 ║
║                     Version %s                           ║
╚═══════════════════════════════════════════════════════════╝
`
)

func main() {
	// 命令行参数
	var (
		configPath  string
		showHelp    bool
		showVersion bool
		showStatus  bool
		initOnly    bool
		migrateOnly bool
		resetDB     bool
		verbose     bool
	)

	flag.StringVar(&configPath, "config", "config/config.yaml", "配置文件路径")
	flag.BoolVar(&showHelp, "help", false, "显示帮助信息")
	flag.BoolVar(&showHelp, "h", false, "显示帮助信息 (简写)")
	flag.BoolVar(&showVersion, "version", false, "显示版本信息")
	flag.BoolVar(&showVersion, "v", false, "显示版本信息 (简写)")
	flag.BoolVar(&showStatus, "status", false, "显示迁移状态")
	flag.BoolVar(&initOnly, "init-only", false, "仅初始化默认数据（不迁移表结构）")
	flag.BoolVar(&migrateOnly, "migrate-only", false, "仅迁移表结构（不初始化数据）")
	flag.BoolVar(&resetDB, "reset", false, "重置数据库（危险：会删除所有数据）")
	flag.BoolVar(&verbose, "verbose", true, "详细输出")

	flag.Parse()

	// 显示帮助
	if showHelp {
		printHelp()
		return
	}

	// 显示版本
	if showVersion {
		fmt.Printf("Exchange Migration Tool v%s\n", Version)
		return
	}

	// 显示Banner
	fmt.Printf(Banner, Version)

	// 加载配置
	fmt.Printf("[信息] 加载配置文件: %s\n", configPath)
	if err := config.Init(configPath); err != nil {
		fmt.Printf("[错误] 加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	if err := logger.Init(&config.GlobalConfig.Log); err != nil {
		fmt.Printf("[错误] 初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// 连接数据库（不执行自动迁移）
	logger.Info("连接数据库...")
	dsn := config.GlobalConfig.Database.GetDSN()
	if err := database.InitWithoutMigration(dsn, &config.GlobalConfig.Database); err != nil {
		logger.Error("数据库连接失败: %v", err)
		os.Exit(1)
	}
	defer database.Close()
	logger.Info("数据库连接成功")

	// 显示状态
	if showStatus {
		printStatus()
		return
	}

	// 重置数据库（危险操作）
	if resetDB {
		if !confirmReset() {
			logger.Info("操作已取消")
			return
		}
		if err := database.ResetMigration(true); err != nil {
			logger.Error("重置数据库失败: %v", err)
			os.Exit(1)
		}
		logger.Info("数据库重置完成")
		return
	}

	// 构建迁移配置
	migrationConfig := database.MigrationConfig{
		AutoMigrate:     !initOnly,
		InitDefaultData: !migrateOnly,
		Verbose:         verbose,
	}

	// 执行迁移
	if err := database.RunMigrationWithConfig(migrationConfig); err != nil {
		logger.Error("迁移失败: %v", err)
		os.Exit(1)
	}

	logger.Info("迁移成功完成!")
}

// printHelp 打印帮助信息
func printHelp() {
	fmt.Println(`
Exchange Database Migration Tool

用法:
  migrate [选项]

选项:
  -config string     配置文件路径 (默认 "config/config.yaml")
  -h, -help          显示帮助信息
  -v, -version       显示版本信息
  -status            显示当前迁移状态
  -migrate-only      仅迁移表结构（不初始化默认数据）
  -init-only         仅初始化默认数据（不迁移表结构）
  -reset             重置数据库（危险：会删除所有数据！）
  -verbose           详细输出模式 (默认 true)

示例:
  # 执行完整迁移
  migrate

  # 使用自定义配置文件
  migrate -config /path/to/config.yaml

  # 仅迁移表结构
  migrate -migrate-only

  # 仅初始化默认数据
  migrate -init-only

  # 查看迁移状态
  migrate -status

  # 重置数据库（谨慎使用）
  migrate -reset

注意:
  - 迁移操作是增量的，不会删除已有数据
  - 默认数据只在不存在时才会创建，不会覆盖已有配置
  - reset 操作会删除所有表数据，请谨慎使用
`)
}

// printStatus 打印迁移状态
func printStatus() {
	fmt.Println("\n========================================")
	fmt.Println("数据库迁移状态报告")
	fmt.Println("========================================")

	status := database.GetMigrationStatus()

	// 表状态
	fmt.Println("\n[表状态]")
	if tables, ok := status["tables"].(map[string]bool); ok {
		for table, exists := range tables {
			statusStr := "✗ 不存在"
			if exists {
				statusStr = "✓ 已创建"
			}
			fmt.Printf("  %-25s %s\n", table, statusStr)
		}
	}

	// 数据统计
	fmt.Println("\n[数据统计]")
	if counts, ok := status["counts"].(map[string]int64); ok {
		fmt.Printf("  系统配置: %d 条\n", counts["settings"])
		fmt.Printf("  管理员: %d 个\n", counts["admins"])
		fmt.Printf("  币种: %d 个\n", counts["currencies"])
	}

	fmt.Println("\n========================================")
}

// confirmReset 确认重置操作
func confirmReset() bool {
	fmt.Println("\n⚠️  警告: 此操作将删除所有数据表并重新创建！")
	fmt.Println("所有数据将会丢失，此操作不可逆！")
	fmt.Print("\n请输入 'YES' 确认执行: ")

	var confirm string
	fmt.Scanln(&confirm)

	return confirm == "YES"
}
