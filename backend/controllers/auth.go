package controllers

import (
	"blog/models"
	"blog/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthController struct {
	DB          *gorm.DB
	LoginLimiter *utils.LoginLimiter
}

func NewAuthController(db *gorm.DB) *AuthController {
	// 创建登录限制器：5次失败后锁定15分钟，每10分钟清理一次过期记录
	limiter := utils.NewLoginLimiter(5, 15*time.Minute, 10*time.Minute)
	return &AuthController{
		DB:          db,
		LoginLimiter: limiter,
	}
}

type LoginInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Captcha  string `json:"captcha"`
	CaptchaID string `json:"captcha_id"`
}

// CaptchaStore 验证码存储（简单内存存储，生产环境可用 Redis）
var CaptchaStore = make(map[string]*utils.Captcha)

// GetCaptcha 获取验证码
func (ac *AuthController) GetCaptcha(c *gin.Context) {
	captcha := utils.GenerateCaptcha(120, 50)

	// 生成唯一ID
	captchaID := strconv.FormatInt(time.Now().UnixNano(), 36)

	// 存储验证码
	CaptchaStore[captchaID] = captcha

	// 清理过期的验证码（5分钟过期）
	go func() {
		time.Sleep(5 * time.Minute)
		delete(CaptchaStore, captchaID)
	}()

	c.JSON(http.StatusOK, gin.H{
		"captcha_id": captchaID,
		"captcha_image": captcha.Image,
	})
}

// Login handles user login
func (ac *AuthController) Login(c *gin.Context) {
	// 获取客户端IP
	clientIP := c.ClientIP()

	// 检查是否被锁定
	locked, remaining := ac.LoginLimiter.CheckLocked(clientIP)
	if locked {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": "登录失败次数过多，请 " + formatDuration(remaining) + " 后再试",
			"locked": true,
			"remaining_seconds": int(remaining.Seconds()),
		})
		return
	}

	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证验证码
	if input.CaptchaID != "" && input.Captcha != "" {
		storedCaptcha, exists := CaptchaStore[input.CaptchaID]
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "验证码已过期，请刷新"})
			return
		}
		if storedCaptcha.Code != input.Captcha {
			// 验证码错误也记录失败次数
			attempts := ac.LoginLimiter.RecordFailure(clientIP)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "验证码错误",
				"attempts_remaining": ac.LoginLimiter.MaxAttempts - attempts,
			})
			return
		}
		// 验证通过，删除验证码
		delete(CaptchaStore, input.CaptchaID)
	}

	var user models.User
	result := ac.DB.Where("username = ?", input.Username).First(&user)
	if result.Error != nil {
		// 记录失败
		attempts := ac.LoginLimiter.RecordFailure(clientIP)
		remaining := ac.LoginLimiter.MaxAttempts - attempts
		if remaining <= 0 {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "登录失败次数过多，账户已被锁定15分钟",
				"locked": true,
			})
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "用户名或密码错误",
				"attempts_remaining": remaining,
			})
		}
		return
	}

	if !models.CheckPassword(input.Password, user.Password) {
		// 记录失败
		attempts := ac.LoginLimiter.RecordFailure(clientIP)
		remaining := ac.LoginLimiter.MaxAttempts - attempts
		if remaining <= 0 {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "登录失败次数过多，账户已被锁定15分钟",
				"locked": true,
			})
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "用户名或密码错误",
				"attempts_remaining": remaining,
			})
		}
		return
	}

	// 登录成功，清除失败记录
	ac.LoginLimiter.RecordSuccess(clientIP)

	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	session.Save()

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

// Logout handles user logout
func (ac *AuthController) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()

	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

// GetCurrentUser returns the current logged in user
func (ac *AuthController) GetCurrentUser(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var user models.User
	result := ac.DB.First(&user, userID)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

// formatDuration 格式化时长
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return strconv.Itoa(int(d.Seconds())) + " 秒"
	}
	return strconv.Itoa(int(d.Minutes())) + " 分钟"
}