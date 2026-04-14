package handler

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"vehivle/internal/domain/enum"
	"vehivle/internal/domain/model"
	"vehivle/internal/infrastructure/oss"
	"vehivle/internal/repository/postgres"
	"vehivle/internal/service/category"
	"vehivle/internal/service/system_setting"
	"vehivle/internal/service/vehicle"
	"vehivle/internal/transport/http/helper"
	"vehivle/pkg/response"
)

type Public struct {
	OSS             oss.MinioClient
	VehicleService  *vehicle.Service
	CategoryService *category.CategoryService
	SystemService   *system_setting.SysService
	mediaRepo       *postgres.MediaAssetRepo
}

func NewPublic(
	vehSvc *vehicle.Service,
	catSvc *category.CategoryService,
	sysSvc *system_setting.SysService,
	mediaRepo *postgres.MediaAssetRepo,
	ossClient oss.MinioClient,
) *Public {
	return &Public{
		OSS:             ossClient,
		VehicleService:  vehSvc,
		CategoryService: catSvc,
		SystemService:   sysSvc,
		mediaRepo:       mediaRepo,
	}
}

type publicContactResp struct {
	CompanyName           string  `json:"companyName"`
	CustomerServicePhone  *string `json:"customerServicePhone"`
	CustomerServiceWechat *string `json:"customerServiceWechat"`
	DisclaimerText        *string `json:"disclaimerText"`
	DefaultShareTitle     *string `json:"defaultShareTitle"`
	DefaultShareImageURL  string  `json:"defaultShareImageUrl,omitempty"`
}

func (p *Public) publishedVehicleQuery(c *gin.Context) (model.VehicleListQuery, bool) {
	var raw vehicleListQueryParams
	if err := c.ShouldBindQuery(&raw); err != nil {
		response.FailParam(c, err.Error())
		return model.VehicleListQuery{}, false
	}
	sf := strings.TrimSpace(raw.SortField)
	so := strings.TrimSpace(strings.ToLower(raw.SortOrder))
	if _, err := validatePublicSort(sf, so); err != nil {
		response.FailParam(c, err.Error())
		return model.VehicleListQuery{}, false
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
		return model.VehicleListQuery{}, false
	}
	status := enum.VehicleStatusPublished
	return model.VehicleListQuery{
		Keyword:    strings.TrimSpace(raw.Keyword),
		CategoryID: raw.CategoryID,
		Status:     &status,
		Page:       page,
		PageSize:   pageSize,
		SortField:  sf,
		SortOrder:  so,
	}, true
}

func validatePublicSort(sortField string, sortOrder string) (bool, error) {
	return helper.IsValidSortFieldAndOrder(sortField, sortOrder)
}

func (p *Public) attachCoverImageURLs(c *gin.Context, items []*model.Vehicle) {
	if len(items) == 0 {
		return
	}
	ids := make([]string, 0)
	for _, item := range items {
		if item != nil && item.CoverMediaID != "" {
			ids = append(ids, item.CoverMediaID)
		}
	}
	m, err := p.mediaRepo.MapStorageKeysByIDs(c.Request.Context(), ids)
	if err != nil {
		return
	}
	for _, item := range items {
		if item == nil || item.CoverMediaID == "" {
			continue
		}
		if storageKey, ok := m[item.CoverMediaID]; ok {
			item.CoverImageURL = p.OSS.ObjectPublicURL(storageKey)
		}
	}
}

func (p *Public) mediaURL(c *gin.Context, mediaID string) string {
	if strings.TrimSpace(mediaID) == "" {
		return ""
	}
	m, err := p.mediaRepo.MapStorageKeysByIDs(c.Request.Context(), []string{mediaID})
	if err != nil {
		return ""
	}
	if storageKey, ok := m[mediaID]; ok {
		return p.OSS.ObjectPublicURL(storageKey)
	}
	return ""
}

func (p *Public) contact(c *gin.Context) (publicContactResp, bool) {
	sys, err := p.SystemService.Detail(c.Request.Context())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return publicContactResp{}, false
		}
		return publicContactResp{}, false
	}
	resp := publicContactResp{
		CompanyName:           sys.CompanyName,
		CustomerServicePhone:  sys.CustomerServicePhone,
		CustomerServiceWechat: sys.CustomerServiceWechat,
		DisclaimerText:        sys.DisclaimerText,
		DefaultShareTitle:     sys.DefaultShareTitle,
	}
	if sys.DefaultShareImage != nil {
		resp.DefaultShareImageURL = p.mediaURL(c, *sys.DefaultShareImage)
	}
	return resp, true
}

