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

/**
 * 判断排序字段和排序顺序是否合法
 * @param sortField 排序字段
 * @param sortOrder 排序顺序
 * @return bool 是否合法
 * @return error 错误信息
 */
func IsValidSortFieldAndOrder(sortField string, sortOrder string) (bool, error) {
	// 如果sortField不为空，并且sortField不等于createdAt，则返回错误
	if sortField != "" && sortField != "createdAt" {
		return false, fmt.Errorf("无效的 sortField，仅支持 createdAt")
	}
	// 如果sortOrder不为空，并且sortOrder不等于asc或desc，则返回错误
	if sortOrder != "" && sortOrder != "asc" && sortOrder != "desc" {
		return false, fmt.Errorf("无效的 sortOrder，仅支持 asc 或 desc")
	}
	return true, nil
}


/**
校验字段字符串必填项
*/

func RequiredField[T ~string](field T) error {
	if field == "" {
		return  fmt.Errorf("字段%s不能为空！")
	}
	return nil
}