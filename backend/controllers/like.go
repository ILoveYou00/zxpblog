package controllers

import (
	"blog/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type LikeController struct {
	DB *gorm.DB
}

func NewLikeController(db *gorm.DB) *LikeController {
	return &LikeController{DB: db}
}

// LikeArticle 点赞文章
func (lc *LikeController) LikeArticle(c *gin.Context) {
	id := c.Param("id")
	ip := c.ClientIP()

	var article models.Article
	if err := lc.DB.First(&article, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	// 检查是否已点赞
	var existingLike models.Like
	result := lc.DB.Where("article_id = ? AND ip = ?", id, ip).First(&existingLike)
	if result.Error == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Already liked"})
		return
	}

	// 创建点赞记录
	like := models.Like{
		ArticleID: article.ID,
		IP:        ip,
		CreatedAt: time.Now(),
	}
	lc.DB.Create(&like)

	// 更新文章点赞数 - 使用原子操作
	lc.DB.Model(&models.Article{}).Where("id = ?", id).UpdateColumn("like_count", gorm.Expr("like_count + 1"))

	// 获取更新后的点赞数
	lc.DB.First(&article, id)

	c.JSON(http.StatusOK, gin.H{
		"message":    "Liked successfully",
		"like_count": article.LikeCount,
	})
}

// UnlikeArticle 取消点赞
func (lc *LikeController) UnlikeArticle(c *gin.Context) {
	id := c.Param("id")
	ip := c.ClientIP()

	var article models.Article
	if err := lc.DB.First(&article, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	// 删除点赞记录
	result := lc.DB.Where("article_id = ? AND ip = ?", id, ip).Delete(&models.Like{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not liked yet"})
		return
	}

	// 更新文章点赞数 - 使用原子操作
	lc.DB.Model(&models.Article{}).Where("id = ?", id).UpdateColumn("like_count", gorm.Expr("GREATEST(like_count - 1, 0)"))

	// 获取更新后的点赞数
	lc.DB.First(&article, id)

	c.JSON(http.StatusOK, gin.H{
		"message":    "Unliked successfully",
		"like_count": article.LikeCount,
	})
}

// CheckLikeStatus 检查是否已点赞
func (lc *LikeController) CheckLikeStatus(c *gin.Context) {
	id := c.Param("id")
	ip := c.ClientIP()

	var like models.Like
	result := lc.DB.Where("article_id = ? AND ip = ?", id, ip).First(&like)

	c.JSON(http.StatusOK, gin.H{
		"liked": result.Error == nil,
	})
}

// FavoriteController 收藏控制器
type FavoriteController struct {
	DB *gorm.DB
}

func NewFavoriteController(db *gorm.DB) *FavoriteController {
	return &FavoriteController{DB: db}
}

// AddFavorite 添加收藏
func (fc *FavoriteController) AddFavorite(c *gin.Context) {
	id := c.Param("id")
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = c.Query("session_id")
	}
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID required"})
		return
	}

	var article models.Article
	if err := fc.DB.First(&article, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	// 检查是否已收藏
	var existing models.Favorite
	result := fc.DB.Where("article_id = ? AND session_id = ?", id, sessionID).First(&existing)
	if result.Error == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Already favorited"})
		return
	}

	favorite := models.Favorite{
		ArticleID: article.ID,
		SessionID: sessionID,
		CreatedAt: time.Now(),
	}
	fc.DB.Create(&favorite)

	c.JSON(http.StatusOK, gin.H{"message": "Favorited successfully"})
}

// RemoveFavorite 取消收藏
func (fc *FavoriteController) RemoveFavorite(c *gin.Context) {
	id := c.Param("id")
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = c.Query("session_id")
	}
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID required"})
		return
	}

	result := fc.DB.Where("article_id = ? AND session_id = ?", id, sessionID).Delete(&models.Favorite{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not favorited yet"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Removed from favorites"})
}

// GetFavorites 获取收藏列表
func (fc *FavoriteController) GetFavorites(c *gin.Context) {
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = c.Query("session_id")
	}
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID required"})
		return
	}

	var favorites []models.Favorite
	fc.DB.Where("session_id = ?", sessionID).Preload("Article.Category").Order("created_at DESC").Find(&favorites)

	c.JSON(http.StatusOK, gin.H{"data": favorites})
}

// CheckFavoriteStatus 检查是否已收藏
func (fc *FavoriteController) CheckFavoriteStatus(c *gin.Context) {
	id := c.Param("id")
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = c.Query("session_id")
	}

	var favorite models.Favorite
	result := fc.DB.Where("article_id = ? AND session_id = ?", id, sessionID).First(&favorite)

	c.JSON(http.StatusOK, gin.H{
		"favorited": result.Error == nil,
	})
}

// HistoryController 阅读历史控制器
type HistoryController struct {
	DB *gorm.DB
}

func NewHistoryController(db *gorm.DB) *HistoryController {
	return &HistoryController{DB: db}
}

// AddHistory 添加阅读历史
func (hc *HistoryController) AddHistory(c *gin.Context) {
	id := c.Param("id")
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = c.Query("session_id")
	}
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID required"})
		return
	}

	var article models.Article
	if err := hc.DB.First(&article, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	// 删除旧记录（如果存在）
	hc.DB.Where("article_id = ? AND session_id = ?", id, sessionID).Delete(&models.ReadHistory{})

	// 创建新记录
	history := models.ReadHistory{
		ArticleID: article.ID,
		SessionID: sessionID,
		CreatedAt: time.Now(),
	}
	hc.DB.Create(&history)

	c.JSON(http.StatusOK, gin.H{"message": "History recorded"})
}

// GetHistory 获取阅读历史
func (hc *HistoryController) GetHistory(c *gin.Context) {
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = c.Query("session_id")
	}
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID required"})
		return
	}

	limit := 20
	var histories []models.ReadHistory
	hc.DB.Where("session_id = ?", sessionID).
		Preload("Article.Category").
		Order("created_at DESC").
		Limit(limit).
		Find(&histories)

	c.JSON(http.StatusOK, gin.H{"data": histories})
}

// ClearHistory 清除阅读历史
func (hc *HistoryController) ClearHistory(c *gin.Context) {
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = c.Query("session_id")
	}
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID required"})
		return
	}

	hc.DB.Where("session_id = ?", sessionID).Delete(&models.ReadHistory{})

	c.JSON(http.StatusOK, gin.H{"message": "History cleared"})
}