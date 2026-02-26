package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 案件状态
const (
	CaseStatusMissing     = "missing"     // 失踪中
	CaseStatusSearching   = "searching"   // 寻找中
	CaseStatusFound       = "found"       // 已找到
	CaseStatusReunited    = "reunited"    // 已团圆
	CaseStatusClosed      = "closed"      // 已结案
	CaseStatusPendingInfo = "pending_info" // 待补充信息
)

// 案件类型
const (
	CaseTypeElderly    = "elderly"    // 老人走失
	CaseTypeChild      = "child"      // 儿童走失
	CaseTypeAdult      = "adult"      // 成年人走失
	CaseTypeDisability = "disability" // 残障人士走失
	CaseTypeOther      = "other"      // 其他
)

// 走失人员信息
const (
	GenderMale   = "male"
	GenderFemale = "female"
	GenderOther  = "other"
)

// MissingPerson 走失人员信息
type MissingPerson struct {
	ID                uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CaseNo            string          `gorm:"size:50;uniqueIndex;comment:案件编号" json:"case_no"`
	Status            string          `gorm:"size:20;default:missing;comment:案件状态" json:"status"`
	CaseType          string          `gorm:"size:20;comment:案件类型" json:"case_type"`
	
	// 基本信息
	Name              string          `gorm:"size:50;not null;comment:姓名" json:"name"`
	Gender            string          `gorm:"size:10;comment:性别" json:"gender"`
	BirthDate         *time.Time      `gorm:"comment:出生日期" json:"birth_date"`
	Age               int             `gorm:"comment:年龄" json:"age"`
	Height            int             `gorm:"comment:身高(cm)" json:"height"`
	Weight            int             `gorm:"comment:体重(kg)" json:"weight"`
	IDCard            string          `gorm:"size:18;comment:身份证号" json:"id_card"`
	
	// 外貌特征
	Appearance        string          `gorm:"type:text;comment:外貌特征描述" json:"appearance"`
	Clothing          string          `gorm:"type:text;comment:衣着描述" json:"clothing"`
	SpecialFeatures   string          `gorm:"type:text;comment:特殊特征(胎记、疤痕等)" json:"special_features"`
	MentalStatus      string          `gorm:"size:50;comment:精神状态" json:"mental_status"`
	PhysicalStatus    string          `gorm:"size:50;comment:身体状态" json:"physical_status"`
	
	// 走失信息
	MissingTime       time.Time       `gorm:"not null;comment:走失时间" json:"missing_time"`
	MissingLocation   string          `gorm:"size:200;comment:走失地点" json:"missing_location"`
	MissingLongitude  float64         `gorm:"comment:经度" json:"missing_longitude"`
	MissingLatitude   float64         `gorm:"comment:纬度" json:"missing_latitude"`
	MissingDetail     string          `gorm:"type:text;comment:走失详情" json:"missing_detail"`
	PossibleLocation  string          `gorm:"size:200;comment:可能去向" json:"possible_location"`
	
	// 照片
	Photos            []MissingPhoto  `gorm:"foreignKey:MissingPersonID" json:"photos,omitempty"`
	
	// 家属信息
	ContactName       string          `gorm:"size:50;comment:联系人姓名" json:"contact_name"`
	ContactPhone      string          `gorm:"size:20;comment:联系人电话" json:"contact_phone"`
	ContactRelation   string          `gorm:"size:20;comment:联系人关系" json:"contact_relation"`
	ContactAddress    string          `gorm:"size:200;comment:联系人地址" json:"contact_address"`
	FamilyDescription string          `gorm:"type:text;comment:家庭情况描述" json:"family_description"`
	
	// 方言语音关联
	DialectIDs        []uuid.UUID     `gorm:"-" json:"dialect_ids,omitempty"`
	Dialects          []Dialect       `gorm:"many2many:missing_person_dialects;" json:"dialects,omitempty"`
	
	// 组织关联
	ReporterID        uuid.UUID       `gorm:"type:uuid;index;comment:报案人ID" json:"reporter_id"`
	Reporter          User            `gorm:"foreignKey:ReporterID" json:"reporter,omitempty"`
	OrgID             uuid.UUID       `gorm:"type:uuid;index;comment:所属组织ID" json:"org_id"`
	Org               Organization    `gorm:"foreignKey:OrgID" json:"org,omitempty"`
	
	// 结果信息
	FoundTime         *time.Time      `gorm:"comment:找到时间" json:"found_time"`
	FoundLocation     string          `gorm:"size:200;comment:找到地点" json:"found_location"`
	FoundDetail       string          `gorm:"type:text;comment:找到详情" json:"found_detail"`
	ReunionPhotos     []MissingPhoto  `gorm:"foreignKey:MissingPersonID" json:"reunion_photos,omitempty"`
	
	// 标签
	Tags              []Tag           `gorm:"many2many:missing_person_tags;" json:"tags,omitempty"`
	
	// 统计
	ViewCount         int             `gorm:"default:0;comment:浏览次数" json:"view_count"`
	ShareCount        int             `gorm:"default:0;comment:分享次数" json:"share_count"`
	TaskCount         int             `gorm:"default:0;comment:关联任务数" json:"task_count"`
	
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	DeletedAt         gorm.DeletedAt  `gorm:"index" json:"-"`
}

// MissingPhoto 走失人员照片
type MissingPhoto struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MissingPersonID uuid.UUID      `gorm:"type:uuid;index;not null" json:"missing_person_id"`
	URL             string         `gorm:"size:500;not null;comment:图片URL" json:"url"`
	ThumbnailURL    string         `gorm:"size:500;comment:缩略图URL" json:"thumbnail_url"`
	Type            string         `gorm:"size:20;default:normal;comment:图片类型(normal/reunion)" json:"type"`
	Description     string         `gorm:"size:200;comment:描述" json:"description"`
	Sort            int            `gorm:"default:0;comment:排序" json:"sort"`
	CreatedAt       time.Time      `json:"created_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// MissingPersonTrack 轨迹记录
type MissingPersonTrack struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MissingPersonID uuid.UUID      `gorm:"type:uuid;index;not null" json:"missing_person_id"`
	ReporterID      uuid.UUID      `gorm:"type:uuid;index;not null" json:"reporter_id"`
	Reporter        User           `gorm:"foreignKey:ReporterID" json:"reporter,omitempty"`
	TrackTime       time.Time      `gorm:"not null;comment:发现时间" json:"track_time"`
	Location        string         `gorm:"size:200;comment:地点" json:"location"`
	Longitude       float64        `gorm:"comment:经度" json:"longitude"`
	Latitude        float64        `gorm:"comment:纬度" json:"latitude"`
	Description     string         `gorm:"type:text;comment:描述" json:"description"`
	Photos          []string       `gorm:"type:jsonb;comment:照片URLs" json:"photos"`
	IsConfirmed     bool           `gorm:"default:false;comment:是否确认" json:"is_confirmed"`
	CreatedAt       time.Time      `json:"created_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (MissingPerson) TableName() string {
	return "missing_persons"
}

func (MissingPhoto) TableName() string {
	return "missing_photos"
}

func (MissingPersonTrack) TableName() string {
	return "missing_person_tracks"
}

// BeforeCreate 创建前钩子
func (m *MissingPerson) BeforeCreate(tx *gorm.DB) error {
	if m.CaseNo == "" {
		// 生成案件编号: CASE + 年月日 + 6位序号
		m.CaseNo = generateCaseNo()
	}
	return nil
}

func generateCaseNo() string {
	return "CASE" + time.Now().Format("20060102") + uuid.New().String()[:6]
}
