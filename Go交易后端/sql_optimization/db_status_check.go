package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	dsn := "root:root123456@tcp(127.0.0.1:3306)/information_schema?charset=utf8&parseTime=True&loc=Local"

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		log.Fatalf("数据库连接测试失败: %v", err)
	}

	fmt.Println("========================================")
	fmt.Println("MySQL 数据库状态检查报告")
	fmt.Println("========================================\n")

	// 检查MySQL版本
	checkVersion(db)

	// 检查InnoDB配置
	checkInnoDBConfig(db)

	// 检查连接配置
	checkConnectionConfig(db)

	// 检查缓冲池使用情况
	checkBufferPool(db)

	// 检查慢查询配置
	checkSlowQuery(db)

	// 检查表统计信息
	checkTableStats(db)

	// 检查索引创建情况
	checkIndexes(db)

	fmt.Println("\n========================================")
	fmt.Println("检查完成")
	fmt.Println("========================================")
}

func checkVersion(db *sql.DB) {
	fmt.Println("【1. MySQL 版本信息】")
	var version string
	err := db.QueryRow("SELECT VERSION()").Scan(&version)
	if err != nil {
		log.Printf("获取版本失败: %v", err)
		return
	}
	fmt.Printf("  版本: %s\n", version)

	var versionComment string
	db.QueryRow("SHOW VARIABLES LIKE 'version_comment'").Scan(&versionComment)
	fmt.Printf("  编译信息: %s\n", versionComment)
	fmt.Println()
}

func checkInnoDBConfig(db *sql.DB) {
	fmt.Println("【2. InnoDB 核心配置】")

	variables := []string{
		"innodb_buffer_pool_size",
		"innodb_log_file_size",
		"innodb_log_buffer_size",
		"innodb_flush_log_at_trx_commit",
		"innodb_file_per_table",
		"innodb_io_capacity",
	}

	for _, varName := range variables {
		var value string
		err := db.QueryRow(fmt.Sprintf("SHOW VARIABLES LIKE '%s'", varName)).Scan(&varName, &value)
		if err == nil {
			fmt.Printf("  %s: %s\n", varName, value)
		}
	}
	fmt.Println()
}

func checkConnectionConfig(db *sql.DB) {
	fmt.Println("【3. 连接配置】")

	variables := []string{
		"max_connections",
		"thread_cache_size",
		"table_open_cache",
		"wait_timeout",
		"interactive_timeout",
	}

	for _, varName := range variables {
		var value string
		err := db.QueryRow(fmt.Sprintf("SHOW VARIABLES LIKE '%s'", varName)).Scan(&varName, &value)
		if err == nil {
			fmt.Printf("  %s: %s\n", varName, value)
		}
	}
	fmt.Println()
}

func checkBufferPool(db *sql.DB) {
	fmt.Println("【4. 缓冲池使用情况】")

	var poolSizeUsed, poolSizeTotal string
	db.QueryRow("SHOW GLOBAL STATUS LIKE 'Innodb_buffer_pool_bytes_data'").Scan(&poolSizeUsed, &poolSizeUsed)
	db.QueryRow("SHOW GLOBAL STATUS LIKE 'Innodb_buffer_pool_size'").Scan(&poolSizeTotal, &poolSizeTotal)

	fmt.Printf("  缓冲池已用: %s bytes\n", poolSizeUsed)
	fmt.Printf("  缓冲池总大小: %s bytes\n", poolSizeTotal)

	// 计算缓冲池命中率
	var readRequests, readDisk string
	db.QueryRow("SHOW GLOBAL STATUS LIKE 'Innodb_buffer_pool_read_requests'").Scan(&readRequests, &readRequests)
	db.QueryRow("SHOW GLOBAL STATUS LIKE 'Innodb_buffer_pool_reads'").Scan(&readDisk, &readDisk)

	var req, disk int64
	fmt.Sscanf(readRequests, "%d", &req)
	fmt.Sscanf(readDisk, "%d", &disk)

	if req > 0 {
		hitRate := float64(req-disk) / float64(req) * 100
		fmt.Printf("  缓冲池命中率: %.2f%%\n", hitRate)
		if hitRate < 95 {
			fmt.Println("  ⚠️  警告: 命中率低于95%，建议增大 innodb_buffer_pool_size")
		} else {
			fmt.Println("  ✅ 正常")
		}
	}
	fmt.Println()
}

