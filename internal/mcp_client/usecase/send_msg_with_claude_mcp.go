package usecase

import (
	"context"
	"fmt"

	"github.com/erlitx/mcp_server/internal/mcp_client/dto"
	"github.com/rs/zerolog/log"
)

// SendMessage orchestrates the flow of:
// 1. Fetching resources from MCP server
// 2. Building Claude request with resources as context
// 3. Sending to Claude API
// 4. Returning formatted response
func (uc *UseCase) SendMessageWithClaudeMCP(ctx context.Context, message string, model string, maxTokens int) (*dto.SendMessageResponse, error) {
	log.Info().Str("message", message).Msg("processing send message request")

	// Step 1: Fetch all resources from MCP server
	resources, err := uc.mcpClient.ListResources(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch resources from MCP server: %w", err)
	}

	log.Debug().Int("resources_count", len(resources)).Msg("fetched resources from MCP server")

	// Step 2: Build system context from resources
	systemContext := buildSystemContext(resources)

	// Step 3: Build Claude request
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

	// Step 4: Send to Claude API
	claudeResp, err := uc.claudeClient.SendMessage(ctx, claudeReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send message to Claude: %w", err)
	}

	// Step 5: Extract response text
	responseText := extractResponseText(claudeResp)

	// Step 6: Build response with resource metadata
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
