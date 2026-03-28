package category

import (
	"context"
	"errors"
	"vehivle/internal/domain/enum"
	"vehivle/internal/domain/model"
)

type CategoryRepo interface {
	GetById(ctx context.Context, id string) (*model.Category, error)
	GetChildListByID(ctx context.Context, id string) ([]*model.Category, error)
	List(ctx context.Context, q model.CategoryListQuery) ([]*model.Category, error)
	Create(ctx context.Context, category *model.Category) (*model.Category, error)
	Update(ctx context.Context, category *model.Category) error
	Delete(ctx context.Context, id string) error
}
type CategoryService struct {
	categoryRepo CategoryRepo
}

func NewCategoryService(categoryRepo CategoryRepo) *CategoryService {
	return &CategoryService{categoryRepo: categoryRepo}
}
func (s *CategoryService) List(ctx context.Context, q model.CategoryListQuery) ([]*model.Category, error) {
	return s.categoryRepo.List(ctx, q)
}

// GetById 按主键查询单条分类（管理端编辑/更新前拉取当前数据）。
func (s *CategoryService) GetById(ctx context.Context, id string) (*model.Category, error) {
	return s.categoryRepo.GetById(ctx, id)
}
func (s *CategoryService) Create(ctx context.Context, category *model.Category) (*model.Category, error) {
	return s.categoryRepo.Create(ctx, category)
}
func (s *CategoryService) Update(ctx context.Context, category *model.Category) error {
	return s.categoryRepo.Update(ctx, category)
}
func (s *CategoryService) Delete(ctx context.Context, id string) (*model.Category, error) {
	// 获取分类
	category, err := s.categoryRepo.GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, errors.New("分类不存在")
	}
	if category.Status == enum.CategoryStatusEnabled {
		return nil, errors.New("分类启用中，不能删除")
	}
	// 删除分类
	childCategoryList, err := s.categoryRepo.GetChildListByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if len(childCategoryList) > 0 {
		return nil, errors.New("分类下有子分类，不能删除")
	}
	err = s.categoryRepo.Delete(ctx, id)
	if err != nil {
		return nil, err
	}
	return category, nil
}
