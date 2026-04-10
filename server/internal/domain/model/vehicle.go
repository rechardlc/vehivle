package model

import (
	"time"

	"vehivle/internal/domain/enum"
)

// VehicleListQuery 管理端 GET 车型列表查询（字段均为可选，未传表示不筛选）。
// Page/PageSize：Page 从 1 起；PageSize 归一化规则由 handler 与分类列表一致。
// SortField/SortOrder：SortField 为 "createdAt" 时按创建时间排；未传 SortField 时保持 sort_order + updated_at 默认序。
// CategoryID：非 nil 且为空串时表示仅「未绑定分类」的车型（category_id IS NULL 或空串）。
type VehicleListQuery struct {
	Keyword    string
	CategoryID *string
	Status     *enum.VehicleStatus
	Page       int
	PageSize   int
	SortField  string
	SortOrder  string
}

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
	// CategoryName 由列表查询后按 category_id 回填，非表字段。
	CategoryName string `json:"categoryName,omitempty" gorm:"-"`
}

// TableName 与 migrations 中 vehicles 表名一致。
func (*Vehicle) TableName() string {
	return "vehicles"
}

/**
 * 发布车型
 * @param v *Vehicle
 */
func (v *Vehicle) Publish() {
	v.Status = enum.VehicleStatusPublished
	v.UpdatedAt = time.Now()
}

/**
 * 下架车型
 * @param v *Vehicle
 */
func (v *Vehicle) Unpublish() {
	v.Status = enum.VehicleStatusUnpublished
	v.UpdatedAt = time.Now()
}
