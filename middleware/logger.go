package middleware

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

var Logger *log.Logger

// 启动日志系统，并在每天12点轮转
func InitLogger(logDir string) {
	os.MkdirAll(logDir, 0755)

	go func() {
		for {
			now := time.Now()
			nextNoon := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, now.Location())
			if now.Hour() >= 12 {
				// 已经过了12点，轮转到明天中午
				nextNoon = nextNoon.Add(24 * time.Hour)
			}
			duration := nextNoon.Sub(now)
			time.Sleep(duration)
			setupLogFile(logDir)
		}
	}()

	// 初始日志文件
	setupLogFile(logDir)
}

func setupLogFile(logDir string) {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logPath := filepath.Join(logDir, fmt.Sprintf("track_%s.log", timestamp))
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	Logger = log.New(f, "[Track] ", log.LstdFlags)
}
