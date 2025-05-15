package storage

import (
	"context"
	"encoding/json"
	"time"

	"github.com/qwy-tacking/model"

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
	event.TimeStamp = time.Now().Unix()
	if event.Count == 0 {
		event.Count = 1 // 默认计数为1
	}
	data, _ := json.Marshal(event)
	return RDB.RPush(context.Background(), "event_queue", data).Err()
}

func PopNEvents(n int) []model.Event {
	ctx := context.Background()

	vals, err := RDB.LPopCount(ctx, "event_queue", n).Result()
	if err != nil || len(vals) == 0 {
		return nil
	}

	var result []model.Event
	for _, v := range vals {
		var e model.Event
		if err := json.Unmarshal([]byte(v), &e); err == nil {
			result = append(result, e)
		}
	}
	return result
}
