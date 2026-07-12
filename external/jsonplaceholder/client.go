package jsonplaceholder

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"clean-template/internal/config"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(cfg config.Config) *Client {
	return &Client{
		baseURL: cfg.ExternalAPIURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type postResponse struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

func (c *Client) FetchExample(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/posts/1", c.baseURL), nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("call external api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("external api returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response body: %w", err)
	}

	var post postResponse
	if err := json.Unmarshal(body, &post); err != nil {
		return "", fmt.Errorf("parse response body: %w", err)
	}

	return post.Title, nil
}
