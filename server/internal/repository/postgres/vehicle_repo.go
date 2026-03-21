package postgres

import (
	"context"
	"vehivle/internal/domain/model"
	"gorm.io/gorm"
)

type VehicleRepo struct {
	db *gorm.DB
}

func NewVehicleRepo(db *gorm.DB) *VehicleRepo {
	return &VehicleRepo{db: db}
}

func (v *VehicleRepo) GetById(ctx context.Context, id string) (*model.Vehicle, error) {
	var row model.Vehicle
	if err := v.db.WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}