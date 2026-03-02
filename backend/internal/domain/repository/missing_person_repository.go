package repository

import (
	"context"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
)

// MissingPersonRepository 走失人员仓储接口
type MissingPersonRepository interface {
	Repository[entity.MissingPerson]

	// List 分页查询
	List(ctx context.Context, query *MissingPersonQuery) (*PageResult[entity.MissingPerson], error)

	// FindByReporterID 根据报告人查找
	FindByReporterID(ctx context.Context, reporterID string, pagination Pagination) (*PageResult[entity.MissingPerson], error)

	// FindByStatus 根据状态查找
	FindByStatus(ctx context.Context, status entity.MissingStatus, pagination Pagination) (*PageResult[entity.MissingPerson], error)

	// FindByRegion 根据地区查找
	FindByRegion(ctx context.Context, province, city, district string, pagination Pagination) (*PageResult[entity.MissingPerson], error)

	// UpdateStatus 更新状态
	UpdateStatus(ctx context.Context, id string, status entity.MissingStatus) error

	// AddTrack 添加轨迹
	AddTrack(ctx context.Context, track *entity.MissingPersonTrack) error

	// GetTracks 获取轨迹
	GetTracks(ctx context.Context, personID string) ([]entity.MissingPersonTrack, error)

	// GetStats 获取统计
	GetStats(ctx context.Context) (*entity.MissingPersonStats, error)

	// Search 全文搜索
	Search(ctx context.Context, keyword string, pagination Pagination) (*PageResult[entity.MissingPerson], error)

	// CountByStatus 按状态统计
	CountByStatus(ctx context.Context, status entity.MissingStatus) (int64, error)

	// CountByDateRange 按日期范围统计
	CountByDateRange(ctx context.Context, start, end string) (int64, error)

	// IncrementViews 增加浏览次数
	IncrementViews(ctx context.Context, id string) error
}

// MissingPersonQuery 走失人员查询参数
type MissingPersonQuery struct {
	Pagination
	Keyword      string             `json:"keyword"`
	Status       entity.MissingStatus `json:"status"`
	Gender       string             `json:"gender"`
	AgeMin       int                `json:"age_min"`
	AgeMax       int                `json:"age_max"`
	Province     string             `json:"province"`
	City         string             `json:"city"`
	District     string             `json:"district"`
	MissingDate  string             `json:"missing_date"`
	UrgencyLevel string             `json:"urgency_level"`
	SortField    string             `json:"sort_field"`
	SortOrder    string             `json:"sort_order"`
}

// NewMissingPersonQuery 创建默认查询
func NewMissingPersonQuery() *MissingPersonQuery {
	return &MissingPersonQuery{
		Pagination: Pagination{
			Page:     1,
			PageSize: 10,
		},
		SortField: "created_at",
		SortOrder: "desc",
	}
}
