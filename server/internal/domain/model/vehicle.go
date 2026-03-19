package model

import (
	"time"
	"vehivle/internal/domain/enum"
)

type Vehicle struct {
	ID string `json:"id"`
	CategoryID string `json:"category_id"`
	Name string `json:"name"`
	CoverMediaID string `json:"cover_media_id"`
	PriceMode enum.PriceMode `json:"price_mode"`
	MSRPPrice int `json:"msrp_price"`
	Status enum.VehicleStatus `json:"status"`
	SellingPoints string `json:"selling_points"`
	SortOrder int `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (v *Vehicle) Publish() {
	v.Status = enum.VehicleStatusPublished
	v.UpdatedAt = time.Now()
}

func (v *Vehicle) Unpublish() {
	v.Status = enum.VehicleStatusUnpublished
	v.UpdatedAt = time.Now()
}
