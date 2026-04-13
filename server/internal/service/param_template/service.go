package param_template

import (
	"fmt"
	"vehivle/internal/domain/model"

	"context"
)

type ParamTemplateRepo interface {
	// 创建模板
	Create(ctx context.Context, p *model.ParamTemplate) (*model.ParamTemplate, error)
	// 更新模板
	Update(ctx context.Context, p *model.ParamTemplate) (*model.ParamTemplate, error)
	// 删除模板
	Delete(ctx context.Context, id string) error
	// 获取模板
	Detail(ctx context.Context, id string) (*model.ParamTemplate, error)
	// 获取模板所有的数据
	AllDetail(ctx context.Context, id string) (*model.ParamTemplate, error)
	// 获取模板总数
	Count(ctx context.Context) (int64, error)
}

type ParamTemplateService struct {
	paramTemplateRepo ParamTemplateRepo
}

func NewParamTemplateService(paramTemplateRepo ParamTemplateRepo) *ParamTemplateService {
	return &ParamTemplateService{paramTemplateRepo: paramTemplateRepo}
}

func (p *ParamTemplateService) Create(ctx context.Context, q *model.ParamTemplate) (*model.ParamTemplate, error) {
	result, err := p.paramTemplateRepo.Create(ctx, q)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (p *ParamTemplateService) Update(ctx context.Context, id string, q *model.ParamTemplate) (*model.ParamTemplate, error) {
	if _, err := p.paramTemplateRepo.Detail(ctx, id); err != nil {
		return nil, fmt.Errorf("更新的模板不存在")
	}
	result, err := p.paramTemplateRepo.Update(ctx, q)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (p *ParamTemplateService) AllDetail(ctx context.Context, id string) (*model.ParamTemplate, error) {
	result, err := p.paramTemplateRepo.AllDetail(ctx, id)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (p *ParamTemplateService) Detail(ctx context.Context, id string) (*model.ParamTemplate, error) {
	result, err := p.paramTemplateRepo.Detail(ctx, id)
	if err != nil {
		return nil, err
	}
	return result, nil
}
