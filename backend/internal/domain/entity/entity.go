package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseEntity 基础实体
type BaseEntity struct {
	ID        string         `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate 创建前自动生成 UUID
func (e *BaseEntity) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	return nil
}

// SetID 设置 ID
func (e *BaseEntity) SetID(id string) {
	e.ID = id
}

// GetID 获取 ID
func (e *BaseEntity) GetID() string {
	return e.ID
}

// GetCreatedAt 获取创建时间
func (e *BaseEntity) GetCreatedAt() time.Time {
	return e.CreatedAt
}

// GetUpdatedAt 获取更新时间
func (e *BaseEntity) GetUpdatedAt() time.Time {
	return e.UpdatedAt
}

// IsDeleted 是否已删除
func (e *BaseEntity) IsDeleted() bool {
	return e.DeletedAt.Valid
}
