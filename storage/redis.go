package storage

import (
	"context"
	"encoding/json"
	"time"

	"github.com/qwy-tacking/middleware"
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
		event.Count = 1
	}
	data, _ := json.Marshal(event)

	ctx := context.Background()
	pipe := RDB.Pipeline()
	pipe.RPush(ctx, "event_queue_stats", data)
	pipe.RPush(ctx, "event_queue_detail", data)
	_, err := pipe.Exec(ctx)
	return err
}


// func PopNEvents(n int) []model.Event {
// 	ctx := context.Background()
// 	middleware.Logger.Printf("尝试取出 %d 个数据",n)

// 	vals, err := RDB.LPopCount(ctx, "event_queue", n).Result()
// 	if err != nil || len(vals) == 0 {
// 		middleware.Logger.Printf("redis取出错误 %v",err)
// 		return nil
// 	}

//		var result []model.Event
//		for _, v := range vals {
//			var e model.Event
//			if err := json.Unmarshal([]byte(v), &e); err == nil {
//				middleware.Logger.Printf("反序列化事件出错： %v，原始数据 %s",err,v)
//				result = append(result, e)
//			}
//		}
//		return result
//	}
const lpopCountScript = `
local count = tonumber(ARGV[1])
local results = {}
for i = 1, count do
    local val = redis.call('LPOP', KEYS[1])
    if not val then break end
    table.insert(results, val)
end
return results
`

func PopNEvents(n int) []model.Event {
	ctx := context.Background()
	// middleware.Logger.Printf("尝试取出 %d 个数据", n)

	// 执行 Lua 脚本
	vals, err := RDB.Eval(ctx, lpopCountScript, []string{"event_queue_stats"}, n).StringSlice()
	if err != nil {
		if err == redis.Nil {
			// middleware.Logger.Println("队列为空")
		} else {
			middleware.Logger.Printf("redis取出错误: %v", err)
		}
		return nil
	}

	if len(vals) == 0 {
		// middleware.Logger.Println("队列为空，没有取出任何数据")
		return nil
	}

	var result []model.Event
	for _, v := range vals {
		var e model.Event
		if err := json.Unmarshal([]byte(v), &e); err != nil {
			middleware.Logger.Printf("反序列化事件出错: %v，原始数据: %s", err, v)
			continue
		}
		result = append(result, e)
	}
	middleware.Logger.Printf("成功取出并反序列化 %d/%d 个事件", len(result), len(vals))
	return result
}

func PopNDetailEvents(n int) []model.Event {
	ctx := context.Background()
	vals, err := RDB.Eval(ctx, lpopCountScript, []string{"event_queue_detail"}, n).StringSlice()
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
