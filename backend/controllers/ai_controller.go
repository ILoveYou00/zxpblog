package controllers

import (
	"blog/config"
	"blog/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AIController struct {
	Config *config.Config
	DB     *gorm.DB
	Client *http.Client
}

func NewAIController(cfg *config.Config, db *gorm.DB) *AIController {
	return &AIController{
		Config: cfg,
		DB:     db,
		Client: &http.Client{Timeout: 60 * time.Second},
	}
}

// AIRequest AI 请求结构
type AIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	MaxTokens int      `json:"max_tokens,omitempty"`
}

// Message 消息结构
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AIResponse AI 响应结构
type AIResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Message string `json:"message"`
	History []Message `json:"history,omitempty"`
}

// WritingRequest 写作请求
type WritingRequest struct {
	Type    string `json:"type"`    // continue, polish, summarize, expand, translate
	Content string `json:"content"` // 原始内容
	Context string `json:"context"` // 上下文（可选）
}

// GenerateSummaryRequest 生成摘要请求
type GenerateSummaryRequest struct {
	Content string `json:"content"`
}

// GetAIStatus 获取 AI 状态
func (a *AIController) GetAIStatus(c *gin.Context) {
	// 检查 AI 是否正确配置
	enabled := a.Config.AIEnabled && a.Config.AIApiUrl != "" && a.Config.AIModel != ""
	c.JSON(http.StatusOK, gin.H{
		"enabled": enabled,
	})
}

// Chat AI 聊天
func (a *AIController) Chat(c *gin.Context) {
	// 检查 AI 是否正确配置
	if !a.Config.AIEnabled || a.Config.AIApiUrl == "" || a.Config.AIModel == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI 功能未配置"})
		return
	}

	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取博客上下文信息（仅公开信息）
	blogContext := a.getPublicBlogContext()

	// 构建系统提示 - 限制 AI 只能回答博客相关公开信息
	systemPrompt := fmt.Sprintf(`你是一个友好的博客助手。以下是这个博客的公开信息：

%s

重要规则：
1. 你只能回答与博客内容相关的问题
2. 不要透露任何管理员信息、系统配置或敏感数据
3. 如果用户询问敏感信息，礼貌地拒绝并说明你只能帮助了解博客内容
4. 对于与技术无关的问题，可以礼貌地说明你的职责

请用简洁、友好的方式回答问题。`, blogContext)

	messages := []Message{
		{Role: "system", Content: systemPrompt},
	}

	// 添加历史消息
	if len(req.History) > 0 {
		messages = append(messages, req.History...)
	}

	// 添加当前消息
	messages = append(messages, Message{Role: "user", Content: req.Message})

	response, err := a.callAI(messages)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": response,
	})
}

// getPublicBlogContext 获取博客公开上下文信息（用于普通用户的 AI 聊天）
func (a *AIController) getPublicBlogContext() string {
	var sb strings.Builder

	// 获取分类
	var categories []models.Category
	a.DB.Find(&categories)
	if len(categories) > 0 {
		sb.WriteString("## 博客分类\n")
		for _, cat := range categories {
			fmt.Fprintf(&sb, "- %s\n", cat.Name)
		}
		sb.WriteString("\n")
	}

	// 获取标签
	var tags []models.Tag
	a.DB.Find(&tags)
	if len(tags) > 0 {
		sb.WriteString("## 博客标签\n")
		for i, tag := range tags {
			if i > 0 {
				sb.WriteString("、")
			}
			sb.WriteString(tag.Name)
		}
		sb.WriteString("\n\n")
	}

	// 获取已发布的最新文章（最多10篇）- 只包含公开信息
	var articles []models.Article
	a.DB.Where("is_published = ?", true).Order("created_at DESC").Limit(10).Find(&articles)
	if len(articles) > 0 {
		sb.WriteString("## 最新文章\n")
		for i, article := range articles {
			fmt.Fprintf(&sb, "%d. **%s** - %s\n", i+1, article.Title, article.CreatedAt.Format("2006-01-02"))
			if article.Summary != "" {
				fmt.Fprintf(&sb, "   简介：%s\n", article.Summary)
			}
		}
	}

	// 只统计已发布文章数
	var publishedCount int64
	a.DB.Model(&models.Article{}).Where("is_published = ?", true).Count(&publishedCount)

	fmt.Fprintf(&sb, "\n## 博客统计\n- 已发布文章：%d\n", publishedCount)

	return sb.String()
}

