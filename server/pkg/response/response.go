package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	CodeSuccess          = "000000" // 成功
	CodeAuthFailed       = "A00001" // 认证失败
	CodeAuthDenied       = "A00004" // 授权失败
	CodeParamError       = "B00001" // 参数错误
	CodeBusinessError    = "C00002" // 业务错误
	CodeMediaError       = "M00001" // 媒体错误
	CodeNotFound         = "D00001" // 路由不存在
	CodeMethodNotAllowed = "D00002" // 方法不允许
)

const RequestIDKey = "request_id"

// Body 统一 JSON 响应体
type Body struct {
	Code      string      `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	RequestID string      `json:"requestId"`
	Timestamp string      `json:"timestamp"`
}

// PageResult 分页信息
type PageResult struct {
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

// ListResult 通用分页响应体
type ListResult[T any] struct {
	List []T         `json:"list"`
	Page *PageResult `json:"page"`
}

// getRequestID 从 gin.Context 中提取链路请求 ID
func getRequestID(c *gin.Context) string {
	if v, ok := c.Get(RequestIDKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	if data == nil {
		data = gin.H{}
	}
	c.JSON(http.StatusOK, Body{
		Code:      CodeSuccess,
		Message:   "success",
		Data:      data,
		RequestID: getRequestID(c),
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// Fail 通用失败响应
func Fail(c *gin.Context, code string, message string) {
	c.JSON(http.StatusOK, Body{
		Code:      code,
		Message:   message,
		Data:      nil,
		RequestID: getRequestID(c),
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// FailNotFound 路由不存在（HTTP 404），用于 NoRoute 通配。
func FailNotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, Body{
		Code:      CodeNotFound,
		Message:   message,
		Data:      nil,
		RequestID: getRequestID(c),
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// FailMethodNotAllowed 方法不允许（HTTP 405），用于 NoMethod 通配。
func FailMethodNotAllowed(c *gin.Context, message string) {
	c.JSON(http.StatusMethodNotAllowed, Body{
		Code:      CodeMethodNotAllowed,
		Message:   message,
		Data:      nil,
		RequestID: getRequestID(c),
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// FailParam 参数错误响应
func FailParam(c *gin.Context, message string) {
	Fail(c, CodeParamError, message)
}

// FailAuth 认证失败响应（HTTP 401），对齐 tech.md §3.2 错误码。
// Axios 仅在非 2xx 时触发 error 拦截器，返回 401 是前端无感刷新链路的前提。
func FailAuth(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Body{
		Code:      CodeAuthFailed,
		Message:   message,
		Data:      nil,
		RequestID: getRequestID(c),
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// FailAuthDenied 授权失败响应（HTTP 403），对齐 tech.md §3.2 错误码。
func FailAuthDenied(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Body{
		Code:      CodeAuthDenied,
		Message:   message,
		Data:      nil,
		RequestID: getRequestID(c),
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// FailBusiness 业务错误响应
func FailBusiness(c *gin.Context, message string) {
	if message == "" {
		message = "服务端异常！"
	}
	Fail(c, CodeBusinessError, message)
}

// FailMedia 媒体错误响应
func FailMedia(c *gin.Context, message string) {
	Fail(c, CodeMediaError, message)
}
