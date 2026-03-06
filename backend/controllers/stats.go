package controllers

import (
	"blog/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type StatsController struct {
	DB *gorm.DB
}

func NewStatsController(db *gorm.DB) *StatsController {
	return &StatsController{DB: db}
}

// GetStats 获取站点统计
func (sc *StatsController) GetStats(c *gin.Context) {
	var articleCount, publishedCount, commentCount, viewCount, likeCount int64

	sc.DB.Model(&models.Article{}).Count(&articleCount)
	sc.DB.Model(&models.Article{}).Where("is_published = ?", true).Count(&publishedCount)
	sc.DB.Model(&models.Comment{}).Where("is_approved = ?", true).Count(&commentCount)

	// 总浏览量
	sc.DB.Model(&models.Article{}).Select("COALESCE(SUM(view_count), 0)").Scan(&viewCount)
	sc.DB.Model(&models.Article{}).Select("COALESCE(SUM(like_count), 0)").Scan(&likeCount)

	// 获取今日访问量
	today := time.Now().Format("2006-01-02")
	var todayStats models.SiteStat
	sc.DB.Where("date = ?", today).First(&todayStats)

	// 热门文章
	var hotArticles []models.Article
	sc.DB.Where("is_published = ?", true).Order("view_count DESC").Limit(5).Find(&hotArticles)

	// 最新文章
	var latestArticles []models.Article
	sc.DB.Where("is_published = ?", true).Order("created_at DESC").Limit(5).Find(&latestArticles)

	// 最近7天访问趋势
	var weeklyStats []models.SiteStat
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	sc.DB.Where("date >= ?", sevenDaysAgo).Order("date ASC").Find(&weeklyStats)

	// 标签数量
	var tagCount int64
	sc.DB.Model(&models.Tag{}).Count(&tagCount)

	// 分类数量
	var categoryCount int64
	sc.DB.Model(&models.Category{}).Count(&categoryCount)

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"article_count":   articleCount,
			"published_count": publishedCount,
			"comment_count":   commentCount,
			"view_count":      viewCount,
			"like_count":      likeCount,
			"today_views":     todayStats.ViewCount,
			"today_visitors":  todayStats.VisitorCount,
			"tag_count":       tagCount,
			"category_count":  categoryCount,
			"hot_articles":    hotArticles,
			"latest_articles": latestArticles,
			"weekly_stats":    weeklyStats,
		},
	})
}

// RecordView 记录访问
func (sc *StatsController) RecordView(c *gin.Context) {
	today := time.Now().Format("2006-01-02")

	// 更新或创建今日统计
	var stats models.SiteStat
	result := sc.DB.Where("date = ?", today).First(&stats)
	if result.Error != nil {
		// 创建新记录
		stats = models.SiteStat{
			Date:         parseDate(today),
			ViewCount:    1,
			VisitorCount: 1,
		}
		sc.DB.Create(&stats)
	} else {
		// 更新访问量
		sc.DB.Model(&stats).Updates(map[string]interface{}{
			"view_count": gorm.Expr("view_count + 1"),
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "View recorded"})
}

// archiveItem 用于统一排序文章和HTML页面
type archiveItem struct {
	ID        uint
	Title     string
	Slug      string
	CreatedAt time.Time
	Category  models.Category
	Type      string
}

// GetArchive 获取文章归档
func (sc *StatsController) GetArchive(c *gin.Context) {
	// 获取文章
	var articles []models.Article
	sc.DB.Where("is_published = ?", true).
		Select("id, title, slug, created_at, category_id").
		Preload("Category").
		Order("created_at DESC").
		Find(&articles)

	// 获取HTML页面
	var htmlPages []models.HtmlPage
	sc.DB.Where("is_published = ?", true).
		Select("id, title, slug, created_at, category_id").
		Preload("Category").
		Order("created_at DESC").
		Find(&htmlPages)

	// 合并文章和HTML页面，统一按时间排序
	var allItems []archiveItem
	for _, article := range articles {
		allItems = append(allItems, archiveItem{
			ID:        article.ID,
			Title:     article.Title,
			Slug:      article.Slug,
			CreatedAt: article.CreatedAt,
			Category:  article.Category,
			Type:      "article",
		})
	}
	for _, htmlPage := range htmlPages {
		allItems = append(allItems, archiveItem{
			ID:        htmlPage.ID,
			Title:     htmlPage.Title,
			Slug:      htmlPage.Slug,
			CreatedAt: htmlPage.CreatedAt,
			Category:  htmlPage.Category,
			Type:      "htmlpage",
		})
	}

	// 按创建时间降序排序
	for i := 0; i < len(allItems); i++ {
		for j := i + 1; j < len(allItems); j++ {
			if allItems[i].CreatedAt.Before(allItems[j].CreatedAt) {
				allItems[i], allItems[j] = allItems[j], allItems[i]
			}
		}
	}

	// 按年月分组
	archive := make(map[string]map[string][]gin.H)

	// 添加到归档
	for _, item := range allItems {
		year := item.CreatedAt.Format("2006")
		month := item.CreatedAt.Format("01月")

		if archive[year] == nil {
			archive[year] = make(map[string][]gin.H)
		}

		archive[year][month] = append(archive[year][month], gin.H{
			"id":         item.ID,
			"title":      item.Title,
			"slug":       item.Slug,
			"created_at": item.CreatedAt,
			"category":   item.Category,
			"type":       item.Type,
		})
	}

	// 转换为数组格式（按年倒序）
	var result []gin.H
	years := getSortedKeys(archive)
	for _, year := range years {
		months := getSortedMonthKeys(archive[year])
		var monthData []gin.H
		articleCount := 0
		for _, month := range months {
			articlesInMonth := len(archive[year][month])
			articleCount += articlesInMonth
			monthData = append(monthData, gin.H{
				"month":    month,
				"articles": archive[year][month],
				"count":    articlesInMonth,
			})
		}
		result = append(result, gin.H{
			"year":   year,
			"months": monthData,
			"count":  articleCount,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

func parseDate(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}

func getSortedKeys(m map[string]map[string][]gin.H) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	// 简单排序（降序）
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] < keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return keys
}

func getSortedMonthKeys(m map[string][]gin.H) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	// 按月份排序（降序）
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return keys
}

// parseInt 辅助函数
func parseInt(s string) int {
	result, _ := strconv.Atoi(s)
	return result
}