package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"gorm.io/gorm"
)

// UserRepositoryImpl 用户仓储实现
type UserRepositoryImpl struct {
	*BaseRepository[entity.User]
}

// NewUserRepository 创建用户仓储
func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &UserRepositoryImpl{
		BaseRepository: NewBaseRepository[entity.User](db),
	}
}

// Create 创建用户（空邮箱写 NULL，避免唯一索引被空字符串占用）
func (r *UserRepositoryImpl) Create(ctx context.Context, user *entity.User) error {
	db := r.db.WithContext(ctx)
	omitFields := make([]string, 0, 3)

	if strings.TrimSpace(user.Email) == "" {
		user.Email = ""
		omitFields = append(omitFields, "Email")
	}
	if strings.TrimSpace(user.WxOpenID) == "" {
		user.WxOpenID = ""
		omitFields = append(omitFields, "WxOpenID")
	}
	if strings.TrimSpace(user.WxUnionID) == "" {
		user.WxUnionID = ""
		omitFields = append(omitFields, "WxUnionID")
	}
	if len(omitFields) > 0 {
		db = db.Omit(omitFields...)
	}
	return db.Create(user).Error
}

// Update 更新用户（空邮箱写 NULL）
func (r *UserRepositoryImpl) Update(ctx context.Context, user *entity.User) error {
	db := r.db.WithContext(ctx).Model(user).Select("*")
	clearNullableCols := make([]string, 0, 3)

	if strings.TrimSpace(user.Email) == "" {
		db = db.Omit("Email")
		clearNullableCols = append(clearNullableCols, "email")
	}
	if strings.TrimSpace(user.WxOpenID) == "" {
		db = db.Omit("WxOpenID")
		clearNullableCols = append(clearNullableCols, "wx_openid")
	}
	if strings.TrimSpace(user.WxUnionID) == "" {
		db = db.Omit("WxUnionID")
		clearNullableCols = append(clearNullableCols, "wx_unionid")
	}

	if err := db.Updates(user).Error; err != nil {
		return err
	}

	for _, col := range clearNullableCols {
		if err := r.db.WithContext(ctx).Model(user).Update(col, nil).Error; err != nil {
			return err
		}
	}

	return nil
}

// FindByPhone 根据手机号查找
func (r *UserRepositoryImpl) FindByPhone(ctx context.Context, phone string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	return &user, nil
}

// FindByEmail 根据邮箱查找
func (r *UserRepositoryImpl) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	return &user, nil
}

// FindByPhoneOrNickname 根据手机号或昵称查找
func (r *UserRepositoryImpl) FindByPhoneOrNickname(ctx context.Context, username string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).
		Where("phone = ? OR nickname = ?", username, username).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	return &user, nil
}

