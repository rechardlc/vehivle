package router

import (
	"log/slog"
	"vehivle/internal/infrastructure/oss"
	"vehivle/internal/transport/http/handler"
	"vehivle/internal/transport/http/middleware"
	"vehivle/pkg/logger"
	"vehivle/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Router 路由注册器，持有 Gin 引擎并负责 API 路由分组与挂载。
type Router struct {
	engine *gin.Engine
	logger logger.Logger
	db     *gorm.DB
	oss    oss.MinioClient
}

// New 创建 Router 实例。
func New(engine *gin.Engine, logger logger.Logger, db *gorm.DB, ossClient oss.MinioClient) *Router {
	return &Router{engine: engine, logger: logger, db: db, oss: ossClient}
}

// Register 注册 API 路由分组（admin、public）、健康检查等路由。
func (r *Router) Register() error {
	// 健康检查：GET /health，用于探活、负载均衡健康检查等。
	r.engine.GET("/health", r.healthHandler)
	// 注册admin路由
	v1 := r.engine.Group("/api/v1")
	// 注册admin路由组
	admin := v1.Group("/admin")
	{
		// 注册user路由组
		user := admin.Group("/user")
		userHandler := handler.NewUser(r.db)
		{
			user.GET("/:user_id", userHandler.Get)
			user.POST("", userHandler.Create)
			user.PUT("/:user_id", userHandler.Update)
			user.DELETE("/:user_id", userHandler.Delete)
		}
		// 注册vehicles路由组
		vehicles := admin.Group("/vehicles")
		vehiclesHandler := handler.NewVehicles(r.db)
		{
			vehicles.GET("", vehiclesHandler.List)
			vehicles.POST("", vehiclesHandler.Create)
			vehicles.PUT("/:vehicle_id", vehiclesHandler.Update)
			vehicles.DELETE("/:vehicle_id", vehiclesHandler.Delete)
		}
		// 注册categories路由组
		categories := admin.Group("/categories")
		categoriesHandler := handler.NewCategories(r.db)
		{
			categoryList := []string{"keyword", "level", "status", "page", "pageSize", "sortField", "sortOrder"}
			categories.GET("", middleware.ValidateParams(categoryList), categoriesHandler.List)
			categories.POST("", categoriesHandler.Create)
			categories.PUT("/:category_id", categoriesHandler.Update)
			categories.DELETE("/:category_id", categoriesHandler.Delete)
		}
		// OSS 上传（TODO: 上线前挂载认证中间件）
		uploadHandler := handler.NewUpload(r.oss)
		admin.POST("/upload/images", uploadHandler.UploadImages)
	}
	// 注册public路由
	public := v1.Group("/public")
	{
		public.GET("/vehicles", handler.NewVehicles(r.db).List)
	}
	// 通配：未注册路由返回统一 404
	r.engine.NoRoute(r.noRouteHandler)
	// 通配：路由存在但方法不匹配返回 405
	r.engine.NoMethod(r.noMethodHandler)
	return nil
}

// noRouteHandler 未匹配到任何路由时返回统一 JSON 404。
func (r *Router) noRouteHandler(c *gin.Context) {
	response.FailNotFound(c, "route not found")
}

// noMethodHandler 路由存在但 HTTP 方法不匹配时返回 405。
func (r *Router) noMethodHandler(c *gin.Context) {
	response.FailMethodNotAllowed(c, "method not allowed")
}

// healthHandler 健康检查接口处理函数。
func (r *Router) healthHandler(c *gin.Context) {
	rid := ""
	if v, ok := c.Get(logger.GinRequestIDKey); ok {
		if s, ok := v.(string); ok {
			rid = s
		}
	}
	r.logger.Info(c.Request.Context(), "health check", slog.String("request_id", rid))
	response.Success(c, gin.H{
		"status":   "ok",
		"message":  "API is running",
		"ossReady": r.oss.Bucket != "",
	})
}
