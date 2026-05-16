package usecase

import (
	"context"
	"fmt"

	"github.com/erlitx/mcp_server/internal/mcp_client/dto"
	"github.com/rs/zerolog/log"
)

// Отправляет сообщение с контекстом MCP
func (uc *UseCase) SendMessageWithClaudeMCP(ctx context.Context, message string, model string, maxTokens int) (*dto.SendMessageResponse, error) {
	log.Info().Str("message", message).Msg("processing send message request")

	resources, err := uc.MCPServer.ListResources(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch resources from MCP server: %w", err)
	}

	log.Debug().Int("resources_count", len(resources)).Msg("fetched resources from MCP server")

	systemContext := buildSystemContext(resources)

	claudeReq := dto.ClaudeRequest{
		Model:     model,
		MaxTokens: maxTokens,
		System:    systemContext,
		Messages: []dto.ClaudeMessage{
			{
				Role:    "user",
				Content: message,
			},
		},
	}

	claudeResp, err := uc.claudeClient.SendMessage(ctx, claudeReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send message to Claude: %w", err)
	}

	responseText := extractResponseText(claudeResp)

	response := &dto.SendMessageResponse{
		Response:      responseText,
		ResourcesUsed: buildResourceInfoList(resources),
		Model:         claudeResp.Model,
		Usage: &dto.UsageInfo{
			InputTokens:  claudeResp.Usage.InputTokens,
			OutputTokens: claudeResp.Usage.OutputTokens,
		},
	}

	log.Info().
		Int("input_tokens", response.Usage.InputTokens).
		Int("output_tokens", response.Usage.OutputTokens).
		Int("resources_used", len(response.ResourcesUsed)).
		Msg("successfully processed message")

	return response, nil
}
