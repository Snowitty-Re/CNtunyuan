package router

import (
	"strings"

	"github.com/Snowitty-Re/CNtunyuan/internal/config"
	"github.com/Snowitty-Re/CNtunyuan/internal/interfaces/http/handler"
	pkgmiddleware "github.com/Snowitty-Re/CNtunyuan/pkg/middleware"
	"github.com/Snowitty-Re/CNtunyuan/pkg/response"
	"github.com/gin-gonic/gin"
)

func NewBootstrapEngine(bootstrapHandler *handler.BootstrapHandler) *gin.Engine {
	engine := gin.New()
	engine.Use(pkgmiddleware.RecoveryMiddleware())
	engine.Use(pkgmiddleware.TraceIDMiddleware())
	engine.Use(pkgmiddleware.SecurityHeadersMiddleware())

	var corsOrigins []string
	cfg := config.GetConfig()
	if cfg != nil && strings.TrimSpace(cfg.Server.CORSOrigins) != "" {
		corsOrigins = strings.Split(cfg.Server.CORSOrigins, ",")
		for i := range corsOrigins {
			corsOrigins[i] = strings.TrimSpace(corsOrigins[i])
		}
	}
	engine.Use(pkgmiddleware.CORSMiddleware(corsOrigins...))
	engine.Use(pkgmiddleware.RequestSizeMiddleware(10 * 1024 * 1024))
	engine.Use(pkgmiddleware.LoggingMiddleware())
	engine.Use(pkgmiddleware.ErrorHandlerMiddleware())

	api := engine.Group("/api/v1")
	api.GET("/health", func(c *gin.Context) {
		response.Success(c, gin.H{
			"status": "UP",
			"mode":   "bootstrap",
		})
	})
	api.GET("/", func(c *gin.Context) {
		response.Success(c, gin.H{
			"name":        "助力团圆志愿者系统",
			"description": "系统处于初始化模式，请先完成首启引导",
			"bootstrap":   "/api/v1/bootstrap/status",
		})
	})
	if bootstrapHandler != nil {
		bootstrapHandler.RegisterRoutes(api)
	}

	engine.NoRoute(func(c *gin.Context) {
		response.NotFound(c, "route not found")
	})
	engine.NoMethod(func(c *gin.Context) {
		response.ErrorCodeWithMessage(c, 405, "method not allowed")
	})

	return engine
}
