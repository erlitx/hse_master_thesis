package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/erlitx/mcp_server/internal/mcp_client/domain"
	"github.com/erlitx/mcp_server/internal/mcp_client/dto"
	"github.com/erlitx/mcp_server/utils"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const maxGatewayToolIterations = 10

// HandleGatewayConversation manually orchestrates Claude <-> MCP tool calls.
//
// Flow:
//  1. Load/create session and append user message.
//  2. Fetch MCP tool schema and send Claude request with `tools`.
//  3. If Claude returns tool_use blocks, execute each tool via MCP tools/call.
//  4. Append tool_result as a user message and continue the loop.
//  5. Stop when no tool_use is returned (final assistant response) or max iterations reached.
//
// This keeps full history in one session so subsequent requests can continue context.
func (uc *UseCase) HandleGatewayConversation(
	ctx context.Context,
	userMessage string,
	sessionID *string,
	model string,
	maxTokens int,
	systemPrompt string,
) (*domain.Session, error) {
	// Step 1: Resolve session (continue existing or create a new one).
	session, err := uc.getOrCreateSession(ctx, sessionID, model)
	if err != nil {
		return nil, err
	}
	log.Debug().Str("session_id", session.ID).Msgf("gateway session start snapshot:\n%s", utils.PrettyJSON(session))

	// Step 2: Append incoming user text.
	session.Messages = append(session.Messages, newMessage(session.ID, "user", []domain.Content{
		{Type: "text", Text: userMessage},
	}))
	session.UpdatedAt = time.Now()
	log.Debug().
		Str("session_id", session.ID).
		Int("messages_count", len(session.Messages)).
		Msgf("session after user message append:\n%s", utils.PrettyJSON(session))

	// Step 3: Load MCP tool definitions and convert to Claude tool schema.
	tools, err := uc.mcpClient.ListTools(ctx)
	if err != nil {
		_ = uc.sessionRepo.SaveSession(ctx, session)
		return nil, fmt.Errorf("failed to list MCP tools: %w", err)
	}

	claudeTools := make([]dto.ClaudeToolDefinition, 0, len(tools))
	for _, t := range tools {
		claudeTools = append(claudeTools, dto.ClaudeToolDefinition{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.InputSchema,
		})
	}

	// Step 4: Run Claude/tool loop until Claude no longer requests tool_use blocks.
	for i := 0; i < maxGatewayToolIterations; i++ {
		log.Debug().
			Str("session_id", session.ID).
			Int("iteration", i+1).
			Int("messages_count", len(session.Messages)).
			Msg("gateway loop iteration start")

		claudeReq := dto.ManualGatewayRequest{
			Model:     model,
			MaxTokens: maxTokens,
			System:    systemPrompt,
			Messages:  dto.FromDomainMessages(session.Messages),
			Tools:     claudeTools,
		}

		log.Info().
			Str("session_id", session.ID).
			Int("iteration", i+1).
			Msg("Sending message to Claude")
			
		log.Debug().
			Str("session_id", session.ID).
			Int("iteration", i+1).
			Msgf("claude request payload:\n%s", utils.PrettyJSON(claudeReq))

		claudeResp, err := uc.claudeClient.SendGatewayMessage(ctx, claudeReq)
		if err != nil {
			_ = uc.sessionRepo.SaveSession(ctx, session)
			return nil, fmt.Errorf("failed to send message to Claude: %w", err)
		}
		log.Debug().
			Str("session_id", session.ID).
			Int("iteration", i+1).
			Str("stop_reason", claudeResp.StopReason).
			Msgf("claude response payload:\n%s", utils.PrettyJSON(claudeResp))

		// Step 4.1: Persist assistant turn in in-memory session state.
		session.TotalTokens += claudeResp.Usage.InputTokens + claudeResp.Usage.OutputTokens
		session.StopReason = claudeResp.StopReason
		session.Model = claudeResp.Model

		assistantContent := dto.ToDomainContent(claudeResp.Content)
		session.Messages = append(session.Messages, domain.Message{
			ID:           uuid.New().String(),
			SessionID:    session.ID,
			Role:         "assistant",
			Content:      assistantContent,
			CreatedAt:    time.Now(),
			InputTokens:  claudeResp.Usage.InputTokens,
			OutputTokens: claudeResp.Usage.OutputTokens,
		})
		session.UpdatedAt = time.Now()
		log.Debug().
			Str("session_id", session.ID).
			Int("iteration", i+1).
			Int("messages_count", len(session.Messages)).
			Msgf("session after assistant append:\n%s", utils.PrettyJSON(session))

		// Step 4.2: Execute requested tools and prepare tool_result blocks.
		toolResults, hasToolUse := uc.handleToolUses(ctx, claudeResp.Content)
		log.Debug().
			Str("session_id", session.ID).
			Int("iteration", i+1).
			Bool("has_tool_use", hasToolUse).
			Int("tool_results_count", len(toolResults)).
			Msgf("tool results payload:\n%s", utils.PrettyJSON(toolResults))
		if !hasToolUse {
			// Final assistant response reached.
			log.Debug().
				Str("session_id", session.ID).
				Int("iteration", i+1).
				Msg("gateway loop end: no tool_use blocks")
			break
		}

		// Step 4.3: Feed tool results back to Claude as the next user turn.
		session.Messages = append(session.Messages, newMessage(session.ID, "user", toolResults))
		session.UpdatedAt = time.Now()
		log.Debug().
			Str("session_id", session.ID).
			Int("iteration", i+1).
			Int("messages_count", len(session.Messages)).
			Msgf("session after tool_result append:\n%s", utils.PrettyJSON(session))
	}

	// Step 5: Save finalized session snapshot.
	if err := uc.sessionRepo.SaveSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}
	log.Debug().Str("session_id", session.ID).Msgf("gateway session final snapshot:\n%s", utils.PrettyJSON(session))

	return session, nil
}

