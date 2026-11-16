// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package notifications

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type SendResult struct {
	Retry  bool
	Reason string
}
type Sender interface {
	Send(context.Context, string, string, string, string) SendResult
}
type DingTalk struct{ transport http.RoundTripper }

func (d DingTalk) Send(ctx context.Context, raw, secret, title, message string) SendResult {
	if ValidateWebhook(raw) != nil || secret == "" {
		return SendResult{Reason: "机器人配置无效"}
	}
	u, _ := url.Parse(raw)
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp + "\n" + secret))
	q := u.Query()
	q.Set("timestamp", timestamp)
	q.Set("sign", base64.StdEncoding.EncodeToString(mac.Sum(nil)))
	u.RawQuery = q.Encode()
	payload, _ := json.Marshal(map[string]any{"msgtype": "markdown", "markdown": map[string]string{"title": title, "text": message}})
	req, err := http.NewRequestWithContext(ctx, "POST", u.String(), bytes.NewReader(payload))
	if err != nil {
		return SendResult{Reason: "无法创建通知请求"}
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 15 * time.Second, Transport: d.transport, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	response, err := client.Do(req)
	if err != nil {
		return SendResult{Retry: true, Reason: "网络错误或请求超时，送达状态未知"}
	}
	defer response.Body.Close()
	if response.StatusCode == 408 || response.StatusCode == 429 || response.StatusCode >= 500 {
		return SendResult{Retry: true, Reason: "钉钉限流或服务暂不可用"}
	}
	if response.StatusCode != 200 {
		return SendResult{Reason: "钉钉拒绝请求，请检查机器人配置"}
	}
	b, err := io.ReadAll(io.LimitReader(response.Body, 65537))
	if err != nil || len(b) > 65536 {
		return SendResult{Retry: true, Reason: "钉钉响应读取失败或超出限制"}
	}
	var result struct {
		Code    *int   `json:"errcode"`
		Message string `json:"errmsg"`
	}
	if json.Unmarshal(b, &result) != nil || result.Code == nil {
		return SendResult{Retry: true, Reason: "钉钉响应格式无效"}
	}
	responseMessage := strings.ToLower(result.Message)
	if *result.Code == 130101 || *result.Code == 90018 || (*result.Code != 0 && (strings.Contains(responseMessage, "too many") || strings.Contains(responseMessage, "frequency") || strings.Contains(responseMessage, "限流") || strings.Contains(responseMessage, "频率"))) {
		return SendResult{Retry: true, Reason: "钉钉发送频率受限"}
	}
	if *result.Code != 0 {
		return SendResult{Reason: "钉钉业务错误，请检查令牌、签名及机器人安全设置"}
	}
	return SendResult{}
}
