package mocks

import (
	"context"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"github.com/stretchr/testify/mock"
)

// UserRepositoryMock 用户仓储Mock
type UserRepositoryMock struct {
	mock.Mock
}

// NewUserRepositoryMock 创建用户仓储Mock
func NewUserRepositoryMock() *UserRepositoryMock {
	return &UserRepositoryMock{}
}

// Create 创建用户
func (m *UserRepositoryMock) Create(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// FindByID 根据ID查找用户
func (m *UserRepositoryMock) FindByID(ctx context.Context, id string) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

// FindByPhone 根据手机号查找用户
func (m *UserRepositoryMock) FindByPhone(ctx context.Context, phone string) (*entity.User, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

// FindByEmail 根据邮箱查找用户
func (m *UserRepositoryMock) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

// FindByPhoneOrNickname 根据手机号或昵称查找用户
func (m *UserRepositoryMock) FindByPhoneOrNickname(ctx context.Context, username string) (*entity.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

// FindByOpenID 根据微信OpenID查找用户
func (m *UserRepositoryMock) FindByOpenID(ctx context.Context, openID string) (*entity.User, error) {
	args := m.Called(ctx, openID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

// FindByOrgID 根据组织ID查找用户
func (m *UserRepositoryMock) FindByOrgID(ctx context.Context, orgID string, pagination repository.Pagination) (*repository.PageResult[entity.User], error) {
	args := m.Called(ctx, orgID, pagination)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.PageResult[entity.User]), args.Error(1)
}

// FindByRole 根据角色查找用户
func (m *UserRepositoryMock) FindByRole(ctx context.Context, role string, pagination repository.Pagination) (*repository.PageResult[entity.User], error) {
	args := m.Called(ctx, role, pagination)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.PageResult[entity.User]), args.Error(1)
}

// List 分页查询用户列表
func (m *UserRepositoryMock) List(ctx context.Context, query *repository.UserQuery) (*repository.PageResult[entity.User], error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.PageResult[entity.User]), args.Error(1)
}

// Update 更新用户
func (m *UserRepositoryMock) Update(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// UpdatePassword 更新密码
func (m *UserRepositoryMock) UpdatePassword(ctx context.Context, userID string, hashedPassword string) error {
	args := m.Called(ctx, userID, hashedPassword)
	return args.Error(0)
}

// UpdateStatus 更新状态
func (m *UserRepositoryMock) UpdateStatus(ctx context.Context, userID string, status entity.UserStatus) error {
	args := m.Called(ctx, userID, status)
	return args.Error(0)
}

// UpdateRole 更新角色
func (m *UserRepositoryMock) UpdateRole(ctx context.Context, userID string, role string) error {
	args := m.Called(ctx, userID, role)
	return args.Error(0)
}

// Delete 删除用户
func (m *UserRepositoryMock) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Count 统计用户数量
func (m *UserRepositoryMock) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// CountByOrg 统计组织用户数量
func (m *UserRepositoryMock) CountByOrg(ctx context.Context, orgID string) (int64, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0).(int64), args.Error(1)
}

// CountByRole 根据角色统计用户数量
func (m *UserRepositoryMock) CountByRole(ctx context.Context, role string) (int64, error) {
	args := m.Called(ctx, role)
	return args.Get(0).(int64), args.Error(1)
}

// ExistsPhone 检查手机号是否存在
func (m *UserRepositoryMock) ExistsPhone(ctx context.Context, phone string) (bool, error) {
	args := m.Called(ctx, phone)
	return args.Bool(0), args.Error(1)
}

// ExistsEmail 检查邮箱是否存在
func (m *UserRepositoryMock) ExistsEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

// SoftDelete 软删除用户
func (m *UserRepositoryMock) SoftDelete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// FindAll 查找所有用户
func (m *UserRepositoryMock) FindAll(ctx context.Context) ([]entity.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]entity.User), args.Error(1)
}

// Exists 检查用户是否存在
func (m *UserRepositoryMock) Exists(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

// Ensure UserRepositoryMock implements UserRepository interface
var _ repository.UserRepository = (*UserRepositoryMock)(nil)
