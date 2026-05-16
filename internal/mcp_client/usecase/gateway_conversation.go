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

// Оркестрирует диалог Claude и MCP-инструментов
func (uc *UseCase) HandleGatewayConversation(ctx context.Context, input dto.GatewayConversationInput) (*domain.Session, error) {
	session, err := uc.getOrCreateSession(ctx, input.SessionID, input.Model)
	if err != nil {
		return nil, err
	}
	log.Debug().Str("session_id", session.ID).Msgf("gateway session start snapshot:\n%s", utils.PrettyJSON(session))

	userMessage := newUserMessageWithContent(session.ID, []domain.Content{{Type: "text", Text: input.Message}})
	uc.appendSessionMessage(session, userMessage)

	log.Debug().
		Str("session_id", session.ID).
		Int("messages_count", len(session.Messages)).
		Msgf("session after user message append:\n%s", utils.PrettyJSON(session))

	toolsDTO, err := uc.MCPServer.ListTools(ctx)
	if err != nil {
		_ = uc.sessionRepo.SaveSession(ctx, session)
		return nil, fmt.Errorf("failed to list MCP tools: %w", err)
	}
	session.Tools = dto.DomainToolsFromMCPTools(toolsDTO)
	session.GatewayModel = input.Model
	session.GatewayMaxTokens = input.MaxTokens
	session.GatewaySystem = input.System

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

// Создаёт пользовательское сообщение с контентом
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

// Добавляет сообщение в сессию
func (uc *UseCase) appendSessionMessage(session *domain.Session, msg domain.Message) {
	if msg.SessionID == "" {
		msg.SessionID = session.ID
	}
	session.Messages = append(session.Messages, msg)
	session.UpdatedAt = time.Now()
}

// Выполняет tool_use и добавляет tool_result
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

		callResp, callErr := uc.MCPServer.CallTool(ctx, block.Name, input)
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

		// Construct the tool_result block
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

// Возвращает указатель на bool
func boolPtr(v bool) *bool {
	return &v
}
