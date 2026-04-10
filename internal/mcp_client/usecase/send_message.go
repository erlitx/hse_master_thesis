package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/erlitx/mcp_server/internal/dto"
	"github.com/rs/zerolog/log"
)

// SendMessage orchestrates the flow of:
// 1. Fetching resources from MCP server
// 2. Building Claude request with resources as context
// 3. Sending to Claude API
// 4. Returning formatted response
func (uc *UseCase) SendMessage(ctx context.Context, message string, model string, maxTokens int) (*dto.SendMessageResponse, error) {
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

// buildSystemContext creates a system message with all MCP resources as context
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

// extractResponseText extracts the text content from Claude's response
func extractResponseText(resp *dto.ClaudeResponse) string {
	if len(resp.Content) == 0 {
		return ""
	}

	// Combine all text content blocks
	var sb strings.Builder
	for _, content := range resp.Content {
		if content.Type == "text" {
			sb.WriteString(content.Text)
		}
	}

	return sb.String()
}

// buildResourceInfoList creates a list of resource metadata for the response
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

// extractNameFromURI extracts a readable name from the resource URI
func extractNameFromURI(uri string) string {
	// Extract the last part of the URI as the name
	parts := strings.Split(uri, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return uri
}

// parseResourceContent attempts to parse resource content as JSON for better formatting
func parseResourceContent(content string) string {
	var data interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		// If not valid JSON, return as-is
		return content
	}

	// Re-marshal with indentation
	formatted, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return content
	}

	return string(formatted)
}
