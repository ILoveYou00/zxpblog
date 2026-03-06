package main

import (
	"fmt"
	"log"
	"blog/config"
	"blog/models"
	"blog/routes"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// Load config
	cfg := config.Load()

	// Connect to database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto migrate
	err = db.AutoMigrate(
		&models.User{}, &models.Article{}, &models.Category{}, &models.Comment{},
		&models.Like{}, &models.Favorite{}, &models.ReadHistory{},
		&models.Tag{}, &models.ArticleTag{}, &models.FriendLink{},
		&models.Announcement{}, &models.Media{}, &models.SiteStat{},
		&models.About{}, &models.OAuthConnection{}, &models.HtmlPage{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Create default admin user if not exists
	var userCount int64
	db.Model(&models.User{}).Count(&userCount)
	if userCount == 0 {
		hashedPassword, _ := models.HashPassword(cfg.AdminPassword)
		admin := models.User{
			Username: cfg.AdminUsername,
			Password: hashedPassword,
			Email:    "admin@localhost",
		}
		db.Create(&admin)
		log.Println("Default admin user created")
	}

	// Create default category if not exists
	var categoryCount int64
	db.Model(&models.Category{}).Count(&categoryCount)
	if categoryCount == 0 {
		defaultCategory := models.Category{
			Name: "默认分类",
			Slug: "default",
		}
		db.Create(&defaultCategory)
		log.Println("Default category created")
	}

	// Setup Gin
	if cfg.GinMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// Setup routes
	routes.Setup(r, db, cfg)

	// Start server
	log.Printf("Server starting on port %s", cfg.AppPort)
	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}