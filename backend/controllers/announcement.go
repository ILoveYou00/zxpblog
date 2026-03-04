package controllers

import (
	"blog/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AnnouncementController struct {
	DB *gorm.DB
}

func NewAnnouncementController(db *gorm.DB) *AnnouncementController {
	return &AnnouncementController{DB: db}
}

// GetAnnouncements 获取公告列表（公开）
func (ac *AnnouncementController) GetAnnouncements(c *gin.Context) {
	var announcements []models.Announcement
	now := time.Now()

	// 获取当前有效的公告
	ac.DB.Where("is_active = ? AND start_time <= ? AND end_time >= ?", true, now, now).
		Order("created_at DESC").
		Find(&announcements)

	c.JSON(http.StatusOK, gin.H{"data": announcements})
}

// GetAllAnnouncements 获取所有公告 (admin)
func (ac *AnnouncementController) GetAllAnnouncements(c *gin.Context) {
	var announcements []models.Announcement
	ac.DB.Order("created_at DESC").Find(&announcements)

	c.JSON(http.StatusOK, gin.H{"data": announcements})
}

// CreateAnnouncement 创建公告 (admin)
func (ac *AnnouncementController) CreateAnnouncement(c *gin.Context) {
	var announcement models.Announcement
	if err := c.ShouldBindJSON(&announcement); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	announcement.CreatedAt = time.Now()
	result := ac.DB.Create(&announcement)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": announcement})
}

// UpdateAnnouncement 更新公告 (admin)
func (ac *AnnouncementController) UpdateAnnouncement(c *gin.Context) {
	id := c.Param("id")

	var announcement models.Announcement
	if err := ac.DB.First(&announcement, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Announcement not found"})
		return
	}

	var input models.Announcement
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	announcement.Title = input.Title
	announcement.Content = input.Content
	announcement.IsActive = input.IsActive
	announcement.StartTime = input.StartTime
	announcement.EndTime = input.EndTime
	ac.DB.Save(&announcement)

	c.JSON(http.StatusOK, gin.H{"data": announcement})
}

// DeleteAnnouncement 删除公告 (admin)
func (ac *AnnouncementController) DeleteAnnouncement(c *gin.Context) {
	id := c.Param("id")

	result := ac.DB.Delete(&models.Announcement{}, id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Announcement deleted"})
}

// ToggleAnnouncement 切换公告状态 (admin)
func (ac *AnnouncementController) ToggleAnnouncement(c *gin.Context) {
	id := c.Param("id")

	var announcement models.Announcement
	if err := ac.DB.First(&announcement, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Announcement not found"})
		return
	}

	announcement.IsActive = !announcement.IsActive
	ac.DB.Save(&announcement)

	c.JSON(http.StatusOK, gin.H{"data": announcement})
}