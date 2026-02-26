package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// WeChatService 微信服务
type WeChatService struct {
	appID     string
	appSecret string
}

// NewWeChatService 创建微信服务
func NewWeChatService(appID, appSecret string) *WeChatService {
	return &WeChatService{
		appID:     appID,
		appSecret: appSecret,
	}
}

// Code2SessionResponse 微信登录响应
type Code2SessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// GetUserInfoResponse 用户信息响应
type GetUserInfoResponse struct {
	OpenID    string `json:"openid"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"headimgurl"`
	UnionID   string `json:"unionid"`
	ErrCode   int    `json:"errcode"`
	ErrMsg    string `json:"errmsg"`
}

// Code2Session 登录凭证校验
func (s *WeChatService) Code2Session(code string) (*Code2SessionResponse, error) {
	apiURL := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		url.QueryEscape(s.appID),
		url.QueryEscape(s.appSecret),
		url.QueryEscape(code),
	)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("请求微信接口失败: %w", err)
	}
	defer resp.Body.Close()

	var result Code2SessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if result.ErrCode != 0 {
		return nil, fmt.Errorf("微信接口错误: %s", result.ErrMsg)
	}

	return &result, nil
}

// GetAccessToken 获取AccessToken
func (s *WeChatService) GetAccessToken() (string, error) {
	apiURL := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		url.QueryEscape(s.appID),
		url.QueryEscape(s.appSecret),
	)

	resp, err := http.Get(apiURL)
	if err != nil {
		return "", fmt.Errorf("请求微信接口失败: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if result.ErrCode != 0 {
		return "", fmt.Errorf("微信接口错误: %s", result.ErrMsg)
	}

	return result.AccessToken, nil
}

// DecryptUserInfo 解密用户信息
func (s *WeChatService) DecryptUserInfo(encryptedData, sessionKey, iv string) (map[string]interface{}, error) {
	// 这里应该使用微信提供的解密算法
	// 简化处理，实际项目中需要实现完整的解密逻辑
	return map[string]interface{}{
		"nickName":  "微信用户",
		"avatarUrl": "",
	}, nil
}
