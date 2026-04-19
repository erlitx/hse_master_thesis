package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/erlitx/mcp_server/internal/mcp_client/domain"
)

// SessionRepository implements domain.SessionRepository with in-memory storage
type SessionRepository struct {
	mu       sync.RWMutex
	sessions map[string]*domain.Session
}

// New creates a new in-memory session repository
func New() *SessionRepository {
	return &SessionRepository{
		sessions: make(map[string]*domain.Session),
	}
}

// CreateSession creates a new session in memory
func (r *SessionRepository) CreateSession(ctx context.Context, session *domain.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if existing, exists := r.sessions[session.ID]; exists {
		// Idempotent create: copy stored canonical session back to caller
		// so caller does not continue with a partial payload.
		sessionCopy := *existing
		sessionCopy.Messages = make([]domain.Message, len(existing.Messages))
		copy(sessionCopy.Messages, existing.Messages)
		sessionCopy.Tools = make([]domain.ToolDefinition, len(existing.Tools))
		copy(sessionCopy.Tools, existing.Tools)
		*session = sessionCopy
		return nil
	}

	// Create a copy to avoid external mutations
	sessionCopy := *session
	sessionCopy.Messages = make([]domain.Message, len(session.Messages))
	copy(sessionCopy.Messages, session.Messages)
	sessionCopy.Tools = make([]domain.ToolDefinition, len(session.Tools))
	copy(sessionCopy.Tools, session.Tools)

	r.sessions[session.ID] = &sessionCopy
	return nil
}

// GetSession retrieves a session by ID
func (r *SessionRepository) GetSession(ctx context.Context, sessionID string) (*domain.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, exists := r.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session with ID %s not found", sessionID)
	}

	// Return a copy to avoid external mutations
	sessionCopy := *session
	sessionCopy.Messages = make([]domain.Message, len(session.Messages))
	copy(sessionCopy.Messages, session.Messages)
	sessionCopy.Tools = make([]domain.ToolDefinition, len(session.Tools))
	copy(sessionCopy.Tools, session.Tools)

	return &sessionCopy, nil
}

// SaveSession updates an existing session
func (r *SessionRepository) SaveSession(ctx context.Context, session *domain.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sessions[session.ID]; !exists {
		return fmt.Errorf("session with ID %s not found", session.ID)
	}

	// Create a copy to avoid external mutations
	sessionCopy := *session
	sessionCopy.Messages = make([]domain.Message, len(session.Messages))
	copy(sessionCopy.Messages, session.Messages)
	sessionCopy.Tools = make([]domain.ToolDefinition, len(session.Tools))
	copy(sessionCopy.Tools, session.Tools)

	r.sessions[session.ID] = &sessionCopy
	return nil
}
