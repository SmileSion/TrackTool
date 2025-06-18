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
	logTableInitialized   bool
	logTableCreateTime    time.Time
	logTableName          = "event_logs"
)

// InitMySQL 初始化数据库连接，并启动时检查旧表并重命名
func InitMySQL(dsn string) {
	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		panic("MySQL连接失败: " + err.Error())
	}
	log.Printf("  InitMySQL DSN: %s\n", dsn)
	if err = DB.Ping(); err != nil {
		panic("MySQL Ping失败: " + err.Error())
	}

	middleware.Logger.Println("MySQL连接成功")

	// 启动时检查旧日志表和明细表，若存在则重命名避免冲突
	checkAndRenameOldTable(logTableName, "event_logs")
	checkAndRenameOldTable(detailTableName, "event_detail")
}

func InsertEvents(events []model.Event) error {
	if len(events) == 0 {
		return nil
	}

	now := time.Now()

	// 初始化（程序启动时第一次）
	if !logTableInitialized {
		if !checkTableExists(logTableName) {
			if err := createLogTable(logTableName); err != nil {
				return fmt.Errorf("创建初始日志表失败: %v", err)
			}
		}
		logTableCreateTime = now
		logTableInitialized = true
	}

	// 每天 0 点切表逻辑
	if !isSameDate(now, logTableCreateTime) {
		oldTable := logTableName
		suffix := logTableCreateTime.Format("20060102")
		newName := fmt.Sprintf("event_logs_%s", suffix)

		if err := renameTable(oldTable, newName); err != nil {
			return fmt.Errorf("重命名旧日志表失败: %v", err)
		}
		middleware.Logger.Printf("旧日志表 %s 已重命名为 %s", oldTable, newName)

		if err := createLogTable(logTableName); err != nil {
			return fmt.Errorf("创建新日志表失败: %v", err)
		}
		logTableCreateTime = now
	}

	// 聚合事件
	eventCountMap := make(map[string]int)
	for _, e := range events {
		key := fmt.Sprintf("%s:%s:%s:%s", e.ClientType, e.Site, e.EventType, e.EventDetail)
		eventCountMap[key] += e.Count
	}

	stmt, err := DB.Prepare(fmt.Sprintf(`
		INSERT INTO %s (event_time, client_type, site, event_type, event_detail, count)
		VALUES (?, ?, ?, ?, ?, ?)`, logTableName))
	if err != nil {
		return fmt.Errorf("prepare失败: %v", err)
	}
	defer stmt.Close()

	timestamp := now.Unix() / 60 * 60 // 当前分钟整点时间戳

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
		}
	}

	middleware.Logger.Printf("事件数据写入 %s 成功，共计 %d 条", logTableName, len(eventCountMap))
	return nil
}

// 判断是否是同一天
func isSameDate(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}


func InsertEventDetails(events []model.Event) error {
	if len(events) == 0 {
		return nil
	}

	now := time.Now()

	if !detailTableInitialized {
		if !checkTableExists(detailTableName) {
			if err := createDetailTable(detailTableName); err != nil {
				return fmt.Errorf("创建明细表失败: %v", err)
			}
		}
		detailTableCreateTime = now
		detailTableInitialized = true
	}

	if now.Day() != detailTableCreateTime.Day() {
		oldTable := detailTableName
		suffix := detailTableCreateTime.Format("20060102")
		newName := fmt.Sprintf("event_detail_%s", suffix)

		if err := renameTable(oldTable, newName); err != nil {
			return fmt.Errorf("重命名旧明细表失败: %v", err)
		}
		middleware.Logger.Printf("明细表 %s 重命名为 %s", oldTable, newName)

		if err := createDetailTable(detailTableName); err != nil {
			return fmt.Errorf("创建新明细表失败: %v", err)
		}
		detailTableCreateTime = now
	}

	// 批量插入
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

	if _, err := DB.Exec(query, valueArgs...); err != nil {
		return fmt.Errorf("明细插入失败: %v", err)
	}

	middleware.Logger.Printf("明细写入 %s 成功，共 %d 条", detailTableName, len(events))
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
    query := fmt.Sprintf("RENAME TABLE `%s` TO `%s`", oldName, newName)
    _, err := DB.Exec(query)
    return err
}

func createLogTable(tableName string) error {
	createSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id BIGINT NOT NULL AUTO_INCREMENT,
			event_time BIGINT NOT NULL,
			client_type VARCHAR(20) NOT NULL,
			site VARCHAR(100) NOT NULL,
			event_type VARCHAR(50) NOT NULL,
			event_detail VARCHAR(255) NOT NULL,
			count INT DEFAULT '1',
			PRIMARY KEY (id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
	`, tableName)

	_, err := DB.Exec(createSQL)
	return err
}

// checkAndRenameOldTable 如果表存在，用当前时间戳重命名，避免重名冲突
func checkAndRenameOldTable(baseTableName, prefix string) {
	if checkTableExists(baseTableName) {
		now := time.Now()
		// 用时间戳拼接新表名，不带下划线
		newName := fmt.Sprintf("%s%d", prefix, now.Unix())
		err := renameTable(baseTableName, newName)
		if err != nil {
			middleware.Logger.Printf("启动时重命名旧表 %s 失败: %v", baseTableName, err)
		} else {
			middleware.Logger.Printf("启动时旧表 %s 重命名为 %s", baseTableName, newName)
		}
	}
}