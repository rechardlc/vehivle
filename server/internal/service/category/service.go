package category

import (
	"context"
	"errors"
	"vehivle/internal/domain/model"
	"vehivle/internal/domain/enum"
)

type CategoryRepo interface {
	GetById(ctx context.Context, id string) (*model.Category, error)
	List(ctx context.Context) ([]*model.Category, error)
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
func (s *CategoryService) List(ctx context.Context) ([]*model.Category, error) {
	return s.categoryRepo.List(ctx)
}
func (s *CategoryService) Create(ctx context.Context, category *model.Category) (*model.Category, error) {
	return s.categoryRepo.Create(ctx, category)
}
func (s *CategoryService) Update(ctx context.Context, category *model.Category) error {
	return s.categoryRepo.Update(ctx, category)
}
func (s *CategoryService) Delete(ctx context.Context, id string) (*model.Category, error) {
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
	err = s.categoryRepo.Delete(ctx, id)
	if err != nil {
		return nil, err
	}
	return category, nil
}