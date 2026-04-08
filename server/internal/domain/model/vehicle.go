package model

import (
	"time"
	"vehivle/internal/domain/enum"
)

type Vehicle struct {
	ID            string             `json:"id" gorm:"column:id;type:uuid"`
	CategoryID    *string            `json:"categoryId,omitempty" gorm:"column:category_id"`
	Name          string             `json:"name" gorm:"column:name"`
	CoverMediaID  string             `json:"coverMediaId" gorm:"column:cover_media_id"`
	PriceMode     enum.PriceMode     `json:"priceMode" gorm:"column:price_mode"`
	MSRPPrice     int                `json:"msrpPrice" gorm:"column:msrp_price"`
	Status        enum.VehicleStatus `json:"status" gorm:"column:status"`
	SellingPoints string             `json:"sellingPoints" gorm:"column:selling_points"`
	SortOrder     int                `json:"sortOrder" gorm:"column:sort_order"`
	CreatedAt     time.Time          `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt     time.Time          `json:"updatedAt" gorm:"column:updated_at"`
	// CoverImageURL 由查询阶段根据 media_assets.storage_key 拼接，非表字段。
	CoverImageURL string `json:"coverImageUrl,omitempty" gorm:"-"`
}

// TableName 与 migrations 中 vehicles 表名一致。
func (*Vehicle) TableName() string {
	return "vehicles"
}

func (v *Vehicle) Publish() {
	v.Status = enum.VehicleStatusPublished
	v.UpdatedAt = time.Now()
}

func (v *Vehicle) Unpublish() {
	v.Status = enum.VehicleStatusUnpublished
	v.UpdatedAt = time.Now()
}
