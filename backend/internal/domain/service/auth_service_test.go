// Package service auth service unit tests
package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/config"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/valueobject"
	repoImpl "github.com/Snowitty-Re/CNtunyuan/internal/infrastructure/repository"
	"github.com/Snowitty-Re/CNtunyuan/internal/infrastructure/sms"
	"github.com/Snowitty-Re/CNtunyuan/pkg/utils"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// MockTokenService is a mock implementation of TokenService for testing
type MockTokenService struct {
	generateTokenPairFunc func(ctx context.Context, user *entity.User) (*TokenPair, error)
	validateTokenFunc     func(ctx context.Context, token string) (*TokenClaims, error)
	revokeTokenFunc       func(ctx context.Context, token string) error
}

func (m *MockTokenService) GenerateTokenPair(ctx context.Context, user *entity.User) (*TokenPair, error) {
	if m.generateTokenPairFunc != nil {
		return m.generateTokenPairFunc(ctx, user)
	}
	return &TokenPair{
		AccessToken:  "mock-access-token",
		RefreshToken: "mock-refresh-token",
		ExpiresIn:    3600,
	}, nil
}

func (m *MockTokenService) ValidateToken(ctx context.Context, token string) (*TokenClaims, error) {
	if m.validateTokenFunc != nil {
		return m.validateTokenFunc(ctx, token)
	}
	return nil, errors.New("token not found")
}

func (m *MockTokenService) RevokeToken(ctx context.Context, token string) error {
	if m.revokeTokenFunc != nil {
		return m.revokeTokenFunc(ctx, token)
	}
	return nil
}

// MockCache is a mock implementation of cache.Cache for testing
type MockCache struct {
	data map[string]interface{}
}

func NewMockCache() *MockCache {
	return &MockCache{
		data: make(map[string]interface{}),
	}
}

func (m *MockCache) Get(ctx context.Context, key string, dest interface{}) error {
	val, exists := m.data[key]
	if !exists {
		return errors.New("key not found")
	}
	// Simple type assertion for string values (used in verify code)
	if s, ok := val.(string); ok {
		if sp, ok := dest.(*string); ok {
			*sp = s
			return nil
		}
	}
	// For int values (used in login attempts)
	if i, ok := val.(int); ok {
		if ip, ok := dest.(*int); ok {
			*ip = i
			return nil
		}
	}
	return errors.New("type mismatch")
}

func (m *MockCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	m.data[key] = value
	return nil
}

func (m *MockCache) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	if _, exists := m.data[key]; exists {
		return false, nil
	}
	m.data[key] = value
	return true, nil
}

func (m *MockCache) IncrWithTTL(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	val, exists := m.data[key]
	if !exists {
		m.data[key] = int64(1)
		return 1, nil
	}
	switch v := val.(type) {
	case int:
		n := int64(v + 1)
		m.data[key] = n
		return n, nil
	case int64:
		n := v + 1
		m.data[key] = n
		return n, nil
	default:
		return 0, errors.New("type mismatch")
	}
}

func (m *MockCache) Delete(ctx context.Context, keys ...string) error {
	for _, key := range keys {
		delete(m.data, key)
	}
	return nil
}

func (m *MockCache) Exists(ctx context.Context, key string) (bool, error) {
	_, exists := m.data[key]
	return exists, nil
}

func (m *MockCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	return 0, nil
}

func (m *MockCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return nil
}

func (m *MockCache) Close() error {
	return nil
}

// MockWechatClient is a mock implementation of WechatClient for testing
type MockWechatClient struct {
	code2SessionFunc func(code string) (*WechatSession, error)
	getAccessTokenFn func() (string, int, error)
	getPhoneNumberFn func(accessToken, code string) (*WechatPhoneInfo, error)
}

func (m *MockWechatClient) Code2Session(code string) (*WechatSession, error) {
	if m.code2SessionFunc != nil {
		return m.code2SessionFunc(code)
	}
	return &WechatSession{
		OpenID:     "mock-openid",
		SessionKey: "mock-session-key",
		UnionID:    "mock-unionid",
	}, nil
}

