package domain

import "errors"

var (
	// Session errors
	ErrSessionNotFound    = errors.New("session not found")
	ErrSessionExists      = errors.New("session already exists")
	ErrInvalidSessionID   = errors.New("invalid session id")

	// Message errors
	ErrEmptyMessage       = errors.New("message cannot be empty")
	ErrInvalidMessageRole = errors.New("invalid message role")

	// Claude API errors
	ErrClaudeAPI          = errors.New("claude api error")
	ErrClaudeRateLimited  = errors.New("claude rate limit exceeded")
	ErrClaudeInvalidModel = errors.New("invalid claude model")

	// Repository errors
	ErrRepositoryFailed   = errors.New("repository operation failed")
)
