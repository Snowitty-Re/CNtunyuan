package wechat

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/service"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
)

// Client 微信小程序客户端
type Client struct {
	appID      string
	appSecret  string
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
	reqURL := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		url.QueryEscape(c.appID),
		url.QueryEscape(c.appSecret),
		url.QueryEscape(code),
	)

	resp, err := c.httpClient.Get(reqURL)
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

// DecryptedUserInfo 解密后的微信用户信息
type DecryptedUserInfo struct {
	OpenID    string `json:"openId"`
	Nickname  string `json:"nickName"`
	Gender    int    `json:"gender"`
	City      string `json:"city"`
	Province  string `json:"province"`
	Country   string `json:"country"`
	AvatarURL string `json:"avatarUrl"`
	Language  string `json:"language"`
	UnionID   string `json:"unionId"`
	Watermark struct {
		AppID     string `json:"appid"`
		Timestamp int64  `json:"timestamp"`
	} `json:"watermark"`
}

// DecryptUserInfo 解密微信用户信息
// sessionKey: 从Code2Session获取的session_key
// encryptedData: 前端wx.getUserInfo返回的encryptedData
// iv: 前端wx.getUserInfo返回的iv
func (c *Client) DecryptUserInfo(sessionKey, encryptedData, iv string) (*DecryptedUserInfo, error) {
	plaintext, err := c.decrypt(sessionKey, encryptedData, iv)
	if err != nil {
		return nil, fmt.Errorf("decrypt user info failed: %w", err)
	}

	var userInfo DecryptedUserInfo
	if err := json.Unmarshal(plaintext, &userInfo); err != nil {
		return nil, fmt.Errorf("parse user info failed: %w", err)
	}

	// 验证watermark中的appid
	if userInfo.Watermark.AppID != c.appID {
		logger.Warn("WeChat watermark appid mismatch",
			logger.String("expected", c.appID),
			logger.String("actual", userInfo.Watermark.AppID),
		)
		return nil, fmt.Errorf("watermark appid mismatch")
	}

	return &userInfo, nil
}

// DecryptedPhoneInfo 解密后的手机号信息
type DecryptedPhoneInfo struct {
	PhoneNumber     string `json:"phoneNumber"`
	PurePhoneNumber string `json:"purePhoneNumber"`
	CountryCode     string `json:"countryCode"`
	Watermark       struct {
		AppID     string `json:"appid"`
		Timestamp int64  `json:"timestamp"`
	} `json:"watermark"`
}

// DecryptPhoneNumber 解密微信手机号
// sessionKey: 从Code2Session获取的session_key
// encryptedData: 前端getPhoneNumber返回的encryptedData
// iv: 前端getPhoneNumber返回的iv
func (c *Client) DecryptPhoneNumber(sessionKey, encryptedData, iv string) (*DecryptedPhoneInfo, error) {
	plaintext, err := c.decrypt(sessionKey, encryptedData, iv)
	if err != nil {
		return nil, fmt.Errorf("decrypt phone number failed: %w", err)
	}

	var phoneInfo DecryptedPhoneInfo
	if err := json.Unmarshal(plaintext, &phoneInfo); err != nil {
		return nil, fmt.Errorf("parse phone info failed: %w", err)
	}

	// 验证watermark中的appid
	if phoneInfo.Watermark.AppID != c.appID {
		logger.Warn("WeChat watermark appid mismatch",
			logger.String("expected", c.appID),
			logger.String("actual", phoneInfo.Watermark.AppID),
		)
		return nil, fmt.Errorf("watermark appid mismatch")
	}

	return &phoneInfo, nil
}

