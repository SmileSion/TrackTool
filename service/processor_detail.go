package service

import (
	"context"
	"sync"
	"time"

	"github.com/qwy-tacking/middleware"
	"github.com/qwy-tacking/storage"
	"github.com/qwy-tacking/model"
)

func StartDetailProcessor(ctx context.Context, workerCount int, batchSize int, wg *sync.WaitGroup) {
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()

			var buffer []model.Event

			for {
				select {
				case <-ctx.Done():
					middleware.Logger.Printf("[DetailWorker-%d] 接收到退出信号，处理剩余 %d 条数据", workerID, len(buffer))
					if len(buffer) > 0 {
						_ = storage.InsertEventDetails(buffer)
					}
					return

				case <-ticker.C:
					if len(buffer) > 0 {
						if err := storage.InsertEventDetails(buffer); err != nil {
							middleware.Logger.Printf("[DetailWorker-%d] 定时批量写入失败: %v", workerID, err)
						} else {
							middleware.Logger.Printf("[DetailWorker-%d] 定时写入明细: %d 条", workerID, len(buffer))
						}
						buffer = buffer[:0] // 清空缓冲区
					}

				default:
					events := storage.PopNDetailEvents(batchSize)
					if len(events) == 0 {
						time.Sleep(500 * time.Millisecond)
						continue
					}

					buffer = append(buffer, events...)
					if len(buffer) >= batchSize {
						if err := storage.InsertEventDetails(buffer); err != nil {
							middleware.Logger.Printf("[DetailWorker-%d] 批量写入失败: %v", workerID, err)
						} else {
							middleware.Logger.Printf("[DetailWorker-%d] 写入明细: %d 条", workerID, len(buffer))
						}
						buffer = buffer[:0]
					}
				}
			}
		}(i)
	}
}
