package main

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/qwy-tacking/config"
	"github.com/qwy-tacking/controller"
	"github.com/qwy-tacking/middleware"
	"github.com/qwy-tacking/service"
	"github.com/qwy-tacking/storage"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
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

	// 设置 context 和 WaitGroup 以支持优雅停机
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	// 启动异步处理器（支持退出）
	service.StartProcessor(ctx, config.Conf.Cap.Worker, config.Conf.Cap.Batch, &wg)

	// 捕获退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 先设置运行模式
	gin.SetMode(gin.ReleaseMode)

	// 初始化 Gin 引擎
	r := gin.Default()

	// CORS 跨域配置
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 路由注册
	r.POST("/track", controller.TrackHandler)

	// 启动服务器（独立 goroutine 以便监听退出）
	serverErr := make(chan error, 1)
	go func() {
		portStr := strconv.Itoa(config.Conf.Server.Port)
		log.Printf("[Server] Listening on port %s\n", portStr)
		serverErr <- r.Run(":" + portStr)
	}()

	select {
	case <-quit:
		log.Println("[Shutdown] 收到退出信号，正在优雅退出...")
		cancel()
		wg.Wait()
		log.Println("[Shutdown] 后台任务已退出，程序结束")
	case err := <-serverErr:
		log.Fatalf("[Server] 启动失败: %v\n", err)
	}
}
