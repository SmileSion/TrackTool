package middleware

import (
	"log"
	"os"
)

var Logger *log.Logger

func InitLogger(filepath string) {
	f, err := os.OpenFile(filepath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	Logger = log.New(f, "[Track] ", log.LstdFlags)
}