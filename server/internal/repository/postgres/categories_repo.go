package postgres

import (
	"context"
	"fmt"

	"vehivle/internal/domain/model"

	"gorm.io/gorm"
)

/**
 * 分类仓库
 */
type CategoryRepo struct {
	db *gorm.DB
}
/**
 * 创建分类仓库实例
 */
func NewCategoryRepo(db *gorm.DB) *CategoryRepo {
	return &CategoryRepo{db: db}
}
/**
 * 根据ID获取分类
 */
func (r *CategoryRepo) GetById(ctx context.Context, id string) (*model.Category, error) {
	fmt.Println("id", id)
	var category model.Category
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&category).Error; err != nil {
		return nil, err
	}
	return &category, nil
}
/**
 * 获取分类列表
 */
func (r *CategoryRepo) List(ctx context.Context) ([]*model.Category, error) {
	var categories []*model.Category
	if err := r.db.WithContext(ctx).Order("sort_order DESC, updated_at DESC").Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

/**
 * 创建分类
 */
func (r *CategoryRepo) Create(ctx context.Context, category *model.Category) (*model.Category, error) {
	if err := r.db.WithContext(ctx).Create(category).Error; err != nil {
		return nil, err
	}
	return category, nil
}

/**
 * 更新分类	
 */
func (r *CategoryRepo) Update(ctx context.Context, category *model.Category) error {
	return r.db.WithContext(ctx).Save(category).Error
}
/**
 * 删除分类
 */
func (r *CategoryRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Category{}).Error
}