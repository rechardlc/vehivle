package handler

import (
	"vehivle/internal/domain/model"
	"vehivle/internal/service/param_template"

	"vehivle/internal/transport/http/helper"

	"vehivle/pkg/response"

	"github.com/gin-gonic/gin"
)

type ParamTemplates struct {
	ParamTemplateService *param_template.ParamTemplateService
}

func NewParamTemplates(paramTemplateService *param_template.ParamTemplateService) *ParamTemplates {
	return &ParamTemplates{ParamTemplateService: paramTemplateService}
}

func (p *ParamTemplates) Create(c *gin.Context) {
	var body model.ParamTemplate
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailParam(c, "参数错误,请检查参数是否符合要求: "+err.Error())
		return
	}
	paramTemplate, err := p.ParamTemplateService.Create(c.Request.Context(), &body)
	if err != nil {
		response.FailBusiness(c, "创建模板失败: "+err.Error())
		return
	}
	response.Success(c, paramTemplate)
}

func (p *ParamTemplates) Update(c *gin.Context) {
	id := c.Param("id")
	if err := helper.RequiredField(id); err != nil {
		response.FailParam(c, err.Error())
		return
	}
	var body model.ParamTemplate
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailParam(c, "参数错误,请检查参数是否符合要求: "+err.Error())
		return
	}
	result, err := p.ParamTemplateService.Update(c.Request.Context(), id, &body)
	if err != nil {
		response.FailBusiness(c, "更新失败！")
		return
	}
	response.Success(c, result)
}

func (p *ParamTemplates) AllDetail(c *gin.Context) {
	id := c.Param("id")
	if err := helper.RequiredField(id); err != nil {
		response.FailParam(c, err.Error())
		return
	}
	result, err := p.ParamTemplateService.AllDetail(c, id)
	if err != nil {
		response.FailBusiness(c, "模板数据为空")
		return
	}
	response.Success(c, result)
}

func (p *ParamTemplates) Detail(c *gin.Context) {
	id := c.Param("id")
	if err := helper.RequiredField(id); err != nil {
		response.FailParam(c, err.Error())
		return
	}
	result, err := p.ParamTemplateService.Detail(c, id)
	if err != nil {
		response.FailBusiness(c, "模板数据为空")
		return
	}
	response.Success(c, result)
}