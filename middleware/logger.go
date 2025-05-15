package middleware

import (
	"log"

	"gopkg.in/natefinch/lumberjack.v2"
	"github.com/qwy-tacking/config"
)

var Logger *log.Logger

func InitLogger(logPath string) {
	Logger = log.New(&lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    config.Conf.Log.MaxSize,
		MaxBackups: config.Conf.Log.MaxBackups,
		MaxAge:     config.Conf.Log.MaxAge,
		Compress:   config.Conf.Log.Compress,
	}, "[Track] ", log.LstdFlags)
}
