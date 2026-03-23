package model

import (
	"time"
	"vehivle/internal/domain/enum"
)

type Vehicle struct {
	ID            string             `json:"id"`
	CategoryID    *string            `json:"categoryId,omitempty"`
	Name          string             `json:"name"`
	CoverMediaID  string             `json:"coverMediaId"`
	PriceMode     enum.PriceMode     `json:"priceMode"`
	MSRPPrice     int                `json:"msrpPrice"`
	Status        enum.VehicleStatus `json:"status"`
	SellingPoints string             `json:"sellingPoints"`
	SortOrder     int                `json:"sortOrder"`
	CreatedAt     time.Time          `json:"createdAt"`
	UpdatedAt     time.Time          `json:"updatedAt"`
}

func (v *Vehicle) Publish() {
	v.Status = enum.VehicleStatusPublished
	v.UpdatedAt = time.Now()
}

func (v *Vehicle) Unpublish() {
	v.Status = enum.VehicleStatusUnpublished
	v.UpdatedAt = time.Now()
}
