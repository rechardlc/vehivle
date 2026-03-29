package model

import (
	"time"

	"vehivle/internal/domain/enum"
)

// CategoryListQuery 管理端 GET 列表查询（字段均为可选，未传表示不筛选）。
// Page/PageSize：Page 从 1 起；PageSize 为 0 表示不分页（返回当前筛选下的全部行）。
// SortField/SortOrder：SortField 为 "createdAt" 时按创建时间排；SortOrder 为 "asc"|"desc"（忽略大小写）；未传 SortField 时保持 sort_order + updated_at 默认序。
type CategoryListQuery struct {
	Keyword   string
	Level     *int
	Status    *enum.CategoryStatus
	Page      int
	PageSize  int
	SortField string
	SortOrder string
}

// CategoryCreateInput 管理端创建分类时可写字段（绑定 JSON 时用此类型，避免接受伪造 id/时间戳）。
// FE 类比：类似 TS 里 Pick<Category, 'name' | 'parentId' | ...>，只暴露可编辑键。
type CategoryCreateInput struct {
	Name      string              `json:"name"`
	ParentID  *string             `json:"parentId"`
	Level     int                 `json:"level"`
	Status    enum.CategoryStatus `json:"status"`
	SortOrder int                 `json:"sortOrder"`
}

// CategoryUpdateBody PATCH 部分更新：仅非 nil 指针字段表示本次提交（omit 的键不覆盖库内原值）。
type CategoryUpdateBody struct {
	Name      *string              `json:"name,omitempty"`
	ParentID  *string              `json:"parentId,omitempty"`
	Level     *int                 `json:"level,omitempty"`
	Status    *enum.CategoryStatus `json:"status,omitempty"`
	SortOrder *int                 `json:"sortOrder,omitempty"`
}

// Category 领域实体；匿名嵌入 CategoryCreateInput，复用同一套字段与 json tag，避免与 Input 重复声明。
// FE 类比：类似 class Category extends CategoryBase { id; createdAt; ... }，序列化时子类字段与基类展平到同一 JSON 对象。
// Go detail: encoding/json 会把嵌入结构体的字段提升到外层；GORM 同样映射嵌入字段到表列。
// 持久化：对应表 categories（见 TableName）；时间戳列显式 gorm column，与 migrations 中列名一致，避免仅依赖默认 NamingStrategy。
type Category struct {
	// ID 主键 UUID，对应 categories.id。
	ID string `json:"id"`
	CategoryCreateInput
	// IconMediaID 分类图标媒体 ID，可空，对应 categories.icon_media_id。
	IconMediaID *string `json:"iconMediaId,omitempty"`
	// CreatedAt 创建时间，对应 categories.created_at。
	CreatedAt time.Time `json:"createdAt" gorm:"column:created_at"`
	// UpdatedAt 最近更新时间，对应 categories.updated_at。
	UpdatedAt time.Time `json:"updatedAt" gorm:"column:updated_at"`
	// ParentName 父级名称，仅列表/展示用，非数据库列（gorm:"-" 不参与读写）。
	ParentName string `json:"parentName,omitempty" gorm:"-"`
}

// TableName 显式指定 GORM 表名，与 migrations 中 CREATE TABLE categories 一致。
// 如果不显示指定，GORM 会默认使用结构体名称的复数形式作为表名，这是一种约定俗成的做法，避免歧义。
func (*Category) TableName() string {
	return "categories"
}
