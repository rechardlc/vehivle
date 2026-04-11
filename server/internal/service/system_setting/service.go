package system_setting

import (
	"context"
	"errors"
	"vehivle/internal/domain/model"
)

type SysRepo interface {
	Detail(ctx context.Context) (*model.SystemSetting, error)
	Create(ctx context.Context, m *model.SystemSetting) error
	Update(ctx context.Context, updates map[string]interface{}) error
	Exists(ctx context.Context) (bool, error)
}

type SysService struct {
	sysRepo SysRepo
}

func NewSysService(sysRepo SysRepo) *SysService {
	return &SysService{sysRepo: sysRepo}
}

func (s *SysService) Detail(ctx context.Context) (*model.SystemSetting, error) {
	return s.sysRepo.Detail(ctx)
}

func (s *SysService) Create(ctx context.Context, m *model.SystemSetting) error {
	exists, err := s.sysRepo.Exists(ctx)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("系统配置已存在，请使用更新接口")
	}
	return s.sysRepo.Create(ctx, m)
}

func (s *SysService) Update(ctx context.Context, updates map[string]interface{}) error {
	exists, err := s.sysRepo.Exists(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("系统配置不存在，请先创建")
	}
	return s.sysRepo.Update(ctx, updates)
}