// Package main 是 API 服务的程序入口，负责启动 HTTP 服务并注册路由。
package main

import (
	"log/slog"
	"strconv"
	"vehivle/configs"
	"vehivle/internal/bootstrap"
)

// main 是程序入口函数，初始化 Gin 引擎、注册健康检查接口并启动 HTTP 服务。
func main() {
	// 加载配置。
	cfg, err := configs.Load()
	if err != nil {
		slog.Error("加载配置失败", "error", err)
		panic(err)
	}
	// 创建 Bootstrap 实例。
	b := bootstrap.New(cfg)
	// 启动 HTTP 服务。
	r, err := b.Run()
	// 如果启动失败，记录错误并退出。
	if err != nil {
		slog.Error("启动服务失败", "error", err)
		panic(err)
	}
	// 启动 HTTP 服务，监听指定端口。
	r.Run(":" + strconv.Itoa(cfg.App.Port))
	// 优雅关闭，后续增加优雅关闭。

	// defer b.Close()
}
