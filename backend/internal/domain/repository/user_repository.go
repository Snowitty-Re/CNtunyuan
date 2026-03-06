package repository

import (
	"context"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
)

// UserRepository 用户仓储接口
type UserRepository interface {
	Repository[entity.User]

	// FindByPhone 根据手机号查找
	FindByPhone(ctx context.Context, phone string) (*entity.User, error)

	// FindByEmail 根据邮箱查找
	FindByEmail(ctx context.Context, email string) (*entity.User, error)

	// FindByPhoneOrNickname 根据手机号或昵称查找
	FindByPhoneOrNickname(ctx context.Context, username string) (*entity.User, error)

	// FindByOpenID 根据微信OpenID查找
	FindByOpenID(ctx context.Context, openID string) (*entity.User, error)

	// FindByOrgID 根据组织ID查找用户
	FindByOrgID(ctx context.Context, orgID string, pagination Pagination) (*PageResult[entity.User], error)

	// FindByRole 根据角色查找用户
	FindByRole(ctx context.Context, role entity.Role, pagination Pagination) (*PageResult[entity.User], error)

	// List 分页查询用户列表
	List(ctx context.Context, query *UserQuery) (*PageResult[entity.User], error)

	// UpdatePassword 更新密码
	UpdatePassword(ctx context.Context, userID string, hashedPassword string) error

	// UpdateStatus 更新状态
	UpdateStatus(ctx context.Context, userID string, status entity.UserStatus) error

	// UpdateRole 更新角色
	UpdateRole(ctx context.Context, userID string, role entity.Role) error

	// CountByOrg 统计组织用户数量
	CountByOrg(ctx context.Context, orgID string) (int64, error)

	// CountByRole 统计角色用户数量
	CountByRole(ctx context.Context, role entity.Role) (int64, error)

	// ExistsPhone 检查手机号是否存在
	ExistsPhone(ctx context.Context, phone string) (bool, error)

	// ExistsEmail 检查邮箱是否存在
	ExistsEmail(ctx context.Context, email string) (bool, error)
}

// UserQuery 用户查询参数
type UserQuery struct {
	Pagination
	Keyword   string           `json:"keyword"`   // 关键词搜索
	Role      entity.Role      `json:"role"`      // 角色筛选
	Status    entity.UserStatus `json:"status"`    // 状态筛选
	OrgID     string           `json:"org_id"`    // 组织筛选
	StartTime string           `json:"start_time"` // 开始时间
	EndTime   string           `json:"end_time"`   // 结束时间
	SortField string           `json:"sort_field"` // 排序字段
	SortOrder string           `json:"sort_order"` // 排序方向
}

// NewUserQuery 创建默认用户查询
func NewUserQuery() *UserQuery {
	return &UserQuery{
		Pagination: Pagination{
			Page:     1,
			PageSize: 10,
		},
		SortField: "created_at",
		SortOrder: "desc",
	}
}
