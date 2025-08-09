package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client HTTP客户端
type Client struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
}

// NewClient 创建新的API客户端
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiKey: apiKey,
	}
}

// request 执行HTTP请求的核心方法
func (c *Client) request(method, endpoint string, payload interface{}, result interface{}) error {
	url := c.baseURL + endpoint

	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("marshal payload failed: %w", err)
		}
		body = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response failed: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(responseBody))
	}

	// 解析通用响应格式
	var apiResp APIResponse
	if err := json.Unmarshal(responseBody, &apiResp); err != nil {
		return fmt.Errorf("unmarshal response failed: %w", err)
	}

	if !apiResp.Success {
		return fmt.Errorf("API error: %s", apiResp.Error)
	}

	// 如果需要解析result
	if result != nil && apiResp.Data != nil {
		dataBytes, err := json.Marshal(apiResp.Data)
		if err != nil {
			return fmt.Errorf("marshal data failed: %w", err)
		}

		if err := json.Unmarshal(dataBytes, result); err != nil {
			return fmt.Errorf("unmarshal result failed: %w", err)
		}
	}

	return nil
}

// GET 执行GET请求
func (c *Client) GET(endpoint string, result interface{}) error {
	return c.request("GET", endpoint, nil, result)
}

// POST 执行POST请求
func (c *Client) POST(endpoint string, payload interface{}, result interface{}) error {
	return c.request("POST", endpoint, payload, result)
}
