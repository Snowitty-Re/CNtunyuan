package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Dialect 方言语音记录
type Dialect struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Title       string         `gorm:"size:100;not null;index:idx_dialect_title;comment:标题" json:"title"`
	Description string         `gorm:"type:text;comment:描述" json:"description"`

	// 语音文件
	AudioURL string `gorm:"size:500;not null;comment:音频URL" json:"audio_url"`
	Duration int    `gorm:"index:idx_dialect_duration;comment:时长(秒)" json:"duration"`
	FileSize int    `gorm:"comment:文件大小(字节)" json:"file_size"`
	Format   string `gorm:"size:10;comment:格式(mp3/wav)" json:"format"`

	// 地区信息
	Province  string  `gorm:"size:50;index:idx_dialect_province;comment:省" json:"province"`
	City      string  `gorm:"size:50;index:idx_dialect_city;comment:市" json:"city"`
	District  string  `gorm:"size:50;index:idx_dialect_district;comment:区" json:"district"`
	Town      string  `gorm:"size:50;comment:镇/街道" json:"town"`
	Village   string  `gorm:"size:50;comment:村/社区" json:"village"`
	Longitude float64 `gorm:"comment:经度" json:"longitude"`
	Latitude  float64 `gorm:"comment:纬度" json:"latitude"`
	Address   string  `gorm:"size:200;comment:详细地址" json:"address"`

	// 标签
	Tags []Tag `gorm:"many2many:dialect_tags;" json:"tags,omitempty"`

	// 采集人信息
	CollectorID uuid.UUID    `gorm:"type:uuid;index:idx_dialect_collector;comment:采集人ID" json:"collector_id"`
	Collector   User         `gorm:"foreignKey:CollectorID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"collector,omitempty"`
	OrgID       uuid.UUID    `gorm:"type:uuid;index:idx_dialect_org;comment:所属组织ID" json:"org_id"`
	Org         Organization `gorm:"foreignKey:OrgID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"org,omitempty"`

	// 采集信息
	RecordTime *time.Time `gorm:"comment:录音时间" json:"record_time"`
	Weather    string     `gorm:"size:20;comment:天气" json:"weather"`
	Device     string     `gorm:"size:100;comment:录音设备" json:"device"`

	// 关联走失人员
	MissingPersons []MissingPerson `gorm:"many2many:missing_person_dialects;" json:"missing_persons,omitempty"`

	// 统计
	PlayCount int    `gorm:"default:0;index:idx_dialect_play;comment:播放次数" json:"play_count"`
	LikeCount int    `gorm:"default:0;comment:点赞次数" json:"like_count"`
	Status    string `gorm:"size:20;default:active;index:idx_dialect_status;comment:状态" json:"status"`

	CreatedAt time.Time      `gorm:"index:idx_dialect_created" json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// DialectComment 方言评论
type DialectComment struct {
	ID         uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	DialectID  uuid.UUID      `gorm:"type:uuid;index:idx_dialectcmt_dialect;not null" json:"dialect_id"`
	Dialect    Dialect        `gorm:"foreignKey:DialectID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	UserID     uuid.UUID      `gorm:"type:uuid;index:idx_dialectcmt_user;not null" json:"user_id"`
	User       User           `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"user,omitempty"`
	Content    string         `gorm:"type:text;not null;comment:内容" json:"content"`
	ParentID   *uuid.UUID     `gorm:"type:uuid;index:idx_dialectcmt_parent;comment:父评论ID" json:"parent_id"`
	ReplyCount int            `gorm:"default:0;comment:回复数" json:"reply_count"`
	CreatedAt  time.Time      `gorm:"index:idx_dialectcmt_created" json:"created_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// DialectLike 方言点赞
type DialectLike struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	DialectID uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_dialectlike_unique;not null" json:"dialect_id"`
	Dialect   Dialect   `gorm:"foreignKey:DialectID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	UserID    uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_dialectlike_unique;not null" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

// DialectPlayLog 方言播放记录
type DialectPlayLog struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	DialectID uuid.UUID `gorm:"type:uuid;index:idx_dialectplay_dialect;not null" json:"dialect_id"`
	Dialect   Dialect   `gorm:"foreignKey:DialectID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	UserID    uuid.UUID `gorm:"type:uuid;index:idx_dialectplay_user;not null" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	IP        string    `gorm:"size:50;comment:IP地址" json:"ip"`
	Duration  int       `gorm:"comment:播放时长" json:"duration"`
	CreatedAt time.Time `gorm:"index:idx_dialectplay_created" json:"created_at"`
}

// TableName 指定表名
func (Dialect) TableName() string {
	return "dialects"
}

func (DialectComment) TableName() string {
	return "dialect_comments"
}

func (DialectLike) TableName() string {
	return "dialect_likes"
}

func (DialectPlayLog) TableName() string {
	return "dialect_play_logs"
}

// IsValidDuration 检查时长是否有效(15-20秒)
func (d *Dialect) IsValidDuration() bool {
	return d.Duration >= 15 && d.Duration <= 20
}

// IncrementPlayCount 增加播放次数
func (d *Dialect) IncrementPlayCount() {
	d.PlayCount++
}

// IncrementLikeCount 增加点赞次数
func (d *Dialect) IncrementLikeCount() {
	d.LikeCount++
}

// DecrementLikeCount 减少点赞次数
func (d *Dialect) DecrementLikeCount() {
	if d.LikeCount > 0 {
		d.LikeCount--
	}
}