// FindByOpenID 根据微信OpenID查找
func (r *UserRepositoryImpl) FindByOpenID(ctx context.Context, openID string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("wx_openid = ?", openID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	return &user, nil
}

// FindByOrgID 根据组织ID查找用户
func (r *UserRepositoryImpl) FindByOrgID(ctx context.Context, orgID string, pagination repository.Pagination) (*repository.PageResult[entity.User], error) {
	var users []entity.User
	var total int64

	db := r.db.WithContext(ctx).Where("org_id = ?", orgID)

	if err := db.Model(&entity.User{}).Count(&total).Error; err != nil {
		return nil, err
	}

	if err := r.Paginate(db, pagination).Find(&users).Error; err != nil {
		return nil, err
	}

	return repository.NewPageResult(users, total, pagination.Page, pagination.PageSize), nil
}

// FindByRole 根据角色查找用户
func (r *UserRepositoryImpl) FindByRole(ctx context.Context, role entity.Role, pagination repository.Pagination) (*repository.PageResult[entity.User], error) {
	var users []entity.User
	var total int64

	db := r.db.WithContext(ctx).Where("role = ?", role)

	if err := db.Model(&entity.User{}).Count(&total).Error; err != nil {
		return nil, err
	}

	if err := r.Paginate(db, pagination).Find(&users).Error; err != nil {
		return nil, err
	}

	return repository.NewPageResult(users, total, pagination.Page, pagination.PageSize), nil
}

// List 分页查询用户列表
func (r *UserRepositoryImpl) List(ctx context.Context, query *repository.UserQuery) (*repository.PageResult[entity.User], error) {
	var users []entity.User
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.User{})

	// 关键词搜索
	if query.Keyword != "" {
		db = db.Where("nickname LIKE ? OR phone LIKE ? OR email LIKE ?",
			"%"+query.Keyword+"%", "%"+query.Keyword+"%", "%"+query.Keyword+"%")
	}

	// 角色筛选
	if query.Role != "" {
		db = db.Where("role = ?", query.Role)
	}

	// 状态筛选
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	// 组织筛选
	if query.OrgID != "" {
		db = db.Where("org_id = ?", query.OrgID)
	}

	// 时间范围
	if query.StartTime != "" {
		db = db.Where("created_at >= ?", query.StartTime)
	}
	if query.EndTime != "" {
		db = db.Where("created_at <= ?", query.EndTime)
	}

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	// 排序（白名单校验防止 SQL 注入）
	order := "created_at DESC"
	if query.SortField != "" {
		allowedSortFields := map[string]bool{
			"created_at":    true,
			"updated_at":    true,
			"nickname":      true,
			"phone":         true,
			"role":          true,
			"status":        true,
			"last_login_at": true,
		}
		if allowedSortFields[query.SortField] {
			sortOrder := "DESC"
			if query.SortOrder == "asc" || query.SortOrder == "ASC" {
				sortOrder = "ASC"
			}
			order = query.SortField + " " + sortOrder
		}
	}

	// 分页查询
	if err := db.Order(order).
		Preload("Org").
		Offset((query.Page - 1) * query.PageSize).
		Limit(query.PageSize).
		Find(&users).Error; err != nil {
		return nil, err
	}

	return repository.NewPageResult(users, total, query.Page, query.PageSize), nil
}

// UpdatePassword 更新密码
func (r *UserRepositoryImpl) UpdatePassword(ctx context.Context, userID string, hashedPassword string) error {
	return r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("id = ?", userID).
		Update("password", hashedPassword).
		Error
}

// UpdateStatus 更新状态
func (r *UserRepositoryImpl) UpdateStatus(ctx context.Context, userID string, status entity.UserStatus) error {
	return r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("id = ?", userID).
		Update("status", status).
		Error
}

// UpdateRole 更新角色
func (r *UserRepositoryImpl) UpdateRole(ctx context.Context, userID string, role entity.Role) error {
	return r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("id = ?", userID).
		Update("role", role).
		Error
}

// CountByOrg 统计组织用户数量
func (r *UserRepositoryImpl) CountByOrg(ctx context.Context, orgID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("org_id = ?", orgID).
		Count(&count).Error
	return count, err
}

// CountByRole 统计角色用户数量
func (r *UserRepositoryImpl) CountByRole(ctx context.Context, role entity.Role) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("role = ?", role).
		Count(&count).Error
	return count, err
}

// ExistsPhone 检查手机号是否存在
func (r *UserRepositoryImpl) ExistsPhone(ctx context.Context, phone string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("phone = ?", phone).
		Count(&count).Error
	return count > 0, err
}

// ExistsEmail 检查邮箱是否存在
func (r *UserRepositoryImpl) ExistsEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("email = ?", email).
		Count(&count).Error
	return count > 0, err
}

// CountByDateRange 按日期范围统计新增用户
func (r *UserRepositoryImpl) CountByDateRange(ctx context.Context, start, end string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.User{}).
		Where("created_at >= ? AND created_at <= ?", start, end).
		Count(&count).Error
	return count, err
}