func (p *Public) Home(c *gin.Context) {
	contact, _ := p.contact(c)
	status := enum.CategoryStatusEnabled
	categories, err := p.CategoryService.List(c.Request.Context(), model.CategoryListQuery{
		Status:   &status,
		Page:     1,
		PageSize: maxVehicleListPageSize,
	})
	if err != nil {
		response.FailBusiness(c, err.Error())
		return
	}
	published := enum.VehicleStatusPublished
	vehicles, err := p.VehicleService.ListAdmin(c.Request.Context(), model.VehicleListQuery{
		Status:   &published,
		Page:     1,
		PageSize: defaultVehicleListPageSize,
	})
	if err != nil {
		response.FailBusiness(c, err.Error())
		return
	}
	p.attachCoverImageURLs(c, vehicles.List)
	response.Success(c, gin.H{
		"banners":    []gin.H{},
		"categories": categories.List,
		"vehicles":   vehicles.List,
		"contact":    contact,
		"zones":      []gin.H{},
	})
}

func (p *Public) Categories(c *gin.Context) {
	status := enum.CategoryStatusEnabled
	result, err := p.CategoryService.List(c.Request.Context(), model.CategoryListQuery{
		Status:   &status,
		Page:     1,
		PageSize: maxVehicleListPageSize,
	})
	if err != nil {
		response.FailBusiness(c, err.Error())
		return
	}
	response.Success(c, result)
}

func (p *Public) Vehicles(c *gin.Context) {
	q, ok := p.publishedVehicleQuery(c)
	if !ok {
		return
	}
	result, err := p.VehicleService.ListAdmin(c.Request.Context(), q)
	if err != nil {
		response.FailBusiness(c, err.Error())
		return
	}
	p.attachCoverImageURLs(c, result.List)
	response.Success(c, result)
}

func (p *Public) VehicleDetail(c *gin.Context) {
	id := c.Param("id")
	target, err := p.VehicleService.GetById(c.Request.Context(), id)
	if err != nil || target.Status != enum.VehicleStatusPublished {
		response.FailNotFound(c, "车型不存在或已下架")
		return
	}
	p.attachCoverImageURLs(c, []*model.Vehicle{target})
	contact, _ := p.contact(c)
	detailMediaIDs, err := p.VehicleService.ListDetailMediaIDs(c.Request.Context(), target.ID)
	if err != nil {
		response.FailBusiness(c, err.Error())
		return
	}
	detailImages := make([]string, 0, len(detailMediaIDs))
	for _, mediaID := range detailMediaIDs {
		if url := p.mediaURL(c, mediaID); url != "" {
			detailImages = append(detailImages, url)
		}
	}
	params := []model.VehicleParamDisplay{}
	if target.CategoryID != nil && strings.TrimSpace(*target.CategoryID) != "" {
		params, err = p.VehicleService.ListPublicParams(c.Request.Context(), target.ID, *target.CategoryID)
		if err != nil {
			response.FailBusiness(c, err.Error())
			return
		}
	}
	response.Success(c, gin.H{
		"vehicle":         target,
		"contact":         contact,
		"detailImages":    detailImages,
		"params":          params,
		"emptyParamsText": "暂无参数，请联系上游获取详细配置",
	})
}

func (p *Public) Contact(c *gin.Context) {
	contact, ok := p.contact(c)
	if !ok {
		response.Success(c, gin.H{})
		return
	}
	response.Success(c, contact)
}

func (p *Public) ShareCheck(c *gin.Context) {
	id := c.Param("id")
	contact, _ := p.contact(c)
	target, err := p.VehicleService.GetById(c.Request.Context(), id)
	if err != nil || target.Status != enum.VehicleStatusPublished {
		title := contact.DefaultShareTitle
		response.Success(c, gin.H{
			"available": false,
			"message":   "车型已下架或不存在",
			"title":     title,
			"imageUrl":  contact.DefaultShareImageURL,
		})
		return
	}
	p.attachCoverImageURLs(c, []*model.Vehicle{target})
	response.Success(c, gin.H{
		"available": true,
		"vehicleId": target.ID,
		"title":     target.Name,
		"imageUrl":  target.CoverImageURL,
	})
}
