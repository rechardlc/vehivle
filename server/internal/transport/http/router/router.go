package router

import (
	"log/slog"
	"vehivle/configs"
	"vehivle/internal/infrastructure/oss"
	"vehivle/internal/transport/http/handler"
	"vehivle/internal/transport/http/middleware"
	"vehivle/pkg/logger"
	"vehivle/pkg/response"

	"github.com/gin-gonic/gin"
)

// Handlers 所有已装配好的 HTTP handler 集合，由 bootstrap 组装后传入。
type Handlers struct {
	Auth           *handler.Auth
	User           *handler.User
	Vehicles       *handler.Vehicles
	Categories     *handler.Categories
	System         *handler.System
	Upload         *handler.Upload
	ParamTemplates *handler.ParamTemplates
	Public         *handler.Public
}

type Router struct {
	engine   *gin.Engine
	logger   logger.Logger
	oss      oss.MinioClient
	jwt      configs.JWTConfig
	handlers *Handlers
}

func New(engine *gin.Engine, logger logger.Logger, ossClient oss.MinioClient, jwtCfg configs.JWTConfig, h *Handlers) *Router {
	return &Router{engine: engine, logger: logger, oss: ossClient, jwt: jwtCfg, handlers: h}
}

func (r *Router) Register() error {
	r.engine.GET("/health", r.healthHandler)

	authGroup := r.engine.Group("/api/v1/admin/auth")
	{
		authGroup.POST("/login", r.handlers.Auth.Login)
		authGroup.POST("/refresh", r.handlers.Auth.Refresh)
		authGroup.POST("/logout", r.handlers.Auth.Logout)
		authGroup.GET("/me", middleware.JWTAuth(r.jwt.Secret), r.handlers.Auth.Me)
	}

	v1 := r.engine.Group("/api/v1")
	admin := v1.Group("/admin")
	admin.Use(middleware.JWTAuth(r.jwt.Secret))
	{
		user := admin.Group("/user")
		{
			user.GET("/:user_id", r.handlers.User.Get)
			user.POST("", r.handlers.User.Create)
			user.PUT("/:user_id", r.handlers.User.Update)
			user.DELETE("/:user_id", r.handlers.User.Delete)
		}

		vehicles := admin.Group("/vehicles")
		vehicleList := []string{"keyword", "categoryId", "status", "page", "pageSize", "sortField", "sortOrder"}
		{
			vehicles.POST("/batch-status", r.handlers.Vehicles.BatchStatus)
			vehicles.POST("/:vehicle_id/publish", r.handlers.Vehicles.Publish)
			vehicles.POST("/:vehicle_id/unpublish", r.handlers.Vehicles.Unpublish)
			vehicles.POST("/:vehicle_id/duplicate", r.handlers.Vehicles.Duplicate)
			vehicles.GET("/:vehicle_id/detail-images", r.handlers.Vehicles.DetailImages)
			vehicles.PUT("/:vehicle_id/detail-images", r.handlers.Vehicles.SaveDetailImages)
			vehicles.GET("", middleware.ValidateParams(vehicleList), r.handlers.Vehicles.List)
			vehicles.POST("", r.handlers.Vehicles.Create)
			vehicles.PUT("/:vehicle_id", r.handlers.Vehicles.Update)
			vehicles.DELETE("/:vehicle_id", r.handlers.Vehicles.Delete)
		}

		categories := admin.Group("/categories")
		{
			categoryList := []string{"keyword", "level", "status", "page", "pageSize", "sortField", "sortOrder"}
			categories.GET("", middleware.ValidateParams(categoryList), r.handlers.Categories.List)
			categories.POST("", r.handlers.Categories.Create)
			categories.PUT("/:category_id", r.handlers.Categories.Update)
			categories.DELETE("/:category_id", r.handlers.Categories.Delete)
		}

		sys := admin.Group("/system-settings")
		{
			sys.GET("", r.handlers.System.Detail)
			sys.POST("", r.handlers.System.Create)
			sys.PUT("", r.handlers.System.Update)
		}

		paramTemplates := admin.Group("/param-templates")
		{
			templateList := []string{"page", "pageSize"}
			paramTemplates.GET("/list", middleware.ValidateParams(templateList), r.handlers.ParamTemplates.List)
			paramTemplates.POST("", r.handlers.ParamTemplates.Create)
			paramTemplates.PUT("/:id", r.handlers.ParamTemplates.Update)
			paramTemplates.GET("/getItemsById/:id", r.handlers.ParamTemplates.GetItemsById)
			paramTemplates.GET("/getItemsbyId/:id", r.handlers.ParamTemplates.GetItemsById)
			paramTemplates.GET("/getById/:id", r.handlers.ParamTemplates.GetById)
			paramTemplates.DELETE("/:id", r.handlers.ParamTemplates.Delete)
		}

		admin.POST("/upload/images", r.handlers.Upload.UploadImages)
	}

	public := v1.Group("/public")
	{
		publicVehicleList := []string{"keyword", "categoryId", "page", "pageSize", "sortField", "sortOrder"}
		public.GET("/home", r.handlers.Public.Home)
		public.GET("/categories", r.handlers.Public.Categories)
		public.GET("/vehicles", middleware.ValidateParams(publicVehicleList), r.handlers.Public.Vehicles)
		public.GET("/vehicles/:id/share-check", r.handlers.Public.ShareCheck)
		public.GET("/vehicles/:id", r.handlers.Public.VehicleDetail)
		public.GET("/contact", r.handlers.Public.Contact)
	}

	r.engine.NoRoute(r.noRouteHandler)
	r.engine.NoMethod(r.noMethodHandler)
	return nil
}

// noRouteHandler 未匹配到任何路由时返回统一 JSON 404。
func (r *Router) noRouteHandler(c *gin.Context) {
	path := c.Request.URL.Path
	response.FailNotFound(c, "路由未找到: "+path)
}

// noMethodHandler 路由存在但 HTTP 方法不匹配时返回 405。
func (r *Router) noMethodHandler(c *gin.Context) {
	method := c.Request.Method
	response.FailMethodNotAllowed(c, "方法不允许: "+method)
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