func (m *MockWechatClient) GetAccessToken() (string, int, error) {
	if m.getAccessTokenFn != nil {
		return m.getAccessTokenFn()
	}
	return "mock-access-token", 7200, nil
}

func (m *MockWechatClient) GetPhoneNumber(accessToken, code string) (*WechatPhoneInfo, error) {
	if m.getPhoneNumberFn != nil {
		return m.getPhoneNumberFn(accessToken, code)
	}
	return &WechatPhoneInfo{
		PhoneNumber:     "13800138000",
		PurePhoneNumber: "13800138000",
		CountryCode:     "86",
	}, nil
}

// TestDB wraps test database
type TestDB struct {
	DB *gorm.DB
}

// NewTestDB creates test database
func NewTestDB(t *testing.T) *TestDB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err, "failed to connect to test database")

	// Auto migrate entities
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

// Close closes test database
func (tdb *TestDB) Close() {
	sqlDB, err := tdb.DB.DB()
	if err == nil {
		sqlDB.Close()
	}
}

// MustCreate creates record and fails test on error
func MustCreate(t *testing.T, db *gorm.DB, value interface{}) {
	err := db.Create(value).Error
	require.NoError(t, err, "failed to create record")
}

// Context returns test context
func Context() context.Context {
	return context.Background()
}

// setupAuthTest creates test dependencies for auth service tests
func setupAuthTest(t *testing.T) (*AuthService, *TestDB, *MockTokenService, *MockCache, *MockWechatClient) {
	tdb := NewTestDB(t)
	userRepo := repoImpl.NewUserRepository(tdb.DB)
	mockTokenService := &MockTokenService{}
	mockCache := NewMockCache()
	mockWechatClient := &MockWechatClient{}

	service := NewAuthService(userRepo, mockTokenService, mockCache, mockWechatClient)
	return service, tdb, mockTokenService, mockCache, mockWechatClient
}

// createTestUser creates a test user with the given parameters
func createTestUser(t *testing.T, tdb *TestDB, phone, nickname, password string, status entity.UserStatus, role entity.Role) *entity.User {
	user := &entity.User{
		BaseEntity: entity.BaseEntity{ID: uuid.New().String()},
		Nickname:   nickname,
		Phone:      phone,
		Email:      phone + "@test.com", // Unique email based on phone
		WxOpenID:   "wx_" + phone,       // Unique wx_openid based on phone
		Role:       role,
		Status:     status,
		OrgID:      "test-org-id",
	}
	user.SetPassword(password)
	MustCreate(t, tdb.DB, user)
	return user
}

// createTestOrg creates a test organization
func createTestOrg(t *testing.T, tdb *TestDB) *entity.Organization {
	org := &entity.Organization{
		BaseEntity: entity.BaseEntity{ID: uuid.New().String()},
		Name:       "测试组织",
		Code:       "TEST001",
		Type:       entity.OrgTypeCity,
	}
	MustCreate(t, tdb.DB, org)
	return org
}

