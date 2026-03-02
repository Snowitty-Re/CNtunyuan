package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// DialectStatus 方言状态
type DialectStatus string

const (
	DialectStatusActive   DialectStatus = "active"
	DialectStatusInactive DialectStatus = "inactive"
	DialectStatusPending  DialectStatus = "pending"
)

// DialectType 方言类型
type DialectType string

const (
	DialectTypePhrase  DialectType = "phrase"  // 短语
	DialectTypeStory   DialectType = "story"   // 故事
	DialectTypeSong    DialectType = "song"    // 歌曲
	DialectTypeDaily   DialectType = "daily"   // 日常用语
	DialectTypeOther   DialectType = "other"   // 其他
)

// Dialect 方言语音领域实体
type Dialect struct {
	BaseEntity
	Title       string        `gorm:"size:100;not null" json:"title"`
	Content     string        `gorm:"type:text" json:"content,omitempty"`
	Region      string        `gorm:"size:100" json:"region"`
	Province    string        `gorm:"size:50" json:"province,omitempty"`
	City        string        `gorm:"size:50" json:"city,omitempty"`
	DialectType DialectType   `gorm:"size:20;default:'phrase'" json:"dialect_type"`
	AudioUrl    string        `gorm:"size:255;not null" json:"audio_url"`
	Duration    int           `json:"duration"` // 秒
	FileSize    int           `json:"file_size"` // 字节
	Format      string        `gorm:"size:10" json:"format,omitempty"` // mp3, wav, etc.
	Status      DialectStatus `gorm:"size:20;default:'active'" json:"status"`
	IsFeatured  bool          `gorm:"default:false" json:"is_featured"`
	PlayCount   int           `gorm:"default:0" json:"play_count"`
	LikeCount   int           `gorm:"default:0" json:"like_count"`
	CommentCount int          `gorm:"default:0" json:"comment_count"`
	Tags        string        `gorm:"type:json" json:"tags,omitempty"`
	Description string        `gorm:"type:text" json:"description,omitempty"`
	
	// 关联
	UploaderID  string         `gorm:"type:uuid;not null;index" json:"uploader_id"`
	OrgID       string         `gorm:"type:uuid;not null;index" json:"org_id"`
	
	Uploader    *User          `gorm:"foreignKey:UploaderID" json:"uploader,omitempty"`
	Org         *Organization  `gorm:"foreignKey:OrgID" json:"org,omitempty"`
}

// TableName 表名
func (Dialect) TableName() string {
	return "ty_dialects"
}

// Validate 验证
func (d *Dialect) Validate() error {
	if d.Title == "" {
		return errors.New("标题不能为空")
	}
	if d.AudioUrl == "" {
		return errors.New("音频URL不能为空")
	}
	if d.Region == "" {
		return errors.New("地区不能为空")
	}
	if d.Duration <= 0 || d.Duration > 300 {
		return errors.New("音频时长必须在1-300秒之间")
	}
	return nil
}

// IsActive 是否活跃
func (d *Dialect) IsActive() bool {
	return d.Status == DialectStatusActive
}

// CanPlay 是否可以播放
func (d *Dialect) CanPlay() bool {
	return d.Status == DialectStatusActive
}

// IncrementPlayCount 增加播放次数
func (d *Dialect) IncrementPlayCount() {
	d.PlayCount++
}

// IncrementLikeCount 增加点赞数
func (d *Dialect) IncrementLikeCount() {
	d.LikeCount++
}

// DecrementLikeCount 减少点赞数
func (d *Dialect) DecrementLikeCount() {
	if d.LikeCount > 0 {
		d.LikeCount--
	}
}

// IncrementCommentCount 增加评论数
func (d *Dialect) IncrementCommentCount() {
	d.CommentCount++
}

// DecrementCommentCount 减少评论数
func (d *Dialect) DecrementCommentCount() {
	if d.CommentCount > 0 {
		d.CommentCount--
	}
}

// Feature 设为精选
func (d *Dialect) Feature() {
	d.IsFeatured = true
}

// Unfeature 取消精选
func (d *Dialect) Unfeature() {
	d.IsFeatured = false
}

// Approve 审核通过
func (d *Dialect) Approve() {
	d.Status = DialectStatusActive
}

// Reject 拒绝
func (d *Dialect) Reject() {
	d.Status = DialectStatusInactive
}

// DialectComment 方言评论
type DialectComment struct {
	BaseEntity
	DialectID   string `gorm:"type:uuid;not null;index" json:"dialect_id"`
	UserID      string `gorm:"type:uuid;not null" json:"user_id"`
	Content     string `gorm:"type:text;not null" json:"content"`
	ParentID    *string `gorm:"type:uuid;index" json:"parent_id,omitempty"`
	ReplyCount  int    `gorm:"default:0" json:"reply_count"`
	LikeCount   int    `gorm:"default:0" json:"like_count"`
	
	User        *User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 表名
func (DialectComment) TableName() string {
	return "ty_dialect_comments"
}

// DialectLike 方言点赞
type DialectLike struct {
	ID        string    `gorm:"type:uuid;primaryKey" json:"id"`
	DialectID string    `gorm:"type:uuid;not null;index:idx_dialect_user,unique" json:"dialect_id"`
	UserID    string    `gorm:"type:uuid;not null;index:idx_dialect_user,unique" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName 表名
func (DialectLike) TableName() string {
	return "ty_dialect_likes"
}

// DialectPlayLog 方言播放记录
type DialectPlayLog struct {
	ID        string    `gorm:"type:uuid;primaryKey" json:"id"`
	DialectID string    `gorm:"type:uuid;not null;index" json:"dialect_id"`
	UserID    *string   `gorm:"type:uuid;index" json:"user_id,omitempty"`
	IP        string    `gorm:"size:50" json:"ip,omitempty"`
	UserAgent string    `gorm:"size:255" json:"user_agent,omitempty"`
	Duration  int       `json:"duration"` // 播放时长（秒）
	CreatedAt time.Time `json:"created_at"`
}

// TableName 表名
func (DialectPlayLog) TableName() string {
	return "ty_dialect_play_logs"
}

// DialectStats 方言统计
type DialectStats struct {
	Total          int64 `json:"total"`
	Active         int64 `json:"active"`
	Pending        int64 `json:"pending"`
	Featured       int64 `json:"featured"`
	TotalPlays     int64 `json:"total_plays"`
	TotalLikes     int64 `json:"total_likes"`
	TotalComments  int64 `json:"total_comments"`
	TodayUploads   int64 `json:"today_uploads"`
	ThisWeekUploads int64 `json:"this_week_uploads"`
}

// NewDialect 创建新方言
func NewDialect(title, region, audioUrl, uploaderID, orgID string, duration int) (*Dialect, error) {
	d := &Dialect{
		BaseEntity: BaseEntity{
			ID: uuid.New().String(),
		},
		Title:       title,
		Region:      region,
		AudioUrl:    audioUrl,
		UploaderID:  uploaderID,
		OrgID:       orgID,
		Duration:    duration,
		Status:      DialectStatusPending,
		DialectType: DialectTypePhrase,
	}

	if err := d.Validate(); err != nil {
		return nil, err
	}

	return d, nil
}
