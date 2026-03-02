package logger

import (
	"os"
	"path/filepath"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var globalLogger *zap.Logger

// Config 简化版日志配置
type Config struct {
	Level    string
	Format   string
	Output   string
	Filename string
}

// InitWithConfig 使用简化配置初始化日志
func InitWithConfig(cfg Config) error {
	logCfg := &config.LogConfig{
		Level:      cfg.Level,
		Format:     cfg.Format,
		OutputPath: "",
	}
	if cfg.Output == "file" && cfg.Filename != "" {
		logCfg.OutputPath = filepath.Dir(cfg.Filename)
	}
	return Init(logCfg)
}

// Init 初始化日志
func Init(cfg *config.LogConfig) error {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 解析日志级别
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	// 创建核心
	var cores []zapcore.Core

	// 控制台输出
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level)
	cores = append(cores, consoleCore)

	// 文件输出
	if cfg.OutputPath != "" {
		if err := os.MkdirAll(cfg.OutputPath, 0755); err != nil {
			return err
		}

		fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
		logFile := filepath.Join(cfg.OutputPath, "app.log")

		// 使用 lumberjack 进行日志切割
		hook := &lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    100, // MB
			MaxBackups: 10,
			MaxAge:     30, // days
			Compress:   true,
		}

		fileCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(hook), level)
		cores = append(cores, fileCore)
	}

	// 创建 Logger
	core := zapcore.NewTee(cores...)
	globalLogger = zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	return nil
}

// GetLogger 获取全局 Logger
func GetLogger() *zap.Logger {
	if globalLogger == nil {
		return zap.NewNop()
	}
	return globalLogger
}

// SetLogger 设置全局 Logger
func SetLogger(logger *zap.Logger) {
	globalLogger = logger
}

// Sync 同步日志
func Sync() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}

// Debug 调试日志
func Debug(msg string, fields ...zap.Field) {
	GetLogger().Debug(msg, fields...)
}

// Info 信息日志
func Info(msg string, fields ...zap.Field) {
	GetLogger().Info(msg, fields...)
}

// Warn 警告日志
func Warn(msg string, fields ...zap.Field) {
	GetLogger().Warn(msg, fields...)
}

// Error 错误日志
func Error(msg string, fields ...zap.Field) {
	GetLogger().Error(msg, fields...)
}

// Fatal 致命日志
func Fatal(msg string, fields ...zap.Field) {
	GetLogger().Fatal(msg, fields...)
}

// With 创建带字段的 Logger
func With(fields ...zap.Field) *zap.Logger {
	return GetLogger().With(fields...)
}

// WithContext 从上下文创建 Logger
func WithContext(ctx interface{}) *zap.Logger {
	// 可以扩展支持从 context.Context 中提取 trace_id 等
	return GetLogger()
}

// LogRequest 记录请求日志
func LogRequest(method, path, clientIP string, status int, duration time.Duration, err error) {
	fields := []zap.Field{
		zap.String("method", method),
		zap.String("path", path),
		zap.String("client_ip", clientIP),
		zap.Int("status", status),
		zap.Duration("duration", duration),
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
		Error("HTTP Request", fields...)
		return
	}

	Info("HTTP Request", fields...)
}

// LogDBOperation 记录数据库操作日志
func LogDBOperation(operation, table string, duration time.Duration, rows int64, err error) {
	fields := []zap.Field{
		zap.String("operation", operation),
		zap.String("table", table),
		zap.Duration("duration", duration),
		zap.Int64("rows", rows),
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
		Error("DB Operation", fields...)
		return
	}

	Debug("DB Operation", fields...)
}

// LogBusiness 记录业务日志
func LogBusiness(action, resource, resourceID string, userID string, err error) {
	fields := []zap.Field{
		zap.String("action", action),
		zap.String("resource", resource),
		zap.String("resource_id", resourceID),
		zap.String("user_id", userID),
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
		Error("Business Operation", fields...)
		return
	}

	Info("Business Operation", fields...)
}
