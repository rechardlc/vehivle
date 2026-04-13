package enum

import (
	"database/sql/driver"
	"fmt"
	"strconv"
)

/**
 * 车型状态枚举
 */
type VehicleStatus string
const (
	VehicleStatusDraft VehicleStatus = "draft" // 草稿
	VehicleStatusPublished VehicleStatus = "published" // 已发布
	VehicleStatusUnpublished VehicleStatus = "unpublished" // 未发布
	VehicleStatusDeleted VehicleStatus = "deleted" // 已删除
)

/**
 * 价格模式枚举
 */
type PriceMode string
const (
	PriceModeShowPrice PriceMode = "show_price" // 显示价格
	PriceModePhoneInquiry PriceMode = "phone_inquiry" // 电话询价
)

/**
 * 分类状态：数据库与 JSON API 均为 SMALLINT 语义（1=启用 0=禁用），序列化为 JSON 数字。
 */
type CategoryStatus int8

const (
	CategoryStatusDisabled CategoryStatus = 0
	CategoryStatusEnabled  CategoryStatus = 1
)

/**
 * 模板状态枚举
 */
type ParamTemplateStatus int8
const (
	ParamTemplateStatusDisabled ParamTemplateStatus = 0
	ParamTemplateStatusEnabled  ParamTemplateStatus = 1
)

func (s ParamTemplateStatus) Value() (driver.Value, error) {
	return int64(s), nil
}

func (s *ParamTemplateStatus) Scan(src interface{}) error {
	if src == nil {
		*s = ParamTemplateStatusDisabled
		return nil
	}
	switch v := src.(type) {
	case int64:
		*s = ParamTemplateStatus(v)
	case int32:
		*s = ParamTemplateStatus(v)
	case int16:
		*s = ParamTemplateStatus(v)
	case int8:
		*s = ParamTemplateStatus(v)
	case []byte:
		n, err := strconv.ParseInt(string(v), 10, 8)
		if err != nil {
			return err
		}
		*s = ParamTemplateStatus(n)
	default:
		return fmt.Errorf("cannot scan %T into ParamTemplateStatus", src)
	}
	return nil
}

// ParseCategoryStatus 解析 URL 表单等场景下的字符串 "0" / "1"。
func ParseCategoryStatus(s string) (CategoryStatus, error) {
	switch s {
	case "1":
		return CategoryStatusEnabled, nil
	case "0":
		return CategoryStatusDisabled, nil
	default:
		return 0, fmt.Errorf("invalid category status: %q", s)
	}
}

func (s CategoryStatus) Value() (driver.Value, error) {
	return int64(s), nil
}

func (s *CategoryStatus) Scan(src interface{}) error {
	if src == nil {
		*s = CategoryStatusDisabled
		return nil
	}
	switch v := src.(type) {
	case int64:
		*s = CategoryStatus(v)
	case int32:
		*s = CategoryStatus(v)
	case int16:
		*s = CategoryStatus(v)
	case int8:
		*s = CategoryStatus(v)
	case []byte:
		n, err := strconv.ParseInt(string(v), 10, 8)
		if err != nil {
			return err
		}
		*s = CategoryStatus(n)
	default:
		return fmt.Errorf("cannot scan %T into CategoryStatus", src)
	}
	return nil
}