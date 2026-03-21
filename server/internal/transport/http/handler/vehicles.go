package handler

import (
	"vehivle/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"vehivle/internal/service/vehicle"
	"vehivle/internal/repository"
)
// Vehicles 车型处理器
type Vehicles struct {
	DB *gorm.DB
	VehicleService *vehicle.Service
}

// NewVehicles 创建车型处理器
func NewVehicles(db *gorm.DB) *Vehicles {
	return &Vehicles{DB: db, VehicleService: vehicle.NewService(postgres.New(db))}
}

// 获取车型列表
func (v *Vehicles) List(c *gin.Context) {
	response.Success(c, gin.H{
		"message": "ok",
	})
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
