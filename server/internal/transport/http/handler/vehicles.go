package handler

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"vehivle/internal/domain/enum"
	"vehivle/internal/domain/model"
	"vehivle/internal/repository/postgres"
	"vehivle/internal/service/vehicle"
	"vehivle/pkg/response"
)

// Vehicles 车型处理器
type Vehicles struct {
	DB             *gorm.DB
	VehicleService *vehicle.Service
}

// NewVehicles 创建车型处理器
func NewVehicles(db *gorm.DB) *Vehicles {
	repo := postgres.NewVehicleRepo(db)
	return &Vehicles{
		DB:             db,
		VehicleService: vehicle.NewService(repo),
	}
}

// vehicleCreateBody 创建车型请求体（camelCase JSON）
type vehicleCreateBody struct {
	Name          string         `json:"name"`
	CategoryID    *string        `json:"categoryId"`
	CoverMediaID  string         `json:"coverMediaId"`
	PriceMode     enum.PriceMode `json:"priceMode"`
	MSRPPrice     int            `json:"msrpPrice"`
	SellingPoints string         `json:"sellingPoints"`
	SortOrder     int            `json:"sortOrder"`
}

// List 获取车型列表：/api/v1/public/vehicles 仅已上架；/api/v1/admin/vehicles 含全部状态。
func (v *Vehicles) List(c *gin.Context) {
	onlyPublished := strings.Contains(c.Request.URL.Path, "/public/")
	items, err := v.VehicleService.List(c.Request.Context(), onlyPublished)
	if err != nil {
		response.FailBusiness(c, err.Error())
		return
	}
	response.Success(c, gin.H{"items": items})
}

// Create 创建车型
func (v *Vehicles) Create(c *gin.Context) {
	var body vehicleCreateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailBusiness(c, err.Error())
		return
	}
	veh, err := v.VehicleService.Create(c.Request.Context(), &model.Vehicle{
		ID:            uuid.New().String(),
		Name:          body.Name,
		CategoryID:    body.CategoryID,
		CoverMediaID:  body.CoverMediaID,
		PriceMode:     body.PriceMode,
		MSRPPrice:     body.MSRPPrice,
		Status:        enum.VehicleStatusDraft,
		SellingPoints: body.SellingPoints,
		SortOrder:     body.SortOrder,
	})
	if err != nil {
		response.FailBusiness(c, err.Error())
		return
	}
	response.Success(c, veh)
}

/** strPtr 将字符串转为 *string；空串返回 nil，对齐数据库 NULL 语义。 */
func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// Update 更新车型
func (v *Vehicles) Update(c *gin.Context) {
	response.Success(c, gin.H{
		"message": "ok",
	})
}

// Delete 删除车型
func (v *Vehicles) Delete(c *gin.Context) {
	response.Success(c, gin.H{
		"message": "ok",
	})
}