func TestAuthService_Login(t *testing.T) {
	service, tdb, mockTokenService, _, _ := setupAuthTest(t)
	defer tdb.Close()

	// Create test organization and user
	createTestOrg(t, tdb)
	createTestUser(t, tdb, "13800138000", "测试用户", "correctpassword", entity.UserStatusActive, entity.RoleVolunteer)

	// Setup mock token service
	mockTokenService.generateTokenPairFunc = func(ctx context.Context, user *entity.User) (*TokenPair, error) {
		return &TokenPair{
			AccessToken:  "test-access-token",
			RefreshToken: "test-refresh-token",
			ExpiresIn:    3600,
		}, nil
	}

	tests := []struct {
		name       string
		creds      valueobject.LoginCredentials
		wantErr    bool
		errCode    int // errors.ErrorCode
		checkToken bool
	}{
		{
			name: "valid credentials with phone",
			creds: valueobject.LoginCredentials{
				Username: "13800138000",
				Password: "correctpassword",
			},
			wantErr:    false,
			checkToken: true,
		},
		{
			name: "valid credentials with nickname",
			creds: valueobject.LoginCredentials{
				Username: "测试用户",
				Password: "correctpassword",
			},
			wantErr:    false,
			checkToken: true,
		},
		{
			name: "wrong password",
			creds: valueobject.LoginCredentials{
				Username: "13800138000",
				Password: "wrongpassword",
			},
			wantErr: true,
			errCode: 1002, // CodeInvalidPassword
		},
		{
			name: "non-existing user",
			creds: valueobject.LoginCredentials{
				Username: "nonexistent",
				Password: "anypassword",
			},
			wantErr: true,
			errCode: 1002, // CodeInvalidPassword
		},
		{
			name: "disabled user",
			creds: valueobject.LoginCredentials{
				Username: "13800138001",
				Password: "password123",
			},
			wantErr: true,
			errCode: 1006, // CodeAccountDisabled
		},
		{
			name: "banned user",
			creds: valueobject.LoginCredentials{
				Username: "13800138002",
				Password: "password123",
			},
			wantErr: true,
			errCode: 1007, // CodeAccountLocked
		},
	}

	// Create disabled and banned users for testing
	createTestUser(t, tdb, "13800138001", "禁用用户", "password123", entity.UserStatusInactive, entity.RoleVolunteer)
	createTestUser(t, tdb, "13800138002", "封禁用户", "password123", entity.UserStatusBanned, entity.RoleVolunteer)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, user, err := service.Login(Context(), tt.creds, "127.0.0.1")

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCode != 0 {
					// Check error code if possible
				}
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, user)
			assert.NotNil(t, result)

			if tt.checkToken {
				assert.NotEmpty(t, result.AccessToken)
				assert.NotEmpty(t, result.RefreshToken)
				assert.Equal(t, "Bearer", result.TokenType)
				assert.Greater(t, result.ExpiresIn, 0)
			}

			// Verify user login info was recorded
			if user != nil {
				var updatedUser entity.User
				err := tdb.DB.First(&updatedUser, "id = ?", user.ID).Error
				require.NoError(t, err)
				assert.NotNil(t, updatedUser.LastLoginAt)
				assert.Equal(t, "127.0.0.1", updatedUser.LastLoginIP)
			}
		})
	}
}

func TestAuthService_Logout(t *testing.T) {
	service, tdb, mockTokenService, _, _ := setupAuthTest(t)
	defer tdb.Close()

	revokedTokens := make(map[string]bool)
	mockTokenService.revokeTokenFunc = func(ctx context.Context, token string) error {
		revokedTokens[token] = true
		return nil
	}

	tests := []struct {
		name        string
		token       string
		shouldError bool
	}{
		{
			name:        "logout with valid token",
			token:       "valid-token-123",
			shouldError: false,
		},
		{
			name:        "logout with empty token",
			token:       "",
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Logout(Context(), tt.token)

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.token != "" {
					assert.True(t, revokedTokens[tt.token], "token should be revoked")
				}
			}
		})
	}
}

func TestAuthService_RefreshToken(t *testing.T) {
	service, tdb, mockTokenService, _, _ := setupAuthTest(t)
	defer tdb.Close()

	// Create test organization and user
	createTestOrg(t, tdb)
	testUser := createTestUser(t, tdb, "13800138000", "测试用户", "password123", entity.UserStatusActive, entity.RoleVolunteer)

	validRefreshToken := "valid-refresh-token"
	invalidToken := "invalid-token"
	expiredToken := "expired-token"

	mockTokenService.validateTokenFunc = func(ctx context.Context, token string) (*TokenClaims, error) {
		switch token {
		case validRefreshToken:
			return &TokenClaims{
				UserID:   testUser.ID,
				Nickname: testUser.Nickname,
				Role:     string(testUser.Role),
				OrgID:    testUser.OrgID,
			}, nil
		case expiredToken:
			return nil, errors.New("token expired")
		default:
			return nil, errors.New("invalid token")
		}
	}

	mockTokenService.generateTokenPairFunc = func(ctx context.Context, user *entity.User) (*TokenPair, error) {
		return &TokenPair{
			AccessToken:  "new-access-token",
			RefreshToken: "new-refresh-token",
			ExpiresIn:    3600,
		}, nil
	}

	tests := []struct {
		name         string
		refreshToken string
		wantErr      bool
		checkToken   bool
	}{
		{
			name:         "valid refresh token",
			refreshToken: validRefreshToken,
			wantErr:      false,
			checkToken:   true,
		},
		{
			name:         "invalid refresh token",
			refreshToken: invalidToken,
			wantErr:      true,
		},
		{
			name:         "expired refresh token",
			refreshToken: expiredToken,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, user, err := service.RefreshToken(Context(), tt.refreshToken)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, user)
			assert.NotNil(t, result)

			if tt.checkToken {
				assert.Equal(t, "new-access-token", result.AccessToken)
				assert.Equal(t, "new-refresh-token", result.RefreshToken)
				assert.Equal(t, "Bearer", result.TokenType)
			}
		})
	}
}

