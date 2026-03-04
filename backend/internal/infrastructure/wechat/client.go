package wechat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/service"
)

// Client 微信小程序客户端
type Client struct {
	appID     string
	appSecret string
	httpClient *http.Client
}

// NewClient 创建微信客户端
func NewClient(appID, appSecret string) *Client {
	return &Client{
		appID:     appID,
		appSecret: appSecret,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Code2Session 登录凭证校验
// https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/user-login/code2Session.html
func (c *Client) Code2Session(code string) (*service.WechatSession, error) {
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		c.appID,
		c.appSecret,
		code,
	)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("wechat api request failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		OpenID     string `json:"openid"`
		SessionKey string `json:"session_key"`
		UnionID    string `json:"unionid"`
		ErrCode    int    `json:"errcode"`
		ErrMsg     string `json:"errmsg"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode wechat response failed: %w", err)
	}

	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wechat api error: code=%d, msg=%s", result.ErrCode, result.ErrMsg)
	}

	return &service.WechatSession{
		OpenID:     result.OpenID,
		SessionKey: result.SessionKey,
		UnionID:    result.UnionID,
	}, nil
}
