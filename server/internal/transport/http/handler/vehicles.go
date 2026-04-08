package handler

import (
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"vehivle/internal/domain/enum"
	"vehivle/internal/domain/model"
	"vehivle/internal/domain/rule"
	"vehivle/internal/infrastructure/oss"
	"vehivle/internal/repository/postgres"
	"vehivle/internal/service/vehicle"
	"vehivle/pkg/response"
)

// Vehicles 车型处理器
type Vehicles struct {
	OSS            oss.MinioClient
	VehicleService *vehicle.Service
	mediaRepo      *postgres.MediaAssetRepo
}

// NewVehicles 创建车型处理器（OSS 用于拼接封面公网 URL）。
func NewVehicles(db *gorm.DB, ossClient oss.MinioClient) *Vehicles {
	repo := postgres.NewVehicleRepo(db)
	return &Vehicles{
		OSS:            ossClient,
		VehicleService: vehicle.NewService(repo),
		mediaRepo:      postgres.NewMediaAssetRepo(db),
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

type vehicleUpdateBody struct {
	Name          *string         `json:"name,omitempty"`
	CategoryID    *string         `json:"categoryId,omitempty"`
	CoverMediaID  *string         `json:"coverMediaId,omitempty"`
	PriceMode     *enum.PriceMode `json:"priceMode,omitempty"`
	MSRPPrice     *int            `json:"msrpPrice,omitempty"`
	SellingPoints *string         `json:"sellingPoints,omitempty"`
	SortOrder     *int            `json:"sortOrder,omitempty"`
}

type batchStatusBody struct {
	IDs    []string           `json:"ids"`
	Status enum.VehicleStatus `json:"status"`
}

// List 获取车型列表：/api/v1/public/vehicles 仅已上架；/api/v1/admin/vehicles 含全部状态。
func (v *Vehicles) List(c *gin.Context) {
	onlyPublished := strings.Contains(c.Request.URL.Path, "/public/")
	items, err := v.VehicleService.List(c.Request.Context(), onlyPublished)
	if err != nil {
		response.FailBusiness(c, err.Error())
		return
	}
	v.attachCoverImageURLs(c, items)
	response.Success(c, gin.H{"items": items})
}

func (v *Vehicles) attachCoverImageURLs(c *gin.Context, items []*model.Vehicle) {
	if len(items) == 0 {
		return
	}
	ids := make([]string, 0)
	for _, it := range items {
		if it != nil && it.CoverMediaID != "" {
			ids = append(ids, it.CoverMediaID)
		}
	}
	if len(ids) == 0 {
		return
	}
	m, err := v.mediaRepo.MapStorageKeysByIDs(c.Request.Context(), ids)
	if err != nil {
		return
	}
	for _, it := range items {
		if it == nil || it.CoverMediaID == "" {
			continue
		}
		if sk, ok := m[it.CoverMediaID]; ok {
			it.CoverImageURL = v.OSS.ObjectPublicURL(sk)
		}
	}
}

// Create 创建车型
func (v *Vehicles) Create(c *gin.Context) {
	var body vehicleCreateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailParam(c, err.Error())
		return
	}
	now := time.Now()
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
		CreatedAt:     now,
		UpdatedAt:     now,
	})
	if err != nil {
		response.FailBusiness(c, err.Error())
		return
	}
	v.attachCoverImageURLs(c, []*model.Vehicle{veh})
	response.Success(c, veh)
}

// Update 更新车型（部分字段）。
func (v *Vehicles) Update(c *gin.Context) {
	id := c.Param("vehicle_id")
	var body vehicleUpdateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailParam(c, err.Error())
		return
	}
	veh, err := v.VehicleService.UpdateVehicle(c.Request.Context(), id, &vehicle.VehicleUpdateInput{
		Name:          body.Name,
		CategoryID:    body.CategoryID,
		CoverMediaID:  body.CoverMediaID,
		PriceMode:     body.PriceMode,
		MSRPPrice:     body.MSRPPrice,
		SellingPoints: body.SellingPoints,
		SortOrder:     body.SortOrder,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.FailBusiness(c, "车型不存在")
			return
		}
		response.FailBusiness(c, err.Error())
		return
	}
	v.attachCoverImageURLs(c, []*model.Vehicle{veh})
	response.Success(c, veh)
}

// Delete 逻辑删除车型
func (v *Vehicles) Delete(c *gin.Context) {
	id := c.Param("vehicle_id")
	if err := v.VehicleService.SoftDelete(c.Request.Context(), id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.FailBusiness(c, "车型不存在")
			return
		}
		response.FailBusiness(c, err.Error())
		return
	}
	response.Success(c, gin.H{"ok": true})
}

// Publish 发布车型（MVP：详情图/参数未接库前放宽校验）。
func (v *Vehicles) Publish(c *gin.Context) {
	id := c.Param("vehicle_id")
	target, err := v.VehicleService.GetById(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.FailBusiness(c, "车型不存在")
			return
		}
		response.FailBusiness(c, err.Error())
		return
	}
	req := rule.PublishRequirements{
		HasCoverImage:     target.CoverMediaID != "",
		HasDetailImages:   true,
		HasRequiredParams: true,
	}
	if err := v.VehicleService.Publish(c.Request.Context(), id, req); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.FailBusiness(c, "车型不存在")
			return
		}
		response.FailBusiness(c, err.Error())
		return
	}
	response.Success(c, gin.H{"ok": true})
}

// Unpublish 下架车型
func (v *Vehicles) Unpublish(c *gin.Context) {
	id := c.Param("vehicle_id")
	if err := v.VehicleService.Unpublish(c.Request.Context(), id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.FailBusiness(c, "车型不存在")
			return
		}
		response.FailBusiness(c, err.Error())
		return
	}
	response.Success(c, gin.H{"ok": true})
}

// Duplicate 复制车型为草稿
func (v *Vehicles) Duplicate(c *gin.Context) {
	id := c.Param("vehicle_id")
	veh, err := v.VehicleService.Duplicate(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.FailBusiness(c, "车型不存在")
			return
		}
		response.FailBusiness(c, err.Error())
		return
	}
	v.attachCoverImageURLs(c, []*model.Vehicle{veh})
	response.Success(c, veh)
}

// BatchStatus 批量上下架
func (v *Vehicles) BatchStatus(c *gin.Context) {
	var body batchStatusBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailParam(c, err.Error())
		return
	}
	if body.Status != enum.VehicleStatusPublished && body.Status != enum.VehicleStatusUnpublished {
		response.FailParam(c, "status 仅支持 published 或 unpublished")
		return
	}
	if err := v.VehicleService.BatchSetStatus(c.Request.Context(), body.IDs, body.Status); err != nil {
		response.FailBusiness(c, err.Error())
		return
	}
	response.Success(c, gin.H{"ok": true})
}
