package vehicle

import (
	"context"
	"fmt"
	"strings"
	"vehivle/internal/domain/model"
	"vehivle/internal/domain/rule"
)

// VehicleRepo 定义了车辆仓库接口
type VehicleRepo interface {
	GetById(ctx context.Context, id string) (*model.Vehicle, error)
	Update(ctx context.Context, vehicle *model.Vehicle) error
	List(ctx context.Context, onlyPublished bool) ([]*model.Vehicle, error)
	Create(ctx context.Context, vehicle *model.Vehicle) (*model.Vehicle, error)
}

// Service 定义了车辆服务
type Service struct {
	vehicles VehicleRepo
}

// NewService 创建车辆服务实例
func NewService(vehicles VehicleRepo) *Service {
	return &Service{vehicles: vehicles}
}

// List 返回车型列表；公开接口传 onlyPublished=true，后台管理传 false。
func (s *Service) List(ctx context.Context, onlyPublished bool) ([]*model.Vehicle, error) {
	return s.vehicles.List(ctx, onlyPublished)
}

// Create 创建车型
func (s *Service) Create(ctx context.Context, vehicle *model.Vehicle) (*model.Vehicle, error) {
	return s.vehicles.Create(ctx, vehicle)
}

// Publish 发布车辆
func (s *Service) Publish(ctx context.Context, id string, req rule.PublishRequirements) error {
	v, err := s.vehicles.GetById(ctx, id)
	if err != nil {
		return err
	}
	if ok, errs := rule.CanPublishVehicle(v, &req); !ok {
		return fmt.Errorf("publish vehicle failed: %s", strings.Join(errs, "\n"))
	}
	v.Publish()
	return s.vehicles.Update(ctx, v)
}
