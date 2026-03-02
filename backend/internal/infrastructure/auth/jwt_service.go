package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/config"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/service"
	"github.com/Snowitty-Re/CNtunyuan/internal/infrastructure/cache"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenExpired  = errors.New("token expired")
	ErrTokenInvalid  = errors.New("token invalid")
	ErrTokenNotFound = errors.New("token not found")
)

// JWTService JWT service implementation
type JWTService struct {
	secret        []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
	cache         cache.Cache
}

// JWTClaims JWT claims
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Nickname string `json:"nickname"`
	Role     string `json:"role"`
	OrgID    string `json:"org_id"`
	jwt.RegisteredClaims
}

// NewJWTService create JWT service
func NewJWTService(cfg *config.JWTConfig, cache cache.Cache) service.TokenService {
	return &JWTService{
		secret:        []byte(cfg.Secret),
		accessExpiry:  time.Duration(cfg.ExpireTime) * time.Second,
		refreshExpiry: time.Duration(cfg.ExpireTime*2) * time.Second,
		cache:         cache,
	}
}

// GenerateTokenPair generate token pair
func (s *JWTService) GenerateTokenPair(ctx context.Context, user *entity.User) (*service.TokenPair, error) {
	accessToken, err := s.generateToken(user, s.accessExpiry)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateToken(user, s.refreshExpiry)
	if err != nil {
		return nil, err
	}

	return &service.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(s.accessExpiry.Seconds()),
	}, nil
}

// generateToken generate token
func (s *JWTService) generateToken(user *entity.User, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		UserID:   user.ID,
		Nickname: user.Nickname,
		Role:     string(user.Role),
		OrgID:    user.OrgID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "cntuanyuan",
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// ValidateToken validate token
func (s *JWTService) ValidateToken(ctx context.Context, tokenString string) (*service.TokenClaims, error) {
	if s.cache != nil {
		blacklisted, err := s.cache.Exists(ctx, cacheKeyBlacklisted(tokenString))
		if err == nil && blacklisted {
			return nil, ErrTokenInvalid
		}
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		logger.Warn("Token parse failed", logger.Err(err))
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}

	return &service.TokenClaims{
		UserID:    claims.UserID,
		Nickname:  claims.Nickname,
		Role:      entity.Role(claims.Role),
		OrgID:     claims.OrgID,
		IssuedAt:  claims.IssuedAt.Time,
		ExpiresAt: claims.ExpiresAt.Time,
	}, nil
}

// RevokeToken revoke token
func (s *JWTService) RevokeToken(ctx context.Context, tokenString string) error {
	if s.cache == nil {
		return nil
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return s.secret, nil
	})

	if err != nil {
		return s.cache.Set(ctx, cacheKeyBlacklisted(tokenString), true, 24*time.Hour)
	}

	if claims, ok := token.Claims.(*JWTClaims); ok {
		remaining := time.Until(claims.ExpiresAt.Time)
		if remaining > 0 {
			return s.cache.Set(ctx, cacheKeyBlacklisted(tokenString), true, remaining)
		}
	}

	return nil
}

// cacheKeyBlacklisted blacklisted token cache key
func cacheKeyBlacklisted(token string) string {
	return cache.CacheKey("blacklist", token)
}
