package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/secure"
	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	limiterGin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"

	"movingManager/database"
	"movingManager/router"
)

func main() {
	// 初始化数据库
	if err := database.InitDB(); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	// 数据迁移
	// migrate.Migrate()

	// 创建Gin引擎
	r := gin.Default()

	// 配置CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 配置安全头部
	r.Use(secure.New(secure.Config{
		BrowserXssFilter:     true,
		ContentTypeNosniff:   true,
		FrameDeny:            true,
		STSSeconds:           31536000,
		STSIncludeSubdomains: false,
		STSPreload:           true,
		AllowedHosts:         []string{},
	}))

	// 配置API限流 (100次/分钟)
	store := memory.NewStore()
	rate := limiter.Rate{Period: 1 * time.Minute, Limit: 100}
	limiterMiddleware := limiterGin.NewMiddleware(limiter.New(store, rate))
	r.Use(limiterMiddleware)

	// 注册路由
	router.RegisterRoutes(r)

	// 启动服务器
	port := "0.0.0.0:8080"
	log.Printf("服务器启动成功，监听端口: %s", port)
	if err := r.Run(port); err != nil && err != http.ErrServerClosed {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
