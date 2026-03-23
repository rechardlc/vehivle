package postgres

import (
	"context"

	"vehivle/internal/domain/enum"
	"vehivle/internal/domain/model"

	"gorm.io/gorm"
)

// VehicleRepo 车辆仓库
type VehicleRepo struct {
	db *gorm.DB
}

// NewVehicleRepo 创建车辆仓库实例
func NewVehicleRepo(db *gorm.DB) *VehicleRepo {
	return &VehicleRepo{db: db}
}

// GetById 根据ID获取车型
func (v *VehicleRepo) GetById(ctx context.Context, id string) (*model.Vehicle, error) {
	var row model.Vehicle
	if err := v.db.WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

// Update 持久化车型变更（全量保存）。
func (v *VehicleRepo) Update(ctx context.Context, vehicle *model.Vehicle) error {
	return v.db.WithContext(ctx).Save(vehicle).Error
}

// List 查询车型列表；onlyPublished 为 true 时仅返回已上架（小程序公开口径）。
func (v *VehicleRepo) List(ctx context.Context, onlyPublished bool) ([]*model.Vehicle, error) {
	q := v.db.WithContext(ctx).Model(&model.Vehicle{}).Order("sort_order DESC, updated_at DESC")
	if onlyPublished {
		q = q.Where("status = ?", string(enum.VehicleStatusPublished))
	}
	var rows []model.Vehicle
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*model.Vehicle, len(rows))
	for i := range rows {
		out[i] = &rows[i]
	}
	return out, nil
}

// Create 创建车型
func (v *VehicleRepo) Create(ctx context.Context, vehicle *model.Vehicle) (*model.Vehicle, error) {
	// 创建车型
	if err := v.db.WithContext(ctx).Create(vehicle).Error; err != nil {
		return nil, err
	}
	// 返回车型
	return vehicle, nil
}
