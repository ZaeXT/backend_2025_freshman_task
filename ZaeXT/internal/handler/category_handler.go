package handler

import (
	"ai-qa-backend/internal/handler/request"
	"ai-qa-backend/internal/handler/response"
	"ai-qa-backend/internal/model"
	"ai-qa-backend/internal/pkg/e"
	"ai-qa-backend/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	categoryService service.CategoryService
}

func NewCategoryHandler(categoryService service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService}
}

func (h *CategoryHandler) Create(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req request.CreateCategory
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, e.InvalidParams, err.Error())
		return
	}

	category, err := h.categoryService.Create(userID.(uint), req.Name, req.ParentID)
	if err != nil {
		response.Fail(c, e.Error, "创建分类失败")
		return
	}

	res := &response.CategoryInfo{
		ID:       category.ID,
		Name:     category.Name,
		ParentID: category.ParentID,
	}

	response.Success(c, res)
}

func (h *CategoryHandler) List(c *gin.Context) {
	userID, _ := c.Get("userID")

	categories, err := h.categoryService.List(userID.(uint))
	if err != nil {
		response.Fail(c, e.Error, "获取分类列表失败")
		return
	}

	responseCategories := transformCategoriesToDTO(categories)

	response.Success(c, responseCategories)
}

func (h *CategoryHandler) Update(c *gin.Context) {
	userID, _ := c.Get("userID")
	categoryID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, e.InvalidParams, "无效的分类ID")
		return
	}

	var req request.UpdateCategory
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, e.InvalidParams, err.Error())
		return
	}

	if err := h.categoryService.Update(uint(categoryID), userID.(uint), req.Name, req.ParentID); err != nil {
		response.Fail(c, e.Error, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	userID, _ := c.Get("userID")
	categoryID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, e.InvalidParams, "无效的分类ID")
		return
	}

	if err := h.categoryService.Delete(uint(categoryID), userID.(uint)); err != nil {
		response.Fail(c, e.Error, err.Error())
		return
	}

	response.Success(c, nil)
}

func transformCategoriesToDTO(categories []*model.Category) []*response.CategoryInfo {
	if categories == nil {
		return nil
	}
	dtos := make([]*response.CategoryInfo, len(categories))
	for i, category := range categories {
		dtos[i] = &response.CategoryInfo{
			ID:       category.ID,
			Name:     category.Name,
			ParentID: category.ParentID,
			Children: transformCategoriesToDTO(category.Children),
		}
	}
	return dtos
}
