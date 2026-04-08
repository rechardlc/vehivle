package category

import (
	"context"
	"errors"
	"vehivle/internal/domain/enum"
	"vehivle/internal/domain/model"
	"vehivle/pkg/response"
)

// CategoryRepo 是分类仓库接口
type CategoryRepo interface {
	// GetById 按主键查询单条分类（管理端编辑/更新前拉取当前数据）。
	GetById(ctx context.Context, id string) (*model.Category, error)
	// GetChildListByID 获取分类的子分类列表
	GetChildListByID(ctx context.Context, id string) ([]*model.Category, error)
	// List 返回分类列表
	List(ctx context.Context, q model.CategoryListQuery) ([]*model.Category, error)
	// Create 创建分类
	Create(ctx context.Context, category *model.Category) (*model.Category, error)
	// Update 更新分类
	Update(ctx context.Context, category *model.Category) error
	// Delete 删除分类
	Delete(ctx context.Context, id string) error
	// Count 获取分类数量
	Count(ctx context.Context, q model.CategoryListQuery) (int64, error)
}

// CategoryService 是分类服务
type CategoryService struct {
	// categoryRepo 是分类仓库
	// 这个是依赖接口注入，而不是依赖具体实现
	categoryRepo CategoryRepo
}

func NewCategoryService(categoryRepo CategoryRepo) *CategoryService {
	return &CategoryService{categoryRepo: categoryRepo}
}

// categoryListPage 根据总数与分页参数构造分页元数据。
// page 已由 handler 层归一化为 >= 1，此处不做参数处理。
// PageSize 为 0 表示未分页，页大小在 JSON 中记为 total。
func categoryListPage(total int64, page, pageSize int) response.PageResult {
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

// List 返回 list + page，与 response.ListResult 及前端 data: { list, page } 对齐。
// 所有参数（page 归一化、keyword trim 等）由 handler 层完成，此处直接使用。
func (s *CategoryService) List(ctx context.Context, q model.CategoryListQuery) (response.ListResult[*model.Category], error) {
	// 获取分类数量
	count, err := s.categoryRepo.Count(ctx, q)
	if err != nil {
		return response.ListResult[*model.Category]{}, err
	}
	// 构造分页元数据
	pageMeta := categoryListPage(count, q.Page, q.PageSize)
	// 如果总数为0，则返回空列表
	if count == 0 {
		return response.ListResult[*model.Category]{
			List: []*model.Category{},
			Page: &pageMeta,
		}, nil
	}
	// 获取分类列表
	items, err := s.categoryRepo.List(ctx, q)
	if err != nil {
		return response.ListResult[*model.Category]{}, err
	}
	// 返回分类列表
	return response.ListResult[*model.Category]{
		List: items,
		Page: &pageMeta,
	}, nil
}
func (s *CategoryService) Count(ctx context.Context, q model.CategoryListQuery) (int64, error) {
	return s.categoryRepo.Count(ctx, q)
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
	// 按主键查询单条分类（管理端编辑/更新前拉取当前数据）。
	category, err := s.categoryRepo.GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	// 如果分类不存在，则返回错误
	if category == nil {
		return nil, errors.New("分类不存在")
	}
	// 如果分类启用中，则返回错误
	if category.Status == enum.CategoryStatusEnabled {
		return nil, errors.New("分类启用中，不能删除")
	}
	// 删除分类
	// 获取分类的子分类列表
	childCategoryList, err := s.categoryRepo.GetChildListByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// 如果分类下有子分类，则返回错误
	if len(childCategoryList) > 0 {
		return nil, errors.New("分类下有子分类，不能删除")
	}
	// 删除分类
	err = s.categoryRepo.Delete(ctx, id)
	if err != nil {
		return nil, err
	}
	// 返回分类
	return category, nil
}
