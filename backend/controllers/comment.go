package controllers

import (
	"blog/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CommentController struct {
	DB *gorm.DB
}

func NewCommentController(db *gorm.DB) *CommentController {
	return &CommentController{DB: db}
}

// GetComments returns approved comments for an article
func (cc *CommentController) GetComments(c *gin.Context) {
	articleID := c.Query("article_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var comments []models.Comment
	var total int64

	query := cc.DB.Model(&models.Comment{}).Where("is_approved = ? AND parent_id IS NULL", true)
	if articleID != "" {
		query = query.Where("article_id = ?", articleID)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	result := query.Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Preload("Replies", "is_approved = ?", true).
		Find(&comments)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": comments,
		"pagination": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
	})
}

// CreateComment creates a new comment
func (cc *CommentController) CreateComment(c *gin.Context) {
	var comment models.Comment
	if err := c.ShouldBindJSON(&comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Auto-approve comments (you can change this to require admin approval)
	comment.IsApproved = true

	result := cc.DB.Create(&comment)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": comment})
}

// ReplyComment 回复评论
func (cc *CommentController) ReplyComment(c *gin.Context) {
	parentID := c.Param("id")

	var parent models.Comment
	if err := cc.DB.First(&parent, parentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Parent comment not found"})
		return
	}

	var input struct {
		Nickname string `json:"nickname" binding:"required"`
		Email    string `json:"email"`
		Content  string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	parentIDUint := parseUintID(parentID)
	comment := models.Comment{
		ArticleID:  parent.ArticleID,
		ParentID:   &parentIDUint,
		Nickname:   input.Nickname,
		Email:      input.Email,
		Content:    input.Content,
		IsApproved: true,
	}

	result := cc.DB.Create(&comment)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": comment})
}

func parseUintID(s string) uint {
	var result uint
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + uint(c-'0')
		}
	}
	return result
}

// GetAllComments returns all comments including unapproved (admin only)
func (cc *CommentController) GetAllComments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var comments []models.Comment
	var total int64

	cc.DB.Model(&models.Comment{}).Count(&total)

	offset := (page - 1) * pageSize
	result := cc.DB.Preload("Article").Order("created_at DESC").
		Offset(offset).Limit(pageSize).Find(&comments)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": comments,
		"pagination": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
	})
}

// ApproveComment approves a comment (admin only)
func (cc *CommentController) ApproveComment(c *gin.Context) {
	id := c.Param("id")

	var comment models.Comment
	result := cc.DB.First(&comment, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		return
	}

	comment.IsApproved = true
	cc.DB.Save(&comment)

	c.JSON(http.StatusOK, gin.H{"data": comment})
}

// DeleteComment deletes a comment (admin only)
func (cc *CommentController) DeleteComment(c *gin.Context) {
	id := c.Param("id")

	result := cc.DB.Delete(&models.Comment{}, id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment deleted"})
}