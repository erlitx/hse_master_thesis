package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/erlitx/mcp_server/internal/mcp_client/domain"
	"github.com/google/uuid"
)

// Загружает или создаёт сессию
func (uc *UseCase) getOrCreateSession(ctx context.Context, sessionID *string, model string) (*domain.Session, error) {
	if sessionID != nil && *sessionID != "" {
		session, err := uc.sessionRepo.GetSession(ctx, *sessionID)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve session %s: %w", *sessionID, err)
		}
		return session, nil
	}

	session := &domain.Session{
		ID:          uuid.New().String(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Messages:    []domain.Message{},
		StopReason:  "",
		Model:       model,
		TotalTokens: 0,
	}

	err := uc.sessionRepo.CreateSession(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}
