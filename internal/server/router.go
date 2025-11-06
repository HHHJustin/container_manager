package server

import (
	"os"

	"github.com/gin-gonic/gin"

	"container-manager/internal/api/handlers"
	"container-manager/internal/middleware"
)

func Run(addr string) error {
	engine := gin.New()
	engine.Use(middleware.Logger())
	engine.Use(middleware.Recover())
	engine.Use(middleware.Cors())

	// 健康檢查
	engine.GET("/healthz", handlers.Health)

	// 登入（回傳 JWT）
	engine.POST("/login", handlers.Login)

	// 受保護的 API v1
	v1 := engine.Group("/v1")
	v1.Use(middleware.Auth())
	{
		v1.POST("/uploads", handlers.Upload)
		v1.POST("/containers", handlers.CreateContainer)
		v1.POST("/containers/:id/start", handlers.StartContainer)
		v1.POST("/containers/:id/stop", handlers.StopContainer)
		v1.DELETE("/containers/:id", handlers.DeleteContainer)
		v1.POST("/jobs", handlers.RunJob)
	}

	// 靜態檔案（上傳存放根目錄由 DATA_DIR 控制）
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}
	_ = os.MkdirAll(dataDir, 0o755)
	engine.Static("/static", dataDir)

	return engine.Run(addr)
}
