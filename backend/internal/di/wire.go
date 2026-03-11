//go:build wireinject
// +build wireinject

// Package di 提供依赖注入容器。
// 注意：本项目的 DI 容器由 wire_gen.go 手动维护，本文件仅为 wire 模板参考。
// 实际的依赖装配请直接修改 wire_gen.go 中的 NewContainer 函数。
package di

import (
	"github.com/Snowitty-Re/CNtunyuan/internal/application/service"
	"github.com/Snowitty-Re/CNtunyuan/internal/config"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/service"
	"github.com/Snowitty-Re/CNtunyuan/internal/infrastructure/auth"
	"github.com/Snowitty-Re/CNtunyuan/internal/infrastructure/cache"
	"github.com/Snowitty-Re/CNtunyuan/internal/infrastructure/database"
	infraRepo "github.com/Snowitty-Re/CNtunyuan/internal/infrastructure/repository"
	"github.com/Snowitty-Re/CNtunyuan/internal/interfaces/http/handler"
	"github.com/Snowitty-Re/CNtunyuan/internal/interfaces/http/middleware"
	"github.com/google/wire"
	"gorm.io/gorm"
)

// Container 依赖容器
type Container struct {
	Config         *config.Config
	DB             *gorm.DB
	Cache          cache.Cache
	AuthService    *service.AuthService
	UserService    *service.UserAppService
	UserHandler    *handler.UserHandler
	AuthHandler    *handler.AuthHandler
	AuthMiddleware *middleware.AuthMiddleware
}

// NewContainer 创建依赖容器
func NewContainer(cfg *config.Config) (*Container, error) {
	wire.Build(
		// 基础设施
		database.NewDatabase,
		provideCache,
		provideJWTService,

		// 仓储
		infraRepo.NewUserRepository,

		// 领域服务
		service.NewAuthService,

		// 应用服务
		service.NewUserAppService,

		// HTTP 处理
		handler.NewAuthHandler,
		handler.NewUserHandler,
		middleware.NewAuthMiddleware,

		// 容器
		wire.Struct(new(Container), "*"),
	)
	return nil, nil
}

// provideCache 提供缓存
func provideCache(cfg *config.Config) (cache.Cache, error) {
	return cache.NewRedis(&cfg.Redis)
}

// provideJWTService 提供JWT服务
func provideJWTService(cfg *config.Config, cache cache.Cache) service.TokenService {
	return auth.NewJWTService(&cfg.JWT, cache)
}
