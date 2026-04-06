// Package sms 腾讯云短信服务实现（通过HTTP API直接调用，无需额外SDK）
package sms

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/config"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
)

// TencentProvider 腾讯云短信提供商
type TencentProvider struct {
	secretID   string
	secretKey  string
	appID      string
	signName   string
	httpClient *http.Client
}

// NewTencentProvider 创建腾讯云短信提供商
func NewTencentProvider(cfg *config.SMSConfig) Provider {
	return &TencentProvider{
		secretID:  cfg.TencentSecretID,
		secretKey: cfg.TencentSecretKey,
		appID:     cfg.TencentAppID,
		signName:  cfg.SignName,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// tencentSMSRequest 腾讯云短信请求体
type tencentSMSRequest struct {
	PhoneNumberSet   []string `json:"PhoneNumberSet"`
	SmsSdkAppId      string   `json:"SmsSdkAppId"`
	SignName         string   `json:"SignName"`
	TemplateId       string   `json:"TemplateId"`
	TemplateParamSet []string `json:"TemplateParamSet"`
}

// tencentSMSResponse 腾讯云短信响应
type tencentSMSResponse struct {
	Response struct {
		SendStatusSet []struct {
			SerialNo    string `json:"SerialNo"`
			PhoneNumber string `json:"PhoneNumber"`
			Fee         int    `json:"Fee"`
			SessionCtx  string `json:"SessionContext"`
			Code        string `json:"Code"`
			Message     string `json:"Message"`
			IsoCode     string `json:"IsoCode"`
		} `json:"SendStatusSet"`
		RequestId string `json:"RequestId"`
		Error     *struct {
			Code    string `json:"Code"`
			Message string `json:"Message"`
		} `json:"Error,omitempty"`
	} `json:"Response"`
}

// SendSMS 发送短信
func (p *TencentProvider) SendSMS(ctx context.Context, phone, signName, templateCode string, params map[string]string) error {
	if p.secretID == "" || p.secretKey == "" {
		return fmt.Errorf("tencent SMS credentials not configured")
	}

	if signName == "" {
		signName = p.signName
	}

	// 手机号需要带国际区号前缀
	if !strings.HasPrefix(phone, "+") {
		phone = "+86" + phone
	}

	// 构建模板参数（按key排序以保证顺序一致）
	templateParams := make([]string, 0, len(params))
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		templateParams = append(templateParams, params[k])
	}

	// 构建请求体
	reqBody := tencentSMSRequest{
		PhoneNumberSet:   []string{phone},
		SmsSdkAppId:      p.appID,
		SignName:         signName,
		TemplateId:       templateCode,
		TemplateParamSet: templateParams,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// 构建签名
	timestamp := time.Now().Unix()
	headers, err := p.buildAuthHeaders(bodyBytes, timestamp)
	if err != nil {
		return fmt.Errorf("failed to build auth headers: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://sms.tencentcloudapi.com", bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("tencent SMS request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var result tencentSMSResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// 检查全局错误
	if result.Response.Error != nil {
		logger.Error("Tencent SMS API error",
			logger.String("code", result.Response.Error.Code),
			logger.String("message", result.Response.Error.Message),
		)
		return fmt.Errorf("tencent SMS API error: %s - %s", result.Response.Error.Code, result.Response.Error.Message)
	}

	// 检查每个号码的发送状态
	for _, status := range result.Response.SendStatusSet {
		if status.Code != "Ok" {
			logger.Error("Tencent SMS send failed",
				logger.String("phone", status.PhoneNumber),
				logger.String("code", status.Code),
				logger.String("message", status.Message),
			)
			return fmt.Errorf("tencent SMS send failed: %s - %s", status.Code, status.Message)
		}
	}

	logger.Info("Tencent SMS sent successfully",
		logger.String("phone", phone),
		logger.String("request_id", result.Response.RequestId),
	)
	return nil
}

// buildAuthHeaders 构建腾讯云API v3签名（TC3-HMAC-SHA256）
func (p *TencentProvider) buildAuthHeaders(payload []byte, timestamp int64) (map[string]string, error) {
	const (
		service = "sms"
		host    = "sms.tencentcloudapi.com"
		action  = "SendSms"
		version = "2021-01-11"
	)

	date := time.Unix(timestamp, 0).UTC().Format("2006-01-02")

	// 1. 拼接规范请求串
	hashedPayload := sha256Hex(payload)
	canonicalRequest := fmt.Sprintf("POST\n/\n\ncontent-type:application/json\nhost:%s\n\ncontent-type;host\n%s",
		host, hashedPayload)

	// 2. 拼接待签名字符串
	credentialScope := fmt.Sprintf("%s/%s/tc3_request", date, service)
	stringToSign := fmt.Sprintf("TC3-HMAC-SHA256\n%d\n%s\n%s",
		timestamp, credentialScope, sha256Hex([]byte(canonicalRequest)))

	// 3. 计算签名
	secretDate := hmacSHA256([]byte("TC3"+p.secretKey), []byte(date))
	secretService := hmacSHA256(secretDate, []byte(service))
	secretSigning := hmacSHA256(secretService, []byte("tc3_request"))
	signature := hex.EncodeToString(hmacSHA256(secretSigning, []byte(stringToSign)))

	// 4. 构建Authorization头
	authorization := fmt.Sprintf("TC3-HMAC-SHA256 Credential=%s/%s, SignedHeaders=content-type;host, Signature=%s",
		p.secretID, credentialScope, signature)

	return map[string]string{
		"Authorization":  authorization,
		"Content-Type":   "application/json",
		"Host":           host,
		"X-TC-Action":    action,
		"X-TC-Version":   version,
		"X-TC-Timestamp": strconv.FormatInt(timestamp, 10),
		"X-TC-Region":    "ap-guangzhou",
	}, nil
}

// hmacSHA256 计算HMAC-SHA256
func hmacSHA256(key, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}

// sha256Hex 计算SHA256并返回hex字符串
func sha256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}
