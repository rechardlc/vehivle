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

// applyVehicleListFilters 将 keyword / category_id / status 追加到 tx；ListByQuery 与 CountByQuery 共用。
func applyVehicleListFilters(tx *gorm.DB, q model.VehicleListQuery) *gorm.DB {
	if q.Keyword != "" {
		// position(lower(?) in lower(name)) > 0 表示name包含q.Keyword
		tx = tx.Where("position(lower(?) in lower(name)) > 0", q.Keyword)
	}
	if q.CategoryID != nil {
		// 如果q.CategoryID为空，则查询category_id为空或空串的车型
		if *q.CategoryID == "" {
			tx = tx.Where("(category_id IS NULL OR category_id = '')")
		} else {
			// 如果q.CategoryID不为空，则查询category_id为q.CategoryID的车型
			tx = tx.Where("category_id = ?", *q.CategoryID)
		}
	}
	if q.Status != nil {
		// 如果q.Status不为空，则查询status为q.Status的车型
		tx = tx.Where("status = ?", string(*q.Status))
	}
	return tx
}

// vehicleListOrderClause 管理端列表排序：仅支持 createdAt；否则 sort_order + updated_at。
func vehicleListOrderClause(q model.VehicleListQuery) string {
	if q.SortField == "createdAt" {
		if q.SortOrder == "asc" {
			return "created_at ASC"
		}
		return "created_at DESC"
	}
	return "sort_order DESC, updated_at DESC"
}

// ListByQuery 管理端分页列表；筛选与排序参数由 handler 归一化后传入。
func (v *VehicleRepo) ListByQuery(ctx context.Context, q model.VehicleListQuery) ([]*model.Vehicle, error) {
	var rows []model.Vehicle
	tx := applyVehicleListFilters(v.db.WithContext(ctx).Model(&model.Vehicle{}), q)
	if q.PageSize > 0 {
		offset := (q.Page - 1) * q.PageSize
		tx = tx.Offset(offset).Limit(q.PageSize)
	}
	if err := tx.Order(vehicleListOrderClause(q)).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*model.Vehicle, len(rows))
	for i := range rows {
		out[i] = &rows[i]
	}
	return out, nil
}

// CountByQuery 管理端列表总数。
func (v *VehicleRepo) CountByQuery(ctx context.Context, q model.VehicleListQuery) (int64, error) {
	var count int64
	tx := applyVehicleListFilters(v.db.WithContext(ctx).Model(&model.Vehicle{}), q)
	if err := tx.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
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
