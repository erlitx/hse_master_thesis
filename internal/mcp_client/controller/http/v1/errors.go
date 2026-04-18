package v1

import (
	"net/http"

	"github.com/erlitx/mcp_server/internal/mcp_client/domain"
	"github.com/erlitx/mcp_server/pkg/render"
)

// mcpClientErrorMappings maps domain errors to HTTP status codes and messages
var mcpClientErrorMappings = map[error]render.ErrorMapping{
	// Session errors
	domain.ErrSessionNotFound: {
		Status:  http.StatusNotFound,
		Message: "session not found",
	},
	domain.ErrSessionExists: {
		Status:  http.StatusConflict,
		Message: "session already exists",
	},
	domain.ErrInvalidSessionID: {
		Status:  http.StatusBadRequest,
		Message: "invalid session id format",
	},

	// Message errors
	domain.ErrEmptyMessage: {
		Status:  http.StatusBadRequest,
		Message: "message cannot be empty",
	},
	domain.ErrInvalidMessageRole: {
		Status:  http.StatusBadRequest,
		Message: "invalid message role - must be 'user' or 'assistant'",
	},

	// Claude API errors
	domain.ErrClaudeAPI: {
		Status:  http.StatusBadGateway,
		Message: "claude api request failed",
	},
	domain.ErrClaudeRateLimited: {
		Status:  http.StatusTooManyRequests,
		Message: "claude rate limit exceeded - please try again later",
	},
	domain.ErrClaudeInvalidModel: {
		Status:  http.StatusBadRequest,
		Message: "invalid claude model specified",
	},

	// Repository errors
	domain.ErrRepositoryFailed: {
		Status:  http.StatusInternalServerError,
		Message: "session storage operation failed",
	},
}
