package domain

import "errors"

var (
	// Ошибки сессии
	ErrSessionNotFound    = errors.New("session not found")
	ErrSessionExists      = errors.New("session already exists")
	ErrInvalidSessionID   = errors.New("invalid session id")

	// Ошибки сообщений
	ErrEmptyMessage       = errors.New("message cannot be empty")
	ErrInvalidMessageRole = errors.New("invalid message role")

	// Ошибки Claude API
	ErrClaudeAPI          = errors.New("claude api error")
	ErrClaudeRateLimited  = errors.New("claude rate limit exceeded")
	ErrClaudeInvalidModel = errors.New("invalid claude model")

	// Ошибки хранилища
	ErrRepositoryFailed   = errors.New("repository operation failed")
)
