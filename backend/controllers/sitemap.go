package controllers

import (
	"blog/models"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SitemapController struct {
	DB *gorm.DB
}

func NewSitemapController(db *gorm.DB) *SitemapController {
	return &SitemapController{DB: db}
}

// GetSitemap 生成sitemap.xml
func (sc *SitemapController) GetSitemap(c *gin.Context) {
	var articles []models.Article
	sc.DB.Where("is_published = ?", true).
		Select("id, slug, updated_at, created_at").
		Order("created_at DESC").
		Find(&articles)

	var categories []models.Category
	sc.DB.Find(&categories)

	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	sb.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")

	// 首页
	sb.WriteString(fmt.Sprintf(`  <url>
    <loc>%s/</loc>
    <lastmod>%s</lastmod>
    <changefreq>daily</changefreq>
    <priority>1.0</priority>
  </url>
`, getBaseURL(c), time.Now().Format("2006-01-02")))

	// 关于页面
	sb.WriteString(fmt.Sprintf(`  <url>
    <loc>%s/about.html</loc>
    <changefreq>monthly</changefreq>
    <priority>0.5</priority>
  </url>
`, getBaseURL(c)))

	// 归档页面
	sb.WriteString(fmt.Sprintf(`  <url>
    <loc>%s/archives.html</loc>
    <changefreq>weekly</changefreq>
    <priority>0.6</priority>
  </url>
`, getBaseURL(c)))

	// 分类页面
	for _, cat := range categories {
		sb.WriteString(fmt.Sprintf(`  <url>
    <loc>%s/?category_id=%d</loc>
    <changefreq>weekly</changefreq>
    <priority>0.6</priority>
  </url>
`, getBaseURL(c), cat.ID))
	}

	// 文章页面
	for _, article := range articles {
		lastMod := article.UpdatedAt
		if lastMod.IsZero() {
			lastMod = article.CreatedAt
		}
		var url string
		if article.Slug != "" {
			url = fmt.Sprintf("%s/article.html?slug=%s", getBaseURL(c), article.Slug)
		} else {
			url = fmt.Sprintf("%s/article.html?id=%d", getBaseURL(c), article.ID)
		}
		sb.WriteString(fmt.Sprintf(`  <url>
    <loc>%s</loc>
    <lastmod>%s</lastmod>
    <changefreq>monthly</changefreq>
    <priority>0.8</priority>
  </url>
`, url, lastMod.Format("2006-01-02")))
	}

	sb.WriteString(`</urlset>`)

	c.Data(http.StatusOK, "application/xml", []byte(sb.String()))
}

func getBaseURL(c *gin.Context) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	host := c.Request.Host
	return fmt.Sprintf("%s://%s", scheme, host)
}