package storage

import (
	"context"
	"encoding/json"
	"github.com/qwy-tacking/model"
	"time"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func InitRedis(addr string, password string, db int) {
	RDB = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
}

func SaveEventToRedis(event model.Event) error {
	// 忽略 user_detail 字段
	event.UserDetail = ""  // 直接清空 user_detail
	event.Timestamp = time.Now().Unix()
	data, _ := json.Marshal(event)
	return RDB.RPush(context.Background(), "event_queue", data).Err()
}

func PopAllEvents() []model.Event {
	ctx := context.Background()
	vals, _ := RDB.LRange(ctx, "event_queue", 0, -1).Result()
	RDB.Del(ctx, "event_queue")

	var result []model.Event
	for _, v := range vals {
		var e model.Event
		json.Unmarshal([]byte(v), &e)
		result = append(result, e)
	}
	return result
}
