package postgres

import (
	"context"

	"github.com/erlitx/mcp_server/internal/mcp_client/domain"
)

// SessionRepository implements domain.SessionRepository with PostgreSQL storage
type SessionRepository struct {
	// TODO: Add database connection pool
	// db *sql.DB or *pgxpool.Pool
}

// New creates a new PostgreSQL session repository
func New( /* db *sql.DB */ ) *SessionRepository {
	return &SessionRepository{
		// db: db,
	}
}

// CreateSession creates a new session in PostgreSQL
func (r *SessionRepository) CreateSession(ctx context.Context, session *domain.Session) error {
	// TODO: Implement PostgreSQL insert
	// INSERT INTO sessions (id, created_at, updated_at, stop_reason, model, total_tokens)
	// VALUES ($1, $2, $3, $4, $5, $6)
	//
	// INSERT INTO messages (id, session_id, role, content, created_at, input_tokens, output_tokens)
	// VALUES ($1, $2, $3, $4, $5, $6, $7)
	panic("not implemented")
}

// GetSession retrieves a session by ID from PostgreSQL
func (r *SessionRepository) GetSession(ctx context.Context, sessionID string) (*domain.Session, error) {
	// TODO: Implement PostgreSQL select
	// SELECT * FROM sessions WHERE id = $1
	// SELECT * FROM messages WHERE session_id = $1 ORDER BY created_at
	panic("not implemented")
}

// SaveSession updates an existing session in PostgreSQL
func (r *SessionRepository) SaveSession(ctx context.Context, session *domain.Session) error {
	// TODO: Implement PostgreSQL update
	// UPDATE sessions SET updated_at = $1, stop_reason = $2, model = $3, total_tokens = $4 WHERE id = $5
	// INSERT new messages into messages table
	panic("not implemented")
}