func TestAuthService_RefreshToken_WithDisabledUser(t *testing.T) {
	service, tdb, mockTokenService, _, _ := setupAuthTest(t)
	defer tdb.Close()

	// Create test organization and disabled user
	createTestOrg(t, tdb)
	disabledUser := createTestUser(t, tdb, "13800138000", "禁用用户", "password123", entity.UserStatusInactive, entity.RoleVolunteer)

	validRefreshToken := "valid-refresh-token"

	mockTokenService.validateTokenFunc = func(ctx context.Context, token string) (*TokenClaims, error) {
		return &TokenClaims{
			UserID:   disabledUser.ID,
			Nickname: disabledUser.Nickname,
			Role:     string(disabledUser.Role),
			OrgID:    disabledUser.OrgID,
		}, nil
	}

	_, _, err := service.RefreshToken(Context(), validRefreshToken)
	assert.Error(t, err)
}

func TestAuthService_ValidateToken(t *testing.T) {
	service, tdb, mockTokenService, _, _ := setupAuthTest(t)
	defer tdb.Close()

	validToken := "valid-token"
	invalidToken := "invalid-token"

	mockTokenService.validateTokenFunc = func(ctx context.Context, token string) (*TokenClaims, error) {
		if token == validToken {
			return &TokenClaims{
				UserID:   "user-id-123",
				Nickname: "Test User",
				Role:     "volunteer",
				OrgID:    "org-id-123",
			}, nil
		}
		return nil, errors.New("invalid token")
	}

	tests := []struct {
		name        string
		token       string
		wantErr     bool
		checkClaims bool
	}{
		{
			name:        "valid token",
			token:       validToken,
			wantErr:     false,
			checkClaims: true,
		},
		{
			name:    "invalid token",
			token:   invalidToken,
			wantErr: true,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := service.ValidateToken(Context(), tt.token)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, claims)

			if tt.checkClaims {
				assert.Equal(t, "user-id-123", claims.UserID)
				assert.Equal(t, "Test User", claims.Nickname)
				assert.Equal(t, "volunteer", claims.Role)
				assert.Equal(t, "org-id-123", claims.OrgID)
			}
		})
	}
}

func TestAuthService_GenerateTokenPair(t *testing.T) {
	service, tdb, mockTokenService, _, _ := setupAuthTest(t)
	defer tdb.Close()

	createTestOrg(t, tdb)
	testUser := createTestUser(t, tdb, "13800138000", "测试用户", "password123", entity.UserStatusActive, entity.RoleVolunteer)

	mockTokenService.generateTokenPairFunc = func(ctx context.Context, user *entity.User) (*TokenPair, error) {
		return &TokenPair{
			AccessToken:  "generated-access-token",
			RefreshToken: "generated-refresh-token",
			ExpiresIn:    7200,
		}, nil
	}

	tokens, err := service.GenerateTokenPair(testUser)
	require.NoError(t, err)
	assert.NotNil(t, tokens)
	assert.Equal(t, "generated-access-token", tokens.AccessToken)
	assert.Equal(t, "generated-refresh-token", tokens.RefreshToken)
	assert.Equal(t, 7200, tokens.ExpiresIn)
}

