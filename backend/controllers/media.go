package controllers

import (
	"blog/models"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MediaController struct {
	DB      *gorm.DB
	UploadDir string
}

func NewMediaController(db *gorm.DB) *MediaController {
	uploadDir := "/app/uploads"
	// 确保上传目录存在
	os.MkdirAll(uploadDir, 0755)
	return &MediaController{DB: db, UploadDir: uploadDir}
}

// UploadMedia 上传媒体文件
func (mc *MediaController) UploadMedia(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	// 检查文件大小 (最大 10MB)
	if header.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File too large (max 10MB)"})
		return
	}

	// 检查文件类型
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
		".webp": true, ".svg": true, ".ico": true,
	}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type"})
		return
	}

	// 生成文件名
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	dateDir := time.Now().Format("2006/01")
	fullDir := filepath.Join(mc.UploadDir, dateDir)
	os.MkdirAll(fullDir, 0755)

	// 保存文件
	filePath := filepath.Join(fullDir, filename)
	dst, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// 获取 MIME 类型
	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// 保存到数据库
	url := fmt.Sprintf("/uploads/%s/%s", dateDir, filename)
	media := models.Media{
		Filename:  header.Filename,
		URL:       url,
		Size:      header.Size,
		MimeType:  mimeType,
		CreatedAt: time.Now(),
	}
	mc.DB.Create(&media)

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"id":       media.ID,
			"filename": media.Filename,
			"url":      media.URL,
			"size":     media.Size,
		},
	})
}

// GetMediaList 获取媒体列表
func (mc *MediaController) GetMediaList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var media []models.Media
	var total int64

	mc.DB.Model(&models.Media{}).Count(&total)

	offset := (page - 1) * pageSize
	mc.DB.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&media)

	c.JSON(http.StatusOK, gin.H{
		"data": media,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// DeleteMedia 删除媒体
func (mc *MediaController) DeleteMedia(c *gin.Context) {
	id := c.Param("id")

	var media models.Media
	if err := mc.DB.First(&media, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Media not found"})
		return
	}

	// 删除文件
	filePath := filepath.Join(mc.UploadDir, strings.TrimPrefix(media.URL, "/uploads/"))
	os.Remove(filePath)

	// 删除数据库记录
	mc.DB.Delete(&media)

	c.JSON(http.StatusOK, gin.H{"message": "Media deleted"})
}

// GetMediaByID 获取单个媒体
func (mc *MediaController) GetMediaByID(c *gin.Context) {
	id := c.Param("id")

	var media models.Media
	if err := mc.DB.First(&media, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Media not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": media})
}