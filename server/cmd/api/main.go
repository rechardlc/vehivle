// Package main 是 API 服务的程序入口，负责启动 HTTP 服务并注册路由。
package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"

	"vehivle/configs"
	"vehivle/pkg/logger"
	"vehivle/pkg/response"
)

// main 是程序入口函数，初始化 Gin 引擎、注册健康检查接口并启动 HTTP 服务。
func main() {
	// gin.Default() 创建默认引擎，已内置 Logger（请求日志）和 Recovery（panic 恢复）中间件
	// 加载配置
	cfg, err := configs.Load()
	fmt.Printf("%+v\n", cfg)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	// 创建 Gin 引擎
	r := gin.New()
	// 使用 Logger 中间件，用于记录请求日志
	r.Use(gin.Logger())
	// 使用 Recovery 中间件，用于捕获 panic 并返回 500 错误
	r.Use(gin.Recovery())
	// 使用 RequestID 中间件，用于生成请求 ID
	r.Use(logger.RequestID())
	// 使用 AccessLog 中间户，用于记录访问日志
	// logger.Env(env) 将 env 转换为 logger.Env 类型
	r.Use(logger.AccessLog(logger.NewLogger(logger.Env(cfg.App.Env))))
	// 注册健康检查接口：GET /health，用于探活、负载均衡健康检查等
	r.GET("/health", func(c *gin.Context) {
		rid, _ := c.Get(logger.GinRequestIDKey)
		logger.NewLogger(logger.Env(cfg.App.Env)).Info(c.Request.Context(), "health check", slog.String("request_id", rid.(string)))
		response.Success(c, gin.H{
			"status":  "ok",
			"message": "API is running",
		})
	})
	r.Run(":" + strconv.Itoa(cfg.App.Port))
}
