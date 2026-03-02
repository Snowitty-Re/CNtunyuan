package logger

import (
	"time"

	"go.uber.org/zap"
)

// String 创建字符串字段
func String(key, val string) zap.Field {
	return zap.String(key, val)
}

// Int 创建整数字段
func Int(key string, val int) zap.Field {
	return zap.Int(key, val)
}

// Int64 创建 int64 字段
func Int64(key string, val int64) zap.Field {
	return zap.Int64(key, val)
}

// Uint 创建 uint 字段
func Uint(key string, val uint) zap.Field {
	return zap.Uint(key, val)
}

// Uint64 创建 uint64 字段
func Uint64(key string, val uint64) zap.Field {
	return zap.Uint64(key, val)
}

// Float64 创建 float64 字段
func Float64(key string, val float64) zap.Field {
	return zap.Float64(key, val)
}

// Bool 创建布尔字段
func Bool(key string, val bool) zap.Field {
	return zap.Bool(key, val)
}

// Duration 创建持续时间字段
func Duration(key string, val time.Duration) zap.Field {
	return zap.Duration(key, val)
}

// Time 创建时间字段
func Time(key string, val time.Time) zap.Field {
	return zap.Time(key, val)
}

// Err 创建错误字段（命名为 Err 避免与 Error 日志函数冲突）
func Err(err error) zap.Field {
	return zap.Error(err)
}

// Any 创建任意类型字段
func Any(key string, val interface{}) zap.Field {
	return zap.Any(key, val)
}

// ByteString 创建字节切片字段
func ByteString(key string, val []byte) zap.Field {
	return zap.ByteString(key, val)
}

// Namespace 创建命名空间
func Namespace(key string) zap.Field {
	return zap.Namespace(key)
}