// getBlogContext 获取博客上下文信息（用于管理员写作助手）
func (a *AIController) getBlogContext() string {
	var sb strings.Builder

	// 获取分类
	var categories []models.Category
	a.DB.Find(&categories)
	if len(categories) > 0 {
		sb.WriteString("## 博客分类\n")
		for _, cat := range categories {
			fmt.Fprintf(&sb, "- %s\n", cat.Name)
		}
		sb.WriteString("\n")
	}

	// 获取标签
	var tags []models.Tag
	a.DB.Find(&tags)
	if len(tags) > 0 {
		sb.WriteString("## 博客标签\n")
		for i, tag := range tags {
			if i > 0 {
				sb.WriteString("、")
			}
			sb.WriteString(tag.Name)
		}
		sb.WriteString("\n\n")
	}

	// 获取最新文章（最多10篇）
	var articles []models.Article
	a.DB.Where("is_published = ?", true).Order("created_at DESC").Limit(10).Find(&articles)
	if len(articles) > 0 {
		sb.WriteString("## 最新文章\n")
		for i, article := range articles {
			fmt.Fprintf(&sb, "%d. **%s** - %s\n", i+1, article.Title, article.CreatedAt.Format("2006-01-02"))
			if article.Summary != "" {
				fmt.Fprintf(&sb, "   简介：%s\n", article.Summary)
			}
		}
	}

	// 获取统计信息
	var articleCount, publishedCount int64
	a.DB.Model(&models.Article{}).Count(&articleCount)
	a.DB.Model(&models.Article{}).Where("is_published = ?", true).Count(&publishedCount)

	fmt.Fprintf(&sb, "\n## 博客统计\n- 文章总数：%d\n- 已发布文章：%d\n", articleCount, publishedCount)

	return sb.String()
}

// WritingAssist 写作助手
func (a *AIController) WritingAssist(c *gin.Context) {
	if !a.Config.AIEnabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI 功能未启用"})
		return
	}

	var req WritingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var prompt string
	switch req.Type {
	case "continue":
		prompt = fmt.Sprintf(`请续写以下内容，保持风格一致，自然衔接：

%s

请直接输出续写内容，不要加任何解释或标记。`, req.Content)
	case "polish":
		prompt = fmt.Sprintf(`请润色以下内容，使其更加流畅、专业，但保持原意不变：

%s

请直接输出润色后的内容，不要加任何解释或标记。`, req.Content)
	case "summarize":
		prompt = fmt.Sprintf(`请为以下内容写一个简洁的摘要（100字以内）：

%s

请直接输出摘要，不要加任何解释或标记。`, req.Content)
	case "expand":
		prompt = fmt.Sprintf(`请扩展以下内容，增加更多细节和深度：

%s

请直接输出扩展后的内容，不要加任何解释或标记。`, req.Content)
	case "title":
		prompt = fmt.Sprintf(`请根据以下内容生成3个吸引人的标题选项：

%s

请直接输出标题，每行一个，不要加任何解释或编号。`, req.Content)
	case "translate":
		targetLang := req.Context
		if targetLang == "" {
			targetLang = "英文"
		}
		prompt = fmt.Sprintf(`请将以下内容翻译成%s：

%s

请直接输出翻译结果，不要加任何解释或标记。`, targetLang, req.Content)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的类型"})
		return
	}

	response, err := a.callAI([]Message{{Role: "user", Content: prompt}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result": response,
	})
}

// GenerateSummary 生成摘要
func (a *AIController) GenerateSummary(c *gin.Context) {
	if !a.Config.AIEnabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI 功能未启用"})
		return
	}

	var req GenerateSummaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	prompt := fmt.Sprintf(`请为以下文章内容生成一个简洁的摘要（150字以内），抓住文章的核心要点：

%s

请直接输出摘要内容，不要加任何解释或标记。`, req.Content)

	response, err := a.callAI([]Message{{Role: "user", Content: prompt}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"summary": response,
	})
}

// callAI 调用 AI API
func (a *AIController) callAI(messages []Message) (string, error) {
	req := AIRequest{
		Model:    a.Config.AIModel,
		Messages: messages,
		MaxTokens: 2000,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// 构建完整的 API URL
	// AI_API_URL 可能是 "https://api.example.com" 或 "https://api.example.com/v1"
	// 我们需要最终 URL 是 "https://api.example.com/v1/chat/completions"
	apiURL := a.Config.AIApiUrl
	if !strings.HasSuffix(apiURL, "/v1") && !strings.HasSuffix(apiURL, "/v1/") {
		apiURL = strings.TrimSuffix(apiURL, "/") + "/v1"
	}
	apiURL = strings.TrimSuffix(apiURL, "/") + "/chat/completions"

	httpReq, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+a.Config.AIApiKey)

	resp, err := a.Client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("AI API returned status %d: %s", resp.StatusCode, string(body))
	}

	var aiResp AIResponse
	if err := json.Unmarshal(body, &aiResp); err != nil {
		return "", fmt.Errorf("failed to parse AI response (invalid JSON): %w, body: %s", err, string(body))
	}

	if aiResp.Error != nil {
		return "", fmt.Errorf("AI API error: %s (type: %s)", aiResp.Error.Message, aiResp.Error.Type)
	}

	if len(aiResp.Choices) == 0 {
		return "", fmt.Errorf("no response choices from AI, body: %s", string(body))
	}

	return aiResp.Choices[0].Message.Content, nil
}