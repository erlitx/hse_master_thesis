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
func (uc *UseCase) HandleGatewayConversation(ctx context.Context, input dto.GatewayConversationInput) (*domain.Session, error) {
	// Step 1: Resolve session (continue existing or create a new one).
	session, err := uc.getOrCreateSession(ctx, input.SessionID, input.Model)
	if err != nil {
		return nil, err
	}
	log.Debug().Str("session_id", session.ID).Msgf("gateway session start snapshot:\n%s", utils.PrettyJSON(session))

	// Step 2: Append incoming user text.
	userMessage := newUserMessageWithContent(session.ID, []domain.Content{{Type: "text", Text: input.Message}})
	uc.appendSessionMessage(session, userMessage)

	log.Debug().
		Str("session_id", session.ID).
		Int("messages_count", len(session.Messages)).
		Msgf("session after user message append:\n%s", utils.PrettyJSON(session))

	// Step 3: Load MCP tool definitions into the session domain model (DTO → domain).
	toolsDTO, err := uc.mcpClient.ListTools(ctx)
	if err != nil {
		_ = uc.sessionRepo.SaveSession(ctx, session)
		return nil, fmt.Errorf("failed to list MCP tools: %w", err)
	}
	session.Tools = dto.DomainToolsFromMCPTools(toolsDTO)
	session.GatewayModel = input.Model
	session.GatewayMaxTokens = input.MaxTokens
	session.GatewaySystem = input.System

	// Step 4: Run Claude/tool loop until Claude no longer requests tool_use blocks.
	for i := 0; i < maxGatewayToolIterations; i++ {
		log.Debug().
			Str("session_id", session.ID).
			Int("iteration", i+1).
			Int("messages_count", len(session.Messages)).
			Msg("Gateway loop iteration start")

		claudeReq := dto.ManualGatewayRequestFromSession(session)

		log.Info().
			Str("session_id", session.ID).
			Int("iteration", i+1).
			Msg("Sending message to Claude")

		claudeResp, err := uc.claudeClient.SendGatewayMessage(ctx, claudeReq)
		if err != nil {
			_ = uc.sessionRepo.SaveSession(ctx, session)
			return nil, fmt.Errorf("failed to send message to Claude: %w", err)
		}

		// Step 4.1: Append assistant turn, then refresh session-level metadata from the response.
		assistantContent := dto.ToDomainContent(claudeResp.Content)
		uc.appendSessionMessage(session, domain.Message{
			ID:           uuid.New().String(),
			SessionID:    session.ID,
			Role:         "assistant",
			Content:      assistantContent,
			CreatedAt:    time.Now(),
			InputTokens:  claudeResp.Usage.InputTokens,
			OutputTokens: claudeResp.Usage.OutputTokens,
		})
		session.TotalTokens += claudeResp.Usage.InputTokens + claudeResp.Usage.OutputTokens
		session.StopReason = claudeResp.StopReason
		session.Model = claudeResp.Model

		log.Debug().
			Str("session_id", session.ID).
			Int("iteration", i+1).
			Int("messages_count", len(session.Messages)).
			Msgf("session after assistant append:\n%s", utils.PrettyJSON(session))

		// Step 4.2: From the latest assistant message only, run tool_use blocks and append tool_result user turn.
		session, hasToolUse := uc.handleToolUses(ctx, session)
		log.Debug().
			Str("session_id", session.ID).
			Int("iteration", i+1).
			Bool("has_tool_use", hasToolUse).
			Int("messages_count", len(session.Messages)).
			Msgf("session after handleToolUses:\n%s", utils.PrettyJSON(session))
		if !hasToolUse {
			// Final assistant response reached.
			log.Debug().
				Str("session_id", session.ID).
				Int("iteration", i+1).
				Msg("gateway loop end: no tool_use blocks")
			break
		}
	}

	// Step 5: Save finalized session snapshot.
	if err := uc.sessionRepo.SaveSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}
	log.Info().
		Str("session_id", session.ID).
		Int("total_tokens", session.TotalTokens).
		Str("stop_reason", session.StopReason).
		Str("model", session.Model).
		Msgf("gateway session final snapshot:\n%s", utils.PrettyJSON(session.Messages[len(session.Messages)-1].Content))

	return session, nil
}

// newUserMessageWithContent builds a user-role message (e.g. plain text or tool_result blocks).
func newUserMessageWithContent(sessionID string, content []domain.Content) domain.Message {
	return domain.Message{
		ID:           uuid.New().String(),
		SessionID:    sessionID,
		Role:         "user",
		Content:      content,
		CreatedAt:    time.Now(),
		InputTokens:  0,
		OutputTokens: 0,
	}
}

// appendSessionMessage appends a transcript message and refreshes the session's UpdatedAt.
func (uc *UseCase) appendSessionMessage(session *domain.Session, msg domain.Message) {
	if msg.SessionID == "" {
		msg.SessionID = session.ID
	}
	session.Messages = append(session.Messages, msg)
	session.UpdatedAt = time.Now()
}

// handleToolUses inspects only the last session message (the current assistant turn) for tool_use blocks,
// executes each via MCP, appends a single user message with tool_result blocks, and returns the updated session.
// Earlier messages are not scanned because their tool_use blocks were already handled on prior iterations.
func (uc *UseCase) handleToolUses(ctx context.Context, session *domain.Session) (*domain.Session, bool) {
	if session == nil || len(session.Messages) == 0 {
		return session, false
	}

	last := session.Messages[len(session.Messages)-1]
	toolResults := make([]domain.Content, 0)
	hasToolUse := false

	for _, block := range last.Content {
		if block.Type != "tool_use" {
			continue
		}

		hasToolUse = true
		input := block.Input
		if input == nil {
			input = map[string]interface{}{}
		}

		callResp, callErr := uc.mcpClient.CallTool(ctx, block.Name, input)
		log.Debug().
			Str("tool_name", block.Name).
			Str("tool_use_id", block.ID).
			Msgf("tool_use input payload:\n%s", utils.PrettyJSON(input))

		var isErr bool
		var toolResultText string
		if callErr != nil {
			isErr = true
			toolResultText = callErr.Error()
			log.Error().Err(callErr).Str("tool", block.Name).Msg("tool call failed")
		} else {
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

	if !hasToolUse {
		return session, false
	}

	uc.appendSessionMessage(session, newUserMessageWithContent(session.ID, toolResults))
	return session, true
}

func boolPtr(v bool) *bool {
	return &v
}
