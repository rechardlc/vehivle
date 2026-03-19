// Package main 是 API 服务的程序入口，负责启动 HTTP 服务并注册路由。
package main

import (
	"log/slog"
	"os"
	"strconv"
	"vehivle/configs"
	"vehivle/internal/bootstrap"
)

// main 是程序入口函数，初始化 Gin 引擎、注册健康检查接口并启动 HTTP 服务。
func main() {
	cfg, err := configs.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	// 创建 Bootstrap 实例。
	b := bootstrap.New(cfg)
	// 启动 HTTP 服务。
	r, err := b.Run()
	// 如果启动失败，记录错误并退出。
	if err != nil {
		slog.Error("failed to run", "error", err)
		os.Exit(1)
	}
	// 启动 HTTP 服务，监听指定端口。
	r.Run(":" + strconv.Itoa(cfg.App.Port))
	// 优雅关闭，后续增加优雅关闭。
}
