package enum

type VehicleStatus string
const (
	VehicleStatusDraft VehicleStatus = "draft" // 草稿
	VehicleStatusPublished VehicleStatus = "published" // 已发布
	VehicleStatusUnpublished VehicleStatus = "unpublished" // 未发布
	VehicleStatusDeleted VehicleStatus = "deleted" // 已删除
)

type PriceMode string
const (
	PriceModeShowPrice PriceMode = "show_price" // 显示价格
	PriceModePhoneInquiry PriceMode = "phone_inquiry" // 电话询价
)