package handler

import (
	"errors"
	"strings"
	"vehivle/internal/domain/enum"
	"vehivle/internal/domain/model"
	"vehivle/internal/service/category"
	"vehivle/internal/transport/http/helper"
	"vehivle/internal/transport/http/constant"
	"vehivle/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)


type Categories struct {
	CategoryService *category.CategoryService
}

func NewCategories(svc *category.CategoryService) *Categories {
	return &Categories{CategoryService: svc}
}

// validateResolvedCategory 校验「已确定」的分类字段：状态、名称、排序、层级、二级父级。
// 创建时传入 body；更新时在合并 patch 与 existing 后传入最终 CategoryCreateInput。
func validateResolvedCategory(in model.CategoryCreateInput) error {
	// 校验状态
	if in.Status != enum.CategoryStatusDisabled && in.Status != enum.CategoryStatusEnabled {
		return errors.New("无效的状态，必须是0或1")
	}
	// 校验名称
	if in.Name == "" {
		return errors.New("名称不能为空")
	}
	// 校验排序
	if in.SortOrder <= 0 {
		return errors.New("排序不能小于等于0")
	}
	// 校验层级
	if in.Level != 1 && in.Level != 2 {
		return errors.New("层级必须是1或2")
	}
	// 校验二级父级
	if in.Level == 2 && (in.ParentID == nil || *in.ParentID == "") {
		return errors.New("二级分类必须选择父级")
	}
	return nil
}

// categoryListQueryParams 绑定 GET /admin/categories 的 query（keyword、level、status、page、pageSize、sortField、sortOrder 均可选）。
type categoryListQueryParams struct {
	Keyword   string `form:"keyword"`   // 分类名称
	Level     *int   `form:"level"`     // 分类层级
	Status    *int   `form:"status"`    // 分类状态
	Page      int    `form:"page"`      // 分页页码
	PageSize  int    `form:"pageSize"`  // 分页大小
	SortField string `form:"sortField"` // 排序字段
	SortOrder string `form:"sortOrder"` // 排序顺序
}

/**
 * 获取分类列表
 */
func (c *Categories) List(ctx *gin.Context) {
	var raw categoryListQueryParams
	// 使用ShouldBindQuery绑定请求体
	if err := ctx.ShouldBindQuery(&raw); err != nil {
		response.FailParam(ctx, err.Error())
		return
	}
	// 查询字段白名单: 只能categoryListQueryParams的字段通过
	// 从ctx中取字段，是否在validFields中，不在则返回错误
	// validFields := [7]string{"keyword", "level", "status", "page", "pageSize", "sortField", "sortOrder"}
	sf := strings.TrimSpace(raw.SortField)
	so := strings.TrimSpace(strings.ToLower(raw.SortOrder))
	_, err := helper.IsValidSortFieldAndOrder(sf, so)
	if err != nil {
		response.FailParam(ctx, err.Error())
		return
	}


	// 如果page小于1，则设置为1
	page := raw.Page
	if page <= 0 {
		page = constant.DEFAULT_CATEGORY_LIST_PAGE
	}
	// 如果pageSize小于1，则设置为DEFAULT_CATEGORY_LIST_PAGE_SIZE
	pageSize := raw.PageSize
	if pageSize <= 0 {
		pageSize = constant.DEFAULT_CATEGORY_LIST_PAGE_SIZE
	}
	// 如果pageSize大于MAX_CATEGORY_LIST_PAGE_SIZE，则设置为MAX_CATEGORY_LIST_PAGE_SIZE
	if pageSize > constant.MAX_CATEGORY_LIST_PAGE_SIZE {
		response.FailParam(ctx, "pageSize 不能大于100")
		return
	}
	q := model.CategoryListQuery{
		Keyword:   strings.TrimSpace(raw.Keyword),
		Level:     raw.Level,
		Page:      page,
		PageSize:  raw.PageSize,
		SortField: sf,
		SortOrder: so,
	}
	if raw.Status != nil {
		st := enum.CategoryStatus(*raw.Status)
		if st != enum.CategoryStatusDisabled && st != enum.CategoryStatusEnabled {
			response.FailParam(ctx, "无效的状态，必须是0或1")
			return
		}
		q.Status = &st
	}
	result, err := c.CategoryService.List(ctx.Request.Context(), q)
	if err != nil {
		response.FailBusiness(ctx, err.Error())
		return
	}
	// 为二级分类填充 parentName
	nameByID := make(map[string]string, len(result.List))
	for _, cat := range result.List {
		nameByID[cat.ID] = cat.Name
	}
	for _, cat := range result.List {
		if cat.ParentID != nil {
			if n, ok := nameByID[*cat.ParentID]; ok {
				cat.ParentName = n
			}
		}
	}
	response.Success(ctx, result)
}

/**
 * 创建分类
 */
func (c *Categories) Create(ctx *gin.Context) {
	// 绑定请求体
	var body model.CategoryCreateInput
	// 使用ShouldBindJSON绑定请求体
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.FailBusiness(ctx, err.Error())
		return
	}
	// 校验分类
	if err := validateResolvedCategory(body); err != nil {
		response.FailBusiness(ctx, err.Error())
		return
	}
	category, err := c.CategoryService.Create(ctx.Request.Context(), &model.Category{
		ID:                  uuid.New().String(),
		CategoryCreateInput: body,
	})
	if err != nil {
		response.FailBusiness(ctx, err.Error())
		return
	}
	response.Success(ctx, category)
}

