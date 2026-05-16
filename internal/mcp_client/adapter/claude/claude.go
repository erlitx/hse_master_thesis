package claude

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/erlitx/mcp_server/internal/mcp_client/dto"
	"github.com/rs/zerolog/log"
)

const (
	claudeAPIURL     = "https://api.anthropic.com/v1/messages"
	anthropicVersion = "2023-06-01"
	anthropicBeta    = "mcp-client-2025-11-20"
)

// HTTP-клиент
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// Создаёт новый экземпляр
func New(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

// Отправляет сообщение через gateway MCP
func (c *Client) SendMessage(ctx context.Context, req dto.ClaudeRequest) (*dto.ClaudeResponse, error) {
	return c.send(ctx, req)
}

// Отправляет запрос в формате gateway
func (c *Client) SendGatewayMessage(ctx context.Context, req dto.ManualGatewayRequest) (*dto.ClaudeResponse, error) {
	return c.send(ctx, req)
}

// Выполняет HTTP-запрос к Claude API
func (c *Client) send(ctx context.Context, req interface{}) (*dto.ClaudeResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, claudeAPIURL, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", anthropicVersion)
	httpReq.Header.Set("anthropic-beta", anthropicBeta)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Error().
			Int("status_code", resp.StatusCode).
			Str("response_body", string(body)).
			Msg("Claude API returned error")
		return nil, fmt.Errorf("Claude API error (status %d): %s", resp.StatusCode, string(body))
	}

	var claudeResp dto.ClaudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	log.Debug().
		Str("id", claudeResp.ID).
		Int("input_tokens", claudeResp.Usage.InputTokens).
		Int("output_tokens", claudeResp.Usage.OutputTokens).
		Msgf("successfully received response from Claude: %+v", claudeResp)

	return &claudeResp, nil
}
