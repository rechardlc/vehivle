package handler

import (
	"vehivle/internal/domain/model"
	"vehivle/internal/service/param_template"

	"vehivle/internal/transport/http/helper"
	"vehivle/internal/transport/http/constant"

	"vehivle/pkg/response"

	"github.com/gin-gonic/gin"
)


// 当query查询参数或者post表单提交时，使用form tag标注
type ParamTemplates struct {
	ParamTemplateService *param_template.ParamTemplateService
}
/*
	1. queryListParams不被外部使用，所以小写就可以了
	2. Page、PageSize字段需要与gin交互，所以需要大写，小写无法反射
*/
// type queryListParams struct {
// 	Page     int `form:"page"`
// 	PageSize int `form:"pageSize"`
// }

func NewParamTemplates(paramTemplateService *param_template.ParamTemplateService) *ParamTemplates {
	return &ParamTemplates{ParamTemplateService: paramTemplateService}
}

func (p *ParamTemplates) List(c *gin.Context) {
	var query model.TmpQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.FailParam(c, "参数异常")
	}
	if query.Page == 0 {
		query.Page = constant.DEFAULT_CATEGORY_LIST_PAGE
	}
	if query.PageSize == 0 {
		query.PageSize = constant.DEFAULT_CATEGORY_LIST_PAGE_SIZE
	}
	if query.PageSize > constant.MAX_CATEGORY_LIST_PAGE_SIZE {
		response.FailParam(c, "pagesize最大值超过100")
		return
	}
	result, err := p.ParamTemplateService.List(c.Request.Context(), &query)
	if err != nil {
		response.FailBusiness(c, err.Error())
		return
	}
	response.Success(c, result)
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

func (p *ParamTemplates) GetItemsById(c *gin.Context) {
	id := c.Param("id")
	if err := helper.RequiredField(id); err != nil {
		response.FailParam(c, err.Error())
		return
	}
	result, err := p.ParamTemplateService.GetItemsById(c.Request.Context(), id)
	if err != nil {
		response.FailBusiness(c, "模板数据为空")
		return
	}
	response.Success(c, result)
}

func (p *ParamTemplates) GetById(c *gin.Context) {
	id := c.Param("id")
	if err := helper.RequiredField(id); err != nil {
		response.FailParam(c, err.Error())
		return
	}
	result, err := p.ParamTemplateService.GetById(c.Request.Context(), id)
	if err != nil {
		response.FailBusiness(c, "模板数据为空")
		return
	}
	response.Success(c, result)
}

func (p *ParamTemplates) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := helper.RequiredField(id); err != nil {
		response.FailParam(c, err.Error())
		return
	}
	err := p.ParamTemplateService.Delete(c.Request.Context(), id)
	if err != nil {
		response.FailBusiness(c, "删除失败！")
		return
	}
	response.Success(c, "")
}
