package v1

import (
	"encoding/json"
	"net/http"

	"github.com/erlitx/mcp_server/internal/mcp_client/dto"
	"github.com/rs/zerolog/log"
)

// SendWithClaudeMCP handles POST /api/v1/claude
// Manages conversation with Claude, supporting both new and existing sessions
func (h *Handlers) SendWithClaudeMCP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req dto.ConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Msg("failed to decode request body")
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate request
	if req.Message == "" {
		respondWithError(w, http.StatusBadRequest, "message is required")
		return
	}

	// Set defaults if not provided
	model := req.Model
	if model == "" {
		model = h.config.Claude.Model
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = h.config.Claude.MaxTokens
	}

	systemPrompt := req.System
	if systemPrompt == "" {
		systemPrompt = "You are a data analyst that generates SQL queries for analytical dashboards. Your task is to return ONLY a valid SQL query that will be executed against a ClickHouse database via MCP. STRICT RULES:- Output ONLY SQL. No explanations, no markdown, no text.- Generate a single SELECT query.- Prefer SIMPLE queries over complex ones. - Return ONLY the minimum required columns to answer the question.- Do NOT include extra metrics unless explicitly requested. - Group ONLY by necessary dimensions. - Use clear and concise aliases. - The query must be production-ready and efficient. Do not explain your reasoning. Do not describe the query. Only output SQL. If the user request is ambiguous and requires clarification, ask a question instead of generating a query"
	}

	log.Info().
		Str("message", req.Message).
		Str("session_id", stringPtrToString(req.SessionID)).
		Msg("received conversation request")

	// Call usecase to handle conversation
	session, err := h.usecase.HandleConversation(
		ctx,
		req.Message,
		req.SessionID,
		model,
		maxTokens,
		systemPrompt,
		h.config.MCPServerConnection.Addr,
		"analytics", // MCP server name
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to handle conversation")
		respondWithError(w, http.StatusInternalServerError, "failed to process conversation: "+err.Error())
		return
	}

	// Convert domain session to DTO response
	response := dto.FromSession(session)

	// Return successful response
	respondWithJSON(w, http.StatusOK, response)
}

// Helper function to convert string pointer to string
func stringPtrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

