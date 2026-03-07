package entity

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// OrgType 组织类型
type OrgType string

const (
	OrgTypeRoot      OrgType = "root"
	OrgTypeProvince  OrgType = "province"
	OrgTypeCity      OrgType = "city"
	OrgTypeDistrict  OrgType = "district"
	OrgTypeStreet    OrgType = "street"
	OrgTypeCommunity OrgType = "community"
	OrgTypeTeam      OrgType = "team"
)

// OrgStatus 组织状态
type OrgStatus string

const (
	OrgStatusActive   OrgStatus = "active"
	OrgStatusInactive OrgStatus = "inactive"
)

// Organization 组织领域实体
type Organization struct {
	BaseEntity
	Name         string         `gorm:"size:100;not null" json:"name"`
	Code         string         `gorm:"size:50;uniqueIndex;not null" json:"code"`
	Type         OrgType        `gorm:"size:20;not null" json:"type"`
	Level        int            `gorm:"not null;default:1" json:"level"`
	ParentID     *string        `gorm:"type:uuid;index" json:"parent_id,omitempty"`
	Description  string         `gorm:"type:text" json:"description,omitempty"`
	Address      string         `gorm:"size:255" json:"address,omitempty"`
	ContactName  string         `gorm:"size:50" json:"contact_name,omitempty"`
	ContactPhone string         `gorm:"size:20" json:"contact_phone,omitempty"`
	Status       OrgStatus      `gorm:"size:20;not null;default:'active'" json:"status"`
	Logo         string         `gorm:"size:255" json:"logo,omitempty"`
	SortOrder    int            `gorm:"default:0" json:"sort_order"`
	Parent       *Organization  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children     []Organization `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Stats        *OrgStats      `gorm:"-" json:"stats,omitempty"`
}

// TableName 表名
func (Organization) TableName() string {
	return "ty_organizations"
}

// Validate 验证组织
func (o *Organization) Validate() error {
	if o.Name == "" {
		return errors.New("组织名称不能为空")
	}
	if o.Code == "" {
		return errors.New("组织编码不能为空")
	}
	if !isValidOrgType(o.Type) {
		return fmt.Errorf("无效的组织类型: %s", o.Type)
	}
	return nil
}

// IsActive 是否活跃
func (o *Organization) IsActive() bool {
	return o.Status == OrgStatusActive
}

// IsRoot 是否是根组织
func (o *Organization) IsRoot() bool {
	return o.Type == OrgTypeRoot || o.ParentID == nil
}

// CanHaveChildren 是否可以拥有子组织
func (o *Organization) CanHaveChildren() bool {
	switch o.Type {
	case OrgTypeRoot, OrgTypeProvince, OrgTypeCity, OrgTypeDistrict, OrgTypeStreet:
		return true
	default:
		return false
	}
}

// GetLevelName 获取层级名称
func (o *Organization) GetLevelName() string {
	switch o.Type {
	case OrgTypeRoot:
		return "总部"
	case OrgTypeProvince:
		return "省级"
	case OrgTypeCity:
		return "市级"
	case OrgTypeDistrict:
		return "区级"
	case OrgTypeStreet:
		return "街道"
	case OrgTypeCommunity:
		return "社区"
	case OrgTypeTeam:
		return "团队"
	default:
		return "未知"
	}
}

// SetParent 设置父组织
func (o *Organization) SetParent(parentID string) {
	o.ParentID = &parentID
	o.Level = o.calculateLevel()
}

// calculateLevel 计算层级
func (o *Organization) calculateLevel() int {
	switch o.Type {
	case OrgTypeRoot:
		return 1
	case OrgTypeProvince:
		return 2
	case OrgTypeCity:
		return 3
	case OrgTypeDistrict:
		return 4
	case OrgTypeStreet:
		return 5
	case OrgTypeCommunity:
		return 6
	default:
		return 7
	}
}

// OrgStats 组织统计
type OrgStats struct {
	ID               string `gorm:"type:uuid;primaryKey" json:"id"`
	OrgID            string `gorm:"type:uuid;uniqueIndex;not null" json:"org_id"`
	TotalVolunteers  int    `json:"total_volunteers"`
	ActiveVolunteers int    `json:"active_volunteers"`
	TotalCases       int    `json:"total_cases"`
	ActiveCases      int    `json:"active_cases"`
	CompletedCases   int    `json:"completed_cases"`
	TotalTasks       int    `json:"total_tasks"`
	PendingTasks     int    `json:"pending_tasks"`
	BaseEntity
}

// TableName 表名
func (OrgStats) TableName() string {
	return "ty_org_stats"
}

// OrgTreeNode 组织树节点
type OrgTreeNode struct {
	Organization
	Children []*OrgTreeNode `json:"children,omitempty"`
}

// isValidOrgType 验证组织类型
func isValidOrgType(t OrgType) bool {
	switch t {
	case OrgTypeRoot, OrgTypeProvince, OrgTypeCity, OrgTypeDistrict, OrgTypeStreet, OrgTypeCommunity, OrgTypeTeam:
		return true
	default:
		return false
	}
}

// NewOrganization 创建新组织
func NewOrganization(name, code string, orgType OrgType, parentID *string) (*Organization, error) {
	org := &Organization{
		BaseEntity: BaseEntity{
			ID: uuid.New().String(),
		},
		Name:     name,
		Code:     code,
		Type:     orgType,
		ParentID: parentID,
		Status:   OrgStatusActive,
	}
	org.Level = org.calculateLevel()

	if err := org.Validate(); err != nil {
		return nil, err
	}

	return org, nil
}

// NewRootOrganization 创建根组织
func NewRootOrganization(name, code string) (*Organization, error) {
	rootID := "00000000-0000-0000-0000-000000000000"
	org := &Organization{
		BaseEntity: BaseEntity{
			ID: rootID,
		},
		Name:   name,
		Code:   code,
		Type:   OrgTypeRoot,
		Level:  1,
		Status: OrgStatusActive,
	}

	if err := org.Validate(); err != nil {
		return nil, err
	}

	return org, nil
}
