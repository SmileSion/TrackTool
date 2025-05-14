package service

import (
	"github.com/qwy-tacking/middleware"
	"github.com/qwy-tacking/storage"
	"time"
)

func StartProcessor() {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			events := storage.PopAllEvents()
			if len(events) > 0 {
				err := storage.InsertEvents(events)
				if err != nil {
					middleware.Logger.Println("写入MySQL失败:", err)
				} else {
					middleware.Logger.Printf("写入MySQL成功：%d 条\n", len(events))
				}
			}
		}
	}()
}
