package postgres

import (
	"context"

	"vehivle/internal/domain/model"

	"gorm.io/gorm"
)

// MediaAssetRepo 媒体资源元数据访问
type MediaAssetRepo struct {
	db *gorm.DB
}

// NewMediaAssetRepo 创建 MediaAssetRepo
func NewMediaAssetRepo(db *gorm.DB) *MediaAssetRepo {
	return &MediaAssetRepo{db: db}
}

// Create 插入一条媒体记录（主键由调用方生成 UUID 或由库端 default 生成）。
func (r *MediaAssetRepo) Create(ctx context.Context, row *model.MediaAsset) error {
	return r.db.WithContext(ctx).Create(row).Error
}

// MapStorageKeysByIDs 批量查询 id -> storage_key，用于列表拼接封面 URL。
func (r *MediaAssetRepo) MapStorageKeysByIDs(ctx context.Context, ids []string) (map[string]string, error) {
	if len(ids) == 0 {
		return map[string]string{}, nil
	}
	var rows []model.MediaAsset
	if err := r.db.WithContext(ctx).Model(&model.MediaAsset{}).
		Select("id", "storage_key").
		Where("id IN ?", ids).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[string]string, len(rows))
	for i := range rows {
		out[rows[i].ID] = rows[i].StorageKey
	}
	return out, nil
}
