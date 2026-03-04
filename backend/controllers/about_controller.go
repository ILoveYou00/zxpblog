package controllers

import (
	"blog/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AboutController struct {
	DB *gorm.DB
}

func NewAboutController(db *gorm.DB) *AboutController {
	return &AboutController{DB: db}
}

// GetAbout 获取关于页面信息（公开）
func (ac *AboutController) GetAbout(c *gin.Context) {
	var about models.About
	result := ac.DB.First(&about)

	if result.Error != nil {
		// 如果没有记录，返回默认值
		c.JSON(http.StatusOK, gin.H{
			"data": models.About{
				Name:    "博主",
				Bio:     "一名热爱技术和生活的开发者",
				AboutMe: "欢迎来到我的个人博客！",
				Skills:  "Go,JavaScript,Vue.js,MySQL,Docker",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": about})
}

// UpdateAbout 更新关于页面信息（管理员）
func (ac *AboutController) UpdateAbout(c *gin.Context) {
	var about models.About
	result := ac.DB.First(&about)

	var input models.About
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 如果记录不存在，创建新记录
	if result.Error != nil {
		about = input
		about.UpdatedAt = time.Now()
		if err := ac.DB.Create(&about).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		// 更新现有记录
		about.Name = input.Name
		about.Bio = input.Bio
		about.Avatar = input.Avatar
		about.AboutMe = input.AboutMe
		about.AboutMeHTML = input.AboutMeHTML
		about.Skills = input.Skills
		about.Email = input.Email
		about.Github = input.Github
		about.Twitter = input.Twitter
		about.Website = input.Website
		about.UpdatedAt = time.Now()

		if err := ac.DB.Save(&about).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": about})
}