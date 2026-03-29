package helper
import (
	"github.com/gin-gonic/gin"
	"fmt"
	"strings"
)

// 判断参数合法性，fields是合法字段，queryParams是请求参数
// 参数白名单控制
func IsValidFields(ctx *gin.Context, fields []string) (bool, error) {
	// 将fields转换为map
	queryParams := ctx.Request.URL.Query()
	// 节省内存，使用struct{}而不是bool
	fieldsMap := make(map[string]struct{})
	for _, field := range fields {
		fieldsMap[field] = struct{}{}
	}
	for key := range queryParams {
		// 判断key是否在fieldsMap中,不在则返回错误
		if _, ok := fieldsMap[key]; !ok {
			return false, fmt.Errorf("无效的字段: %s，仅支持 %s", key, strings.Join(fields, ", "))
		}
	}
	return true, nil // 所有字段都合法
}