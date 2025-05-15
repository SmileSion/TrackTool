package storage

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
	"log"

	"github.com/qwy-tacking/middleware"
	"github.com/qwy-tacking/model"
	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

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
		eventCountMap[key]++
	}

	stmt, err := DB.Prepare(`
		INSERT INTO event_logs (timestamp, client_type, site, event_type, event_detail, count)
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

