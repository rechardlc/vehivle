package logger

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// 在Gin中使用的请求ID的Key
const (
	GinRequestIDKey = "request_id"
)

// gin.HandlerFunc是Gin框架的中间件类型，用于构建gin中间件，用于生成请求ID
// RequestID是生成请求ID的中间件
// 返回一个gin.HandlerFunc类型的匿名函数，匿名函数中包含gin.Context类型的参数c
func RequestID() gin.HandlerFunc {
	// 返回一个gin.HandlerFunc类型的匿名函数
	return func(c *gin.Context) {
		// 从请求头中获取请求ID
		requestID := c.GetHeader("X-Request-ID")
		// 如果请求ID为空，则生成一个
		if requestID == "" {
			// uuid.New()生成一个唯一的UUID
			requestID = uuid.New().String()
		}
		// 将请求ID设置到gin.Context上下文中
		c.Set(GinRequestIDKey, requestID)
		// 将请求ID设置到响应头中
		c.Header("X-Request-ID", requestID)
		// 继续执行下一个中间件
		c.Next()
	}
}

// 访问日志
func AccessLog(logger Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取开始时间
		start := time.Now()
		// 获取请求路径
		path := c.Request.URL.Path
		// 获取请求方法
		method := c.Request.Method
		// 获取客户端IP
		clientIP := c.ClientIP()
		// 获取请求ID
		requestID := ""
		// 获取请求ID
		if v, ok := c.Get(GinRequestIDKey); ok {
			if s, ok := v.(string); ok {
				requestID = s
			}
		}
		// 将请求ID设置到标准库context.Context上下文中
		ctx := WithRequestID(c.Request.Context(), requestID)
		// Middleware是洋葱模型：将c.Next()放在中间，可以实现请求前和请求后的操作
		c.Next()
		// 获取延迟时间
		latency := time.Since(start)
		// 获取响应状态码
		status := c.Writer.Status()
		// 记录日志
		ginErr := c.Errors.ByType(gin.ErrorTypePublic).String()
		if ginErr != "" {
			logger.Error(ctx, "request", slog.String("error", ginErr))
		}
		logger.Info(
			ctx,
			"request",
			slog.String("path", path),
			slog.String("method", method),
			slog.String("client_ip", clientIP),
			slog.Int("status", status),
			slog.Duration("latency_ms", latency),
		)
	}
}
