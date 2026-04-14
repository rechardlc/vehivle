package param_template

import (
	"context"
	"errors"
	"fmt"

	"vehivle/internal/domain/model"
	"vehivle/pkg/response"
)

type ParamTemplateRepo interface {
	List(ctx context.Context, page int, pageSize int) ([]model.ParamTemplate, error)
	// 创建模板
	Create(ctx context.Context, p *model.ParamTemplate) (*model.ParamTemplate, error)
	// 更新模板
	Update(ctx context.Context, id string, p *model.ParamTemplate) (*model.ParamTemplate, error)
	// 删除模板
	Delete(ctx context.Context, id string) error
	// 获取模板
	GetById(ctx context.Context, id string) (*model.ParamTemplate, error)
	// 获取模板所有的数据
	GetItemsById(ctx context.Context, id string) (*model.ParamTemplate, error)
	// 获取模板总数
	Count(ctx context.Context) (int64, error)

	ItemsCount(ctx context.Context, id string) (int64, error)
}

// ListQuery 是参数模板列表的服务层 DTO。
// FE analogy: 类似前端 API client 的 query type，只描述列表接口需要的 page/pageSize。
// Go detail: DTO 放在 service 边界，避免 domain model 混入 HTTP 查询语义。
type ListQuery struct {
	Page     int
	PageSize int
}

// ListItem 是参数模板列表响应 DTO，比领域实体多一个 itemNum 统计字段。
type ListItem struct {
	model.ParamTemplate
	ItemNum int `json:"itemNum"`
}

type ParamTemplateService struct {
	paramTemplateRepo ParamTemplateRepo
}

func NewParamTemplateService(paramTemplateRepo ParamTemplateRepo) *ParamTemplateService {
	return &ParamTemplateService{paramTemplateRepo: paramTemplateRepo}
}

func (p *ParamTemplateService) List(ctx context.Context, q ListQuery) (response.ListResult[ListItem], error) {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 10
	}
	c, err := p.paramTemplateRepo.Count(ctx)
	if err != nil {
		return response.ListResult[ListItem]{}, err
	}
	pageMeta := response.PageResult{
		Page:       q.Page,
		PageSize:   q.PageSize,
		Total:      int(c),
		TotalPages: int((c + int64(q.PageSize) - 1) / int64(q.PageSize)),
	}
	if c == 0 {
		return response.ListResult[ListItem]{
			List: []ListItem{},
			Page: &pageMeta,
		}, nil
	}
	rows, err := p.paramTemplateRepo.List(ctx, q.Page, q.PageSize)
	if err != nil {
		return response.ListResult[ListItem]{}, err
	}
	result := make([]ListItem, len(rows))
	for i, row := range rows {
		result[i] = ListItem{ParamTemplate: row}
		if row.ID == nil {
			continue
		}
		ic, err := p.paramTemplateRepo.ItemsCount(ctx, *row.ID)
		if err != nil {
			return response.ListResult[ListItem]{}, err
		}
		result[i].ItemNum = int(ic)
	}
	return response.ListResult[ListItem]{
		List: result,
		Page: &pageMeta,
	}, nil
}

func (p *ParamTemplateService) Create(ctx context.Context, q *model.ParamTemplate) (*model.ParamTemplate, error) {
	result, err := p.paramTemplateRepo.Create(ctx, q)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (p *ParamTemplateService) Update(ctx context.Context, id string, q *model.ParamTemplate) (*model.ParamTemplate, error) {
	if _, err := p.paramTemplateRepo.GetById(ctx, id); err != nil {
		return nil, fmt.Errorf("更新的模板不存在")
	}
	if q == nil {
		return nil, errors.New("参数模板不能为空")
	}
	result, err := p.paramTemplateRepo.Update(ctx, id, q)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (p *ParamTemplateService) GetItemsById(ctx context.Context, id string) (*model.ParamTemplate, error) {
	result, err := p.paramTemplateRepo.GetItemsById(ctx, id)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (p *ParamTemplateService) GetById(ctx context.Context, id string) (*model.ParamTemplate, error) {
	result, err := p.paramTemplateRepo.GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (p *ParamTemplateService) Delete(ctx context.Context, id string) error {
	err := p.paramTemplateRepo.Delete(ctx, id)
	if err != nil {
		return err
	}
	return nil
}
