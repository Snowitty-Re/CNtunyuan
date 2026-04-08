package repository

import (
	"context"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
)

// DialectRepository 方言仓储接口
type DialectRepository interface {
	Repository[entity.Dialect]

	// List 分页查询
	List(ctx context.Context, query *DialectQuery) (*PageResult[entity.Dialect], error)

	// FindByRegion 根据地区查找
	FindByRegion(ctx context.Context, province, city string, pagination Pagination) (*PageResult[entity.Dialect], error)

	// FindByUploader 根据上传者查找
	FindByUploader(ctx context.Context, uploaderID string, pagination Pagination) (*PageResult[entity.Dialect], error)

	// FindFeatured 查找精选
	FindFeatured(ctx context.Context, pagination Pagination) (*PageResult[entity.Dialect], error)

	// FindByType 根据类型查找
	FindByType(ctx context.Context, dialectType entity.DialectType, pagination Pagination) (*PageResult[entity.Dialect], error)

	// Search 搜索
	Search(ctx context.Context, keyword string, pagination Pagination) (*PageResult[entity.Dialect], error)

	// IncrementPlayCount 增加播放次数
	IncrementPlayCount(ctx context.Context, id string) error

	// IncrementLikeCount 增加点赞数
	IncrementLikeCount(ctx context.Context, id string) error

	// DecrementLikeCount 减少点赞数
	DecrementLikeCount(ctx context.Context, id string) error

	// AddComment 添加评论
	AddComment(ctx context.Context, comment *entity.DialectComment) error

	// GetComments 获取评论
	GetComments(ctx context.Context, dialectID string, pagination Pagination) (*PageResult[entity.DialectComment], error)

	// AddLike 添加点赞
	AddLike(ctx context.Context, like *entity.DialectLike) error

	// RemoveLike 取消点赞
	RemoveLike(ctx context.Context, dialectID, userID string) error

	// HasLiked 是否已点赞
	HasLiked(ctx context.Context, dialectID, userID string) (bool, error)

	// AddPlayLog 添加播放记录
	AddPlayLog(ctx context.Context, log *entity.DialectPlayLog) error

	// GetStats 获取统计
	GetStats(ctx context.Context) (*entity.DialectStats, error)

	// ListCardGroups 获取卡片分组列表
	ListCardGroups(ctx context.Context, includeInactive bool) ([]entity.DialectCardGroup, error)

	// CreateCardGroup 创建卡片分组
	CreateCardGroup(ctx context.Context, group *entity.DialectCardGroup) error

	// UpdateCardGroup 更新卡片分组
	UpdateCardGroup(ctx context.Context, group *entity.DialectCardGroup) error

	// DeleteCardGroup 删除卡片分组
	DeleteCardGroup(ctx context.Context, id string) error

	// FindCardGroupByID 根据ID获取卡片分组
	FindCardGroupByID(ctx context.Context, id string) (*entity.DialectCardGroup, error)

	// ListCards 获取卡片列表
	ListCards(ctx context.Context, query *DialectCardQuery) ([]entity.DialectCard, error)

	// CreateCard 创建卡片
	CreateCard(ctx context.Context, card *entity.DialectCard) error

	// UpdateCard 更新卡片
	UpdateCard(ctx context.Context, card *entity.DialectCard) error

	// DeleteCard 删除卡片
	DeleteCard(ctx context.Context, id string) error

	// FindCardByID 根据ID获取卡片
	FindCardByID(ctx context.Context, id string) (*entity.DialectCard, error)
}

// DialectQuery 方言查询参数
type DialectQuery struct {
	Pagination
	Keyword    string               `json:"keyword"`
	Region     string               `json:"region"`
	Province   string               `json:"province"`
	City       string               `json:"city"`
	Type       entity.DialectType   `json:"type"`
	Status     entity.DialectStatus `json:"status"`
	UploaderID string               `json:"uploader_id"`
	OrgID      string               `json:"org_id"`
	OrgIDs     []string             `json:"org_ids"`
	IsFeatured *bool                `json:"is_featured,omitempty"`
	SortBy     string               `json:"sort_by"` // play_count, like_count, created_at
	SortOrder  string               `json:"sort_order"`
}

// DialectCardQuery 方言卡片查询参数
type DialectCardQuery struct {
	GroupID          string `json:"group_id"`
	IncludeInactive  bool   `json:"include_inactive"`
	IncludeGroupInfo bool   `json:"include_group_info"`
}

// NewDialectQuery 创建默认方言查询
func NewDialectQuery() *DialectQuery {
	return &DialectQuery{
		Pagination: Pagination{
			Page:     1,
			PageSize: 10,
		},
		SortBy:    "created_at",
		SortOrder: "desc",
	}
}
