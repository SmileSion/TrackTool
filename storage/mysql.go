package storage

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/qwy-tacking/middleware"
	"github.com/qwy-tacking/model"
)

var (
	DB *sql.DB
	detailTableName        = "event_detail" // 当前明细表名
	detailTableCreateTime  time.Time        // 明细表创建时间（首次插入时设定）
	detailTableInitialized = false
)

// InitMySQL 初始化数据库连接
func InitMySQL(dsn string) {
	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		panic("MySQL连接失败: " + err.Error())
	}
	log.Printf("  InitMySQL DSN: %s\n", dsn)
	// 验证连接是否成功
	if err = DB.Ping(); err != nil {
		panic("MySQL Ping失败: " + err.Error())
	}

	middleware.Logger.Println("MySQL连接成功")
}

// InsertEvents 插入事件数据到 MySQL（按分钟聚合）
func InsertEvents(events []model.Event) error {
	eventCountMap := make(map[string]int)

	for _, e := range events {
		key := fmt.Sprintf("%s:%s:%s:%s", e.ClientType, e.Site, e.EventType, e.EventDetail)
		eventCountMap[key] += e.Count
	}

	stmt, err := DB.Prepare(`
		INSERT INTO event_logs (event_time, client_type, site, event_type, event_detail, count)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare失败: %v", err)
	}
	defer stmt.Close()

	timestamp := time.Now().Unix() / 60 * 60

	for key, count := range eventCountMap {
		parts := strings.SplitN(key, ":", 4)
		if len(parts) < 4 {
			middleware.Logger.Printf("非法Key跳过: %s", key)
			continue
		}
		clientType, site, eventType, eventDetail := parts[0], parts[1], parts[2], parts[3]

		_, err := stmt.Exec(timestamp, clientType, site, eventType, eventDetail, count)
		if err != nil {
			middleware.Logger.Printf("插入失败 -> client=%s site=%s event=%s detail=%s count=%d，错误：%v",
				clientType, site, eventType, eventDetail, count, err)
			continue
		}
	}

	return nil
}

// InsertEventDetails 写入明细表，每5天自动切表（改名+新建）
func InsertEventDetails(events []model.Event) error {
	if len(events) == 0 {
		return nil
	}

	now := time.Now()

	// 第一次写入时初始化表
	if !detailTableInitialized {
		if !checkTableExists(detailTableName) {
			if err := createDetailTable(detailTableName); err != nil {
				return fmt.Errorf("创建初始明细表失败: %v", err)
			}
		}
		detailTableCreateTime = now
		detailTableInitialized = true
	}

	// 判断是否满5天切表
	if now.Sub(detailTableCreateTime) >= 5*24*time.Hour {
		oldTable := detailTableName
		suffix := detailTableCreateTime.Format("20060102")
		newName := fmt.Sprintf("event_detail_%s", suffix)

		if err := renameTable(oldTable, newName); err != nil {
			return fmt.Errorf("切换旧表失败: %v", err)
		}
		middleware.Logger.Printf("已将旧表 %s 重命名为 %s", oldTable, newName)

		if err := createDetailTable(detailTableName); err != nil {
			return fmt.Errorf("创建新 event_detail 表失败: %v", err)
		}
		detailTableCreateTime = now
	}

	// 构建批量插入 SQL
	valueStrings := make([]string, 0, len(events))
	valueArgs := make([]interface{}, 0, len(events)*6)

	for _, e := range events {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?)")
		valueArgs = append(valueArgs,
			e.TimeStamp,
			e.ClientType,
			e.Site,
			e.EventType,
			e.EventDetail,
			e.UserDetail,
		)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (event_time, client_type, site, event_type, event_detail, user_detail)
		VALUES %s`, detailTableName, strings.Join(valueStrings, ","))

	_, err := DB.Exec(query, valueArgs...)
	if err != nil {
		return fmt.Errorf("批量插入失败: %v", err)
	}

	middleware.Logger.Printf("明细数据批量写入 %s 成功，共 %d 条", detailTableName, len(events))
	return nil
}

// 判断表是否存在
func checkTableExists(table string) bool {
	query := `SELECT COUNT(*) FROM information_schema.tables WHERE table_name = ?`
	var count int
	err := DB.QueryRow(query, table).Scan(&count)
	return err == nil && count > 0
}

// 创建表
func createDetailTable(table string) error {
	query := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		event_time BIGINT NOT NULL,
		client_type VARCHAR(20) NOT NULL,
		site VARCHAR(100) NOT NULL,
		event_type VARCHAR(50) NOT NULL,
		event_detail TEXT,
		user_detail TEXT
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
	`, table)

	_, err := DB.Exec(query)
	return err
}

// 表改名
func renameTable(oldName, newName string) error {
	query := fmt.Sprintf(`RENAME TABLE %s TO %s`, oldName, newName)
	_, err := DB.Exec(query)
	return err
}