func TestAuthService_GetCurrentUser(t *testing.T) {
	service, tdb, _, _, _ := setupAuthTest(t)
	defer tdb.Close()

	createTestOrg(t, tdb)
	testUser := createTestUser(t, tdb, "13800138000", "测试用户", "password123", entity.UserStatusActive, entity.RoleVolunteer)

	tests := []struct {
		name    string
		userID  string
		wantErr bool
	}{
		{
			name:    "existing user",
			userID:  testUser.ID,
			wantErr: false,
		},
		{
			name:    "non-existing user",
			userID:  "non-existing-id",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := service.GetCurrentUser(Context(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, user)
			assert.Equal(t, tt.userID, user.ID)
		})
	}
}

func TestAuthService_WechatLogin(t *testing.T) {
	service, tdb, mockTokenService, _, mockWechatClient := setupAuthTest(t)
	defer tdb.Close()

	createTestOrg(t, tdb)

	mockWechatClient.code2SessionFunc = func(code string) (*WechatSession, error) {
		if code == "invalid-code" {
			return nil, errors.New("invalid code")
		}
		return &WechatSession{
			OpenID:     "wx-openid-123",
			SessionKey: "session-key-123",
			UnionID:    "unionid-123",
		}, nil
	}

	mockTokenService.generateTokenPairFunc = func(ctx context.Context, user *entity.User) (*TokenPair, error) {
		return &TokenPair{
			AccessToken:  "wx-access-token",
			RefreshToken: "wx-refresh-token",
			ExpiresIn:    3600,
		}, nil
	}

	tests := []struct {
		name         string
		code         string
		userInfo     *valueobject.WechatUserInfo
		wantErr      bool
		wantNeedBind bool
		checkToken   bool
	}{
		{
			name:         "new user login - need bind phone",
			code:         "valid-code-new",
			userInfo:     &valueobject.WechatUserInfo{Nickname: "微信用户", Avatar: "avatar.jpg"},
			wantErr:      false,
			wantNeedBind: true,
			checkToken:   false,
		},
		{
			name:         "existing user login",
			code:         "valid-code-existing",
			userInfo:     nil,
			wantErr:      false,
			wantNeedBind: false,
			checkToken:   true,
		},
		{
			name:     "invalid wechat code",
			code:     "invalid-code",
			userInfo: nil,
			wantErr:  true,
		},
	}

	// Create existing wechat user
	existingUser := &entity.User{
		BaseEntity: entity.BaseEntity{ID: uuid.New().String()},
		Nickname:   "Existing Wechat User",
		Phone:      "13800138001",
		Email:      "existing@wechat.com",
		Role:       entity.RoleVolunteer,
		Status:     entity.UserStatusActive,
		OrgID:      "test-org-id",
		WxOpenID:   "wx-openid-existing",
	}
	existingUser.SetPassword("password123")
	MustCreate(t, tdb.DB, existingUser)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Override the mock for specific test cases
			switch tt.name {
			case "existing user login":
				mockWechatClient.code2SessionFunc = func(code string) (*WechatSession, error) {
					return &WechatSession{
						OpenID:     "wx-openid-existing",
						SessionKey: "session-key",
						UnionID:    "unionid",
					}, nil
				}
			case "invalid wechat code":
				mockWechatClient.code2SessionFunc = func(code string) (*WechatSession, error) {
					return nil, errors.New("invalid code")
				}
			default:
				// Reset to default mock for new user login
				mockWechatClient.code2SessionFunc = func(code string) (*WechatSession, error) {
					return &WechatSession{
						OpenID:     "wx-openid-" + code,
						SessionKey: "session-key-123",
						UnionID:    "unionid-123",
					}, nil
				}
			}

			result, user, needBind, err := service.WechatLogin(Context(), tt.code, "127.0.0.1", tt.userInfo)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantNeedBind, needBind)

			if needBind {
				// For new users, result should be nil but user should be created
				assert.Nil(t, result)
				assert.NotNil(t, user)
				assert.NotEmpty(t, user.ID)
			} else {
				// For existing users, both result and user should be set
				assert.NotNil(t, result)
				assert.NotNil(t, user)

				if tt.checkToken {
					assert.NotEmpty(t, result.AccessToken)
					assert.NotEmpty(t, result.RefreshToken)
				}
			}
		})
	}
}

