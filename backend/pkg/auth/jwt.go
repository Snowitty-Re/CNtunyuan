package auth

import (
	"errors"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenType 令牌类型
const (
	AccessToken  = "access"
	RefreshToken = "refresh"
)

// CustomClaims 自定义JWT Claims
type CustomClaims struct {
	UserID   string `json:"user_id"`
	OpenID   string `json:"open_id"`
	UnionID  string `json:"union_id"`
	Role     string `json:"role"`
	OrgID    string `json:"org_id"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

// JWTAuth JWT认证
type JWTAuth struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewJWTAuth 创建JWT认证实例
func NewJWTAuth(cfg *config.JWTConfig) *JWTAuth {
	return &JWTAuth{
		secret:     []byte(cfg.Secret),
		accessTTL:  time.Duration(cfg.ExpireTime) * time.Second,
		refreshTTL: time.Duration(cfg.ExpireTime*2) * time.Second,
	}
}

// GenerateToken 生成访问令牌
func (j *JWTAuth) GenerateToken(userID, openID, unionID, role, orgID string) (string, error) {
	return j.generateToken(userID, openID, unionID, role, orgID, AccessToken, j.accessTTL)
}

// GenerateRefreshToken 生成刷新令牌
func (j *JWTAuth) GenerateRefreshToken(userID, openID, unionID, role, orgID string) (string, error) {
	return j.generateToken(userID, openID, unionID, role, orgID, RefreshToken, j.refreshTTL)
}

// generateToken 生成令牌
func (j *JWTAuth) generateToken(userID, openID, unionID, role, orgID, tokenType string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := CustomClaims{
		UserID:    userID,
		OpenID:    openID,
		UnionID:   unionID,
		Role:      role,
		OrgID:     orgID,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "CNtunyuan",
			Subject:   userID,
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

// ParseToken 解析令牌
func (j *JWTAuth) ParseToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token expired")
		}
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshToken 刷新令牌
func (j *JWTAuth) RefreshToken(refreshToken string) (string, string, error) {
	claims, err := j.ParseToken(refreshToken)
	if err != nil {
		return "", "", err
	}

	if claims.TokenType != RefreshToken {
		return "", "", errors.New("invalid token type")
	}

	accessToken, err := j.GenerateToken(claims.UserID, claims.OpenID, claims.UnionID, claims.Role, claims.OrgID)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := j.GenerateRefreshToken(claims.UserID, claims.OpenID, claims.UnionID, claims.Role, claims.OrgID)
	if err != nil {
		return "", "", err
	}

	return accessToken, newRefreshToken, nil
}

// ValidateToken 验证令牌
func (j *JWTAuth) ValidateToken(tokenString string) (*CustomClaims, error) {
	claims, err := j.ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != AccessToken {
		return nil, errors.New("invalid token type")
	}

	return claims, nil
}

// GetUserIDFromToken 从令牌中获取用户ID
func (j *JWTAuth) GetUserIDFromToken(tokenString string) (uuid.UUID, error) {
	claims, err := j.ParseToken(tokenString)
	if err != nil {
		return uuid.Nil, err
	}

	return uuid.Parse(claims.UserID)
}
