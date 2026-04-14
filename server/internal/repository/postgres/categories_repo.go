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

// applyListFilters 将 keyword / level / status 筛选条件追加到 tx 上，
// List 和 Count 共用同一份 WHERE 逻辑，确保结果一致。
// 调用方需保证 q.Keyword 已 TrimSpace（handler 层统一处理）。
func applyListFilters(tx *gorm.DB, q model.CategoryListQuery) *gorm.DB {
	if q.Keyword != "" {
		// position(lower(?) in lower(name))：精确子串匹配，避免 LIKE 通配符 %/_ 与用户输入冲突
		tx = tx.Where("position(lower(?) in lower(name)) > 0", q.Keyword)
	}
	if q.Level != nil {
		tx = tx.Where("level = ?", *q.Level)
	}
	if q.Status != nil {
		tx = tx.Where("status = ?", *q.Status)
	}
	return tx
}

// categoryListOrderClause 列表排序：仅支持 createdAt；否则沿用业务默认（sort_order + updated_at）。
// 调用方需保证 SortField / SortOrder 已 TrimSpace + ToLower（handler 层统一处理）。
func categoryListOrderClause(q model.CategoryListQuery) string {
	if q.SortField == "createdAt" {
		if q.SortOrder == "asc" {
			return "created_at ASC"
		}
		return "created_at DESC"
	}
	return "sort_order DESC, updated_at DESC"
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
 * 获取分类列表（可选 keyword / level / status，与 GET 查询参数一致）。
 * 所有参数（page / keyword trim / sort 等）由 handler 层统一处理后传入，此处直接使用。
 */
func (r *CategoryRepo) List(ctx context.Context, q model.CategoryListQuery) ([]*model.Category, error) {
	// 应用筛选条件
	var categories []*model.Category
	tx := applyListFilters(r.db.WithContext(ctx).Model(&model.Category{}), q)
	// 应用分页条件
	if q.PageSize > 0 {
		offset := (q.Page - 1) * q.PageSize
		tx = tx.Offset(offset).Limit(q.PageSize)
	}
	if err := tx.Order(categoryListOrderClause(q)).Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

// Count 获取满足筛选条件的分类总数，用于分页计算。
func (r *CategoryRepo) Count(ctx context.Context, q model.CategoryListQuery) (int64, error) {
	var count int64
	tx := applyListFilters(r.db.WithContext(ctx).Model(&model.Category{}), q)
	if err := tx.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
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
