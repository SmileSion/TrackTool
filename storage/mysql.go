package storage

import (
	"database/sql"
	"github.com/qwy-tacking/model"
	"time"
	"fmt"
)

var DB *sql.DB

func InitMySQL(dsn string) {
	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
}

func InsertEvents(events []model.Event) error {
	// 按照 client_type、site、event_type、event_detail 分组统计
	eventCountMap := make(map[string]int)
	for _, e := range events {
		key := fmt.Sprintf("%s:%s:%s:%s", e.ClientType, e.Site, e.EventType, e.EventDetail)
		eventCountMap[key]++
	}

	// 插入数据库
	stmt, err := DB.Prepare("INSERT INTO event_logs (timestamp, client_type, site, event_type, event_detail, count) VALUES (?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE count = count + VALUES(count)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for key, count := range eventCountMap {
		// 分离出 key 中的字段
		var clientType, site, eventType, eventDetail string
		_, err := fmt.Sscanf(key, "%s:%s:%s:%s", &clientType, &site, &eventType, &eventDetail)
		if err != nil {
			return err
		}

		// 执行插入语句
		_, err = stmt.Exec(time.Now().Unix(), clientType, site, eventType, eventDetail, count)
		if err != nil {
			continue
		}
	}
	return nil
}
