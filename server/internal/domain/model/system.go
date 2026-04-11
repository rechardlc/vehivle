package model

import "time"

// SystemSetting 系统全局配置（单行表，id 恒为 1）。
// 必填字段用值类型 + NOT NULL；可选字段用指针 + NULL。
type SystemSetting struct {
	ID                    int       `json:"id"                      gorm:"primaryKey"`
	CompanyName           string    `json:"company_name"            gorm:"column:company_name;not null"`
	CustomerServicePhone  *string   `json:"customer_service_phone"  gorm:"column:customer_service_phone"`
	CustomerServiceWechat *string   `json:"customer_service_wechat" gorm:"column:customer_service_wechat"`
	DefaultPriceMode      string    `json:"default_price_mode"      gorm:"column:default_price_mode;not null"`
	DisclaimerText        *string   `json:"disclaimer_text"         gorm:"column:disclaimer_text"`
	DefaultShareTitle     *string   `json:"default_share_title"     gorm:"column:default_share_title"`
	DefaultShareImage     *string   `json:"default_share_image"     gorm:"column:default_share_image"`
	CreatedAt             time.Time `json:"created_at"              gorm:"autoCreateTime"`
	UpdatedAt             time.Time `json:"updated_at"              gorm:"autoUpdateTime"`
}

func (*SystemSetting) TableName() string {
	return "system_settings"
}
