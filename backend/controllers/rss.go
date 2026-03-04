package controllers

import (
	"blog/models"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RSSController struct {
	DB *gorm.DB
}

func NewRSSController(db *gorm.DB) *RSSController {
	return &RSSController{DB: db}
}

// GetRSS returns RSS feed
func (rc *RSSController) GetRSS(c *gin.Context) {
	var articles []models.Article
	rc.DB.Where("is_published = ?", true).Order("created_at DESC").Limit(20).Preload("Category").Find(&articles)

	now := time.Now().Format(time.RFC1123)

	rss := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
<channel>
<title>My Blog</title>
<link>http://localhost/</link>
<description>Personal Blog</description>
<language>zh-CN</language>
<lastBuildDate>` + now + `</lastBuildDate>
<atom:link href="http://localhost/api/rss" rel="self" type="application/rss+xml"/>
`

	for _, article := range articles {
		description := article.Summary
		if description == "" && len(article.Content) > 200 {
			description = article.Content[:200] + "..."
		} else if description == "" {
			description = article.Content
		}

		rss += `
<item>
<title><![CDATA[` + article.Title + `]]></title>
<link>http://localhost/article/` + fmt.Sprintf("%d", article.ID) + `</link>
<description><![CDATA[` + description + `]]></description>
<pubDate>` + article.CreatedAt.Format(time.RFC1123) + `</pubDate>
<guid isPermaLink="true">http://localhost/article/` + fmt.Sprintf("%d", article.ID) + `</guid>
</item>
`
	}

	rss += `
</channel>
</rss>`

	c.Data(http.StatusOK, "application/rss+xml; charset=utf-8", []byte(rss))
}