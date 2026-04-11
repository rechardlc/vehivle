package postgres

import (
	"context"

	"gorm.io/gorm"

	"vehivle/internal/domain/model"
)

type System struct {
	db *gorm.DB
}

func NewSysSettings(db *gorm.DB) *System {
	return &System{db: db}
}

func (s *System) Detail(ctx context.Context) (*model.SystemSetting, error) {
	var row model.SystemSetting
	if err := s.db.WithContext(ctx).First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *System) Create(ctx context.Context, m *model.SystemSetting) error {
	m.ID = 1
	return s.db.WithContext(ctx).Create(m).Error
}

func (s *System) Update(ctx context.Context, updates map[string]interface{}) error {
	return s.db.WithContext(ctx).Model(&model.SystemSetting{}).Where("id = 1").Updates(updates).Error
}

func (s *System) Exists(ctx context.Context) (bool, error) {
	var count int64
	if err := s.db.WithContext(ctx).Model(&model.SystemSetting{}).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
