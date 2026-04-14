package postgres

import (
	"vehivle/internal/domain/model"

	"gorm.io/gorm"

	"context"
)

type ParamTemplateRepo struct {
	db *gorm.DB
}

func NewParamTemplateRepo(db *gorm.DB) *ParamTemplateRepo {
	return &ParamTemplateRepo{db: db}
}

// 查询模板列表（分页）
func (r *ParamTemplateRepo) List(ctx context.Context, q *model.TmpQuery) (*[]model.ParamTemplate, error) {
	var rows []model.ParamTemplate
	offset := (q.Page - 1) * q.PageSize
	err := r.db.WithContext(ctx).
		Model(&model.ParamTemplate{}).
		Order("created_at DESC").
		Offset(offset).
		Limit(q.PageSize).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return &rows, nil
}

// 创建模板
func (r *ParamTemplateRepo) Create(ctx context.Context, p *model.ParamTemplate) (*model.ParamTemplate, error) {
	if err := r.db.WithContext(ctx).Create(p).Error; err != nil {
		return nil, err
	}
	return p, nil
}

// 更新模板
func (r *ParamTemplateRepo) Update(ctx context.Context, p *model.ParamTemplate) (*model.ParamTemplate, error) {
	return p, r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 除去Items字段，更新所有的字段
		if err := tx.Model(p).Omit("Items").Updates(p).Error; err != nil {
			return err
		}
		nItemIds := make([]string, 0)
		for i := range p.Items {
			// 有id的情况，通过save直接更新，没有id，会创建id并且插入，然后回填id
			p.Items[i].TemplateID = *p.ID
			if err := tx.Save(&p.Items[i]).Error; err != nil {
				return err
			}
			// 将新的id放入newItems中，用于后续删除
			if p.Items[i].ID != nil {
				nItemIds = append(nItemIds, *p.Items[i].ID)
			}
		}
		// 这个是在构造查询器
		delQuery := tx.Where("template_id=?", *p.ID)
		if len(nItemIds) > 0 {
			delQuery = delQuery.Where("id not in ?", nItemIds)
		}
		if err := delQuery.Delete(&model.ParamTemplateItem{}).Error; err != nil {
			return err
		}
		return nil
	})
}

// 获取模板总数
func (r *ParamTemplateRepo) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.ParamTemplate{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// 只查询模板本身的数据
func (r *ParamTemplateRepo) GetById(ctx context.Context, id string) (*model.ParamTemplate, error) {
	var row model.ParamTemplate
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

// 查询完整的模板（模板本身的数据+items的数据）

func (r *ParamTemplateRepo) GetItemsById(ctx context.Context, id string) (*model.ParamTemplate, error) {
	var row model.ParamTemplate
	// Preload： 会连带查询关联的子表的数据，字段为必须ParamTemplate的结构体的名称
	err := r.db.WithContext(ctx).Preload("Items").Where("id=?", id).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// 创建模板项
func (r *ParamTemplateRepo) CreateItem(ctx context.Context, i *model.ParamTemplateItem) (*model.ParamTemplateItem, error) {
	if err := r.db.WithContext(ctx).Create(i).Error; err != nil {
		return nil, err
	}
	return i, nil
}

// 删除模板
func (r *ParamTemplateRepo) Delete(ctx context.Context, id string) error {
	// 在设计表时，关联了    category_id UUID NOT NULL REFERENCES categories (id) ON DELETE RESTRICT,
	// 所以执行时，默认删除主表，自动删除子表
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.ParamTemplate{}).Error; err != nil {
		return err
	}
	return nil
}

func (r *ParamTemplateRepo) ItemsCount(ctx context.Context, id string) (int64, error) {
	var count int64
	if err := r.db.
			WithContext(ctx).
			Model(&model.ParamTemplateItem{}).
			Where("template_id = ?", id).
			Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}