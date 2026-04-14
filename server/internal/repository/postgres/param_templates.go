package postgres

import (
	"context"
	"errors"
	"time"

	"vehivle/internal/domain/model"

	"gorm.io/gorm"
)

type ParamTemplateRepo struct {
	db *gorm.DB
}

func NewParamTemplateRepo(db *gorm.DB) *ParamTemplateRepo {
	return &ParamTemplateRepo{db: db}
}

// 查询模板列表（分页）
func (r *ParamTemplateRepo) List(ctx context.Context, page int, pageSize int) ([]model.ParamTemplate, error) {
	var rows []model.ParamTemplate
	offset := (page - 1) * pageSize
	err := r.db.WithContext(ctx).
		Model(&model.ParamTemplate{}).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// 创建模板
func (r *ParamTemplateRepo) Create(ctx context.Context, p *model.ParamTemplate) (*model.ParamTemplate, error) {
	if err := r.db.WithContext(ctx).Create(p).Error; err != nil {
		return nil, err
	}
	return p, nil
}

// 更新模板
func (r *ParamTemplateRepo) Update(ctx context.Context, id string, p *model.ParamTemplate) (*model.ParamTemplate, error) {
	if id == "" {
		return nil, errors.New("模板 id 不能为空")
	}
	return p, r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{
			"name":        p.Name,
			"category_id": p.CategoryID,
			"status":      p.Status,
			"updated_at":  time.Now(),
		}
		if err := tx.Model(&model.ParamTemplate{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return err
		}
		nItemIds := make([]string, 0)
		for i := range p.Items {
			p.Items[i].TemplateID = id
			p.Items[i].UpdatedAt = time.Now()
			if p.Items[i].ID == nil || *p.Items[i].ID == "" {
				p.Items[i].ID = nil
				if err := tx.Create(&p.Items[i]).Error; err != nil {
					return err
				}
			} else {
				itemID := *p.Items[i].ID
				itemUpdates := map[string]interface{}{
					"field_key":  p.Items[i].FieldKey,
					"field_name": p.Items[i].FieldName,
					"field_type": p.Items[i].FieldType,
					"unit":       p.Items[i].Unit,
					"required":   p.Items[i].Required,
					"display":    p.Items[i].Display,
					"sort_order": p.Items[i].SortOrder,
					"updated_at": time.Now(),
				}
				res := tx.Model(&model.ParamTemplateItem{}).
					Where("id = ? AND template_id = ?", itemID, id).
					Updates(itemUpdates)
				if res.Error != nil {
					return res.Error
				}
				if res.RowsAffected == 0 {
					return errors.New("参数项不存在或不属于当前模板")
				}
			}
			if p.Items[i].ID == nil || *p.Items[i].ID == "" {
				return errors.New("参数项保存后未生成 id")
			}
			nItemIds = append(nItemIds, *p.Items[i].ID)
		}
		delQuery := tx.Where("template_id = ?", id)
		if len(nItemIds) > 0 {
			delQuery = delQuery.Where("id NOT IN ?", nItemIds)
		}
		if err := delQuery.Delete(&model.ParamTemplateItem{}).Error; err != nil {
			return err
		}
		p.ID = &id
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
	// param_template_items.template_id 使用 ON DELETE CASCADE，删除模板会自动清理子项。
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
