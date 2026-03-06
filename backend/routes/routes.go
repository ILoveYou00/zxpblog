package routes

import (
	"blog/config"
	"blog/controllers"
	"blog/middleware"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Setup(r *gin.Engine, db *gorm.DB, cfg *config.Config) {
	// Session middleware
	store := cookie.NewStore([]byte(cfg.SessionSecret))
	r.Use(sessions.Sessions("blog_session", store))

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Session-ID")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Initialize controllers
	articleCtrl := controllers.NewArticleController(db)
	categoryCtrl := controllers.NewCategoryController(db)
	commentCtrl := controllers.NewCommentController(db)
	authCtrl := controllers.NewAuthController(db)
	rssCtrl := controllers.NewRSSController(db)
	likeCtrl := controllers.NewLikeController(db)
	favoriteCtrl := controllers.NewFavoriteController(db)
	historyCtrl := controllers.NewHistoryController(db)
	tagCtrl := controllers.NewTagController(db)
	friendCtrl := controllers.NewFriendLinkController(db)
	announcementCtrl := controllers.NewAnnouncementController(db)
	mediaCtrl := controllers.NewMediaController(db)
	statsCtrl := controllers.NewStatsController(db)
	sitemapCtrl := controllers.NewSitemapController(db)
	aboutCtrl := controllers.NewAboutController(db)
	aiCtrl := controllers.NewAIController(cfg, db)
	oauthCtrl := controllers.NewOAuthController(cfg, db)
	htmlCtrl := controllers.NewHtmlPageController(db)

	// API routes
	api := r.Group("/api")
	{
		// Public routes
		api.GET("/articles", articleCtrl.GetArticles)
		api.GET("/articles/:id", articleCtrl.GetArticle)
		api.GET("/articles/slug/:slug", articleCtrl.GetArticleBySlug)
		api.GET("/articles/:id/related", articleCtrl.GetRelatedArticles)
		api.GET("/categories", categoryCtrl.GetCategories)
		api.GET("/comments", commentCtrl.GetComments)
		api.POST("/comments", commentCtrl.CreateComment)
		api.POST("/comments/:id/reply", commentCtrl.ReplyComment)
		api.GET("/rss", rssCtrl.GetRSS)
		api.GET("/session", articleCtrl.GetSession)

		// 点赞/收藏/历史
		api.POST("/articles/:id/like", likeCtrl.LikeArticle)
		api.DELETE("/articles/:id/like", likeCtrl.UnlikeArticle)
		api.GET("/articles/:id/liked", likeCtrl.CheckLikeStatus)
		api.POST("/articles/:id/favorite", favoriteCtrl.AddFavorite)
		api.DELETE("/articles/:id/favorite", favoriteCtrl.RemoveFavorite)
		api.GET("/articles/:id/favorited", favoriteCtrl.CheckFavoriteStatus)
		api.GET("/favorites", favoriteCtrl.GetFavorites)
		api.POST("/history", historyCtrl.AddHistory)
		api.GET("/history", historyCtrl.GetHistory)
		api.DELETE("/history", historyCtrl.ClearHistory)

		// 标签
		api.GET("/tags", tagCtrl.GetTags)
		api.GET("/tags/:id/articles", tagCtrl.GetTagArticles)

		// 友链
		api.GET("/friend-links", friendCtrl.GetFriendLinks)

		// 公告
		api.GET("/announcements", announcementCtrl.GetAnnouncements)

		// 统计
		api.GET("/stats", statsCtrl.GetStats)
		api.POST("/stats/record", statsCtrl.RecordView)
		api.GET("/archives", statsCtrl.GetArchive)

		// 关于页面
		api.GET("/about", aboutCtrl.GetAbout)

		// HTML页面
		api.GET("/html-pages", htmlCtrl.GetHtmlPages)
		api.GET("/html-pages/:id", htmlCtrl.GetHtmlPage)

		// AI 功能
		api.GET("/ai/status", aiCtrl.GetAIStatus)
		api.POST("/ai/chat", aiCtrl.Chat)
		api.POST("/ai/summary", aiCtrl.GenerateSummary)

		// Auth routes
		auth := api.Group("/auth")
		{
			auth.GET("/captcha", authCtrl.GetCaptcha)
			auth.POST("/login", authCtrl.Login)
			auth.POST("/logout", authCtrl.Logout)
			auth.GET("/me", authCtrl.GetCurrentUser)
			// OAuth routes
			auth.GET("/oauth/status", oauthCtrl.GetOAuthStatus)
			auth.GET("/oauth/github", oauthCtrl.GitHubOAuthStart)
			auth.GET("/oauth/github/callback", oauthCtrl.GitHubOAuthCallback)
		}

		// Admin routes (protected)
		admin := api.Group("/admin")
		admin.Use(middleware.AuthRequired(cfg))
		{
			// Articles
			admin.GET("/articles", articleCtrl.GetAllArticles)
			admin.POST("/articles", articleCtrl.CreateArticle)
			admin.PUT("/articles/:id", articleCtrl.UpdateArticle)
			admin.DELETE("/articles/:id", articleCtrl.DeleteArticle)

			// Categories
			admin.POST("/categories", categoryCtrl.CreateCategory)
			admin.PUT("/categories/:id", categoryCtrl.UpdateCategory)
			admin.DELETE("/categories/:id", categoryCtrl.DeleteCategory)

			// Comments
			admin.GET("/comments", commentCtrl.GetAllComments)
			admin.PUT("/comments/:id/approve", commentCtrl.ApproveComment)
			admin.DELETE("/comments/:id", commentCtrl.DeleteComment)

			// Tags
			admin.GET("/tags", tagCtrl.GetTags)
			admin.POST("/tags", tagCtrl.CreateTag)
			admin.PUT("/tags/:id", tagCtrl.UpdateTag)
			admin.DELETE("/tags/:id", tagCtrl.DeleteTag)
			admin.GET("/articles/:id/tags", tagCtrl.GetArticleTags)
			admin.PUT("/articles/:id/tags", tagCtrl.SetArticleTags)

			// Friend links
			admin.GET("/friend-links/all", friendCtrl.GetAllFriendLinks)
			admin.POST("/friend-links", friendCtrl.CreateFriendLink)
			admin.PUT("/friend-links/:id", friendCtrl.UpdateFriendLink)
			admin.DELETE("/friend-links/:id", friendCtrl.DeleteFriendLink)
			admin.PUT("/friend-links/:id/toggle", friendCtrl.ToggleFriendLink)

			// Announcements
			admin.GET("/announcements/all", announcementCtrl.GetAllAnnouncements)
			admin.POST("/announcements", announcementCtrl.CreateAnnouncement)
			admin.PUT("/announcements/:id", announcementCtrl.UpdateAnnouncement)
			admin.DELETE("/announcements/:id", announcementCtrl.DeleteAnnouncement)
			admin.PUT("/announcements/:id/toggle", announcementCtrl.ToggleAnnouncement)

			// Media
			admin.POST("/media", mediaCtrl.UploadMedia)
			admin.GET("/media", mediaCtrl.GetMediaList)
			admin.GET("/media/:id", mediaCtrl.GetMediaByID)
			admin.DELETE("/media/:id", mediaCtrl.DeleteMedia)

			// About
			admin.PUT("/about", aboutCtrl.UpdateAbout)

			// HTML Pages
			admin.GET("/html-pages/all", htmlCtrl.GetAllHtmlPages)
			admin.POST("/html-pages", htmlCtrl.CreateHtmlPage)
			admin.PUT("/html-pages/:id", htmlCtrl.UpdateHtmlPage)
			admin.DELETE("/html-pages/:id", htmlCtrl.DeleteHtmlPage)

			// AI 写作助手（管理员）
			admin.POST("/ai/writing", aiCtrl.WritingAssist)
		}
	}

	// Sitemap
	r.GET("/sitemap.xml", sitemapCtrl.GetSitemap)

	// Static files for uploads
	r.Static("/uploads", "/app/uploads")
}