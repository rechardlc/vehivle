package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"vehivle/configs"
	"vehivle/internal/infrastructure/oss"
	"vehivle/internal/transport/http/router"
	"vehivle/pkg/logger"

	"gorm.io/gorm"
	"vehivle/internal/infrastructure/postgres"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Bootstrap 是应用启动器，负责初始化应用依赖并启动 HTTP 服务。
type Bootstrap struct {
	cfg       *configs.Conf
	logger    logger.Logger
	db        *gorm.DB
	ossClient oss.MinioClient
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
		context.Background(),
		"bootstrap run",
		slog.String("env", b.cfg.App.Env),
		slog.Int("port", b.cfg.App.Port),
		slog.String("config", fmt.Sprintf("conf.%s.yaml", b.cfg.App.Env)),
	)
	r := gin.New()
	// 前后端分离项目无需尾斜杠自动重定向
	r.RedirectTrailingSlash = false
	// 限制 multipart 表单内存占用（10MB）
	r.MaxMultipartMemory = 10 << 20
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(logger.RequestID())
	r.Use(logger.AccessLog(b.logger))
	// 打开数据库连接
	if err := b.pgsqlConnPool(); err != nil {
		return nil, err
	}
	// 检查数据库连接
	if err := postgres.Ping(context.Background(), b.db, false); err != nil {
		return nil, err
	}
	// 对象存储：MinIO/S3 为强依赖，启动时必须连通并确保 Bucket 存在
	if err := b.ossConnPool(); err != nil {
		return nil, err
	}
	b.logger.Info(context.Background(), "OSS 连接成功", slog.String("bucket", b.cfg.Oss.Bucket))
	// 注册 API 路由（admin、public 分组）及健康检查
	if err := b.injectRouter(r); err != nil {
		return nil, err
	}
	b.logger.Info(context.Background(), "启动成功", slog.String("listen", ":"+strconv.Itoa(b.cfg.App.Port)))
	return r, nil
}

func (b *Bootstrap) injectRouter(gin *gin.Engine) error {
	if err := router.New(gin, b.logger, b.db, b.ossClient).Register(); err != nil {
		b.logger.Error(context.Background(), "注册路由失败", slog.String("error", err.Error()))
		return fmt.Errorf("注册路由失败: %w", err)
	}
	return nil
}

// pgsqlConnPool 初始化 PostgreSQL 连接池
func (b *Bootstrap) pgsqlConnPool() error {
	db, err := postgres.Open(&b.cfg.Database, b.logger)
	if err != nil {
		return fmt.Errorf("打开数据库失败: %w", err)
	}
	b.db = db
	return nil
}

// normalizeOssEndpoint 将配置中的 URL 转为 minio.New 所需的 host:port（无 scheme），并根据 http/https 推导 TLS。
func normalizeOssEndpoint(raw string, cfgUseSSL bool) (endpoint string, useSSL bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", useSSL
	}
	lower := strings.ToLower(raw)
	useSSL = cfgUseSSL
	switch {
	case strings.HasPrefix(lower, "https://"):
		useSSL = true
		raw = raw[8:]
	case strings.HasPrefix(lower, "http://"):
		useSSL = false
		raw = raw[7:]
	}
	if i := strings.Index(raw, "/"); i >= 0 {
		raw = raw[:i]
	}
	return strings.TrimSpace(raw), useSSL
}

// ossConnPool 初始化 MinIO 客户端并确保 Bucket 存在。
func (b *Bootstrap) ossConnPool() error {
	ctx := context.Background()
	endpoint, useSSL := normalizeOssEndpoint(b.cfg.Oss.Endpoint, b.cfg.Oss.UseSSL)
	if endpoint == "" {
		return fmt.Errorf("OSS 连接地址为空")
	}
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(b.cfg.Oss.AccessKey, b.cfg.Oss.SecretKey, ""),
		Region: b.cfg.Oss.Region,
		Secure: useSSL,
	})
	if err != nil {
		return fmt.Errorf("创建 OSS 客户端失败: %w", err)
	}
	exists, err := minioClient.BucketExists(ctx, b.cfg.Oss.Bucket)
	if err != nil {
		return fmt.Errorf("判断存储桶是否存在失败: %w", err)
	}
	if !exists {
		err = minioClient.MakeBucket(ctx, b.cfg.Oss.Bucket, minio.MakeBucketOptions{Region: b.cfg.Oss.Region})
		if err != nil {
			return fmt.Errorf("创建存储桶失败: %w", err)
		}
	}
	// 根据 TLS 推导 scheme，构建前端可用的公开访问基址
	scheme := "http"
	if useSSL {
		scheme = "https"
	}
	b.ossClient = oss.MinioClient{
		Endpoint:  endpoint,
		PublicURL: fmt.Sprintf("%s://%s", scheme, endpoint),
		Bucket:    b.cfg.Oss.Bucket,
		Client:    minioClient,
	}
	return nil
}

// 后续可增加 redis、jwt、Ping 等检查。
