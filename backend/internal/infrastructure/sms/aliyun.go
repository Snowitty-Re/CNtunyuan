// Package sms 阿里云短信服务实现（通过HTTP API直接调用，无需额外SDK）
package sms

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/config"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
	"github.com/google/uuid"
)

// AliyunProvider 阿里云短信提供商
type AliyunProvider struct {
	accessKeyID     string
	accessKeySecret string
	signName        string
	httpClient      *http.Client
}

// NewAliyunProvider 创建阿里云短信提供商
func NewAliyunProvider(cfg *config.SMSConfig) Provider {
	return &AliyunProvider{
		accessKeyID:     cfg.AliyunAccessKeyID,
		accessKeySecret: cfg.AliyunAccessSecret,
		signName:        cfg.SignName,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// aliyunResponse 阿里云短信API响应
type aliyunResponse struct {
	Code      string `json:"Code"`
	Message   string `json:"Message"`
	RequestID string `json:"RequestId"`
	BizID     string `json:"BizId"`
}

// SendSMS 发送短信
func (p *AliyunProvider) SendSMS(ctx context.Context, phone, signName, templateCode string, params map[string]string) error {
	if p.accessKeyID == "" || p.accessKeySecret == "" {
		return fmt.Errorf("aliyun SMS credentials not configured")
	}

	if signName == "" {
		signName = p.signName
	}

	// 序列化模板参数
	paramBytes, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("failed to marshal SMS params: %w", err)
	}

	// 构建公共请求参数
	queryParams := map[string]string{
		"AccessKeyId":      p.accessKeyID,
		"Action":           "SendSms",
		"Format":           "JSON",
		"PhoneNumbers":     phone,
		"RegionId":         "cn-hangzhou",
		"SignName":         signName,
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureNonce":   uuid.New().String(),
		"SignatureVersion": "1.0",
		"TemplateCode":     templateCode,
		"TemplateParam":    string(paramBytes),
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"Version":          "2017-05-25",
	}

	// 计算签名
	signature := p.computeSignature(queryParams)
	queryParams["Signature"] = signature

	// 构建请求URL
	reqURL := "https://dysmsapi.aliyuncs.com/?" + p.encodeParams(queryParams)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("aliyun SMS request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var result aliyunResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Code != "OK" {
		logger.Error("Aliyun SMS send failed",
			logger.String("code", result.Code),
			logger.String("message", result.Message),
			logger.String("phone", phone),
		)
		return fmt.Errorf("aliyun SMS error: %s - %s", result.Code, result.Message)
	}

	logger.Info("Aliyun SMS sent successfully",
		logger.String("phone", phone),
		logger.String("biz_id", result.BizID),
	)
	return nil
}

// computeSignature 计算阿里云API签名（HMAC-SHA1）
func (p *AliyunProvider) computeSignature(params map[string]string) string {
	// 1. 按参数名排序
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 2. 构建规范化查询字符串
	var pairs []string
	for _, k := range keys {
		pairs = append(pairs, specialURLEncode(k)+"="+specialURLEncode(params[k]))
	}
	canonicalQuery := strings.Join(pairs, "&")

	// 3. 构建待签名字符串: GET&%2F&<encoded_query>
	stringToSign := "GET&" + specialURLEncode("/") + "&" + specialURLEncode(canonicalQuery)

	// 4. HMAC-SHA1 签名
	mac := hmac.New(sha1.New, []byte(p.accessKeySecret+"&"))
	mac.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// encodeParams 编码URL参数
func (p *AliyunProvider) encodeParams(params map[string]string) string {
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}
	return values.Encode()
}

// specialURLEncode 阿里云特殊URL编码（RFC 3986）
func specialURLEncode(s string) string {
	encoded := url.QueryEscape(s)
	encoded = strings.ReplaceAll(encoded, "+", "%20")
	encoded = strings.ReplaceAll(encoded, "*", "%2A")
	encoded = strings.ReplaceAll(encoded, "%7E", "~")
	return encoded
}