func TestAuthService_WechatLogin_WithDisabledUser(t *testing.T) {
	service, tdb, _, _, mockWechatClient := setupAuthTest(t)
	defer tdb.Close()

	createTestOrg(t, tdb)

	// Create disabled wechat user
	disabledUser := &entity.User{
		BaseEntity: entity.BaseEntity{ID: uuid.New().String()},
		Nickname:   "Disabled Wechat User",
		Phone:      "13800138002",
		Email:      "disabled@wechat.com",
		Role:       entity.RoleVolunteer,
		Status:     entity.UserStatusInactive,
		OrgID:      "test-org-id",
		WxOpenID:   "wx-openid-disabled",
	}
	disabledUser.SetPassword("password123")
	MustCreate(t, tdb.DB, disabledUser)

	mockWechatClient.code2SessionFunc = func(code string) (*WechatSession, error) {
		return &WechatSession{
			OpenID:     "wx-openid-disabled",
			SessionKey: "session-key",
			UnionID:    "unionid",
		}, nil
	}

	_, _, _, err := service.WechatLogin(Context(), "valid-code", "127.0.0.1", nil)
	assert.Error(t, err)
}

func TestAuthService_BindPhone(t *testing.T) {
	service, tdb, mockTokenService, mockCache, _ := setupAuthTest(t)
	defer tdb.Close()

	createTestOrg(t, tdb)

	// Set up SMS service with dev mode for testing
	devSMSConfig := &config.SMSConfig{DevMode: true}
	devSMS := sms.NewService(devSMSConfig, mockCache)
	service.SetSMSService(devSMS)

	mockTokenService.generateTokenPairFunc = func(ctx context.Context, user *entity.User) (*TokenPair, error) {
		return &TokenPair{
			AccessToken:  "bound-access-token",
			RefreshToken: "bound-refresh-token",
			ExpiresIn:    3600,
		}, nil
	}

	// Create temp user for binding
	tempUser := &entity.User{
		BaseEntity: entity.BaseEntity{ID: uuid.New().String()},
		Nickname:   "Temp User",
		Phone:      "", // Empty phone
		Email:      "temp@bind.com",
		Role:       entity.RoleVolunteer,
		Status:     entity.UserStatusActive,
		OrgID:      "test-org-id",
		WxOpenID:   "wx-openid-temp",
	}
	tempUser.SetPassword("password123")
	MustCreate(t, tdb.DB, tempUser)

	// Create user with existing phone
	existingPhoneUser := &entity.User{
		BaseEntity: entity.BaseEntity{ID: uuid.New().String()},
		Nickname:   "Existing Phone User",
		Phone:      "13800138999",
		Email:      "existing@phone.com",
		Role:       entity.RoleVolunteer,
		Status:     entity.UserStatusActive,
		OrgID:      "test-org-id",
	}
	existingPhoneUser.SetPassword("password123")
	MustCreate(t, tdb.DB, existingPhoneUser)

	tests := []struct {
		name    string
		userID  string
		phone   string
		code    string
		wantErr bool
	}{
		{
			name:    "bind phone to existing user with valid code",
			userID:  tempUser.ID,
			phone:   "13800138003",
			code:    "", // Will be set after sending verify code
			wantErr: false,
		},
		{
			name:    "bind existing phone to new user",
			userID:  "",
			phone:   "13800138999", // Already exists
			code:    "123456",
			wantErr: true,
		},
		{
			name:    "bind phone with invalid code",
			userID:  tempUser.ID,
			phone:   "13800138004",
			code:    "000000", // Invalid code
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Send verify code first if code is empty (needs to be generated)
			if tt.code == "" && !tt.wantErr {
				code, err := devSMS.SendVerifyCode(Context(), tt.phone)
				require.NoError(t, err)
				tt.code = code
			}

			result, err := service.BindPhone(Context(), tt.userID, tt.phone, tt.code)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.NotEmpty(t, result.AccessToken)
			assert.NotEmpty(t, result.RefreshToken)

			// Verify phone was bound
			if tt.userID != "" {
				var updatedUser entity.User
				err := tdb.DB.First(&updatedUser, "id = ?", tt.userID).Error
				require.NoError(t, err)
				assert.Equal(t, tt.phone, updatedUser.Phone)
			}
		})
	}
}

