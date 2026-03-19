package router

import (
	"github.com/gin-gonic/gin"
	"log/slog"
	"vehivle/pkg/logger"
	"vehivle/pkg/response"
	"vehivle/internal/transport/http/handler"
)

// Router 路由注册器，持有 Gin 引擎并负责 API 路由分组与挂载。
type Router struct {
	engine *gin.Engine
	logger logger.Logger
}

// New 创建 Router 实例，engine 为已完成中间件配置的 Gin 引擎，logger 用于请求日志。
func New(engine *gin.Engine, logger logger.Logger) *Router {
	return &Router{engine: engine, logger: logger}
}

// Register 注册 API 路由分组（admin、public）、健康检查等路由。
func (r *Router) Register() error {
	// 健康检查：GET /health，用于探活、负载均衡健康检查等。
	r.engine.GET("/health", r.healthHandler)

	v1 := r.engine.Group("/api/v1")
	admin := v1.Group("/admin")
	{
		user := admin.Group("/user")
		{
			user.GET("/:user_id", handler.UserHandler)
			user.POST("/", handler.UserCreateHandler)
			user.PUT("/:user_id", handler.UserUpdateHandler)
			user.DELETE("/:user_id", handler.UserDeleteHandler)
		}
		vehicles := admin.Group("/vehicles")
		{
			vehicles.GET("/", handler.VehiclesListHandler)
			vehicles.POST("/", handler.VehiclesCreateHandler)
			vehicles.PUT("/:vehicle_id", handler.VehiclesUpdateHandler)
			vehicles.DELETE("/:vehicle_id", handler.VehiclesDeleteHandler)
		}
	}
	public := v1.Group("/public")
	{
		public.GET("/vehicles", handler.VehiclesListHandler)
	}

	// 通配：未注册路由返回统一 404（类似前端 catch-all）
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
		"status":     "ok",
		"message":    "API is running",
		"request_id": rid,
	})
}