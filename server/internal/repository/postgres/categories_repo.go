package postgres

import (
	"context"

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
	var category model.Category
	// withContext: 使用上下文管理数据库连接
	// where: 查询条件
	// First: 获取第一条记录，find：获取所有记录， take：获取指定数量的记录
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&category).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

/**
 * 根据ID获取子类的父级分类 list
 */
func (r *CategoryRepo) GetChildListByID(ctx context.Context, id string) ([]*model.Category, error) {
	var categories []*model.Category
	if err := r.db.WithContext(ctx).Where("parent_id = ?", id).Order("sort_order DESC, updated_at DESC").Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
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
