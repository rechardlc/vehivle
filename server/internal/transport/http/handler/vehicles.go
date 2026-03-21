package handler

import (
	"strings"

	"vehivle/internal/repository/postgres"
	"vehivle/internal/service/vehicle"
	"vehivle/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Vehicles 车型处理器
type Vehicles struct {
	DB             *gorm.DB
	VehicleService *vehicle.Service
}

// NewVehicles 创建车型处理器
func NewVehicles(db *gorm.DB) *Vehicles {
	// 创建车辆仓库实例
	repo := postgres.NewVehicleRepo(db)
	return &Vehicles{
		DB: db, 
		VehicleService: vehicle.NewService(repo),
	}
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

// 创建车型
func (v *Vehicles) Create(c *gin.Context) {
	response.Success(c, gin.H{
		"message": "ok",
	})
}

// 更新车型
func (v *Vehicles) Update(c *gin.Context) {
	response.Success(c, gin.H{
		"message": "ok",
	})
}

// 删除车型
func (v *Vehicles) Delete(c *gin.Context) {
	response.Success(c, gin.H{
		"message": "ok",
	})
}
