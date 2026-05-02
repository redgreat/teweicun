/**
 * 功能：API流程测试 HTTP 客户端（统一响应解析/鉴权/错误输出）
 * 创建时间：2026-04-28
 * 创建人：GPT-5.2
 */

package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	BaseURL string
	Token   string
	HTTP    *http.Client
}

type APIError struct {
	HTTPStatus int
	Code       int
	Msg        string
	Body       string
}

func (e *APIError) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf("api error: http=%d code=%d msg=%q body=%s", e.HTTPStatus, e.Code, e.Msg, truncate(e.Body, 1000))
}

type envelope struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

type pageEnvelope struct {
	Total int64           `json:"total"`
	List  json.RawMessage `json:"list"`
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		HTTP: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

type LoginResp struct {
	Token    string   `json:"token"`
	UserID   int64    `json:"user_id"`
	Username string   `json:"username"`
	RealName string   `json:"real_name"`
	Roles    []string `json:"roles"`
}

func (c *Client) Login(ctx context.Context, username, password string) (*LoginResp, error) {
	reqBody := map[string]any{"username": username, "password": password}
	var out LoginResp
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/auth/login", nil, reqBody, &out); err != nil {
		return nil, err
	}
	c.Token = out.Token
	return &out, nil
}

func (c *Client) DoJSON(ctx context.Context, method, path string, query url.Values, in any, out any) error {
	fullURL := c.BaseURL + path
	if query != nil && len(query) > 0 {
		fullURL += "?" + query.Encode()
	}

	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)

	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return &APIError{HTTPStatus: resp.StatusCode, Code: -1, Msg: "invalid json response", Body: string(raw)}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || env.Code != 0 {
		return &APIError{HTTPStatus: resp.StatusCode, Code: env.Code, Msg: env.Msg, Body: string(raw)}
	}

	if out == nil {
		return nil
	}
	if len(env.Data) == 0 || string(env.Data) == "null" {
		return nil
	}
	return json.Unmarshal(env.Data, out)
}

func (c *Client) DoPage(ctx context.Context, method, path string, query url.Values, outList any, outTotal *int64) error {
	var page pageEnvelope
	if err := c.DoJSON(ctx, method, path, query, nil, &page); err != nil {
		return err
	}
	if outTotal != nil {
		*outTotal = page.Total
	}
	if outList == nil {
		return nil
	}
	return json.Unmarshal(page.List, outList)
}

func truncate(s string, n int) string {
	if n <= 0 || len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

