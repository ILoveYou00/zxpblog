package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Username  string         `json:"username" gorm:"uniqueIndex;size:50;not null"`
	Password  string         `json:"-" gorm:"size:255;not null"`
	Email     string         `json:"email" gorm:"size:100"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type Article struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	Title        string         `json:"title" gorm:"size:200;not null;index"`
	Slug         string         `json:"slug" gorm:"uniqueIndex;size:200"`
	Content      string         `json:"content" gorm:"type:longtext"`
	ContentFormat string         `json:"content_format" gorm:"size:20;default:'html'"` // html 或 markdown
	Summary      string         `json:"summary" gorm:"size:500"`
	CoverImage   string         `json:"cover_image" gorm:"size:255"`
	CategoryID   uint           `json:"category_id" gorm:"index"`
	Category     Category       `json:"category"`
	Tags         string         `json:"tags" gorm:"size:255"` // JSON array (deprecated, use ArticleTags)
	ViewCount    int            `json:"view_count" gorm:"default:0"`
	LikeCount    int            `json:"like_count" gorm:"default:0"`
	IsPublished  bool           `json:"is_published" gorm:"default:false;index"`
	IsPinned     bool           `json:"is_pinned" gorm:"default:false;index"`
	ReadTime     int            `json:"read_time" gorm:"default:0"` // 阅读时长(分钟)
	CreatedAt    time.Time      `json:"created_at" gorm:"index"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

type Category struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"size:50;not null"`
	Slug      string         `json:"slug" gorm:"uniqueIndex;size:50"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type Comment struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	ArticleID  uint           `json:"article_id" gorm:"index;not null"`
	Article    Article        `json:"article"`
	ParentID   *uint          `json:"parent_id" gorm:"index"`
	Replies    []Comment      `json:"replies" gorm:"foreignKey:ParentID"`
	Nickname   string         `json:"nickname" gorm:"size:50;not null"`
	Email      string         `json:"email" gorm:"size:100"`
	Content    string         `json:"content" gorm:"type:text;not null"`
	IsApproved bool           `json:"is_approved" gorm:"default:false"`
	CreatedAt  time.Time      `json:"created_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

// Like 文章点赞
type Like struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ArticleID uint      `json:"article_id" gorm:"index;not null"`
	IP        string    `json:"ip" gorm:"size:45"`
	CreatedAt time.Time `json:"created_at"`
}

// Favorite 文章收藏
type Favorite struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ArticleID uint      `json:"article_id" gorm:"index;not null"`
	SessionID string    `json:"session_id" gorm:"size:100;index"`
	CreatedAt time.Time `json:"created_at"`
}

// ReadHistory 阅读历史
type ReadHistory struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ArticleID uint      `json:"article_id" gorm:"index;not null"`
	SessionID string    `json:"session_id" gorm:"size:100;index"`
	CreatedAt time.Time `json:"created_at"`
}

// Tag 标签
type Tag struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"uniqueIndex;size:50;not null"`
	Slug      string         `json:"slug" gorm:"uniqueIndex;size:50"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// ArticleTag 文章-标签关联
type ArticleTag struct {
	ID        uint `json:"id" gorm:"primaryKey"`
	ArticleID uint `json:"article_id" gorm:"index"`
	TagID     uint `json:"tag_id" gorm:"index"`
}

// FriendLink 友情链接
type FriendLink struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"size:100;not null"`
	URL       string         `json:"url" gorm:"size:255;not null"`
	Logo      string         `json:"logo" gorm:"size:255"`
	Desc      string         `json:"desc" gorm:"size:200"`
	SortOrder int            `json:"sort_order" gorm:"default:0"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// Announcement 公告
type Announcement struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Title     string         `json:"title" gorm:"size:200;not null"`
	Content   string         `json:"content" gorm:"type:text"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	StartTime time.Time      `json:"start_time"`
	EndTime   time.Time      `json:"end_time"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// Media 媒体文件
type Media struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Filename  string         `json:"filename" gorm:"size:255;not null"`
	URL       string         `json:"url" gorm:"size:500;not null"`
	Size      int64          `json:"size"`
	MimeType  string         `json:"mime_type" gorm:"size:50"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// SiteStat 站点统计
type SiteStat struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Date         time.Time `json:"date" gorm:"uniqueIndex"`
	ViewCount    int       `json:"view_count" gorm:"default:0"`
	VisitorCount int       `json:"visitor_count" gorm:"default:0"`
}

// About 关于页面配置
type About struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	Name         string `json:"name" gorm:"size:100"`                            // 博主名字
	Bio          string `json:"bio" gorm:"size:255"`                             // 简介
	Avatar       string `json:"avatar" gorm:"size:255"`                          // 头像URL
	AboutMe      string `json:"about_me" gorm:"type:text"`                       // 关于我内容
	AboutMeHTML  string `json:"about_me_html" gorm:"type:text"`                  // 关于我HTML内容
	Skills       string `json:"skills" gorm:"size:500"`                          // 技术栈，逗号分隔
	Email        string `json:"email" gorm:"size:100"`                           // 邮箱
	Github       string `json:"github" gorm:"size:255"`                          // GitHub链接
	Twitter      string `json:"twitter" gorm:"size:255"`                         // Twitter链接
	Website      string `json:"website" gorm:"size:255"`                         // 个人网站
	UpdatedAt    time.Time `json:"updated_at"`
}

// OAuthConnection OAuth 关联表
type OAuthConnection struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UserID       uint      `json:"user_id" gorm:"index;not null"`
	User         User      `json:"user"`
	Provider     string    `json:"provider" gorm:"size:50;not null"`  // github, google, etc.
	ProviderID   string    `json:"provider_id" gorm:"size:255;not null"`
	AccessToken  string    `json:"-" gorm:"size:512"`
	RefreshToken string    `json:"-" gorm:"size:512"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// HtmlPage 导入的HTML页面
type HtmlPage struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Title       string         `json:"title" gorm:"size:200;not null;index"`
	Slug        string         `json:"slug" gorm:"uniqueIndex;size:200"`
	Content     string         `json:"content" gorm:"type:longtext"` // 完整 HTML 内容
	Summary     string         `json:"summary" gorm:"size:500"`
	CoverImage  string         `json:"cover_image" gorm:"size:255"`
	CategoryID  uint           `json:"category_id" gorm:"index"`
	Category    Category       `json:"category"`
	Tags        string         `json:"tags" gorm:"size:255"` // 逗号分隔的标签名
	ViewCount   int            `json:"view_count" gorm:"default:0"`
	IsPublished bool           `json:"is_published" gorm:"default:false;index"`
	CreatedAt   time.Time      `json:"created_at" gorm:"index"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword checks if the provided password matches the hashed password
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}