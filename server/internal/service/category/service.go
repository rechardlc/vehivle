package category

import (
	"context"
	"errors"
	"vehivle/internal/domain/enum"
	"vehivle/internal/domain/model"
	"vehivle/pkg/response"
)

type CategoryRepo interface {
	GetById(ctx context.Context, id string) (*model.Category, error)
	GetChildListByID(ctx context.Context, id string) ([]*model.Category, error)
	List(ctx context.Context, q model.CategoryListQuery) ([]*model.Category, error)
	Create(ctx context.Context, category *model.Category) (*model.Category, error)
	Update(ctx context.Context, category *model.Category) error
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context, q model.CategoryListQuery) (int64, error)
}
type CategoryService struct {
	categoryRepo CategoryRepo
}

func NewCategoryService(categoryRepo CategoryRepo) *CategoryService {
	return &CategoryService{categoryRepo: categoryRepo}
}

// categoryListPage 根据总数与分页参数构造分页元数据（PageSize 为 0 表示未分页，页大小在 JSON 中记为 total）。
func categoryListPage(total int64, page, pageSize int) response.PageResult {
	normPage := page
	if normPage <= 0 {
		normPage = 1
	}
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
		Page:       normPage,
		PageSize:   sizeInJSON,
		Total:      int(total),
		TotalPages: totalPages,
	}
}

// List 返回 list + page，与 response.ListResult 及前端 data: { list, page } 对齐。
func (s *CategoryService) List(ctx context.Context, q model.CategoryListQuery) (response.ListResult[*model.Category], error) {
	nq := q
	if nq.Page <= 0 {
		nq.Page = 1
	}
	count, err := s.categoryRepo.Count(ctx, nq)
	if err != nil {
		return response.ListResult[*model.Category]{}, err
	}
	pageMeta := categoryListPage(count, nq.Page, nq.PageSize)
	if count == 0 {
		return response.ListResult[*model.Category]{
			List: []*model.Category{},
			Page: &pageMeta,
		}, nil
	}
	items, err := s.categoryRepo.List(ctx, nq)
	if err != nil {
		return response.ListResult[*model.Category]{}, err
	}
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
