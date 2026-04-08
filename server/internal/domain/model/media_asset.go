package model

import "time"

// MediaAsset 对象存储媒体元数据，对应表 media_assets。
type MediaAsset struct {
	ID          string    `json:"id" gorm:"column:id;type:uuid"`
	StorageKey  string    `json:"storageKey" gorm:"column:storage_key"`
	MimeType    string    `json:"mimeType" gorm:"column:mime_type"`
	FileSize    int64     `json:"fileSize" gorm:"column:file_size"`
	AssetType   string    `json:"assetType" gorm:"column:asset_type"`
	CreatedAt   time.Time `json:"createdAt" gorm:"column:created_at"`
}

// TableName 与 migrations 中 CREATE TABLE media_assets 一致。
func (*MediaAsset) TableName() string {
	return "media_assets"
}
