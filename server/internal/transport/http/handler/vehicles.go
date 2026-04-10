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
	"vehivle/internal/service/category"
	"vehivle/internal/service/vehicle"
	"vehivle/internal/transport/http/helper"
	"vehivle/pkg/response"
)

const (
	// 管理端车型列表默认页码
	defaultVehicleListPage = 1
	// 管理端车型列表默认每页条数
	defaultVehicleListPageSize = 10
	// 管理端车型列表最大每页条数
	maxVehicleListPageSize = 100
)

// Vehicles 车型处理器
type Vehicles struct {
	OSS             oss.MinioClient
	VehicleService  *vehicle.Service
	CategoryService *category.CategoryService
	mediaRepo       *postgres.MediaAssetRepo
}

// NewVehicles 创建车型处理器（OSS 用于拼接封面公网 URL；分类服务用于列表回填与写入校验）。
func NewVehicles(db *gorm.DB, ossClient oss.MinioClient) *Vehicles {
	repo := postgres.NewVehicleRepo(db)
	catRepo := postgres.NewCategoryRepo(db)
	return &Vehicles{
		OSS:             ossClient,
		VehicleService:  vehicle.NewService(repo),
		CategoryService: category.NewCategoryService(catRepo),
		mediaRepo:       postgres.NewMediaAssetRepo(db),
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

// vehicleListQueryParams 绑定 GET /admin/vehicles 的 query（与分类列表风格一致）。
type vehicleListQueryParams struct {
	Keyword    string  `form:"keyword"`
	CategoryID *string `form:"categoryId"`
	Status     string  `form:"status"`
	Page       int     `form:"page"`
	PageSize   int     `form:"pageSize"`
	SortField  string  `form:"sortField"`
	SortOrder  string  `form:"sortOrder"`
}

// enrichVehicleCategoryNames 按 category_id 回填 CategoryName，供管理端表格展示。
func (v *Vehicles) enrichVehicleCategoryNames(c *gin.Context, list []*model.Vehicle) {
	seen := make(map[string]struct{})
	ids := make([]string, 0)
	for _, item := range list {
		if item == nil || item.CategoryID == nil {
			continue
		}
		id := strings.TrimSpace(*item.CategoryID)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return
	}
	nameByID := make(map[string]string, len(ids))
	for _, id := range ids {
		cat, err := v.CategoryService.GetById(c.Request.Context(), id)
		if err != nil || cat == nil {
			continue
		}
		nameByID[id] = cat.Name
	}
	for _, item := range list {
		if item == nil || item.CategoryID == nil {
			continue
		}
		id := strings.TrimSpace(*item.CategoryID)
		if n, ok := nameByID[id]; ok {
			item.CategoryName = n
		}
	}
}

// ensureCategoryExists 当 categoryId 非空时校验分类存在（创建/更新车型用）。
func (v *Vehicles) ensureCategoryExists(c *gin.Context, categoryID *string) error {
	if categoryID == nil {
		return nil
	}
	id := strings.TrimSpace(*categoryID)
	if id == "" {
		return nil
	}
	_, err := v.CategoryService.GetById(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("分类不存在")
		}
		return err
	}
	return nil
}

// List 获取车型列表：公开路径仅返回已上架，支持 keyword/categoryId 与分页（滚动加载）；管理端额外支持 status 筛选。响应均为 { list, page }。
func (v *Vehicles) List(c *gin.Context) {
	onlyPublished := strings.Contains(c.Request.URL.Path, "/public/")

	var raw vehicleListQueryParams
	// ShouldBindQuery 会自动将请求参数绑定到 raw 结构体中
	// &raw 是raw的地址，所以ShouldBindQuery会自动将请求参数绑定到raw中
	if err := c.ShouldBindQuery(&raw); err != nil {
		response.FailParam(c, err.Error())
		return
	}
	// stings.TrimSpace 去除空格
	// strings.ToLower 转换为小写
	sf := strings.TrimSpace(raw.SortField)
	so := strings.TrimSpace(strings.ToLower(raw.SortOrder))
	// 如果sf不为空，并且sf不等于createdAt，则返回错误
	_, err := helper.IsValidSortFieldAndOrder(sf, so)
	if err != nil {
		response.FailParam(c, err.Error())
		return
	}
	page := raw.Page
	if page <= 0 {
		page = defaultVehicleListPage
	}
	pageSize := raw.PageSize
	if pageSize <= 0 {
		pageSize = defaultVehicleListPageSize
	}
	if pageSize > maxVehicleListPageSize {
		response.FailParam(c, "pageSize 不能大于100")
		return
	}

	q := model.VehicleListQuery{
		Keyword:    strings.TrimSpace(raw.Keyword),
		CategoryID: raw.CategoryID,
		Page:       page,
		PageSize:   pageSize,
		SortField:  sf,
		SortOrder:  so,
	}
	if onlyPublished {
		pub := enum.VehicleStatusPublished
		// &pub 是pub的地址，所以q.Status是pub的地址
		q.Status = &pub
	} else if strings.TrimSpace(raw.Status) != "" {
		// 类型转换
		st := enum.VehicleStatus(strings.TrimSpace(raw.Status))
		// enum.VehicleStatus 是 VehicleStatus 的类型，所以st是 VehicleStatus 的类型
		switch st {
		case enum.VehicleStatusDraft, enum.VehicleStatusPublished, enum.VehicleStatusUnpublished, enum.VehicleStatusDeleted:
			q.Status = &st
			// &st 是st的地址，所以q.Status是st的地址
		default:
			response.FailParam(c, "无效的 status")
			return
		}
	}

	result, err := v.VehicleService.ListAdmin(c.Request.Context(), q)
	if err != nil {
		response.FailBusiness(c, err.Error())
		return
	}
	v.attachCoverImageURLs(c, result.List)
	if !onlyPublished {
		v.enrichVehicleCategoryNames(c, result.List)
	}
	response.Success(c, result)
}
/**
 * 附件封面图片URL
 * @param c *gin.Context
 * @param items []*model.Vehicle
 */
func (v *Vehicles) attachCoverImageURLs(c *gin.Context, items []*model.Vehicle) {
	// 如果items为空，则返回
	if len(items) == 0 {
		return
	}
	// 创建一个空数组
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
	// 确保分类存在
	if err := v.ensureCategoryExists(c, body.CategoryID); err != nil {
		response.FailBusiness(c, err.Error())
		return
	}
	// 创建车型
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
	if body.CategoryID != nil {
		if err := v.ensureCategoryExists(c, body.CategoryID); err != nil {
			response.FailBusiness(c, err.Error())
			return
		}
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
	// 根据id查询车型
	target, err := v.VehicleService.GetById(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.FailBusiness(c, "车型不存在")
			return
		}
		response.FailBusiness(c, err.Error())
		return
	}
	// 创建发布要求
	req := rule.PublishRequirements{
		HasCoverImage:     target.CoverMediaID != "",
		HasDetailImages:   true,
		HasRequiredParams: true,
	}
	if err := v.VehicleService.Publish(c.Request.Context(), id, req); err != nil {
		// errors.Is 判断err是否是gorm.ErrRecordNotFound
		// gorm.ErrRecordNotFound 是gorm.ErrRecordNotFound类型的错误
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
