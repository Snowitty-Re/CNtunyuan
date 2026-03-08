// Package testutil 测试工具包
package testutil

import (
	"context"
	"fmt"
	"testing"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestDB 测试数据库
type TestDB struct {
	DB *gorm.DB
}

// NewTestDB 创建测试数据库（使用 SQLite 内存模式）
func NewTestDB(t *testing.T) *TestDB {
	// Use unique database name for each test to avoid conflicts
	// Each test gets its own isolated in-memory database
	dbName := fmt.Sprintf("file:%s?mode=memory&cache=shared", uuid.New().String())
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err, "failed to connect to test database")

	// 自动迁移所有实体
	err = db.AutoMigrate(
		&entity.User{},
		&entity.Organization{},
		&entity.MissingPerson{},
		&entity.Task{},
		&entity.Dialect{},
		&entity.File{},
		&entity.AuditLog{},
		&entity.TaskLog{},
		&entity.TaskAttachment{},
		&entity.MissingPersonTrack{},
		&entity.MissingPhoto{},
		&entity.DialectComment{},
		&entity.DialectLike{},
		&entity.DialectPlayLog{},
		&entity.Permission{},
		&entity.UserPermission{},
		&entity.OrgStats{},
	)
	require.NoError(t, err, "failed to migrate test database")

	return &TestDB{DB: db}
}

// Close 关闭测试数据库
func (tdb *TestDB) Close() {
	sqlDB, err := tdb.DB.DB()
	if err == nil {
		sqlDB.Close()
	}
}

// Tx 在事务中执行操作
func (tdb *TestDB) Tx(fn func(tx *gorm.DB) error) error {
	return tdb.DB.Transaction(fn)
}

// Context 返回测试上下文
func Context() context.Context {
	return context.Background()
}

// AssertError 断言错误
func AssertError(t *testing.T, err error, expectedError string) {
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), expectedError)
	}
}

// AssertNoError 断言无错误
func AssertNoError(t *testing.T, err error) {
	assert.NoError(t, err)
}

// MustCreate 必须创建成功
func MustCreate(t *testing.T, db *gorm.DB, value interface{}) {
	err := db.Create(value).Error
	require.NoError(t, err, "failed to create record")
}

// MustFind 必须查找成功
func MustFind(t *testing.T, db *gorm.DB, dest interface{}, conds ...interface{}) {
	err := db.First(dest, conds...).Error
	require.NoError(t, err, "failed to find record")
}
