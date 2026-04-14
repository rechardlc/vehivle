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
	pginfra "vehivle/internal/infrastructure/postgres"
	pgrepo "vehivle/internal/repository/postgres"
	"vehivle/internal/service/auth"
	"vehivle/internal/service/category"
	"vehivle/internal/service/param_template"
	"vehivle/internal/service/system_setting"
	"vehivle/internal/service/vehicle"
	"vehivle/internal/transport/http/handler"
	"vehivle/internal/transport/http/router"
	"vehivle/pkg/logger"

	"gorm.io/gorm"

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
	if err := b.pgsqlConnPool(); err != nil {
		return nil, err
	}
	if err := pginfra.Ping(context.Background(), b.db, false); err != nil {
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

// buildHandlers 构建 handlers 集合
func (b *Bootstrap) buildHandlers() *router.Handlers {
	// repos
	userRepo := pgrepo.NewUserRepo(b.db)
	catRepo := pgrepo.NewCategoryRepo(b.db)
	vehRepo := pgrepo.NewVehicleRepo(b.db)
	sysRepo := pgrepo.NewSysSettings(b.db)
	mediaRepo := pgrepo.NewMediaAssetRepo(b.db)
	plateTemRepo := pgrepo.NewParamTemplateRepo(b.db)

	// services
	authSvc := auth.NewService(userRepo, auth.JWTConfig{
		Secret:             b.cfg.JWT.Secret,
		RefreshSecret:      b.cfg.JWT.RefreshSecret,
		ExpireHours:        b.cfg.JWT.ExpireHours,
		RefreshExpireHours: b.cfg.JWT.RefreshExpireHours,
	})
	catSvc := category.NewCategoryService(catRepo)
	vehSvc := vehicle.NewService(vehRepo)
	sysSvc := system_setting.NewSysService(sysRepo)
	plateTemSvc := param_template.NewParamTemplateService(plateTemRepo)

	// handlers
	return &router.Handlers{
		Auth:           handler.NewAuth(authSvc, b.cfg.JWT),
		User:           handler.NewUser(),
		Vehicles:       handler.NewVehicles(vehSvc, catSvc, mediaRepo, b.ossClient),
		Categories:     handler.NewCategories(catSvc),
		System:         handler.NewSysSettings(sysSvc, b.ossClient, mediaRepo),
		Upload:         handler.NewUpload(b.ossClient, mediaRepo),
		ParamTemplates: handler.NewParamTemplates(plateTemSvc),
		Public:         handler.NewPublic(vehSvc, catSvc, sysSvc, mediaRepo, b.ossClient),
	}
}

func (b *Bootstrap) injectRouter(gin *gin.Engine) error {
	h := b.buildHandlers()
	if err := router.New(gin, b.logger, b.ossClient, b.cfg.JWT, h).Register(); err != nil {
		b.logger.Error(context.Background(), "注册路由失败", slog.String("error", err.Error()))
		return fmt.Errorf("注册路由失败: %w", err)
	}
	return nil
}

// pgsqlConnPool 初始化 PostgreSQL 连接池
func (b *Bootstrap) pgsqlConnPool() error {
	db, err := pginfra.Open(&b.cfg.Database, b.logger)
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
	// 直链基址：优先 public_url（解决 endpoint 为 Docker 内部主机名时浏览器无法访问）
	publicURL := strings.TrimSpace(b.cfg.Oss.PublicURL)
	publicURL = strings.TrimSuffix(publicURL, "/")
	if publicURL == "" {
		scheme := "http"
		if useSSL {
			scheme = "https"
		}
		publicURL = fmt.Sprintf("%s://%s", scheme, endpoint)
	}
	if b.cfg.Oss.EnablePublicRead {
		if err := applyOssPublicReadPolicy(ctx, minioClient, b.cfg.Oss.Bucket); err != nil {
			return fmt.Errorf("设置 OSS 桶公开读策略失败（直链需匿名 GetObject，或关闭 oss.enable_public_read 改用签名 URL）: %w", err)
		}
	}
	b.ossClient = oss.MinioClient{
		Endpoint:  endpoint,
		PublicURL: publicURL,
		Bucket:    b.cfg.Oss.Bucket,
		Client:    minioClient,
	}
	return nil
}

// applyOssPublicReadPolicy 为 Bucket 设置匿名 s3:GetObject，使上传接口返回的 HTTP 直链可在浏览器访问（本地 MinIO 常用）。
func applyOssPublicReadPolicy(ctx context.Context, client *minio.Client, bucket string) error {
	policy := fmt.Sprintf(
		`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::%s/*"]}]}`,
		bucket,
	)
	return client.SetBucketPolicy(ctx, bucket, policy)
}
