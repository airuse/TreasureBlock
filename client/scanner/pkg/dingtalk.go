package pkg

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

// DingTalkConfig 钉钉机器人配置
type DingTalkConfig struct {
	WebhookURL string `yaml:"webhook_url"`
	Secret     string `yaml:"secret"`
	Enabled    bool   `yaml:"enabled"`
}

// DingTalkMessage 钉钉消息结构
type DingTalkMessage struct {
	MsgType string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
	At struct {
		AtMobiles []string `json:"atMobiles,omitempty"`
		IsAtAll   bool     `json:"isAtAll,omitempty"`
	} `json:"at,omitempty"`
}

// DingTalkResponse 钉钉响应结构
type DingTalkResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// DingTalkNotifier 钉钉通知器
type DingTalkNotifier struct {
	config *DingTalkConfig
	client *http.Client
}

// NewDingTalkNotifier 创建钉钉通知器
func NewDingTalkNotifier(config *DingTalkConfig) *DingTalkNotifier {
	return &DingTalkNotifier{
		config: config,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendAlert 发送告警消息
func (d *DingTalkNotifier) SendAlert(title, message string, isFatal bool) error {
	if !d.config.Enabled {
		logrus.Debug("DingTalk notification is disabled")
		return nil
	}

	// 构建消息内容
	content := d.buildMessageContent(title, message, isFatal)

	// 构建钉钉消息
	dingTalkMsg := DingTalkMessage{
		MsgType: "text",
		Text: struct {
			Content string `json:"content"`
		}{
			Content: content,
		},
	}

	// 如果是致命错误，@所有人
	if isFatal {
		dingTalkMsg.At.IsAtAll = true
	}

	// 发送消息
	return d.sendMessage(dingTalkMsg)
}

// buildMessageContent 构建消息内容
func (d *DingTalkNotifier) buildMessageContent(title, message string, isFatal bool) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	var emoji string
	if isFatal {
		emoji = "🚨"
	} else {
		emoji = "⚠️"
	}

	content := fmt.Sprintf(`%s **%s**
⏰ 时间: %s
📝 详情: %s

---
💡 请及时处理相关问题`, emoji, title, timestamp, message)

	return content
}

// sendMessage 发送消息到钉钉
func (d *DingTalkNotifier) sendMessage(msg DingTalkMessage) error {
	// 序列化消息
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// 构建请求URL（包含签名）
	url, err := d.buildSignedURL()
	if err != nil {
		return fmt.Errorf("failed to build signed URL: %w", err)
	}

	// 发送POST请求
	resp, err := d.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var dingTalkResp DingTalkResponse
	if err := json.NewDecoder(resp.Body).Decode(&dingTalkResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// 检查响应状态
	if dingTalkResp.ErrCode != 0 {
		return fmt.Errorf("dingtalk API error: %d - %s", dingTalkResp.ErrCode, dingTalkResp.ErrMsg)
	}

	logrus.Infof("DingTalk notification sent successfully: %s", msg.Text.Content)
	return nil
}

// buildSignedURL 构建带签名的URL
func (d *DingTalkNotifier) buildSignedURL() (string, error) {
	if d.config.Secret == "" {
		// 没有密钥，直接返回原始URL
		return d.config.WebhookURL, nil
	}

	// 生成时间戳（毫秒）
	timestamp := time.Now().UnixNano() / 1e6
	timestampStr := strconv.FormatInt(timestamp, 10)

	// 构建签名字符串
	signString := fmt.Sprintf("%s\n%s", timestampStr, d.config.Secret)

	// 计算HMAC-SHA256签名
	h := hmac.New(sha256.New, []byte(d.config.Secret))
	h.Write([]byte(signString))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// 构建最终URL
	url := fmt.Sprintf("%s&timestamp=%s&sign=%s", d.config.WebhookURL, timestampStr, signature)

	return url, nil
}

// SendFatalError 发送致命错误通知
func (d *DingTalkNotifier) SendFatalError(operation string, err error) error {
	title := fmt.Sprintf("扫块程序致命错误 - %s", operation)
	message := fmt.Sprintf("程序遇到致命错误，即将退出\n\n错误详情: %v", err)

	return d.SendAlert(title, message, true)
}

// SendWarning 发送警告通知
func (d *DingTalkNotifier) SendWarning(title, message string) error {
	return d.SendAlert(title, message, false)
}

// SendInfo 发送信息通知
func (d *DingTalkNotifier) SendInfo(title, message string) error {
	return d.SendAlert(title, message, false)
}
