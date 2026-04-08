package middleware

import (
	"fmt"
	"strings"
	"vehivle/pkg/response"

	"github.com/gin-gonic/gin"
)
/**
 * 参数白名单控制中间件
 * @param allowFields 允许的字段列表
 * @return gin.HandlerFunc
 */
func ValidateParams(allowFields []string) gin.HandlerFunc {
	// 创建一个map，用于存储允许的字段
	// 使用make函数来创建一个map，map的key是string，value是struct{}
	// struct{}是go语言中的一种空结构体，用于占位，不占用内存
	allowMaps := make(map[string]struct{})
	// 遍历allowFields，将allowFields中的每个字段添加到allowMaps中
	for _, field := range allowFields {
		allowMaps[field] = struct{}{}
	}
	// 返回一个gin.HandlerFunc类型的函数，用于验证参数
	return func(c *gin.Context) {
		queryParams := c.Request.URL.Query()
		// 遍历queryParams，将queryParams中的每个字段添加到allowMaps中
		for key := range queryParams {
			// 如果allowMaps中不包含key，则返回错误
			if _, ok := allowMaps[key]; !ok {
				response.FailParam(c, fmt.Sprintf("无效的字段: %s，仅支持 %s", key, strings.Join(allowFields, ", ")))
				// 中断请求，不再执行后续的中间件和路由处理
				c.Abort()
				return
			}	
		}
		// 继续执行后续的中间件和路由处理
		c.Next()
	}
}