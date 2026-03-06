package controllers

import (
	"blog/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HtmlPageController struct {
	DB *gorm.DB
}

func NewHtmlPageController(db *gorm.DB) *HtmlPageController {
	return &HtmlPageController{DB: db}
}

// GetHtmlPages returns a list of published HTML pages
func (h *HtmlPageController) GetHtmlPages(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	var htmlPages []models.HtmlPage
	var total int64

	query := h.DB.Model(&models.HtmlPage{}).Where("is_published = ?", true)

	query.Count(&total)

	offset := (page - 1) * pageSize
	result := query.
		Preload("Category").
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&htmlPages)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": htmlPages,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetHtmlPage returns a single HTML page
func (h *HtmlPageController) GetHtmlPage(c *gin.Context) {
	id := c.Param("id")

	var htmlPage models.HtmlPage
	result := h.DB.Preload("Category").First(&htmlPage, id)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "HTML page not found"})
		return
	}

	// 异步更新浏览量
	go func() {
		h.DB.Model(&models.HtmlPage{}).Where("id = ?", id).UpdateColumn("view_count", gorm.Expr("view_count + 1"))
	}()

	c.JSON(http.StatusOK, gin.H{"data": htmlPage})
}

// GetAllHtmlPages returns all HTML pages including unpublished (admin only)
func (h *HtmlPageController) GetAllHtmlPages(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	var total int64
	h.DB.Model(&models.HtmlPage{}).Count(&total)

	offset := (page - 1) * pageSize

	var htmlPages []models.HtmlPage
	result := h.DB.Preload("Category").Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&htmlPages)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": htmlPages,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// CreateHtmlPage creates a new HTML page (admin only)
func (h *HtmlPageController) CreateHtmlPage(c *gin.Context) {
	var htmlPage models.HtmlPage
	if err := c.ShouldBindJSON(&htmlPage); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate slug if not provided
	if htmlPage.Slug == "" {
		htmlPage.Slug = generateHtmlSlug(htmlPage.Title)
	}

	htmlPage.CreatedAt = time.Now()
	htmlPage.UpdatedAt = time.Now()

	result := h.DB.Create(&htmlPage)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": htmlPage})
}

// UpdateHtmlPage updates an HTML page (admin only)
func (h *HtmlPageController) UpdateHtmlPage(c *gin.Context) {
	id := c.Param("id")

	var htmlPage models.HtmlPage
	result := h.DB.First(&htmlPage, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "HTML page not found"})
		return
	}

	var input models.HtmlPage
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	htmlPage.Title = input.Title
	htmlPage.Slug = input.Slug
	htmlPage.Content = input.Content
	htmlPage.Summary = input.Summary
	htmlPage.CoverImage = input.CoverImage
	htmlPage.CategoryID = input.CategoryID
	htmlPage.Tags = input.Tags
	htmlPage.IsPublished = input.IsPublished
	htmlPage.UpdatedAt = time.Now()

	h.DB.Save(&htmlPage)

	c.JSON(http.StatusOK, gin.H{"data": htmlPage})
}

// DeleteHtmlPage deletes an HTML page (admin only)
func (h *HtmlPageController) DeleteHtmlPage(c *gin.Context) {
	id := c.Param("id")

	result := h.DB.Delete(&models.HtmlPage{}, id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "HTML page deleted"})
}

// Helper function to generate slug from title
func generateHtmlSlug(title string) string {
	// Simple slug generation - replace spaces with dashes
	slug := title
	return slug
}