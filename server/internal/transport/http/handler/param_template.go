package handler

import (
	"strings"

	"vehivle/internal/domain/enum"
	"vehivle/internal/domain/model"
	"vehivle/internal/service/param_template"

	"vehivle/internal/transport/http/constant"
	"vehivle/internal/transport/http/helper"

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

type paramTemplateListQuery struct {
	Page     int `form:"page"`
	PageSize int `form:"pageSize"`
}

type paramTemplateItemDTO struct {
	ID        *string `json:"id,omitempty"`
	FieldKey  string  `json:"fieldKey" binding:"required"`
	FieldName string  `json:"fieldName" binding:"required"`
	FieldType string  `json:"fieldType" binding:"required,oneof=text number single_select"`
	Unit      *string `json:"unit"`
	Required  *int8   `json:"required" binding:"required,oneof=0 1"`
	Display   *int8   `json:"display" binding:"required,oneof=0 1"`
	SortOrder *int    `json:"sortOrder" binding:"required,min=0"`
}

type paramTemplateBody struct {
	Name       string                   `json:"name" binding:"required"`
	CategoryID string                   `json:"categoryId" binding:"required"`
	Status     enum.ParamTemplateStatus `json:"status" binding:"oneof=0 1"`
	Items      []paramTemplateItemDTO   `json:"items"`
}

func (body paramTemplateBody) toModel() *model.ParamTemplate {
	status := body.Status
	items := make([]model.ParamTemplateItem, 0, len(body.Items))
	for _, item := range body.Items {
		id := item.ID
		if id != nil {
			trimmed := strings.TrimSpace(*id)
			if trimmed == "" {
				id = nil
			} else {
				id = &trimmed
			}
		}
		items = append(items, model.ParamTemplateItem{
			ID:        id,
			FieldKey:  strings.TrimSpace(item.FieldKey),
			FieldName: strings.TrimSpace(item.FieldName),
			FieldType: item.FieldType,
			Unit:      item.Unit,
			Required:  item.Required,
			Display:   item.Display,
			SortOrder: item.SortOrder,
		})
	}
	return &model.ParamTemplate{
		Name:       strings.TrimSpace(body.Name),
		CategoryID: strings.TrimSpace(body.CategoryID),
		Status:     &status,
		Items:      items,
	}
}

func (p *ParamTemplates) List(c *gin.Context) {
	var query paramTemplateListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.FailParam(c, "参数异常")
		return
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
	result, err := p.ParamTemplateService.List(c.Request.Context(), param_template.ListQuery{
		Page:     query.Page,
		PageSize: query.PageSize,
	})
	if err != nil {
		response.FailBusiness(c, err.Error())
		return
	}
	response.Success(c, result)
}

func (p *ParamTemplates) Create(c *gin.Context) {
	var body paramTemplateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailParam(c, "参数错误,请检查参数是否符合要求: "+err.Error())
		return
	}
	paramTemplate, err := p.ParamTemplateService.Create(c.Request.Context(), body.toModel())
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
	var body paramTemplateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailParam(c, "参数错误,请检查参数是否符合要求: "+err.Error())
		return
	}
	result, err := p.ParamTemplateService.Update(c.Request.Context(), id, body.toModel())
	if err != nil {
		response.FailBusiness(c, "更新失败："+err.Error())
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
