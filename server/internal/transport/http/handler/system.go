package handler

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"vehivle/internal/domain/model"
	"vehivle/internal/infrastructure/oss"
	"vehivle/internal/service/system_setting"
	"vehivle/pkg/response"
)

type System struct {
	OSS        oss.MinioClient
	SysService *system_setting.SysService
}

func NewSysSettings(svc *system_setting.SysService, ossClient oss.MinioClient) *System {
	return &System{
		OSS:        ossClient,
		SysService: svc,
	}
}

type sysCreateBody struct {
	CompanyName           string  `json:"companyName" binding:"required"`
	CustomerServicePhone  *string `json:"customerServicePhone"`
	CustomerServiceWechat *string `json:"customerServiceWechat"`
	DefaultPriceMode      string  `json:"defaultPriceMode" binding:"required,oneof=show_price phone_inquiry"`
	DisclaimerText        *string `json:"disclaimerText"`
	DefaultShareTitle     *string `json:"defaultShareTitle"`
	DefaultShareImage     *string `json:"defaultShareImage"`
}

type sysUpdateBody struct {
	CompanyName           *string `json:"companyName"`
	CustomerServicePhone  *string `json:"customerServicePhone"`
	CustomerServiceWechat *string `json:"customerServiceWechat"`
	DefaultPriceMode      *string `json:"defaultPriceMode" binding:"omitempty,oneof=show_price phone_inquiry"`
	DisclaimerText        *string `json:"disclaimerText"`
	DefaultShareTitle     *string `json:"defaultShareTitle"`
	DefaultShareImage     *string `json:"defaultShareImage"`
}

func (s *System) Detail(c *gin.Context) {
	sys, err := s.SysService.Detail(c.Request.Context())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Success(c, nil)
			return
		}
		response.FailBusiness(c, "系统数据异常！")
		return
	}

	response.Success(c, sysToResp(sys, s.OSS))
}

func (s *System) Create(c *gin.Context) {
	var body sysCreateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailParam(c, "参数校验失败："+err.Error())
		return
	}

	m := &model.SystemSetting{
		CompanyName:           strings.TrimSpace(body.CompanyName),
		CustomerServicePhone:  body.CustomerServicePhone,
		CustomerServiceWechat: body.CustomerServiceWechat,
		DefaultPriceMode:      body.DefaultPriceMode,
		DisclaimerText:        body.DisclaimerText,
		DefaultShareTitle:     body.DefaultShareTitle,
		DefaultShareImage:     body.DefaultShareImage,
	}

	if err := s.SysService.Create(c.Request.Context(), m); err != nil {
		response.FailBusiness(c, err.Error())
		return
	}

	response.Success(c, sysToResp(m, s.OSS))
}

func (s *System) Update(c *gin.Context) {
	var body sysUpdateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailParam(c, "参数校验失败："+err.Error())
		return
	}

	updates := make(map[string]interface{})
	if body.CompanyName != nil {
		v := strings.TrimSpace(*body.CompanyName)
		if v == "" {
			response.FailParam(c, "公司名称不能为空")
			return
		}
		updates["company_name"] = v
	}
	if body.DefaultPriceMode != nil {
		updates["default_price_mode"] = *body.DefaultPriceMode
	}
	if body.CustomerServicePhone != nil {
		updates["customer_service_phone"] = *body.CustomerServicePhone
	}
	if body.CustomerServiceWechat != nil {
		updates["customer_service_wechat"] = *body.CustomerServiceWechat
	}
	if body.DisclaimerText != nil {
		updates["disclaimer_text"] = *body.DisclaimerText
	}
	if body.DefaultShareTitle != nil {
		updates["default_share_title"] = *body.DefaultShareTitle
	}
	if body.DefaultShareImage != nil {
		updates["default_share_image"] = *body.DefaultShareImage
	}

	if len(updates) == 0 {
		response.FailParam(c, "未提供任何需更新的字段")
		return
	}

	if err := s.SysService.Update(c.Request.Context(), updates); err != nil {
		response.FailBusiness(c, err.Error())
		return
	}

	detail, err := s.SysService.Detail(c.Request.Context())
	if err != nil {
		response.FailBusiness(c, "更新成功但获取详情失败")
		return
	}
	response.Success(c, sysToResp(detail, s.OSS))
}

type sysResp struct {
	ID                    int     `json:"id"`
	CompanyName           string  `json:"companyName"`
	CustomerServicePhone  *string `json:"customerServicePhone"`
	CustomerServiceWechat *string `json:"customerServiceWechat"`
	DefaultPriceMode      string  `json:"defaultPriceMode"`
	DisclaimerText        *string `json:"disclaimerText"`
	DefaultShareTitle     *string `json:"defaultShareTitle"`
	DefaultShareImage     *string `json:"defaultShareImage"`
	DefaultShareImageURL  string  `json:"defaultShareImageUrl,omitempty"`
	UpdatedAt             string  `json:"updatedAt"`
}

func sysToResp(m *model.SystemSetting, ossClient oss.MinioClient) sysResp {
	r := sysResp{
		ID:                    m.ID,
		CompanyName:           m.CompanyName,
		CustomerServicePhone:  m.CustomerServicePhone,
		CustomerServiceWechat: m.CustomerServiceWechat,
		DefaultPriceMode:      m.DefaultPriceMode,
		DisclaimerText:        m.DisclaimerText,
		DefaultShareTitle:     m.DefaultShareTitle,
		DefaultShareImage:     m.DefaultShareImage,
		UpdatedAt:             m.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if m.DefaultShareImage != nil && *m.DefaultShareImage != "" {
		r.DefaultShareImageURL = ossClient.ObjectPublicURL(*m.DefaultShareImage)
	}
	return r
}
