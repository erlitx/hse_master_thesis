package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/erlitx/mcp_server/internal/mcp_client/domain"
	"github.com/erlitx/mcp_server/internal/mcp_client/dto"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// HandleConversation manages the conversation flow with Claude
// It handles both new conversations and continuation of existing sessions
func (uc *UseCase) HandleConversation(
	ctx context.Context,
	userMessage string,
	sessionID *string,
	model string,
	maxTokens int,
	systemPrompt string,
	mcpServerURL string,
	mcpServerName string,
) (*domain.Session, error) {
	log.Info().
		Str("message", userMessage).
		Str("session_id", stringOrEmpty(sessionID)).
		Msg("handling conversation request")

	var session *domain.Session
	var err error

	// Step 1: Get or create session
	if sessionID != nil && *sessionID != "" {
		// Continue existing conversation
		session, err = uc.sessionRepo.GetSession(ctx, *sessionID)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve session %s: %w", *sessionID, err)
		}
		log.Debug().Str("session_id", *sessionID).Int("messages_count", len(session.Messages)).Msg("retrieved existing session")
	} else {
		// Create new session
		session = &domain.Session{
			ID:          uuid.New().String(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Messages:    []domain.Message{},
			StopReason:  "",
			Model:       model,
			TotalTokens: 0,
		}
		if err := uc.sessionRepo.CreateSession(ctx, session); err != nil {
			return nil, fmt.Errorf("failed to create session: %w", err)
		}
		log.Debug().Str("session_id", session.ID).Msg("created new session")
	}

	// Step 2: Add user message to session
	userMsg := domain.Message{
		ID:           uuid.New().String(),
		SessionID:    session.ID,
		Role:         "user",
		Content:      userMessage,
		CreatedAt:    time.Now(),
		InputTokens:  0,
		OutputTokens: 0,
	}
	session.Messages = append(session.Messages, userMsg)
	session.UpdatedAt = time.Now()

	log.Debug().Str("message_id", userMsg.ID).Msg("added user message to session")

	// Step 3: Build Claude request with full conversation history
	claudeMessages := dto.FromDomainMessages(session.Messages)
	claudeReq := dto.ClaudeRequest{
		Model:     model,
		MaxTokens: maxTokens,
		System:    systemPrompt,
		Messages:  claudeMessages,
	}

	// Add MCP server configuration if provided
	if mcpServerURL != "" && mcpServerName != "" {
		claudeReq.MCPServers = []dto.MCPServerInfo{
			{
				Type: "url",
				URL:  mcpServerURL,
				Name: mcpServerName,
			},
		}
		claudeReq.Tools = []dto.ClaudeToolInfo{
			{
				Type:          "mcp_toolset",
				MCPServerName: mcpServerName,
			},
		}
	}

	log.Debug().Msgf("MESSAGE TO CLAUDE: %+v", claudeReq)
	return session, nil

	// // Step 4: Send to Claude API
	// claudeResp, err := uc.claudeClient.SendMessage(ctx, claudeReq)
	// if err != nil {
	// 	// Save session even on error (to preserve user message)
	// 	_ = uc.sessionRepo.SaveSession(ctx, session)
	// 	return nil, fmt.Errorf("failed to send message to Claude: %w", err)
	// }

	// log.Debug().
	// 	Str("stop_reason", claudeResp.StopReason).
	// 	Int("input_tokens", claudeResp.Usage.InputTokens).
	// 	Int("output_tokens", claudeResp.Usage.OutputTokens).
	// 	Msg("received response from Claude")

	// // Step 5: Extract assistant response
	// assistantContent := extractResponseText(claudeResp)

	// // Step 6: Add assistant message to session
	// assistantMsg := domain.Message{
	// 	ID:           uuid.New().String(),
	// 	SessionID:    session.ID,
	// 	Role:         "assistant",
	// 	Content:      assistantContent,
	// 	CreatedAt:    time.Now(),
	// 	InputTokens:  claudeResp.Usage.InputTokens,
	// 	OutputTokens: claudeResp.Usage.OutputTokens,
	// }
	// session.Messages = append(session.Messages, assistantMsg)
	// session.UpdatedAt = time.Now()
	// session.StopReason = claudeResp.StopReason
	// session.Model = claudeResp.Model
	// session.TotalTokens += claudeResp.Usage.InputTokens + claudeResp.Usage.OutputTokens

	// log.Debug().Str("message_id", assistantMsg.ID).Msg("added assistant message to session")

	// // Step 7: Save updated session
	// if err := uc.sessionRepo.SaveSession(ctx, session); err != nil {
	// 	return nil, fmt.Errorf("failed to save session: %w", err)
	// }

	// log.Info().
	// 	Str("session_id", session.ID).
	// 	Int("total_messages", len(session.Messages)).
	// 	Int("total_tokens", session.TotalTokens).
	// 	Str("stop_reason", session.StopReason).
	// 	Msg("conversation handled successfully")

	// return session, nil
}

// Helper function to safely get string value or empty string
func stringOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
