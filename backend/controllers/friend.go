package controllers

import (
	"blog/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FriendLinkController struct {
	DB *gorm.DB
}

func NewFriendLinkController(db *gorm.DB) *FriendLinkController {
	return &FriendLinkController{DB: db}
}

// GetFriendLinks 获取友情链接列表
func (fc *FriendLinkController) GetFriendLinks(c *gin.Context) {
	var links []models.FriendLink
	fc.DB.Where("is_active = ?", true).Order("sort_order ASC, created_at DESC").Find(&links)

	c.JSON(http.StatusOK, gin.H{"data": links})
}

// GetAllFriendLinks 获取所有友链 (admin)
func (fc *FriendLinkController) GetAllFriendLinks(c *gin.Context) {
	var links []models.FriendLink
	fc.DB.Order("sort_order ASC, created_at DESC").Find(&links)

	c.JSON(http.StatusOK, gin.H{"data": links})
}

// CreateFriendLink 创建友链 (admin)
func (fc *FriendLinkController) CreateFriendLink(c *gin.Context) {
	var link models.FriendLink
	if err := c.ShouldBindJSON(&link); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	link.CreatedAt = time.Now()
	result := fc.DB.Create(&link)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": link})
}

// UpdateFriendLink 更新友链 (admin)
func (fc *FriendLinkController) UpdateFriendLink(c *gin.Context) {
	id := c.Param("id")

	var link models.FriendLink
	if err := fc.DB.First(&link, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Friend link not found"})
		return
	}

	var input models.FriendLink
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	link.Name = input.Name
	link.URL = input.URL
	link.Logo = input.Logo
	link.Desc = input.Desc
	link.SortOrder = input.SortOrder
	link.IsActive = input.IsActive
	fc.DB.Save(&link)

	c.JSON(http.StatusOK, gin.H{"data": link})
}

// DeleteFriendLink 删除友链 (admin)
func (fc *FriendLinkController) DeleteFriendLink(c *gin.Context) {
	id := c.Param("id")

	result := fc.DB.Delete(&models.FriendLink{}, id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Friend link deleted"})
}

// ToggleFriendLink 切换友链状态 (admin)
func (fc *FriendLinkController) ToggleFriendLink(c *gin.Context) {
	id := c.Param("id")

	var link models.FriendLink
	if err := fc.DB.First(&link, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Friend link not found"})
		return
	}

	link.IsActive = !link.IsActive
	fc.DB.Save(&link)

	c.JSON(http.StatusOK, gin.H{"data": link})
}

// ParseUint helper
func ParseUint(s string) uint {
	result, _ := strconv.ParseUint(s, 10, 32)
	return uint(result)
}