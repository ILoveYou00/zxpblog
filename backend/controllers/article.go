package controllers

import (
	"blog/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ArticleController struct {
	DB *gorm.DB
}

func NewArticleController(db *gorm.DB) *ArticleController {
	return &ArticleController{DB: db}
}

// ArticleListItem 用于列表展示的简化结构（不含 content）
type ArticleListItem struct {
	ID            uint      `json:"id"`
	Title         string    `json:"title"`
	Slug          string    `json:"slug"`
	Summary       string    `json:"summary"`
	CoverImage    string    `json:"cover_image"`
	CategoryID    uint      `json:"category_id"`
	Category      Category  `json:"category"`
	Tags          string    `json:"tags"`
	ViewCount     int       `json:"view_count"`
	LikeCount     int       `json:"like_count"`
	IsPublished   bool      `json:"is_published"`
	IsPinned      bool      `json:"is_pinned"`
	ReadTime      int       `json:"read_time"`
	CreatedAt     time.Time `json:"created_at"`
}

type Category struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// GetArticles returns a list of published articles
func (ac *ArticleController) GetArticles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	categoryID := c.Query("category_id")
	search := c.Query("search")
	tagID := c.Query("tag_id")

	var articles []ArticleListItem
	var total int64

	query := ac.DB.Model(&models.Article{}).Where("is_published = ?", true)

	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}

	if search != "" {
		query = query.Where("title LIKE ?", "%"+search+"%")
	}

	// 按标签过滤
	if tagID != "" {
		var articleIDs []uint
		ac.DB.Model(&models.ArticleTag{}).Where("tag_id = ?", tagID).Pluck("article_id", &articleIDs)
		if len(articleIDs) > 0 {
			query = query.Where("id IN ?", articleIDs)
		} else {
			query = query.Where("1 = 0") // 无结果
		}
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	// 使用模型查询以正确处理软删除，同时排除 content 字段提升性能
	type ArticleListResult struct {
		ID          uint      `json:"id"`
		Title       string    `json:"title"`
		Slug        string    `json:"slug"`
		Summary     string    `json:"summary"`
		CoverImage  string    `json:"cover_image"`
		CategoryID  uint      `json:"category_id"`
		Category    Category  `json:"category"`
		Tags        string    `json:"tags"`
		ViewCount   int       `json:"view_count"`
		LikeCount   int       `json:"like_count"`
		IsPublished bool      `json:"is_published"`
		IsPinned    bool      `json:"is_pinned"`
		ReadTime    int       `json:"read_time"`
		CreatedAt   time.Time `json:"created_at"`
	}

	var articleResults []ArticleListResult
	result := query.
		Select("id, title, slug, summary, cover_image, category_id, tags, view_count, like_count, is_published, is_pinned, read_time, created_at").
		Preload("Category", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name, slug")
		}).
		Order("is_pinned DESC, created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&articleResults)

	// 转换为 ArticleListItem
	articles = make([]ArticleListItem, len(articleResults))
	for i, r := range articleResults {
		articles[i] = ArticleListItem{
			ID:          r.ID,
			Title:       r.Title,
			Slug:        r.Slug,
			Summary:     r.Summary,
			CoverImage:  r.CoverImage,
			CategoryID:  r.CategoryID,
			Category:    r.Category,
			Tags:        r.Tags,
			ViewCount:   r.ViewCount,
			LikeCount:   r.LikeCount,
			IsPublished: r.IsPublished,
			IsPinned:    r.IsPinned,
			ReadTime:    r.ReadTime,
			CreatedAt:   r.CreatedAt,
		}
	}

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": articles,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetArticle returns a single article
func (ac *ArticleController) GetArticle(c *gin.Context) {
	id := c.Param("id")

	var article models.Article
	result := ac.DB.Preload("Category").First(&article, id)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	// 异步更新浏览量，不阻塞响应
	go func() {
		ac.DB.Model(&models.Article{}).Where("id = ?", id).UpdateColumn("view_count", gorm.Expr("view_count + 1"))
	}()

	c.JSON(http.StatusOK, gin.H{"data": article})
}

// GetArticleBySlug returns article by slug
func (ac *ArticleController) GetArticleBySlug(c *gin.Context) {
	slug := c.Param("slug")

	var article models.Article
	result := ac.DB.Preload("Category").Where("slug = ? AND is_published = ?", slug, true).First(&article)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	// 异步更新浏览量
	go func() {
		ac.DB.Model(&models.Article{}).Where("id = ?", article.ID).UpdateColumn("view_count", gorm.Expr("view_count + 1"))
	}()

	c.JSON(http.StatusOK, gin.H{"data": article})
}

// CreateArticle creates a new article (admin only)
func (ac *ArticleController) CreateArticle(c *gin.Context) {
	var article models.Article
	if err := c.ShouldBindJSON(&article); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate slug if not provided
	if article.Slug == "" {
		article.Slug = generateSlug(article.Title)
	}

	article.CreatedAt = time.Now()
	article.UpdatedAt = time.Now()

	result := ac.DB.Create(&article)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": article})
}

// UpdateArticle updates an article (admin only)
func (ac *ArticleController) UpdateArticle(c *gin.Context) {
	id := c.Param("id")

	var article models.Article
	result := ac.DB.First(&article, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	var input models.Article
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	article.Title = input.Title
	article.Slug = input.Slug
	article.Content = input.Content
	article.ContentFormat = input.ContentFormat
	if article.ContentFormat == "" {
		article.ContentFormat = "html"
	}
	article.Summary = input.Summary
	article.CoverImage = input.CoverImage
	article.CategoryID = input.CategoryID
	article.Tags = input.Tags
	article.IsPublished = input.IsPublished
	article.IsPinned = input.IsPinned
	article.ReadTime = calculateReadTime(input.Content)
	article.UpdatedAt = time.Now()

	ac.DB.Save(&article)

	c.JSON(http.StatusOK, gin.H{"data": article})
}

// DeleteArticle deletes an article (admin only)
func (ac *ArticleController) DeleteArticle(c *gin.Context) {
	id := c.Param("id")

	// 开始事务，确保删除关联数据
	tx := ac.DB.Begin()

	// 删除关联数据
	tx.Where("article_id = ?", id).Delete(&models.Comment{})
	tx.Where("article_id = ?", id).Delete(&models.Like{})
	tx.Where("article_id = ?", id).Delete(&models.Favorite{})
	tx.Where("article_id = ?", id).Delete(&models.ReadHistory{})
	tx.Where("article_id = ?", id).Delete(&models.ArticleTag{})

	// 删除文章
	result := tx.Delete(&models.Article{}, id)
	if result.Error != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Article deleted"})
}

// GetAllArticles returns all articles including drafts (admin only)
func (ac *ArticleController) GetAllArticles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	var total int64
	ac.DB.Model(&models.Article{}).Count(&total)

	offset := (page - 1) * pageSize
	// 管理后台列表使用模型查询以正确处理软删除
	type ArticleListResult struct {
		ID          uint      `json:"id"`
		Title       string    `json:"title"`
		Slug        string    `json:"slug"`
		Summary     string    `json:"summary"`
		CoverImage  string    `json:"cover_image"`
		CategoryID  uint      `json:"category_id"`
		Category    Category  `json:"category"`
		Tags        string    `json:"tags"`
		ViewCount   int       `json:"view_count"`
		LikeCount   int       `json:"like_count"`
		IsPublished bool      `json:"is_published"`
		IsPinned    bool      `json:"is_pinned"`
		ReadTime    int       `json:"read_time"`
		CreatedAt   time.Time `json:"created_at"`
	}

	var articleResults []ArticleListResult
	result := ac.DB.Model(&models.Article{}).
		Select("id, title, slug, summary, cover_image, category_id, tags, view_count, like_count, is_published, is_pinned, read_time, created_at").
		Preload("Category", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name, slug")
		}).
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&articleResults)

	// 转换为 ArticleListItem
	var articles []ArticleListItem
	articles = make([]ArticleListItem, len(articleResults))
	for i, r := range articleResults {
		articles[i] = ArticleListItem{
			ID:          r.ID,
			Title:       r.Title,
			Slug:        r.Slug,
			Summary:     r.Summary,
			CoverImage:  r.CoverImage,
			CategoryID:  r.CategoryID,
			Category:    r.Category,
			Tags:        r.Tags,
			ViewCount:   r.ViewCount,
			LikeCount:   r.LikeCount,
			IsPublished: r.IsPublished,
			IsPinned:    r.IsPinned,
			ReadTime:    r.ReadTime,
			CreatedAt:   r.CreatedAt,
		}
	}

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": articles,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// Helper function to generate slug from title
func generateSlug(title string) string {
	// Simple slug generation - replace spaces with dashes
	slug := title
	// This is a simplified version; in production, you'd want more robust slug generation
	return slug
}

