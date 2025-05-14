package main

import (
	"github.com/qwy-tacking/config"
	"github.com/qwy-tacking/controller"
	"github.com/qwy-tacking/middleware"
	"github.com/qwy-tacking/service"
	"github.com/qwy-tacking/storage"
	"github.com/gin-gonic/gin"
	"strconv"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	config.InitConfig()
	middleware.InitLogger(config.Conf.Log.Filepath)
	storage.InitRedis(config.Conf.Redis.Addr, config.Conf.Redis.Password, config.Conf.Redis.DB)
	storage.InitMySQL(config.Conf.Mysql.DSN)
	service.StartProcessor()

	r := gin.Default()
	r.POST("/track", controller.TrackHandler)

	portStr := strconv.Itoa(config.Conf.Server.Port)
	r.Run(":" + portStr)
}
