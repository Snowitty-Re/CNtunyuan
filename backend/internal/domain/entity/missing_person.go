package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// MissingStatus 走失状态
type MissingStatus string

const (
	MissingStatusMissing   MissingStatus = "missing"
	MissingStatusSearching MissingStatus = "searching"
	MissingStatusFound     MissingStatus = "found"
	MissingStatusReunited  MissingStatus = "reunited"
	MissingStatusClosed    MissingStatus = "closed"
)

// UrgencyLevel 紧急程度
type UrgencyLevel string

const (
	UrgencyLevelCritical UrgencyLevel = "critical" // 紧急
	UrgencyLevelHigh     UrgencyLevel = "high"     // 高
	UrgencyLevelMedium   UrgencyLevel = "medium"   // 中
	UrgencyLevelLow      UrgencyLevel = "low"      // 低
)

// MissingPerson 走失人员领域实体
type MissingPerson struct {
	BaseEntity
	Name         string           `gorm:"size:50;not null" json:"name"`
	Gender       string           `gorm:"size:10;not null" json:"gender"`
	BirthDate    *time.Time       `json:"birth_date,omitempty"`
	Age          int              `json:"age"`
	Height       int              `json:"height,omitempty"`
	Weight       int              `json:"weight,omitempty"`
	Description  string           `gorm:"type:text" json:"description,omitempty"`
	PhotoUrl     string           `gorm:"size:255" json:"photo_url,omitempty"`
	
	// 走失信息
	MissingTime  time.Time        `json:"missing_time"`
	Province     string           `gorm:"size:50" json:"province,omitempty"`
	City         string           `gorm:"size:50" json:"city,omitempty"`
	District     string           `gorm:"size:50" json:"district,omitempty"`
	Address      string           `gorm:"size:255" json:"address,omitempty"`
	Clothes      string           `gorm:"type:text" json:"clothes,omitempty"`
	Features     string           `gorm:"type:text" json:"features,omitempty"`
	
	// 联系人信息
	ContactName  string           `gorm:"size:50" json:"contact_name"`
	ContactPhone string           `gorm:"size:20" json:"contact_phone"`
	ContactRel   string           `gorm:"size:20" json:"contact_rel"`
	AltContact   string           `gorm:"size:20" json:"alt_contact,omitempty"`
	
	// 状态和统计
	Status       MissingStatus    `gorm:"size:20;not null;default:'missing'" json:"status"`
	Urgency      UrgencyLevel     `gorm:"size:20;default:'medium'" json:"urgency"`
	Views        int              `gorm:"default:0" json:"views"`
	ShareCount   int              `gorm:"default:0" json:"share_count"`
	
	// 关联
	ReporterID   string           `gorm:"type:uuid;not null;index" json:"reporter_id"`
	OrgID        string           `gorm:"type:uuid;not null;index" json:"org_id"`
	AssignedTo   *string          `gorm:"type:uuid;index" json:"assigned_to,omitempty"`
	
	// 找到信息
	FoundTime    *time.Time       `json:"found_time,omitempty"`
	FoundLocation string          `gorm:"size:255" json:"found_location,omitempty"`
	FoundNote    string           `gorm:"type:text" json:"found_note,omitempty"`
	
	Reporter     *User            `gorm:"foreignKey:ReporterID" json:"reporter,omitempty"`
	Org          *Organization    `gorm:"foreignKey:OrgID" json:"org,omitempty"`
	Assignee     *User            `gorm:"foreignKey:AssignedTo" json:"assignee,omitempty"`
	Tracks       []MissingPersonTrack `json:"tracks,omitempty"`
}

// TableName 表名
func (MissingPerson) TableName() string {
	return "ty_missing_persons"
}

// Validate 验证
func (m *MissingPerson) Validate() error {
	if m.Name == "" {
		return errors.New("姓名不能为空")
	}
	if m.Gender == "" {
		return errors.New("性别不能为空")
	}
	if m.ContactName == "" {
		return errors.New("联系人姓名不能为空")
	}
	if m.ContactPhone == "" {
		return errors.New("联系人电话不能为空")
	}
	if m.MissingTime.IsZero() {
		return errors.New("走失时间不能为空")
	}
	return nil
}

// IsActive 是否活跃（寻找中）
func (m *MissingPerson) IsActive() bool {
	return m.Status == MissingStatusMissing || m.Status == MissingStatusSearching
}

// IsFound 是否已找到
func (m *MissingPerson) IsFound() bool {
	return m.Status == MissingStatusFound || m.Status == MissingStatusReunited
}