func newMessage(sessionID string, role string, content []domain.Content) domain.Message {
	return domain.Message{
		ID:           uuid.New().String(),
		SessionID:    sessionID,
		Role:         role,
		Content:      content,
		CreatedAt:    time.Now(),
		InputTokens:  0,
		OutputTokens: 0,
	}
}

func (uc *UseCase) handleToolUses(ctx context.Context, content []dto.ClaudeContent) ([]domain.Content, bool) {
	// toolResults is returned back to Claude as the next user-role message content.
	toolResults := make([]domain.Content, 0)
	// hasToolUse tells the caller whether we should continue the Claude loop.
	hasToolUse := false

	for _, block := range content {
		// Ignore non-tool blocks (e.g. plain assistant text).
		if block.Type != "tool_use" {
			continue
		}

		hasToolUse = true
		// Execute the requested MCP tool with Claude-provided input payload.
		callResp, callErr := uc.mcpClient.CallTool(ctx, block.Name, block.Input)
		log.Debug().
			Str("tool_name", block.Name).
			Str("tool_use_id", block.ID).
			Msgf("tool_use input payload:\n%s", utils.PrettyJSON(block.Input))

		var isErr bool
		var toolResultText string
		if callErr != nil {
			// Tool execution error is forwarded back to Claude as tool_result(is_error=true).
			isErr = true
			toolResultText = callErr.Error()
			log.Error().Err(callErr).Str("tool", block.Name).Msg("tool call failed")
		} else {
			// Claude expects tool_result content to be textual; encode structured output as JSON string.
			b, marshalErr := json.Marshal(callResp)
			if marshalErr != nil {
				isErr = true
				toolResultText = marshalErr.Error()
			} else {
				toolResultText = string(b)
			}
			log.Debug().
				Str("tool_name", block.Name).
				Str("tool_use_id", block.ID).
				Msgf("raw tool response payload:\n%s", utils.PrettyJSON(callResp))
		}

		// Build one tool_result block linked to the original tool_use id.
		toolResults = append(toolResults, domain.Content{
			Type:      "tool_result",
			ToolUseID: block.ID,
			IsError:   boolPtr(isErr),
			Content: []domain.Content{
				{
					Type: "text",
					Text: toolResultText,
				},
			},
		})
		log.Debug().
			Str("tool_name", block.Name).
			Str("tool_use_id", block.ID).
			Bool("is_error", isErr).
			Msgf("constructed tool_result block:\n%s", utils.PrettyJSON(toolResults[len(toolResults)-1]))
	}

	return toolResults, hasToolUse
}

func boolPtr(v bool) *bool {
	return &v
}