// GetPhoneNumber 通过code获取手机号（新版API，推荐使用）
// https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/user-info/phone-number/getPhoneNumber.html
func (c *Client) GetPhoneNumber(accessToken, code string) (*service.WechatPhoneInfo, error) {
	reqURL := fmt.Sprintf(
		"https://api.weixin.qq.com/wxa/business/getuserphonenumber?access_token=%s",
		url.QueryEscape(accessToken),
	)

	bodyStr := fmt.Sprintf(`{"code":"%s"}`, code)
	resp, err := c.httpClient.Post(reqURL, "application/json", bytes.NewBufferString(bodyStr))
	if err != nil {
		return nil, fmt.Errorf("wechat getPhoneNumber request failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		ErrCode   int    `json:"errcode"`
		ErrMsg    string `json:"errmsg"`
		PhoneInfo struct {
			PhoneNumber     string `json:"phoneNumber"`
			PurePhoneNumber string `json:"purePhoneNumber"`
			CountryCode     string `json:"countryCode"`
		} `json:"phone_info"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode wechat response failed: %w", err)
	}

	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wechat getPhoneNumber error: code=%d, msg=%s", result.ErrCode, result.ErrMsg)
	}

	return &service.WechatPhoneInfo{
		PhoneNumber:     result.PhoneInfo.PhoneNumber,
		PurePhoneNumber: result.PhoneInfo.PurePhoneNumber,
		CountryCode:     result.PhoneInfo.CountryCode,
	}, nil
}

// GetAccessToken 获取接口调用凭据（用于服务端API调用）
// https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/mp-access-token/getAccessToken.html
func (c *Client) GetAccessToken() (string, int, error) {
	reqURL := fmt.Sprintf(
		"https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		url.QueryEscape(c.appID),
		url.QueryEscape(c.appSecret),
	)

	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return "", 0, fmt.Errorf("wechat getAccessToken request failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", 0, fmt.Errorf("decode wechat response failed: %w", err)
	}

	if result.ErrCode != 0 {
		return "", 0, fmt.Errorf("wechat getAccessToken error: code=%d, msg=%s", result.ErrCode, result.ErrMsg)
	}

	return result.AccessToken, result.ExpiresIn, nil
}

// GetWebUserByCode 通过网页扫码登录 code 获取微信用户信息
// 适用于微信开放平台网站应用（scope=snsapi_login）
func (c *Client) GetWebUserByCode(code string) (*service.WechatWebUserInfo, error) {
	tokenURL := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code",
		url.QueryEscape(c.appID),
		url.QueryEscape(c.appSecret),
		url.QueryEscape(code),
	)

	tokenResp, err := c.httpClient.Get(tokenURL)
	if err != nil {
		return nil, fmt.Errorf("wechat web access_token request failed: %w", err)
	}
	defer tokenResp.Body.Close()

	body, _ := io.ReadAll(tokenResp.Body)
	var tokenResult struct {
		AccessToken string `json:"access_token"`
		OpenID      string `json:"openid"`
		UnionID     string `json:"unionid"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &tokenResult); err != nil {
		return nil, fmt.Errorf("decode wechat web token response failed: %w", err)
	}
	if tokenResult.ErrCode != 0 {
		return nil, fmt.Errorf("wechat web token error: code=%d, msg=%s", tokenResult.ErrCode, tokenResult.ErrMsg)
	}

	user := &service.WechatWebUserInfo{
		OpenID:  tokenResult.OpenID,
		UnionID: tokenResult.UnionID,
	}
	if strings.TrimSpace(tokenResult.AccessToken) == "" || strings.TrimSpace(tokenResult.OpenID) == "" {
		return user, nil
	}

	userInfoURL := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s&lang=zh_CN",
		url.QueryEscape(tokenResult.AccessToken),
		url.QueryEscape(tokenResult.OpenID),
	)
	userInfoResp, err := c.httpClient.Get(userInfoURL)
	if err != nil {
		return user, nil
	}
	defer userInfoResp.Body.Close()

	userInfoBody, _ := io.ReadAll(userInfoResp.Body)
	var userInfoResult struct {
		OpenID     string `json:"openid"`
		UnionID    string `json:"unionid"`
		Nickname   string `json:"nickname"`
		HeadImgURL string `json:"headimgurl"`
		ErrCode    int    `json:"errcode"`
		ErrMsg     string `json:"errmsg"`
	}
	if err := json.Unmarshal(userInfoBody, &userInfoResult); err != nil {
		return user, nil
	}
	if userInfoResult.ErrCode != 0 {
		return user, nil
	}

	if strings.TrimSpace(userInfoResult.OpenID) != "" {
		user.OpenID = strings.TrimSpace(userInfoResult.OpenID)
	}
	if strings.TrimSpace(userInfoResult.UnionID) != "" {
		user.UnionID = strings.TrimSpace(userInfoResult.UnionID)
	}
	user.Nickname = strings.TrimSpace(userInfoResult.Nickname)
	user.Avatar = strings.TrimSpace(userInfoResult.HeadImgURL)
	return user, nil
}

// decrypt AES-128-CBC解密微信加密数据
func (c *Client) decrypt(sessionKey, encryptedData, iv string) ([]byte, error) {
	keyBytes, err := base64.StdEncoding.DecodeString(sessionKey)
	if err != nil {
		return nil, fmt.Errorf("decode session key failed: %w", err)
	}

	dataBytes, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("decode encrypted data failed: %w", err)
	}

	ivBytes, err := base64.StdEncoding.DecodeString(iv)
	if err != nil {
		return nil, fmt.Errorf("decode iv failed: %w", err)
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("create cipher failed: %w", err)
	}

	if len(dataBytes)%block.BlockSize() != 0 {
		return nil, fmt.Errorf("encrypted data is not a multiple of block size")
	}

	mode := cipher.NewCBCDecrypter(block, ivBytes)
	plaintext := make([]byte, len(dataBytes))
	mode.CryptBlocks(plaintext, dataBytes)

	// PKCS#7 去填充
	plaintext, err = pkcs7Unpad(plaintext)
	if err != nil {
		return nil, fmt.Errorf("pkcs7 unpad failed: %w", err)
	}

	return plaintext, nil
}

// pkcs7Unpad 移除PKCS#7填充
func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}
	padding := int(data[len(data)-1])
	if padding > len(data) || padding == 0 {
		return nil, fmt.Errorf("invalid padding size")
	}
	for i := len(data) - padding; i < len(data); i++ {
		if int(data[i]) != padding {
			return nil, fmt.Errorf("invalid padding bytes")
		}
	}
	return data[:len(data)-padding], nil
}
