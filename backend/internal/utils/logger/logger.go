package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/config"
)

// Logger 日志接口
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	With(fields ...Field) Logger
}

// Field 日志字段
type Field struct {
	Key   string
	Value interface{}
}

// Any 创建任意类型字段
func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// String 创建字符串字段
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int 创建整数字段
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Int64 创建64位整数字段
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

// Float64 创建浮点数字段
func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

// Bool 创建布尔字段
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// ErrorField 创建错误字段
func ErrorField(err error) Field {
	return Field{Key: "error", Value: err}
}

// simpleLogger 简单日志实现
type simpleLogger struct {
	cfg    *config.LogConfig
	fields []Field
}

// NewLogger 创建日志实例
func NewLogger(cfg *config.LogConfig) (Logger, error) {
	// 确保日志目录存在
	if cfg.OutputPath != "" {
		if err := os.MkdirAll(cfg.OutputPath, 0755); err != nil {
			return nil, fmt.Errorf("创建日志目录失败: %w", err)
		}
	}

	return &simpleLogger{
		cfg:    cfg,
		fields: make([]Field, 0),
	}, nil
}

func (l *simpleLogger) log(level, msg string, fields []Field) {
	if !l.shouldLog(level) {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// 构建字段字符串
	var fieldStr string
	allFields := append(l.fields, fields...)
	if len(allFields) > 0 {
		fieldStr = " {"
		for i, f := range allFields {
			if i > 0 {
				fieldStr += ", "
			}
			fieldStr += fmt.Sprintf("%s: %v", f.Key, f.Value)
		}
		fieldStr += "}"
	}

	// 格式化输出
	var output string
	if l.cfg.Format == "json" {
		output = fmt.Sprintf(`{"timestamp":"%s","level":"%s","message":"%s"`, timestamp, level, msg)
		for _, f := range allFields {
			output += fmt.Sprintf(",\"%s\":%v", f.Key, f.Value)
		}
		output += "}\n"
	} else {
		output = fmt.Sprintf("[%s] [%s] %s%s\n", timestamp, level, msg, fieldStr)
	}

	// 输出到控制台
	fmt.Print(output)

	// 输出到文件
	if l.cfg.OutputPath != "" {
		l.writeToFile(output)
	}
}

func (l *simpleLogger) writeToFile(content string) {
	fileName := filepath.Join(l.cfg.OutputPath, l.cfg.FileName)
	
	// 打开文件（追加模式）
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "打开日志文件失败: %v\n", err)
		return
	}
	defer f.Close()

	if _, err := f.WriteString(content); err != nil {
		fmt.Fprintf(os.Stderr, "写入日志文件失败: %v\n", err)
	}
}

func (l *simpleLogger) shouldLog(level string) bool {
	levels := map[string]int{
		"debug": 0,
		"info":  1,
		"warn":  2,
		"error": 3,
		"fatal": 4,
	}

	configLevel := levels[l.cfg.Level]
	msgLevel := levels[level]

	return msgLevel >= configLevel
}

func (l *simpleLogger) Debug(msg string, fields ...Field) {
	l.log("debug", msg, fields)
}

func (l *simpleLogger) Info(msg string, fields ...Field) {
	l.log("info", msg, fields)
}

func (l *simpleLogger) Warn(msg string, fields ...Field) {
	l.log("warn", msg, fields)
}

func (l *simpleLogger) Error(msg string, fields ...Field) {
	l.log("error", msg, fields)
}

func (l *simpleLogger) Fatal(msg string, fields ...Field) {
	l.log("fatal", msg, fields)
	os.Exit(1)
}

func (l *simpleLogger) With(fields ...Field) Logger {
	return &simpleLogger{
		cfg:    l.cfg,
		fields: append(l.fields, fields...),
	}
}

// 全局日志实例
var globalLogger Logger

// Init 初始化全局日志
func Init(cfg *config.LogConfig) error {
	logger, err := NewLogger(cfg)
	if err != nil {
		return err
	}
	globalLogger = logger
	return nil
}

// GetLogger 获取全局日志实例
func GetLogger() Logger {
	if globalLogger == nil {
		// 返回一个默认的日志实例
		return &simpleLogger{
			cfg: &config.LogConfig{
				Level:  "info",
				Format: "text",
			},
		}
	}
	return globalLogger
}

// Debug 全局Debug日志
func Debug(msg string, fields ...Field) {
	GetLogger().Debug(msg, fields...)
}

// Info 全局Info日志
func Info(msg string, fields ...Field) {
	GetLogger().Info(msg, fields...)
}

// Warn 全局Warn日志
func Warn(msg string, fields ...Field) {
	GetLogger().Warn(msg, fields...)
}

// Error 全局Error日志
func ErrorLog(msg string, fields ...Field) {
	GetLogger().Error(msg, fields...)
}

// Fatal 全局Fatal日志
func FatalLog(msg string, fields ...Field) {
	GetLogger().Fatal(msg, fields...)
}

// With 创建带字段的日志实例
func With(fields ...Field) Logger {
	return GetLogger().With(fields...)
}
