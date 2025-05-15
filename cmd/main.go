package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/qwy-tacking/config"
	"github.com/qwy-tacking/controller"
	"github.com/qwy-tacking/middleware"
	"github.com/qwy-tacking/service"
	"github.com/qwy-tacking/storage"
	"log"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 初始化配置
	config.InitConfig()
	log.Printf("[Init] MySQL DSN: %s\n", config.Conf.Mysql.DSN)

	// 初始化日志、数据库、Redis
	middleware.InitLogger(config.Conf.Log.Filepath)
	storage.InitRedis(config.Conf.Redis.Addr, config.Conf.Redis.Password, config.Conf.Redis.DB)
	storage.InitMySQL(config.Conf.Mysql.DSN)

	// 启动异步处理器
	service.StartProcessor()

	// 初始化 Gin 引擎
	r := gin.Default()

	// CORS 跨域配置
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // 建议部署时改为固定前端域名
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 路由注册
	r.POST("/track", controller.TrackHandler)

	// 启动服务
	portStr := strconv.Itoa(config.Conf.Server.Port)
	log.Printf("[Server] Listening on port %s\n", portStr)
	if err := r.Run(":" + portStr); err != nil {
		log.Fatalf("[Server] 启动失败: %v\n", err)
	}
}
