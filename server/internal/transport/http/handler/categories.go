package handler

import (
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

/**
 * 获取分类列表
 */
func (c *Categories) List(ctx *gin.Context) {
	items, err := c.CategoryService.List(ctx.Request.Context())
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
	// 校验状态
	if body.Status != enum.CategoryStatusDisabled && body.Status != enum.CategoryStatusEnabled {
		response.FailBusiness(ctx, "无效的状态，必须是0或1")
		return
	}
	// 校验名称
	if body.Name == "" {
		response.FailBusiness(ctx, "名称不能为空")
		return
	}
	// 校验排序
	if body.SortOrder <= 0 {
		response.FailBusiness(ctx, "排序不能小于等于0")
		return
	}
	// 校验层级
	if body.Level != 1 && body.Level != 2 {
		response.FailBusiness(ctx, "层级必须是1或2")
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

// /**
//  * 更新分类
//  */
// func (c *Categories) Update(ctx *gin.Context) {
// 	category, err := c.CategoryService.Update(ctx.Request.Context(), &model.Category{
// 		ID: ctx.Param("id"),
// 		Name: ctx.PostForm("name"),
// 		ParentID: strPtr(ctx.PostForm("parent_id")),
// 		Level: ctx.PostForm("level"),
// 		Status: ctx.PostForm("status"),
// 		SortOrder: ctx.PostForm("sort_order"),
// 	})
// 	if err != nil {
// 		response.FailBusiness(ctx, err.Error())
// 		return
// 	}
// 	response.Success(ctx, gin.H{"item": category})
// }

/**
 * 删除分类
 */
func (c *Categories) Delete(ctx *gin.Context) {
	category, err := c.CategoryService.Delete(ctx.Request.Context(), ctx.Param("category_id"))
	if err != nil {
		response.FailBusiness(ctx, err.Error())
		return
	}
	response.Success(ctx, category)
}