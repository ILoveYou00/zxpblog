package utils

import (
	"strings"
	"sync"
	"time"
)

// CommentRateLimiter 评论频率限制器
type CommentRateLimiter struct {
	records    map[string]*CommentRecord
	mu         sync.RWMutex
	interval   time.Duration // 评论间隔
	cleanupTick time.Duration // 清理周期
}

// CommentRecord 评论记录
type CommentRecord struct {
	LastComment time.Time
	Count       int
}

// NewCommentRateLimiter 创建评论限制器
func NewCommentRateLimiter(interval, cleanupTick time.Duration) *CommentRateLimiter {
	limiter := &CommentRateLimiter{
		records:     make(map[string]*CommentRecord),
		interval:    interval,
		cleanupTick: cleanupTick,
	}

	go limiter.cleanupLoop()

	return limiter
}

// CheckAllowed 检查是否允许评论
func (l *CommentRateLimiter) CheckAllowed(ip string) (bool, time.Duration) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	record, exists := l.records[ip]
	if !exists {
		return true, 0
	}

	elapsed := time.Since(record.LastComment)
	if elapsed < l.interval {
		return false, l.interval - elapsed
	}

	return true, 0
}

// RecordComment 记录评论
func (l *CommentRateLimiter) RecordComment(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	record, exists := l.records[ip]
	if !exists {
		record = &CommentRecord{}
		l.records[ip] = record
	}

	record.LastComment = time.Now()
	record.Count++
}

// cleanupLoop 定期清理过期记录
func (l *CommentRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(l.cleanupTick)
	for range ticker.C {
		l.cleanup()
	}
}

// cleanup 清理过期记录
func (l *CommentRateLimiter) cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	for ip, record := range l.records {
		if now.Sub(record.LastComment) > 10*time.Minute {
			delete(l.records, ip)
		}
	}
}

// SpamKeywords 垃圾评论关键词
var SpamKeywords = []string{
	"viagra", "casino", "porn", "sex", "xxx", "成人", "赌博", "博彩",
	"时时彩", "彩票", "六合彩", "澳门", "威尼斯人", "金沙",
	"代写论文", "论文代写", "办理证件", "办证", "发票",
	"贷款", "信用卡套现", "刷卡", "提现", "套现",
	"seo服务", "外链", "买链接", "卖链接",
	"加微信", "加QQ", "qq群", "微信群",
	"赚大钱", "日赚", "月入", "躺赚",
	"http://", "https://", "www.", ".com", ".cn", ".net", ".org",
}

// IsSpam 检测是否为垃圾评论
func IsSpam(content, nickname string) (bool, string) {
	contentLower := strings.ToLower(content)
	nicknameLower := strings.ToLower(nickname)

	// 检查内容中的垃圾关键词
	for _, keyword := range SpamKeywords {
		if strings.Contains(contentLower, strings.ToLower(keyword)) {
			return true, "评论内容包含不允许的关键词"
		}
		if strings.Contains(nicknameLower, strings.ToLower(keyword)) {
			return true, "昵称包含不允许的关键词"
		}
	}

	// 检查连续重复字符
	if hasRepeatedChars(content, 5) {
		return true, "评论内容包含过多重复字符"
	}

	// 检查链接数量
	linkCount := strings.Count(contentLower, "http://") +
		strings.Count(contentLower, "https://") +
		strings.Count(contentLower, "www.")
	if linkCount > 2 {
		return true, "评论包含过多链接"
	}

	return false, ""
}

// hasRepeatedChars 检查是否有连续重复字符
func hasRepeatedChars(s string, maxRepeat int) bool {
	if len(s) < maxRepeat {
		return false
	}

	count := 1
	for i := 1; i < len(s); i++ {
		if s[i] == s[i-1] {
			count++
			if count >= maxRepeat {
				return true
			}
		} else {
			count = 1
		}
	}
	return false
}