func TestAuthService_SendVerifyCode(t *testing.T) {
	service, tdb, _, mockCache, _ := setupAuthTest(t)
	defer tdb.Close()

	// Set up SMS service with dev mode for testing
	devSMSConfig := &config.SMSConfig{DevMode: true}
	devSMS := sms.NewService(devSMSConfig, mockCache)
	service.SetSMSService(devSMS)

	// This test mainly verifies the method doesn't panic
	// The actual SMS sending is mocked
	err := service.SendVerifyCode(Context(), "13800138000")
	// Should not error even without SMS service configured
	assert.NoError(t, err)
}

func TestAuthService_Login_TokenGenerationFailure(t *testing.T) {
	service, tdb, mockTokenService, _, _ := setupAuthTest(t)
	defer tdb.Close()

	createTestOrg(t, tdb)
	testUser := createTestUser(t, tdb, "13800138000", "测试用户", "correctpassword", entity.UserStatusActive, entity.RoleVolunteer)

	// Mock token service to return error
	mockTokenService.generateTokenPairFunc = func(ctx context.Context, user *entity.User) (*TokenPair, error) {
		return nil, errors.New("token generation failed")
	}

	creds := valueobject.LoginCredentials{
		Username: "13800138000",
		Password: "correctpassword",
	}

	result, _, err := service.Login(Context(), creds, "127.0.0.1")
	assert.Error(t, err)
	assert.Nil(t, result)

	// Verify user login was still recorded even if token generation failed
	var updatedUser entity.User
	err = tdb.DB.First(&updatedUser, "id = ?", testUser.ID).Error
	require.NoError(t, err)
	assert.NotNil(t, updatedUser.LastLoginAt)
}

func TestAuthService_RefreshToken_NonExistingUser(t *testing.T) {
	service, tdb, mockTokenService, _, _ := setupAuthTest(t)
	defer tdb.Close()

	validRefreshToken := "valid-token-for-deleted-user"

	mockTokenService.validateTokenFunc = func(ctx context.Context, token string) (*TokenClaims, error) {
		return &TokenClaims{
			UserID:   "non-existing-user-id",
			Nickname: "Deleted User",
			Role:     "volunteer",
			OrgID:    "org-id",
		}, nil
	}

	_, _, err := service.RefreshToken(Context(), validRefreshToken)
	assert.Error(t, err)
}

func TestAuthService_RefreshToken_TokenGenerationFailure(t *testing.T) {
	service, tdb, mockTokenService, _, _ := setupAuthTest(t)
	defer tdb.Close()

	createTestOrg(t, tdb)
	testUser := createTestUser(t, tdb, "13800138000", "测试用户", "password123", entity.UserStatusActive, entity.RoleVolunteer)

	validRefreshToken := "valid-refresh-token"

	mockTokenService.validateTokenFunc = func(ctx context.Context, token string) (*TokenClaims, error) {
		return &TokenClaims{
			UserID:   testUser.ID,
			Nickname: testUser.Nickname,
			Role:     string(testUser.Role),
			OrgID:    testUser.OrgID,
		}, nil
	}

	mockTokenService.generateTokenPairFunc = func(ctx context.Context, user *entity.User) (*TokenPair, error) {
		return nil, errors.New("token generation failed")
	}

	_, _, err := service.RefreshToken(Context(), validRefreshToken)
	assert.Error(t, err)
}

// Helper function to hash password for test setup
func hashPassword(password string) string {
	hash, _ := utils.HashPassword(password)
	return hash
}
