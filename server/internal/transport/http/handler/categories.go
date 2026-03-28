package handler

import (
	"errors"
	"strings"
	"vehivle/internal/domain/enum"
	"vehivle/internal/domain/model"
	"vehivle/internal/repository/postgres"
	"vehivle/internal/service/category"
	"vehivle/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Categories struct {
	DB              *gorm.DB                  // 数据库连接
	CategoryService *category.CategoryService // 分类服务
}

/**
 * 创建分类处理器
 */
func NewCategories(db *gorm.DB) *Categories {
	// 创建分类仓库实例
	repo := postgres.NewCategoryRepo(db)
	// 创建分类服务实例
	return &Categories{
		DB:              db,
		CategoryService: category.NewCategoryService(repo),
	}
}

// validateResolvedCategory 校验「已确定」的分类字段：状态、名称、排序、层级、二级父级。
// 创建时传入 body；更新时在合并 patch 与 existing 后传入最终 CategoryCreateInput。
func validateResolvedCategory(in model.CategoryCreateInput) error {
	if in.Status != enum.CategoryStatusDisabled && in.Status != enum.CategoryStatusEnabled {
		return errors.New("无效的状态，必须是0或1")
	}
	if in.Name == "" {
		return errors.New("名称不能为空")
	}
	if in.SortOrder <= 0 {
		return errors.New("排序不能小于等于0")
	}
	if in.Level != 1 && in.Level != 2 {
		return errors.New("层级必须是1或2")
	}
	if in.Level == 2 && (in.ParentID == nil || *in.ParentID == "") {
		return errors.New("二级分类必须选择父级")
	}
	return nil
}

// categoryListQueryParams 绑定 GET /admin/categories 的 query（keyword、level、status 均可选）。
type categoryListQueryParams struct {
	Keyword string `form:"keyword"`
	Level   *int   `form:"level"`
	Status  *int   `form:"status"`
}

/**
 * 获取分类列表
 */
func (c *Categories) List(ctx *gin.Context) {
	var raw categoryListQueryParams
	if err := ctx.ShouldBindQuery(&raw); err != nil {
		response.FailParam(ctx, err.Error())
		return
	}
	q := model.CategoryListQuery{Keyword: strings.TrimSpace(raw.Keyword), Level: raw.Level}
	if raw.Status != nil {
		st := enum.CategoryStatus(*raw.Status)
		if st != enum.CategoryStatusDisabled && st != enum.CategoryStatusEnabled {
			response.FailParam(ctx, "无效的状态，必须是0或1")
			return
		}
		q.Status = &st
	}
	items, err := c.CategoryService.List(ctx.Request.Context(), q)
	if err != nil {
		response.FailBusiness(ctx, err.Error())
		return
	}
	// 为二级分类填充 parentName
	nameByID := make(map[string]string, len(items))
	for _, cat := range items {
		nameByID[cat.ID] = cat.Name
	}
	for _, cat := range items {
		if cat.ParentID != nil {
			if n, ok := nameByID[*cat.ParentID]; ok {
				cat.ParentName = n
			}
		}
	}
	response.Success(ctx, items)
}

/**
 * 创建分类
 */
func (c *Categories) Create(ctx *gin.Context) {
	var body model.CategoryCreateInput
	// 使用ShouldBindJSON绑定请求体
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.FailBusiness(ctx, err.Error())
		return
	}
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
	// 获取分类
	existing, err := c.CategoryService.GetById(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.FailNotFound(ctx, "分类不存在")
			return
		}
		response.FailBusiness(ctx, err.Error())
		return
	}
	// 绑定请求体
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
