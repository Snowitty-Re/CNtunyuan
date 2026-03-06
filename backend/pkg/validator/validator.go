// Package validator 提供请求参数验证
package validator

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/Snowitty-Re/CNtunyuan/pkg/errors"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	
	// 注册自定义验证规则
	validate.RegisterValidation("phone", validatePhone)
	validate.RegisterValidation("idcard", validateIDCard)
}

// Validate 获取验证器实例
func Validate() *validator.Validate {
	return validate
}

// ValidateStruct 验证结构体
func ValidateStruct(s interface{}) error {
	if err := validate.Struct(s); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			return formatValidationErrors(validationErrors)
		}
		return errors.Wrap(err, errors.CodeInvalidParam, "validation failed")
	}
	return nil
}

// ValidateVar 验证单个变量
func ValidateVar(field interface{}, tag string) error {
	if err := validate.Var(field, tag); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			return formatValidationErrors(validationErrors)
		}
		return errors.Wrap(err, errors.CodeInvalidParam, "validation failed")
	}
	return nil
}

// formatValidationErrors 格式化验证错误
func formatValidationErrors(errs validator.ValidationErrors) error {
	var messages []string
	for _, err := range errs {
		messages = append(messages, formatError(err))
	}
	return errors.New(errors.CodeInvalidParam, strings.Join(messages, "; "))
}

// formatError 格式化单个验证错误
func formatError(err validator.FieldError) string {
	field := err.Field()
	tag := err.Tag()
	param := err.Param()

	switch tag {
	case "required":
		return field + " 为必填项"
	case "min":
		return field + " 最小长度为 " + param
	case "max":
		return field + " 最大长度为 " + param
	case "len":
		return field + " 长度必须为 " + param
	case "email":
		return field + " 格式不正确"
	case "phone":
		return field + " 格式不正确"
	case "idcard":
		return field + " 格式不正确"
	case "gte":
		return field + " 必须大于等于 " + param
	case "lte":
		return field + " 必须小于等于 " + param
	case "gt":
		return field + " 必须大于 " + param
	case "lt":
		return field + " 必须小于 " + param
	case "oneof":
		return field + " 必须是以下之一: " + param
	case "numeric":
		return field + " 必须是数字"
	case "alphanum":
		return field + " 只能包含字母和数字"
	case "uuid":
		return field + " 格式不正确"
	case "url":
		return field + " URL格式不正确"
	case "datetime":
		return field + " 日期时间格式不正确"
	default:
		return field + " 验证失败: " + tag
	}
}

// validatePhone 验证手机号
func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	if phone == "" {
		return true
	}
	// 中国大陆手机号正则
	pattern := `^1[3-9]\d{9}$`
	matched, _ := regexp.MatchString(pattern, phone)
	return matched
}

// validateIDCard 验证身份证号
func validateIDCard(fl validator.FieldLevel) bool {
	idcard := fl.Field().String()
	if idcard == "" {
		return true
	}
	// 15位或18位身份证
	pattern := `(^\d{15}$)|(^\d{18}$)|(^\d{17}(\d|X|x)$)`
	matched, _ := regexp.MatchString(pattern, idcard)
	return matched
}

// IsValidPhone 验证手机号是否有效
func IsValidPhone(phone string) bool {
	if phone == "" {
		return false
	}
	pattern := `^1[3-9]\d{9}$`
	matched, _ := regexp.MatchString(pattern, phone)
	return matched
}

// IsValidEmail 验证邮箱是否有效
func IsValidEmail(email string) bool {
	if email == "" {
		return false
	}
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	return matched
}

// IsValidIDCard 验证身份证号是否有效
func IsValidIDCard(idcard string) bool {
	if idcard == "" {
		return false
	}
	pattern := `(^\d{15}$)|(^\d{18}$)|(^\d{17}(\d|X|x)$)`
	matched, _ := regexp.MatchString(pattern, idcard)
	return matched
}

// IsValidUUID 验证 UUID 是否有效
func IsValidUUID(uuid string) bool {
	if uuid == "" {
		return false
	}
	pattern := `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`
	matched, _ := regexp.MatchString(pattern, uuid)
	return matched
}

// SanitizeString 清理字符串（防止 XSS）
func SanitizeString(s string) string {
	// 去除前后空白
	s = strings.TrimSpace(s)
	
	// 限制长度
	if utf8.RuneCountInString(s) > 10000 {
		runes := []rune(s)
		s = string(runes[:10000])
	}
	
	return s
}

// TruncateString 截断字符串
func TruncateString(s string, maxLen int) string {
	if maxLen <= 0 {
		return s
	}
	
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	
	return string(runes[:maxLen])
}