func checkSlowQuery(db *sql.DB) {
	fmt.Println("【5. 慢查询配置】")

	var slowQueryLog, longQueryTime, slowQueries string
	db.QueryRow("SHOW VARIABLES LIKE 'slow_query_log'").Scan(&slowQueryLog, &slowQueryLog)
	db.QueryRow("SHOW VARIABLES LIKE 'long_query_time'").Scan(&longQueryTime, &longQueryTime)
	db.QueryRow("SHOW GLOBAL STATUS LIKE 'Slow_queries'").Scan(&slowQueries, &slowQueries)

	fmt.Printf("  slow_query_log: %s\n", slowQueryLog)
	fmt.Printf("  long_query_time: %s\n", longQueryTime)
	fmt.Printf("  慢查询数量: %s\n", slowQueries)

	if slowQueryLog == "OFF" {
		fmt.Println("  ⚠️  建议: 开启慢查询日志以便监控性能问题")
	}
	fmt.Println()
}

func checkTableStats(db *sql.DB) {
	fmt.Println("【6. 数据表统计】")

	rows, err := db.Query(`
		SELECT TABLE_NAME, TABLE_ROWS, DATA_LENGTH, INDEX_LENGTH
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = 'bibi2022'
		ORDER BY DATA_LENGTH DESC
		LIMIT 20
	`)
	if err != nil {
		log.Printf("查询表统计失败: %v", err)
		return
	}
	defer rows.Close()

	fmt.Printf("  %-30s %15s %15s %15s\n", "表名", "行数", "数据大小", "索引大小")
	fmt.Println("  " + strings.Repeat("-", 75))

	totalRows := 0
	totalDataSize := int64(0)
	totalIndexSize := int64(0)

	for rows.Next() {
		var tableName string
		var tableRows, dataLength, indexLength int64
		rows.Scan(&tableName, &tableRows, &dataLength, &indexLength)

		fmt.Printf("  %-30s %15d %15s %15s\n",
			tableName,
			tableRows,
			formatBytes(dataLength),
			formatBytes(indexLength))

		totalRows += int(tableRows)
		totalDataSize += dataLength
		totalIndexSize += indexLength
	}

	fmt.Println("  " + strings.Repeat("-", 75))
	fmt.Printf("  %-30s %15d %15s %15s\n",
		"合计",
		totalRows,
		formatBytes(totalDataSize),
		formatBytes(totalIndexSize))
	fmt.Println()
}

func checkIndexes(db *sql.DB) {
	fmt.Println("【7. 索引统计】")

	rows, err := db.Query(`
		SELECT 
			t.TABLE_NAME,
			COUNT(i.INDEX_NAME) as index_count,
			SUM(i.NON_UNIQUE) as non_unique_count
		FROM information_schema.TABLES t
		LEFT JOIN information_schema.STATISTICS i ON t.TABLE_SCHEMA = i.TABLE_SCHEMA 
			AND t.TABLE_NAME = i.TABLE_NAME
		WHERE t.TABLE_SCHEMA = 'bibi2022'
		GROUP BY t.TABLE_NAME
		ORDER BY index_count DESC
	`)
	if err != nil {
		log.Printf("查询索引统计失败: %v", err)
		return
	}
	defer rows.Close()

	fmt.Printf("  %-30s %15s %15s\n", "表名", "索引数量", "非唯一索引")
	fmt.Println("  " + strings.Repeat("-", 60))

	totalIndexes := 0
	for rows.Next() {
		var tableName string
		var indexCount, nonUniqueCount int
		rows.Scan(&tableName, &indexCount, &nonUniqueCount)
		fmt.Printf("  %-30s %15d %15d\n", tableName, indexCount, nonUniqueCount)
		totalIndexes += indexCount
	}

	fmt.Println("  " + strings.Repeat("-", 60))
	fmt.Printf("  总索引数: %d\n", totalIndexes)
	fmt.Println()
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