// GetSession returns current session info
func (ac *ArticleController) GetSession(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	if userID == nil {
		c.JSON(http.StatusOK, gin.H{"authenticated": false})
		return
	}

	var user models.User
	ac.DB.First(&user, userID)

	c.JSON(http.StatusOK, gin.H{
		"authenticated": true,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

// GetRelatedArticles 获取相关文章
func (ac *ArticleController) GetRelatedArticles(c *gin.Context) {
	id := c.Param("id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))

	var article models.Article
	if err := ac.DB.First(&article, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	// 获取同分类的文章
	var relatedArticles []models.Article
	ac.DB.Where("id != ? AND category_id = ? AND is_published = ?", id, article.CategoryID, true).
		Order("created_at DESC").
		Limit(limit).
		Preload("Category").
		Find(&relatedArticles)

	// 如果同分类文章不够，补充其他文章
	if len(relatedArticles) < limit {
		var moreArticles []models.Article
		excludeIDs := []uint{article.ID}
		for _, a := range relatedArticles {
			excludeIDs = append(excludeIDs, a.ID)
		}
		remaining := limit - len(relatedArticles)
		ac.DB.Where("id NOT IN ? AND is_published = ?", excludeIDs, true).
			Order("view_count DESC").
			Limit(remaining).
			Preload("Category").
			Find(&moreArticles)
		relatedArticles = append(relatedArticles, moreArticles...)
	}

	c.JSON(http.StatusOK, gin.H{"data": relatedArticles})
}

// calculateReadTime 计算阅读时长（分钟）
func calculateReadTime(content string) int {
	if content == "" {
		return 1
	}
	// 简单计算：直接统计字符数
	wordCount := len([]rune(content))
	// 假设每分钟阅读300字
	readTime := wordCount / 300
	if readTime < 1 {
		readTime = 1
	}
	return readTime
}