// CanUpdate 是否可以更新
func (m *MissingPerson) CanUpdate() bool {
	return m.Status != MissingStatusClosed
}

// StartSearch 开始搜索
func (m *MissingPerson) StartSearch() error {
	if m.Status != MissingStatusMissing {
		return errors.New("只有待寻找状态才能开始搜索")
	}
	m.Status = MissingStatusSearching
	return nil
}

// MarkFound 标记为已找到
func (m *MissingPerson) MarkFound(location, note string) error {
	if m.IsFound() {
		return errors.New("该案件已被标记为找到")
	}
	now := time.Now()
	m.Status = MissingStatusFound
	m.FoundTime = &now
	m.FoundLocation = location
	m.FoundNote = note
	return nil
}

// MarkReunited 标记为已团聚
func (m *MissingPerson) MarkReunited() error {
	if m.Status != MissingStatusFound {
		return errors.New("只有已找到状态才能标记为团聚")
	}
	m.Status = MissingStatusReunited
	return nil
}

// Close 关闭案件
func (m *MissingPerson) Close(reason string) error {
	if m.Status == MissingStatusClosed {
		return errors.New("案件已关闭")
	}
	m.Status = MissingStatusClosed
	return nil
}

// AssignTo 分配给某人
func (m *MissingPerson) AssignTo(userID string) {
	m.AssignedTo = &userID
}

// IncrementViews 增加浏览次数
func (m *MissingPerson) IncrementViews() {
	m.Views++
}

// GetAgeAtMissing 计算走失时的年龄
func (m *MissingPerson) GetAgeAtMissing() int {
	if m.BirthDate == nil {
		return m.Age
	}
	age := m.MissingTime.Year() - m.BirthDate.Year()
	if m.MissingTime.YearDay() < m.BirthDate.YearDay() {
		age--
	}
	return age
}

// MissingPersonTrack 轨迹记录
type MissingPersonTrack struct {
	BaseEntity
	MissingPersonID string    `gorm:"type:uuid;not null;index" json:"missing_person_id"`
	ReporterID      string    `gorm:"type:uuid;not null" json:"reporter_id"`
	Location        string    `gorm:"size:255" json:"location"`
	Province        string    `gorm:"size:50" json:"province,omitempty"`
	City            string    `gorm:"size:50" json:"city,omitempty"`
	District        string    `gorm:"size:50" json:"district,omitempty"`
	Address         string    `gorm:"size:255" json:"address,omitempty"`
	Time            time.Time `json:"time"`
	Description     string    `gorm:"type:text" json:"description"`
	Photos          string    `gorm:"type:json" json:"photos,omitempty"`
	VideoUrl        string    `gorm:"size:255" json:"video_url,omitempty"`
	AudioUrl        string    `gorm:"size:255" json:"audio_url,omitempty"`
	Lat             float64   `json:"lat,omitempty"`
	Lng             float64   `json:"lng,omitempty"`
	Status          string    `gorm:"size:20;default:'pending'" json:"status"`
	IsKeyPoint      bool      `gorm:"default:false" json:"is_key_point"`
	
	Reporter        *User           `gorm:"foreignKey:ReporterID" json:"reporter,omitempty"`
}

// TableName 表名
func (MissingPersonTrack) TableName() string {
	return "ty_missing_person_tracks"
}

// MissingPersonStats 走失人员统计
type MissingPersonStats struct {
	Total           int64 `json:"total"`
	Missing         int64 `json:"missing"`
	Searching       int64 `json:"searching"`
	Found           int64 `json:"found"`
	Reunited        int64 `json:"reunited"`
	Closed          int64 `json:"closed"`
	TodayNew        int64 `json:"today_new"`
	ThisWeekNew     int64 `json:"this_week_new"`
	ThisMonthNew    int64 `json:"this_month_new"`
}

// NewMissingPerson 创建新案件
func NewMissingPerson(name, gender, contactName, contactPhone, reporterID, orgID string) (*MissingPerson, error) {
	mp := &MissingPerson{
		BaseEntity: BaseEntity{
			ID: uuid.New().String(),
		},
		Name:         name,
		Gender:       gender,
		ContactName:  contactName,
		ContactPhone: contactPhone,
		ReporterID:   reporterID,
		OrgID:        orgID,
		Status:       MissingStatusMissing,
		Urgency:      UrgencyLevelMedium,
	}

	if err := mp.Validate(); err != nil {
		return nil, err
	}

	return mp, nil
}
