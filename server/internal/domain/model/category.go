package model

import (
	"time"

	"vehivle/internal/domain/enum"
)

// CategoryCreateInput 管理端创建分类时可写字段（绑定 JSON 时用此类型，避免接受伪造 id/时间戳）。
// FE 类比：类似 TS 里 Pick<Category, 'name' | 'parentId' | ...>，只暴露可编辑键。
type CategoryCreateInput struct {
	Name      string              `json:"name"`
	ParentID  *string             `json:"parentId"`
	Level     int                 `json:"level"`
	Status    enum.CategoryStatus `json:"status"`
	SortOrder int                 `json:"sortOrder"`
}

// Category 领域实体；匿名嵌入 CategoryCreateInput，复用同一套字段与 json tag，避免与 Input 重复声明。
// FE 类比：类似 class Category extends CategoryBase { id; createdAt; ... }，序列化时子类字段与基类展平到同一 JSON 对象。
// Go detail: encoding/json 会把嵌入结构体的字段提升到外层；GORM 同样映射嵌入字段到表列。
type Category struct {
	ID string `json:"id"`
	CategoryCreateInput
	IconMediaID *string   `json:"iconMediaId,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	ParentName  string    `json:"parentName,omitempty" gorm:"-"`
}
