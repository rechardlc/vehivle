package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/gin-gonic/gin"

	"vehivle/configs"
	"vehivle/internal/transport/http/router"
	"vehivle/pkg/logger"
)

// Bootstrap 是应用启动器，负责初始化应用依赖并启动 HTTP 服务。
type Bootstrap struct {
	cfg    *configs.Conf
	logger logger.Logger
}

// New 创建一个新的 Bootstrap 实例。
func New(cfg *configs.Conf) *Bootstrap {
	return &Bootstrap{
		cfg:    cfg,
		logger: logger.NewLogger(logger.Env(cfg.App.Env)),
	}
}

// Run 启动 HTTP 服务。
func (b *Bootstrap) Run() (*gin.Engine, error) {
	b.logger.Info(
		// 使用 context.Background() 创建一个空的 context，用于记录日志。
		context.Background(),
		 "bootstrap run",
		 slog.String("env", b.cfg.App.Env), // 记录环境变量。
		 slog.Int("port", b.cfg.App.Port), // 记录端口号。
		 slog.String("config", fmt.Sprintf("conf.%s.yaml", b.cfg.App.Env)), // 记录配置文件路径。
		)
	// 创建一个 Gin 引擎。
	r := gin.New()
	// 使用 Gin 的 Logger 中间件，用于记录请求日志。
	r.Use(gin.Logger())
	// 使用 Gin 的 Recovery 中间件，用于捕获 panic 并返回 500 错误。
	r.Use(gin.Recovery())
	// 使用 logger.RequestID() 中间件，用于生成请求 ID。
	r.Use(logger.RequestID())
	// 使用 logger.AccessLog(b.logger) 中间件，用于记录访问日志。
	r.Use(logger.AccessLog(b.logger))
	// 注册 API 路由（admin、public 分组）及健康检查。
	if err := router.New(r, b.logger).Register(); err != nil {
		return nil, err
	}
	// 记录启动成功日志。
	b.logger.Info(context.Background(), "bootstrap ready", slog.String("listen", ":"+strconv.Itoa(b.cfg.App.Port)))
	// 返回 Gin 引擎。
	return r, nil
}

// 后续增加DB、redis、oss、jwt、Ping等检查。