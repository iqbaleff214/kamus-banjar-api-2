package openrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/iqbaleff214/kamus-banjar-api-2/internal/ai/domain"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/config"
)

type Client struct {
	apiKey  string
	baseURL string
	model   string
	http    *http.Client
}

func New(cfg *config.Config) (*Client, error) {
	if cfg.OpenRouterAPIKey == "" {
		return nil, fmt.Errorf("OPENROUTER_API_KEY is required")
	}
	return &Client{
		apiKey:  cfg.OpenRouterAPIKey,
		baseURL: cfg.OpenRouterBaseURL,
		model:   cfg.OpenRouterModel,
		http:    &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// openRouterRequest is the wire format sent to OpenRouter.
type openRouterRequest struct {
	Model       string           `json:"model"`
	Messages    []domain.Message `json:"messages"`
	Temperature float64          `json:"temperature"`
}

// openRouterResponse is the wire format received from OpenRouter.
type openRouterResponse struct {
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (c *Client) Complete(ctx context.Context, req domain.CompletionRequest) (*domain.CompletionResponse, error) {
	body, err := json.Marshal(openRouterRequest{
		Model:       req.Model,
		Messages:    req.Messages,
		Temperature: req.Temperature,
	})
	if err != nil {
		return nil, domain.ErrAIUnavailable
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, domain.ErrAIUnavailable
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("HTTP-Referer", "https://github.com/iqbaleff214/kamus-banjar-api-2")
	httpReq.Header.Set("X-Title", "Kamus Banjar API")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, domain.ErrAIUnavailable
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, domain.ErrAIUnavailable
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, domain.ErrAIUnavailable
	}

	var orResp openRouterResponse
	if err := json.Unmarshal(raw, &orResp); err != nil {
		return nil, domain.ErrAIUnavailable
	}
	if len(orResp.Choices) == 0 {
		return nil, domain.ErrAIUnavailable
	}

	return &domain.CompletionResponse{
		Model:   orResp.Model,
		Content: orResp.Choices[0].Message.Content,
	}, nil
}
