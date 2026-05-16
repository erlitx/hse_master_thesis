package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/erlitx/mcp_server/internal/mcp_client/dto"
	"github.com/rs/zerolog/log"
)

// Отправляет сообщение через gateway MCP
func (uc *UseCase) SendMessage(ctx context.Context, message string, model string, maxTokens int) (*dto.SendMessageResponse, error) {
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

// Формирует системный контекст из ресурсов
func buildSystemContext(resources []dto.MCPResource) string {
	if len(resources) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("You have access to the following data warehouse resources:\n\n")

	for _, resource := range resources {
		sb.WriteString(fmt.Sprintf("## Resource: %s\n", resource.URI))
		if resource.Name != "" {
			sb.WriteString(fmt.Sprintf("Name: %s\n", resource.Name))
		}
		if resource.Description != "" {
			sb.WriteString(fmt.Sprintf("Description: %s\n", resource.Description))
		}
		sb.WriteString(fmt.Sprintf("Content:\n```json\n%s\n```\n\n", resource.Content))
	}

	sb.WriteString("Use this information to answer questions about the data warehouse structure and available models.\n")

	return sb.String()
}

// Извлекает текст из ответа Claude
func extractResponseText(resp *dto.ClaudeResponse) string {
	if len(resp.Content) == 0 {
		return ""
	}

	var sb strings.Builder
	for _, content := range resp.Content {
		if content.Type == "text" {
			sb.WriteString(content.Text)
		}
	}

	return sb.String()
}

// Собирает список метаданных ресурсов
func buildResourceInfoList(resources []dto.MCPResource) []dto.ResourceInfo {
	result := make([]dto.ResourceInfo, len(resources))
	for i, resource := range resources {
		// Try to extract name from JSON if not set
		name := resource.Name
		if name == "" {
			name = extractNameFromURI(resource.URI)
		}

		result[i] = dto.ResourceInfo{
			URI:         resource.URI,
			Name:        name,
			Description: resource.Description,
		}
	}
	return result
}

// Извлекает имя ресурса из URI
func extractNameFromURI(uri string) string {
	parts := strings.Split(uri, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return uri
}

// Парсит содержимое ресурса
func parseResourceContent(content string) string {
	var data interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		// If not valid JSON, return as-is
		return content
	}

	formatted, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return content
	}

	return string(formatted)
}
