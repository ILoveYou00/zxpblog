package controllers

import (
	"blog/config"
	"blog/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OAuthController struct {
	DB  *gorm.DB
	Cfg *config.Config
}

func NewOAuthController(cfg *config.Config, db *gorm.DB) *OAuthController {
	return &OAuthController{
		DB:  db,
		Cfg: cfg,
	}
}

// GitHubOAuthStart 开始 GitHub OAuth 流程
func (oc *OAuthController) GitHubOAuthStart(c *gin.Context) {
	if oc.Cfg.GitHubClientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "GitHub OAuth 未配置"})
		return
	}

	// 生成随机 state 用于防止 CSRF
	state := generateRandomState()
	session := sessions.Default(c)
	session.Set("oauth_state", state)
	session.Save()

	// 构建 GitHub 授权 URL
	authURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=user:email&state=%s",
		oc.Cfg.GitHubClientID,
		url.QueryEscape(oc.Cfg.GitHubRedirectURL),
		state,
	)

	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// GitHubOAuthCallback GitHub OAuth 回调
func (oc *OAuthController) GitHubOAuthCallback(c *gin.Context) {
	// 验证 state
	session := sessions.Default(c)
	savedState := session.Get("oauth_state")
	receivedState := c.Query("state")

	errorMsg := ""
	if savedState == nil || savedState.(string) != receivedState {
		errorMsg = "OAuth 验证失败，请重新登录"
	}

	// 清除 state
	session.Delete("oauth_state")
	session.Save()

	if errorMsg != "" {
		c.Redirect(http.StatusTemporaryRedirect, "/login.html?error="+url.QueryEscape(errorMsg))
		return
	}

	code := c.Query("code")
	if code == "" {
		c.Redirect(http.StatusTemporaryRedirect, "/login.html?error="+url.QueryEscape("缺少授权码，请重新登录"))
		return
	}

	// 用授权码换取 access token
	token, err := oc.exchangeCodeForToken(code)
	if err != nil {
		fmt.Printf("GitHub OAuth token exchange failed: %v\n", err)
		c.Redirect(http.StatusTemporaryRedirect, "/login.html?error="+url.QueryEscape(fmt.Sprintf("GitHub 登录失败: %v", err)))
		return
	}

	// 获取 GitHub 用户信息
	githubUser, err := oc.getGitHubUserInfo(token)
	if err != nil {
		fmt.Printf("GitHub OAuth get user info failed: %v\n", err)
		c.Redirect(http.StatusTemporaryRedirect, "/login.html?error="+url.QueryEscape("获取用户信息失败，请重试"))
		return
	}

	// 查找或创建用户
	user, err := oc.findOrCreateUser(githubUser, token)
	if err != nil {
		fmt.Printf("GitHub OAuth find/create user failed: %v\n", err)
		c.Redirect(http.StatusTemporaryRedirect, "/login.html?error="+url.QueryEscape("用户处理失败，请重试"))
		return
	}

	// 设置登录 session
	session.Set("user_id", user.ID)
	session.Save()

	// 重定向到管理页面
	c.Redirect(http.StatusTemporaryRedirect, "/admin.html")
}

// exchangeCodeForToken 用授权码换取 access token
func (oc *OAuthController) exchangeCodeForToken(code string) (string, error) {
	tokenURL := "https://github.com/login/oauth/access_token"

	data := url.Values{}
	data.Set("client_id", oc.Cfg.GitHubClientID)
	data.Set("client_secret", oc.Cfg.GitHubClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", oc.Cfg.GitHubRedirectURL)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("create request failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request GitHub failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
		Error       string `json:"error"`
		ErrorURI    string `json:"error_uri"`
		ErrorDesc   string `json:"error_description"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("parse response failed: %v, body: %s", err, string(body))
	}

	if result.Error != "" {
		// 常见错误：bad_verification_code（授权码已使用或无效）
		fmt.Printf("GitHub OAuth error: %s - %s (URI: %s)\n", result.Error, result.ErrorDesc, result.ErrorURI)
		return "", fmt.Errorf("GitHub OAuth 错误: %s", result.ErrorDesc)
	}

	if result.AccessToken == "" {
		return "", fmt.Errorf("响应中没有 access_token: %s", string(body))
	}

	return result.AccessToken, nil
}

// GitHubUserInfo GitHub 用户信息
type GitHubUserInfo struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

// getGitHubUserInfo 获取 GitHub 用户信息
func (oc *OAuthController) getGitHubUserInfo(token string) (*GitHubUserInfo, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user GitHubUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	// 如果没有公开邮箱，尝试获取私有邮箱
	if user.Email == "" {
		email, _ := oc.getGitHubUserEmail(token)
		user.Email = email
	}

	return &user, nil
}

// getGitHubUserEmail 获取 GitHub 用户邮箱
func (oc *OAuthController) getGitHubUserEmail(token string) (string, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var emails []struct {
		Email   string `json:"email"`
		Primary bool   `json:"primary"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	for _, e := range emails {
		if e.Primary {
			return e.Email, nil
		}
	}

	if len(emails) > 0 {
		return emails[0].Email, nil
	}

	return "", nil
}

// findOrCreateUser 查找或创建用户
func (oc *OAuthController) findOrCreateUser(githubUser *GitHubUserInfo, token string) (*models.User, error) {
	// 查找 OAuth 关联
	var conn models.OAuthConnection
	result := oc.DB.Where("provider = ? AND provider_id = ?", "github", fmt.Sprintf("%d", githubUser.ID)).First(&conn)

	if result.Error == nil {
		// 已有关联，获取用户
		var user models.User
		if err := oc.DB.First(&user, conn.UserID).Error; err != nil {
			return nil, err
		}
		// 更新 token
		conn.AccessToken = token
		oc.DB.Save(&conn)
		return &user, nil
	}

	// 没有关联，检查是否有同名用户
	var user models.User
	result = oc.DB.Where("username = ? OR email = ?", githubUser.Login, githubUser.Email).First(&user)

	if result.Error == gorm.ErrRecordNotFound {
		// 创建新用户
		user = models.User{
			Username: githubUser.Login,
			Email:    githubUser.Email,
		}
		if err := oc.DB.Create(&user).Error; err != nil {
			return nil, err
		}
	}

	// 创建 OAuth 关联
	conn = models.OAuthConnection{
		UserID:      user.ID,
		Provider:    "github",
		ProviderID:  fmt.Sprintf("%d", githubUser.ID),
		AccessToken: token,
	}
	if err := oc.DB.Create(&conn).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// GetOAuthStatus 获取 OAuth 配置状态
func (oc *OAuthController) GetOAuthStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"github_enabled": oc.Cfg.GitHubClientID != "",
	})
}

// generateRandomState 生成随机 state
func generateRandomState() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}