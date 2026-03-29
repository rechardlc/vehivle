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
	allowMaps := make(map[string]struct{})
	for _, field := range allowFields {
		allowMaps[field] = struct{}{}
	}
	return func(c *gin.Context) {
		queryParams := c.Request.URL.Query()
		for key := range queryParams {
			if _, ok := allowMaps[key]; !ok {
				response.FailParam(c, fmt.Sprintf("无效的字段: %s，仅支持 %s", key, strings.Join(allowFields, ", ")))
				// 中断请求
				c.Abort()
				return
			}
		}
		c.Next()
	}
}