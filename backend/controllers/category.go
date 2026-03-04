package controllers

import (
	"blog/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CategoryController struct {
	DB *gorm.DB
}

func NewCategoryController(db *gorm.DB) *CategoryController {
	return &CategoryController{DB: db}
}

// GetCategories returns all categories
func (cc *CategoryController) GetCategories(c *gin.Context) {
	var categories []models.Category
	result := cc.DB.Find(&categories)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": categories})
}

// CreateCategory creates a new category (admin only)
func (cc *CategoryController) CreateCategory(c *gin.Context) {
	var category models.Category
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := cc.DB.Create(&category)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": category})
}

// UpdateCategory updates a category (admin only)
func (cc *CategoryController) UpdateCategory(c *gin.Context) {
	id := c.Param("id")

	var category models.Category
	result := cc.DB.First(&category, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	var input models.Category
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category.Name = input.Name
	category.Slug = input.Slug

	cc.DB.Save(&category)

	c.JSON(http.StatusOK, gin.H{"data": category})
}

// DeleteCategory deletes a category (admin only)
func (cc *CategoryController) DeleteCategory(c *gin.Context) {
	id := c.Param("id")

	result := cc.DB.Delete(&models.Category{}, id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category deleted"})
}