package pkg

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Client HTTP客户端
type Client struct {
	baseURL     string
	httpClient  *http.Client
	apiKey      string
	secretKey   string
	accessToken string
	tokenExpiry time.Time
	mutex       sync.RWMutex
	environment string
}

// NewClient 创建新的API客户端
func NewClient(baseURL, apiKey, secretKey, environment string) *Client {
	// 创建支持HTTPS的HTTP客户端
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 如果使用HTTPS，配置TLS
	if len(baseURL) >= 5 && baseURL[:5] == "https" {
		// 生产环境使用系统默认证书，开发环境尝试使用自定义CA证书
		if environment == "production" {
			// 生产环境：使用系统默认的根证书，不跳过验证
			logrus.Info("Production environment detected, using system default certificates")
			transport := &http.Transport{
				TLSClientConfig: &tls.Config{
					// 使用系统默认的根证书
				},
			}
			httpClient.Transport = transport
		} else {
			// 开发环境：尝试使用自定义CA证书，支持多个可能的路径
			caCertPaths := []string{
				"../server/certs/localhost.crt",
				"../../server/certs/localhost.crt",
				"./certs/localhost.crt",
				"../certs/localhost.crt",
				"../../certs/localhost.crt",
			}

			var caCertPool *x509.CertPool
			for _, certPath := range caCertPaths {
				if caCert, err := os.ReadFile(certPath); err == nil {
					caCertPool = x509.NewCertPool()
					if caCertPool.AppendCertsFromPEM(caCert) {
						logrus.Infof("Successfully loaded CA certificate from: %s", certPath)
						break
					}
				}
			}

			if caCertPool != nil {
				// 使用自定义CA证书
				transport := &http.Transport{
					TLSClientConfig: &tls.Config{
						RootCAs: caCertPool,
					},
				}
				httpClient.Transport = transport
			} else {
				// 如果无法加载证书，使用跳过验证（仅用于开发环境）
				logrus.Warn("Failed to load CA certificate, using InsecureSkipVerify (not recommended for production)")
				transport := &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				}
				httpClient.Transport = transport
			}
		}
	}

	return &Client{
		baseURL:     baseURL,
		httpClient:  httpClient,
		apiKey:      apiKey,
		secretKey:   secretKey,
		environment: environment,
	}
}

// GetAccessTokenRequest 获取访问令牌请求
type GetAccessTokenRequest struct {
	APIKey    string `json:"api_key"`
	SecretKey string `json:"secret_key"`
}

// GetAccessTokenResponse 获取访问令牌响应
type GetAccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	ExpiresAt   int64  `json:"expires_at"`
}

// getAccessToken 获取访问令牌
func (c *Client) getAccessToken() error {
	req := GetAccessTokenRequest{
		APIKey:    c.apiKey,
		SecretKey: c.secretKey,
	}

	var resp GetAccessTokenResponse
	err := c.requestWithoutAuth("POST", "/api/auth/token", req, &resp)
	if err != nil {
		return fmt.Errorf("获取访问令牌失败: %w", err)
	}

	c.mutex.Lock()
	c.accessToken = resp.AccessToken
	c.tokenExpiry = time.Unix(resp.ExpiresAt, 0).Add(-5 * time.Minute) // 提前5分钟刷新
	c.mutex.Unlock()

	return nil
}

// ensureValidToken 确保有有效的访问令牌
func (c *Client) ensureValidToken() error {
	c.mutex.RLock()
	needRefresh := c.accessToken == "" || time.Now().After(c.tokenExpiry)
	c.mutex.RUnlock()

	if needRefresh {
		return c.getAccessToken()
	}
	return nil
}

// requestWithoutAuth 执行不需要认证的HTTP请求
func (c *Client) requestWithoutAuth(method, endpoint string, payload interface{}, result interface{}) error {
	prefix := ""
	if c.environment == "production" {
		// 线上Nginx已挂了/api，后端服务自身也带/api，最终需要/api/api
		prefix = ""
	}
	url := c.baseURL + prefix + endpoint

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

// request 执行HTTP请求的核心方法
func (c *Client) request(method, endpoint string, payload interface{}, result interface{}) error {
	// 确保有有效的访问令牌
	if err := c.ensureValidToken(); err != nil {
		return fmt.Errorf("确保访问令牌失败: %w", err)
	}

	prefix := ""
	if c.environment == "production" {
		// 线上Nginx已挂了/api，后端服务自身也带/api，最终需要/api/api
		prefix = ""
	}
	url := c.baseURL + prefix + endpoint

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
	c.mutex.RLock()
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}
	c.mutex.RUnlock()

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

// PUT 执行PUT请求
func (c *Client) PUT(endpoint string, payload interface{}, result interface{}) error {
	return c.request("PUT", endpoint, payload, result)
}

// DELETE 执行DELETE请求
func (c *Client) DELETE(endpoint string, result interface{}) error {
	return c.request("DELETE", endpoint, nil, result)
}
