package postgres

import (
	"context"
	"time"

	"vehivle/internal/domain/enum"
	"vehivle/internal/domain/model"

	"github.com/google/uuid"
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

func (v *VehicleRepo) HasDetailImages(ctx context.Context, id string) (bool, error) {
	var count int64
	if err := v.db.WithContext(ctx).
		Model(&model.VehicleDetailMedia{}).
		Where("vehicle_id = ?", id).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (v *VehicleRepo) RequiredParamsComplete(ctx context.Context, vehicleID string, categoryID string) (bool, error) {
	var templateCount int64
	if err := v.db.WithContext(ctx).
		Table("param_templates").
		Where("category_id = ? AND status = 1", categoryID).
		Count(&templateCount).Error; err != nil {
		return false, err
	}
	if templateCount == 0 {
		return false, nil
	}
	var requiredCount int64
	err := v.db.WithContext(ctx).
		Table("param_template_items AS pti").
		Joins("JOIN param_templates AS pt ON pt.id = pti.template_id").
		Where("pt.category_id = ? AND pt.status = 1 AND pti.required = 1", categoryID).
		Count(&requiredCount).Error
	if err != nil {
		return false, err
	}
	if requiredCount == 0 {
		return true, nil
	}
	var filledCount int64
	err = v.db.WithContext(ctx).
		Table("vehicle_param_values AS vpv").
		Joins("JOIN param_template_items AS pti ON pti.id = vpv.template_item_id").
		Joins("JOIN param_templates AS pt ON pt.id = pti.template_id").
		Where("vpv.vehicle_id = ? AND pt.category_id = ? AND pt.status = 1 AND pti.required = 1 AND btrim(vpv.value_text) <> ''", vehicleID, categoryID).
		Count(&filledCount).Error
	if err != nil {
		return false, err
	}
	return filledCount >= requiredCount, nil
}

func (v *VehicleRepo) ListDetailMediaIDs(ctx context.Context, vehicleID string) ([]string, error) {
	rows, err := v.ListDetailMedia(ctx, vehicleID)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.MediaID)
	}
	return out, nil
}

func (v *VehicleRepo) ListDetailMedia(ctx context.Context, vehicleID string) ([]model.VehicleDetailMedia, error) {
	var rows []model.VehicleDetailMedia
	if err := v.db.WithContext(ctx).
		Model(&model.VehicleDetailMedia{}).
		Where("vehicle_id = ?", vehicleID).
		Order("sort_order DESC, updated_at DESC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (v *VehicleRepo) ReplaceDetailMedia(ctx context.Context, vehicleID string, mediaIDs []string) error {
	return v.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// FE analogy: replace the whole detail image array in one state update.
		// Go detail: the DB transaction keeps delete + reinsert atomic for this request.
		if err := tx.Where("vehicle_id = ?", vehicleID).Delete(&model.VehicleDetailMedia{}).Error; err != nil {
			return err
		}
		now := time.Now()
		for index, mediaID := range mediaIDs {
			row := &model.VehicleDetailMedia{
				ID:        uuid.New().String(),
				VehicleID: vehicleID,
				MediaID:   mediaID,
				SortOrder: len(mediaIDs) - index,
				CreatedAt: now,
				UpdatedAt: now,
			}
			if err := tx.Create(row).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (v *VehicleRepo) ListPublicParams(ctx context.Context, vehicleID string, categoryID string) ([]model.VehicleParamDisplay, error) {
	var rows []model.VehicleParamDisplay
	err := v.db.WithContext(ctx).
		Table("param_template_items AS pti").
		Select("pti.field_key, pti.field_name, pti.field_type, pti.unit, vpv.value_text, pti.sort_order").
		Joins("JOIN param_templates AS pt ON pt.id = pti.template_id").
		Joins("JOIN vehicle_param_values AS vpv ON vpv.template_item_id = pti.id AND vpv.vehicle_id = ?", vehicleID).
		Where("pt.category_id = ? AND pt.status = 1 AND pti.display = 1 AND btrim(vpv.value_text) <> ''", categoryID).
		Order("pti.sort_order DESC, pti.updated_at DESC").
		Scan(&rows).Error
	return rows, err
}
