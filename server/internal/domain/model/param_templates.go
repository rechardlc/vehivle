package model

import (
	"time"

	"vehivle/internal/domain/enum"
)

// ParamTemplate 对应表 param_templates（见 migrations/000005）。
// status：SMALLINT 0/1，与 categories.status 语义一致。
type ParamTemplate struct {
	// ID 可选：创建时可不传，由数据库 DEFAULT gen_random_uuid() 生成；查询/更新后非空。
	ID *string `json:"id,omitempty" gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	// 模板名称：必填项，不能为空字符串
	Name string `json:"name" binding:"required" gorm:"column:name;not null"`
	// 绑定的一级分类
	CategoryID string `json:"categoryId" binding:"required" gorm:"column:category_id;not null;type:uuid"`
	// 参数项：一对多关系，由 handler/service 组装（handler 传入，service 关联创建）。_ 表示不映射到数据库。
	Items []ParamTemplateItem `json:"items" gorm:"foreignKey:TemplateID;constraint:OnDelete:CASCADE"`
	// 状态
	Status enum.ParamTemplateStatus `json:"status" binding:"required" gorm:"column:status;not null"`
	// 创建时间
	CreatedAt time.Time `json:"createdAt" gorm:"column:created_at;autoCreateTime"`
	// 更新时间
	UpdatedAt time.Time `json:"updatedAt" gorm:"column:updated_at;autoUpdateTime"`
}

func (*ParamTemplate) TableName() string {
	return "param_templates"
}

// ParamBody 模板更新（模板 ID 由路由 :id 提供）。
// 若管理端 JSON 使用 enabled/disabled 字符串，在 handler/service 与 ParamTemplateStatus 互转。
type ParamBody struct {
	Name       *string                   `json:"name,omitempty"`
	CategoryID *string                   `json:"categoryId,omitempty"`
	Status     *enum.ParamTemplateStatus `json:"status,omitempty"`
}

// ParamTemplateItem 对应表 param_template_items。
// field_type 约束：text | number | single_select（与 DB CHECK 一致；API 层勿写入 single）。
// required/display：SMALLINT 0/1。
type ParamTemplateItem struct {
	ID         *string   `json:"id,omitempty" gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	TemplateID string    `json:"templateId" binding:"required" gorm:"column:template_id;not null;type:uuid"`
	FieldKey   string    `json:"fieldKey" binding:"required" gorm:"column:field_key;not null"`
	FieldName  string    `json:"fieldName" binding:"required" gorm:"column:field_name;not null"`
	FieldType  string    `json:"fieldType" binding:"required,oneof=text number single_select" gorm:"column:field_type;not null"`
	Unit       *string   `json:"unit,omitempty" gorm:"column:unit"`
	Required   *int8      `json:"required" binding:"required,oneof=0 1" gorm:"column:required;not null"`
	Display    *int8      `json:"display" binding:"required,oneof=0 1" gorm:"column:display;not null"`
	SortOrder  *int       `json:"sortOrder" binding:"required,min=0" gorm:"column:sort_order;not null"`
	CreatedAt  time.Time `json:"createdAt" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt  time.Time `json:"updatedAt" gorm:"column:updated_at;autoUpdateTime"`
}

func (*ParamTemplateItem) TableName() string {
	return "param_template_items"
}
