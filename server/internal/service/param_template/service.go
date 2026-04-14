package param_template

import (
	"context"
	"fmt"
	"math"
	"sync"
	"vehivle/internal/domain/model"
	"vehivle/pkg/response"
)

type ParamTemplateRepo interface {
	List(ctx context.Context, q *model.TmpQuery) (*[]model.ParamTemplate, error)
	// 创建模板
	Create(ctx context.Context, p *model.ParamTemplate) (*model.ParamTemplate, error)
	// 更新模板
	Update(ctx context.Context, p *model.ParamTemplate) (*model.ParamTemplate, error)
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

type ParamTemplateService struct {
	paramTemplateRepo ParamTemplateRepo
}

func NewParamTemplateService(paramTemplateRepo ParamTemplateRepo) *ParamTemplateService {
	return &ParamTemplateService{paramTemplateRepo: paramTemplateRepo}
}

func (p *ParamTemplateService) List(ctx context.Context, q *model.TmpQuery) (response.ListResult[model.PamListTmp], error) {
	c, err := p.paramTemplateRepo.Count(ctx)
	if err != nil {
		return response.ListResult[model.PamListTmp]{}, err
	}
	if c == 0 {
		return response.ListResult[model.PamListTmp]{}, nil
	}
	rows, err := p.paramTemplateRepo.List(ctx, q)
	if err != nil {
		return response.ListResult[model.PamListTmp]{}, err
	}
	result := make([]model.PamListTmp, len(*rows))
	var wg sync.WaitGroup
	for i, row := range *rows {
		result[i] = model.PamListTmp{ParamTemplate: row}
		if row.ID == nil {
			continue
		}

		wg.Add(1)
		// FE analogy: 类似 Promise.all 并发拉取每条模板的 item 统计。
		// Go detail: goroutine + WaitGroup 显式等待，并把 index/id 作为参数传入避免闭包捕获循环变量。
		go func(index int, id string) {
			defer wg.Done()
			ic, _ := p.paramTemplateRepo.ItemsCount(ctx, id)
			result[index].ItemNum = int(ic)
		}(i, *row.ID)
	}
	wg.Wait()
	return response.ListResult[model.PamListTmp]{
		List: result,
		Page: &response.PageResult{
			Page:       q.Page,
			PageSize:   q.PageSize,
			TotalPages: int(math.Ceil(float64(c) / float64(q.PageSize))),
			Total:      int(c),
		},
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
	result, err := p.paramTemplateRepo.Update(ctx, q)
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
