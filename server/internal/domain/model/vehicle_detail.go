package model

import "time"

// VehicleDetailMedia 对应车型详情图集表。
type VehicleDetailMedia struct {
	ID        string    `json:"id" gorm:"column:id;type:uuid"`
	VehicleID string    `json:"vehicleId" gorm:"column:vehicle_id;type:uuid"`
	MediaID   string    `json:"mediaId" gorm:"column:media_id;type:uuid"`
	SortOrder int       `json:"sortOrder" gorm:"column:sort_order"`
	CreatedAt time.Time `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"column:updated_at"`
}

func (*VehicleDetailMedia) TableName() string {
	return "vehicle_detail_media"
}

// VehicleParamValue 对应车型参数值表。
type VehicleParamValue struct {
	ID             string    `json:"id" gorm:"column:id;type:uuid"`
	VehicleID      string    `json:"vehicleId" gorm:"column:vehicle_id;type:uuid"`
	TemplateItemID string    `json:"templateItemId" gorm:"column:template_item_id;type:uuid"`
	ValueText      string    `json:"valueText" gorm:"column:value_text"`
	CreatedAt      time.Time `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt      time.Time `json:"updatedAt" gorm:"column:updated_at"`
}

func (*VehicleParamValue) TableName() string {
	return "vehicle_param_values"
}

// VehicleParamDisplay 是公开详情页参数表读取模型。
type VehicleParamDisplay struct {
	FieldKey  string  `json:"fieldKey"`
	FieldName string  `json:"fieldName"`
	FieldType string  `json:"fieldType"`
	Unit      *string `json:"unit,omitempty"`
	ValueText string  `json:"valueText"`
	SortOrder int     `json:"sortOrder"`
}
