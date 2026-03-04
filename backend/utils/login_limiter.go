package utils

import (
	"sync"
	"time"
)

// LoginAttempt 登录尝试记录
type LoginAttempt struct {
	Count     int
	LockedUntil time.Time
	LastAttempt time.Time
}

// LoginLimiter 登录限制器
type LoginLimiter struct {
	attempts    map[string]*LoginAttempt
	mu          sync.RWMutex
	MaxAttempts int           // 最大尝试次数（公开）
	lockDuration time.Duration // 锁定时长
	cleanupTick  time.Duration // 清理周期
}

// NewLoginLimiter 创建登录限制器
func NewLoginLimiter(maxAttempts int, lockDuration, cleanupTick time.Duration) *LoginLimiter {
	limiter := &LoginLimiter{
		attempts:     make(map[string]*LoginAttempt),
		MaxAttempts:  maxAttempts,
		lockDuration: lockDuration,
		cleanupTick:  cleanupTick,
	}

	// 启动定期清理过期记录的协程
	go limiter.cleanupLoop()

	return limiter
}

// CheckLocked 检查IP是否被锁定
func (l *LoginLimiter) CheckLocked(ip string) (bool, time.Duration) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	attempt, exists := l.attempts[ip]
	if !exists {
		return false, 0
	}

	// 检查是否还在锁定期内
	if !attempt.LockedUntil.IsZero() && time.Now().Before(attempt.LockedUntil) {
		remaining := time.Until(attempt.LockedUntil)
		return true, remaining
	}

	return false, 0
}

// RecordFailure 记录登录失败
func (l *LoginLimiter) RecordFailure(ip string) int {
	l.mu.Lock()
	defer l.mu.Unlock()

	attempt, exists := l.attempts[ip]
	if !exists {
		attempt = &LoginAttempt{Count: 0}
		l.attempts[ip] = attempt
	}

	attempt.Count++
	attempt.LastAttempt = time.Now()

	// 达到最大尝试次数，锁定
	if attempt.Count >= l.MaxAttempts {
		attempt.LockedUntil = time.Now().Add(l.lockDuration)
	}

	return attempt.Count
}

// RecordSuccess 记录登录成功，清除失败记录
func (l *LoginLimiter) RecordSuccess(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.attempts, ip)
}

// GetAttempts 获取尝试次数
func (l *LoginLimiter) GetAttempts(ip string) int {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if attempt, exists := l.attempts[ip]; exists {
		return attempt.Count
	}
	return 0
}

// cleanupLoop 定期清理过期记录
func (l *LoginLimiter) cleanupLoop() {
	ticker := time.NewTicker(l.cleanupTick)
	for range ticker.C {
		l.cleanup()
	}
}

// cleanup 清理过期记录
func (l *LoginLimiter) cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	for ip, attempt := range l.attempts {
		// 清理已解锁且超过30分钟未活动的记录
		if (attempt.LockedUntil.IsZero() || now.After(attempt.LockedUntil)) &&
			now.Sub(attempt.LastAttempt) > 30*time.Minute {
			delete(l.attempts, ip)
		}
	}
}