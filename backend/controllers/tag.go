package controllers

import (
	"blog/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TagController struct {
	DB *gorm.DB
}

func NewTagController(db *gorm.DB) *TagController {
	return &TagController{DB: db}
}

// GetTags 获取标签列表
func (tc *TagController) GetTags(c *gin.Context) {
	// 获取每个标签的文章数
	type TagWithCount struct {
		models.Tag
		ArticleCount int `json:"article_count"`
	}

	var tagsWithCount []TagWithCount

	tc.DB.Model(&models.Tag{}).
		Select("tags.*, COUNT(article_tags.article_id) as article_count").
		Joins("LEFT JOIN article_tags ON article_tags.tag_id = tags.id").
		Group("tags.id").
		Order("article_count DESC, tags.name ASC").
		Find(&tagsWithCount)

	c.JSON(http.StatusOK, gin.H{"data": tagsWithCount})
}

// GetTagArticles 获取标签下的文章
func (tc *TagController) GetTagArticles(c *gin.Context) {
	tagID := c.Param("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	var tag models.Tag
	if err := tc.DB.First(&tag, tagID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tag not found"})
		return
	}

	var articleIDs []uint
	tc.DB.Model(&models.ArticleTag{}).Where("tag_id = ?", tagID).Pluck("article_id", &articleIDs)

	if len(articleIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"data": []interface{}{},
			"tag":  tag,
			"pagination": gin.H{
				"page":       page,
				"page_size":  pageSize,
				"total":      0,
				"total_page": 0,
			},
		})
		return
	}

	var articles []models.Article
	var total int64

	query := tc.DB.Model(&models.Article{}).Where("id IN ? AND is_published = ?", articleIDs, true)
	query.Count(&total)

	offset := (page - 1) * pageSize
	tc.DB.Preload("Category").Where("id IN ? AND is_published = ?", articleIDs, true).
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&articles)

	c.JSON(http.StatusOK, gin.H{
		"data": articles,
		"tag":  tag,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// CreateTag 创建标签 (admin)
func (tc *TagController) CreateTag(c *gin.Context) {
	var tag models.Tag
	if err := c.ShouldBindJSON(&tag); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tag.CreatedAt = time.Now()
	result := tc.DB.Create(&tag)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": tag})
}

// UpdateTag 更新标签 (admin)
func (tc *TagController) UpdateTag(c *gin.Context) {
	id := c.Param("id")

	var tag models.Tag
	if err := tc.DB.First(&tag, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tag not found"})
		return
	}

	var input models.Tag
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tag.Name = input.Name
	tag.Slug = input.Slug
	tc.DB.Save(&tag)

	c.JSON(http.StatusOK, gin.H{"data": tag})
}

// DeleteTag 删除标签 (admin)
func (tc *TagController) DeleteTag(c *gin.Context) {
	id := c.Param("id")

	// 删除关联
	tc.DB.Where("tag_id = ?", id).Delete(&models.ArticleTag{})

	result := tc.DB.Delete(&models.Tag{}, id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tag deleted"})
}

// SetArticleTags 设置文章标签 (admin)
func (tc *TagController) SetArticleTags(c *gin.Context) {
	articleID := c.Param("id")

	var input struct {
		TagIDs []uint `json:"tag_ids"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 删除旧关联
	tc.DB.Where("article_id = ?", articleID).Delete(&models.ArticleTag{})

	// 创建新关联
	for _, tagID := range input.TagIDs {
		articleTag := models.ArticleTag{
			ArticleID: parseUint(articleID),
			TagID:     tagID,
		}
		tc.DB.Create(&articleTag)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tags updated"})
}

// GetArticleTags 获取文章的标签
func (tc *TagController) GetArticleTags(c *gin.Context) {
	articleID := c.Param("id")

	var tagIDs []uint
	tc.DB.Model(&models.ArticleTag{}).Where("article_id = ?", articleID).Pluck("tag_id", &tagIDs)

	var tags []models.Tag
	if len(tagIDs) > 0 {
		tc.DB.Where("id IN ?", tagIDs).Find(&tags)
	}

	c.JSON(http.StatusOK, gin.H{"data": tags})
}

func parseUint(s string) uint {
	var result uint
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + uint(c-'0')
		}
	}
	return result
}