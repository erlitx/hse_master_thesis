package v1

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/erlitx/mcp_server/internal/mcp_client/dto"
	"github.com/erlitx/mcp_server/pkg/render"
	"github.com/rs/zerolog/log"
)

// SendWithGatewayMCP handles POST /api/v1/chat/gateway
// Manually orchestrates Claude <-> MCP tool calls with session persistence.
func (h *Handlers) SendWithGatewayMCP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req dto.GatewayConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Error(w, err, http.StatusBadRequest, "failed to decode request body")
		return
	}

	if req.Message == "" {
		render.Error(w, fmt.Errorf("message field is empty"), http.StatusBadRequest, "message is required")
		return
	}

	model := req.Model
	if model == "" {
		model = h.config.Claude.Model
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = h.config.Claude.MaxTokens
	}

	system := req.System
	if system == "" {
		system = "You are a data analyst that generates SQL queries for analytical dashboards. Your task is to return ONLY a valid SQL query that will be executed against a ClickHouse database. STRICT RULES:- Output ONLY SQL. No explanations, no markdown, no text.- Generate a single SELECT query.- Prefer SIMPLE queries over complex ones. - Return ONLY the minimum required columns to answer the question.- Do NOT include extra metrics unless explicitly requested. - Group ONLY by necessary dimensions. - Use clear and concise aliases. - The query must be production-ready and efficient. Do not explain your reasoning. Do not describe the query. Only output SQL. If the user request is ambiguous and requires clarification, ask a question instead of generating a query"
	}

	input := dto.GatewayConversationInput{
		Message:   req.Message,
		SessionID: req.SessionID,
		Model:     model,
		MaxTokens: maxTokens,
		System:    system,
	}

	log.Info().
		Str("message", input.Message).
		Str("session_id", stringPtrToString(input.SessionID)).
		Msg("received gateway conversation request")

	session, err := h.usecase.HandleGatewayConversation(ctx, input)
	if err != nil {
		render.ErrorSpecific(w, err, mcpClientErrorMappings)
		return
	}

	render.JSON(w, dto.FromSession(session), http.StatusOK)
}