// categoryUpdateBody 部分更新请求体；指针字段表示「本次是否提交该键」（类似 TS Partial + 仅发送变更字段）。
// FE analogy: Pick<Category, 'name' | ...> 里每个键都是可选的，未传的键保持数据库原值。
// Go detail: 用 *T 区分「未出现」与「出现」；JSON null 对 *string 会解成 nil，与「未传」在指针层面同为 nil，二级 parentId 清空需后续若需要可改用自定义类型。

/**
 * 更新分类（支持部分字段，如仅改 status）
 */
func (c *Categories) Update(ctx *gin.Context) {
	id := ctx.Param("category_id")
	// 按主键查询单条分类（管理端编辑/更新前拉取当前数据）。
	existing, err := c.CategoryService.GetById(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.FailNotFound(ctx, "分类不存在")
			return
		}
		response.FailBusiness(ctx, err.Error())
		return
	}
	// 部分更新请求体；指针字段表示「本次是否提交该键」（类似 TS Partial + 仅发送变更字段）。
	// FE analogy: Pick<Category, 'name' | ...> 里每个键都是可选的，未传的键保持数据库原值。
	// Go detail: 用 *T 区分「未出现」与「出现」；JSON null 对 *string 会解成 nil，与「未传」在指针层面同为 nil，二级 parentId 清空需后续若需要可改用自定义类型。
	var body model.CategoryUpdateBody
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.FailBusiness(ctx, err.Error())
		return
	}
	// 更新分类
	if body.Name != nil {
		existing.Name = *body.Name
	}
	if body.ParentID != nil {
		existing.ParentID = body.ParentID
	}
	if body.Level != nil {
		existing.Level = *body.Level
	}
	if body.Status != nil {
		existing.Status = *body.Status
	}
	if body.SortOrder != nil {
		existing.SortOrder = *body.SortOrder
	}

	if err := validateResolvedCategory(existing.CategoryCreateInput); err != nil {
		response.FailBusiness(ctx, err.Error())
		return
	}

	if err := c.CategoryService.Update(ctx.Request.Context(), existing); err != nil {
		response.FailBusiness(ctx, err.Error())
		return
	}
	updated, err := c.CategoryService.GetById(ctx.Request.Context(), id)
	if err != nil {
		response.FailBusiness(ctx, err.Error())
		return
	}
	response.Success(ctx, updated)
}

/**
 * 删除分类
 */
func (c *Categories) Delete(ctx *gin.Context) {
	id := ctx.Param("category_id")
	if id == "" {
		response.FailParam(ctx, "缺少 category_id")
		return
	}
	category, err := c.CategoryService.Delete(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.FailNotFound(ctx, "分类不存在")
			return
		}
		response.FailBusiness(ctx, err.Error())
		return
	}
	response.Success(ctx, category)
}
