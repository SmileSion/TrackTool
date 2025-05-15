package service

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/qwy-tacking/middleware"
	"github.com/qwy-tacking/model"
	"github.com/qwy-tacking/storage"
	"github.com/qwy-tacking/config"
)

var (
	eventChan chan []model.Event
	countMap  map[string]int
	mu        sync.Mutex
)

func init() {
	eventChan = make(chan []model.Event, config.Conf.Cap.Chan) // 带缓冲防止阻塞
	countMap = make(map[string]int)
}

func StartProcessor(ctx context.Context, workerCount int, batchSize int, wg *sync.WaitGroup) {
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					events := storage.PopNEvents(batchSize)
					if len(events) == 0 {
						time.Sleep(500 * time.Millisecond)
						continue
					}
					select {
					case eventChan <- events:
					case <-ctx.Done():
						return
					}
				}
			}
		}(i)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case events := <-eventChan:
				mu.Lock()
				for _, e := range events {
					key := strings.Join([]string{e.ClientType, e.Site, e.EventType, e.EventDetail}, ":")
					// oldCount := countMap[key] //测试使用
					countMap[key] += e.Count
					// middleware.Logger.Printf("累加事件: %s, 旧值: %d, 新值: %d", key, oldCount, countMap[key])
				}
				mu.Unlock()

			case <-ticker.C:
				mu.Lock()
				if len(countMap) > 0 {
					toInsert := make(map[string]int, len(countMap))
					for k, v := range countMap {
						toInsert[k] = v
					}
					countMap = make(map[string]int)
					mu.Unlock()

					var batch []model.Event
					timestamp := time.Now().Unix() / 60 * 60
					for key, count := range toInsert {
						parts := strings.SplitN(key, ":", 4)
						if len(parts) < 4 {
							middleware.Logger.Printf("非法Key跳过: %s", key)
							continue
						}
						batch = append(batch, model.Event{
							ClientType:  parts[0],
							Site:        parts[1],
							EventType:   parts[2],
							EventDetail: parts[3],
							Count:       count,
							TimeStamp:   timestamp,
						})
					}

					if err := storage.InsertEvents(batch); err != nil {
						middleware.Logger.Println("写入MySQL失败:", err)
					} else {
						middleware.Logger.Printf("写入MySQL成功：%d 条\n", len(batch))
					}
				} else {
					mu.Unlock()
				}
			}
		}
	}()
}
