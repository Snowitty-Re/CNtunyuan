package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 组织类型
const (
	OrgTypeRoot     = "root"     // 团圆机构(根)
	OrgTypeProvince = "province" // 省级
	OrgTypeCity     = "city"     // 市级
	OrgTypeDistrict = "district" // 区级
	OrgTypeStreet   = "street"   // 街道
)

// 组织状态
const (
	OrgStatusActive   = "active"   // 正常
	OrgStatusInactive = "inactive" // 停用
)

// Organization 组织架构模型
type Organization struct {
	ID             uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name           string          `gorm:"size:100;not null;index:idx_org_name;comment:组织名称" json:"name"`
	Code           string          `gorm:"size:50;uniqueIndex:idx_org_code;comment:组织编码" json:"code"`
	Type           string          `gorm:"size:20;not null;index:idx_org_type;comment:组织类型" json:"type"`
	Level          int             `gorm:"not null;index:idx_org_level;comment:层级(1-5)" json:"level"`
	ParentID       *uuid.UUID      `gorm:"type:uuid;index:idx_org_parent;comment:父级ID" json:"parent_id"`
	Parent         *Organization   `gorm:"foreignKey:ParentID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"parent,omitempty"`
	Children       []Organization  `gorm:"foreignKey:ParentID;references:ID;" json:"children,omitempty"`
	LeaderID       *uuid.UUID      `gorm:"type:uuid;index:idx_org_leader;comment:负责人ID" json:"leader_id"`
	Leader         *User           `gorm:"foreignKey:LeaderID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"leader,omitempty"`
	Province       string          `gorm:"size:50;index:idx_org_province;comment:省" json:"province"`
	City           string          `gorm:"size:50;index:idx_org_city;comment:市" json:"city"`
	District       string          `gorm:"size:50;index:idx_org_district;comment:区" json:"district"`
	Street         string          `gorm:"size:50;comment:街道" json:"street"`
	Address        string          `gorm:"size:200;comment:详细地址" json:"address"`
	Contact        string          `gorm:"size:50;comment:联系人" json:"contact"`
	Phone          string          `gorm:"size:20;comment:联系电话" json:"phone"`
	Email          string          `gorm:"size:100;comment:邮箱" json:"email"`
	Description    string          `gorm:"type:text;comment:描述" json:"description"`
	Sort           int             `gorm:"default:0;index:idx_org_sort;comment:排序" json:"sort"`
	Status         string          `gorm:"size:20;default:active;index:idx_org_status;comment:状态" json:"status"`
	VolunteerCount int             `gorm:"default:0;comment:志愿者数量" json:"volunteer_count"`
	CaseCount      int             `gorm:"default:0;comment:案件数量" json:"case_count"`
	CreatedAt      time.Time       `gorm:"index:idx_org_created" json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	DeletedAt      gorm.DeletedAt  `gorm:"index" json:"-"`
}

// OrgStats 组织统计
type OrgStats struct {
	ID               uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OrgID            uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_orgstats_org;not null" json:"org_id"`
	Org              Organization `gorm:"foreignKey:OrgID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	TotalCases       int       `gorm:"default:0;comment:总案件数" json:"total_cases"`
	ResolvedCases    int       `gorm:"default:0;comment:已解决案件数" json:"resolved_cases"`
	PendingCases     int       `gorm:"default:0;comment:待处理案件数" json:"pending_cases"`
	TotalTasks       int       `gorm:"default:0;comment:总任务数" json:"total_tasks"`
	CompletedTasks   int       `gorm:"default:0;comment:已完成任务数" json:"completed_tasks"`
	TotalVolunteers  int       `gorm:"default:0;comment:总志愿者数" json:"total_volunteers"`
	ActiveVolunteers int       `gorm:"default:0;comment:活跃志愿者数" json:"active_volunteers"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// TableName 指定表名
func (Organization) TableName() string {
	return "organizations"
}

func (OrgStats) TableName() string {
	return "org_stats"
}

// GetFullPath 获取完整路径
func (o *Organization) GetFullPath(db *gorm.DB) (string, error) {
	var path []string
	current := o
	for current != nil {
		path = append([]string{current.Name}, path...)
		if current.ParentID == nil {
			break
		}
		var parent Organization
		if err := db.First(&parent, current.ParentID).Error; err != nil {
			return "", err
		}
		current = &parent
	}
	result := ""
	for i, p := range path {
		if i > 0 {
			result += " > "
		}
		result += p
	}
	return result, nil
}
