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

const RequestIDKey = "request_id" // 请求ID键名

// 响应体（字段需导出才能被 encoding/json 序列化）
// 定义结构体时，字段名必须大写才能被导出
// 字段名小写则无法被导出，无法被 encoding/json 序列化
type Body struct {
	Code      string      `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	RequestID string      `json:"requestId"`
	Timestamp string      `json:"timestamp"`
}

// 分页响应体
type PageResult struct {
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}


// 通用分页响应体
type ListResult[T any] struct {
	List []T `json:"list"`
	Page *PageResult `json:"page"`
}

// 获取请求ID
func getRequestID(c *gin.Context) string {
	if v, ok := c.Get(RequestIDKey); ok {
		// v.(string) 类型断言写法（避免写成 v == nil 判断）
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// 成功响应
// interface{} 表示任意类型（类似 JS 的 any）
func Success(c *gin.Context, data interface{}) {
	// 如果 data 为 nil，则返回空对象
	if data == nil {
		data = gin.H{} // make(map[string]interface{})
	}
	c.JSON(http.StatusOK, Body{
		Code:      CodeSuccess,
		Message:   "success",
		Data:      data,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// 失败响应
func Fail(c *gin.Context, code string, message string) {
	c.JSON(http.StatusOK, Body{
		Code:      code,
		Message:   message,
		Data:      nil,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// FailNotFound 路由不存在（HTTP 404），用于 NoRoute 通配。
func FailNotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, Body{
		Code:      CodeNotFound,
		Message:   message,
		Data:      nil,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// FailMethodNotAllowed 方法不允许（HTTP 405），用于 NoMethod 通配。
func FailMethodNotAllowed(c *gin.Context, message string) {
	c.JSON(http.StatusMethodNotAllowed, Body{
		Code:      CodeMethodNotAllowed,
		Message:   message,
		Data:      nil,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// 参数错误响应
func FailParam(c *gin.Context, message string) {
	Fail(c, CodeParamError, message)
}

// 认证失败响应
func FailAuth(c *gin.Context, message string) {
	Fail(c, CodeAuthFailed, message)
}

// 授权失败响应
func FailAuthDenied(c *gin.Context, message string) {
	Fail(c, CodeAuthDenied, message)
}

// 业务错误响应
func FailBusiness(c *gin.Context, message string) {
	if message == "" {
		message = "服务端异常！"
	}
	Fail(c, CodeBusinessError, message)
}

// 媒体错误响应
func FailMedia(c *gin.Context, message string) {
	Fail(c, CodeMediaError, message)
}
