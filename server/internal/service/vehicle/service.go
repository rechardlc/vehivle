package vehicle

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"vehivle/internal/domain/enum"
	"vehivle/internal/domain/model"
	"vehivle/internal/domain/rule"
	"vehivle/pkg/response"
)

// VehicleRepo 定义了车辆仓库接口
type VehicleRepo interface {
	GetById(ctx context.Context, id string) (*model.Vehicle, error)
	Update(ctx context.Context, vehicle *model.Vehicle) error
	List(ctx context.Context, onlyPublished bool) ([]*model.Vehicle, error)
	ListByQuery(ctx context.Context, q model.VehicleListQuery) ([]*model.Vehicle, error)
	CountByQuery(ctx context.Context, q model.VehicleListQuery) (int64, error)
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

// GetById 按主键查询车型。
func (s *Service) GetById(ctx context.Context, id string) (*model.Vehicle, error) {
	return s.vehicles.GetById(ctx, id)
}

// List 返回车型列表；公开接口传 onlyPublished=true，后台管理传 false。
func (s *Service) List(ctx context.Context, onlyPublished bool) ([]*model.Vehicle, error) {
	return s.vehicles.List(ctx, onlyPublished)
}

// vehicleListPage 根据总数与分页参数构造分页元数据（与分类列表 service 行为一致）。
func vehicleListPage(total int64, page, pageSize int) response.PageResult {
	var totalPages int
	var sizeInJSON int
	if pageSize > 0 {
		sizeInJSON = pageSize
		if total == 0 {
			totalPages = 0
		} else {
			totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
		}
	} else {
		if total == 0 {
			totalPages = 0
			sizeInJSON = 0
		} else {
			totalPages = 1
			sizeInJSON = int(total)
		}
	}
	return response.PageResult{
		Page:       page,
		PageSize:   sizeInJSON,
		Total:      int(total),
		TotalPages: totalPages,
	}
}

// ListAdmin 管理端分页列表，返回 list + page。
func (s *Service) ListAdmin(ctx context.Context, q model.VehicleListQuery) (response.ListResult[*model.Vehicle], error) {
	count, err := s.vehicles.CountByQuery(ctx, q)
	if err != nil {
		return response.ListResult[*model.Vehicle]{}, err
	}
	pageMeta := vehicleListPage(count, q.Page, q.PageSize)
	if count == 0 {
		return response.ListResult[*model.Vehicle]{
			List: []*model.Vehicle{},
			Page: &pageMeta,
		}, nil
	}
	items, err := s.vehicles.ListByQuery(ctx, q)
	if err != nil {
		return response.ListResult[*model.Vehicle]{}, err
	}
	return response.ListResult[*model.Vehicle]{
		List: items,
		Page: &pageMeta,
	}, nil
}

// Create 创建车型
func (s *Service) Create(ctx context.Context, vehicle *model.Vehicle) (*model.Vehicle, error) {
	return s.vehicles.Create(ctx, vehicle)
}

// VehicleUpdateInput 管理端部分更新：非 nil 字段写入；CategoryID 非 nil 且为空串表示清空分类。
type VehicleUpdateInput struct {
	Name          *string
	CategoryID    *string
	CoverMediaID  *string
	PriceMode     *enum.PriceMode
	MSRPPrice     *int
	SellingPoints *string
	SortOrder     *int
}

func applyVehicleUpdate(v *model.Vehicle, p *VehicleUpdateInput) {
	if p.Name != nil {
		v.Name = *p.Name
	}
	if p.CategoryID != nil {
		if *p.CategoryID == "" {
			v.CategoryID = nil
		} else {
			s := *p.CategoryID
			v.CategoryID = &s
		}
	}
	if p.CoverMediaID != nil {
		v.CoverMediaID = *p.CoverMediaID
	}
	if p.PriceMode != nil {
		v.PriceMode = *p.PriceMode
	}
	if p.MSRPPrice != nil {
		v.MSRPPrice = *p.MSRPPrice
	}
	if p.SellingPoints != nil {
		v.SellingPoints = *p.SellingPoints
	}
	if p.SortOrder != nil {
		v.SortOrder = *p.SortOrder
	}
}

// UpdateVehicle 部分更新车型（非删除状态）。
func (s *Service) UpdateVehicle(ctx context.Context, id string, input *VehicleUpdateInput) (*model.Vehicle, error) {
	v, err := s.vehicles.GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	if v.Status == enum.VehicleStatusDeleted {
		return nil, fmt.Errorf("车型已删除，无法修改")
	}
	applyVehicleUpdate(v, input)
	v.UpdatedAt = time.Now()
	if err := s.vehicles.Update(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}

// SoftDelete 将车型标记为 deleted。
func (s *Service) SoftDelete(ctx context.Context, id string) error {
	v, err := s.vehicles.GetById(ctx, id)
	if err != nil {
		return err
	}
	if v.Status == enum.VehicleStatusDeleted {
		return fmt.Errorf("车型已删除")
	}
	v.Status = enum.VehicleStatusDeleted
	v.UpdatedAt = time.Now()
	return s.vehicles.Update(ctx, v)
}

// Publish 发布车辆
/**
 * 发布车辆
 * @param ctx context.Context
 * @param id string
 * @param req rule.PublishRequirements
 * @return error
 */
func (s *Service) Publish(ctx context.Context, id string, req rule.PublishRequirements) error {
	// 根据id查询车型
	v, err := s.vehicles.GetById(ctx, id)
	if err != nil {
		return err
	}
	// 判断是否可以发布
	if ok, errs := rule.CanPublishVehicle(v, &req); !ok {
		return fmt.Errorf("%s", strings.Join(errs, "; "))
	}
	// 发布车型
	v.Publish()
	return s.vehicles.Update(ctx, v)
}

// Unpublish 下架车辆（仅已发布可下架）。
func (s *Service) Unpublish(ctx context.Context, id string) error {
	v, err := s.vehicles.GetById(ctx, id)
	if err != nil {
		return err
	}
	if v.Status != enum.VehicleStatusPublished {
		return fmt.Errorf("仅已发布车型可下架")
	}
	v.Unpublish()
	return s.vehicles.Update(ctx, v)
}

func cloneStrPtr(p *string) *string {
	if p == nil {
		return nil
	}
	s := *p
	return &s
}

// Duplicate 复制为新的草稿车型。
func (s *Service) Duplicate(ctx context.Context, id string) (*model.Vehicle, error) {
	src, err := s.vehicles.GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	if src.Status == enum.VehicleStatusDeleted {
		return nil, fmt.Errorf("已删除车型不可复制")
	}
	now := time.Now()
	dup := &model.Vehicle{
		ID:            uuid.New().String(),
		CategoryID:    cloneStrPtr(src.CategoryID),
		Name:          src.Name + " (副本)",
		CoverMediaID:  src.CoverMediaID,
		PriceMode:     src.PriceMode,
		MSRPPrice:     src.MSRPPrice,
		Status:        enum.VehicleStatusDraft,
		SellingPoints: src.SellingPoints,
		SortOrder:     src.SortOrder,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	return s.vehicles.Create(ctx, dup)
}

// BatchSetStatus 批量设置状态（跳过已删除）。
func (s *Service) BatchSetStatus(ctx context.Context, ids []string, status enum.VehicleStatus) error {
	if len(ids) == 0 {
		return fmt.Errorf("ids 不能为空")
	}
	for _, id := range ids {
		v, err := s.vehicles.GetById(ctx, id)
		if err != nil {
			return fmt.Errorf("车型 %s: %w", id, err)
		}
		if v.Status == enum.VehicleStatusDeleted {
			continue
		}
		v.Status = status
		v.UpdatedAt = time.Now()
		if err := s.vehicles.Update(ctx, v); err != nil {
			return err
		}
	}
	return nil
}
