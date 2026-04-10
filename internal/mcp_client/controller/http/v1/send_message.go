package v1

import (
	"encoding/json"
	"net/http"

	"github.com/erlitx/mcp_server/internal/dto"
	"github.com/rs/zerolog/log"
)

// SendMessage handles POST /api/v1/send_message
// Accepts a user message, fetches MCP resources, and sends to Claude
func (h *Handlers) SendMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req dto.SendMessageRequest
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

	log.Info().Str("message", req.Message).Msg("received send_message request")

	// Call usecase to process the message
	response, err := h.usecase.SendMessage(
		ctx,
		req.Message,
		h.config.Claude.Model,
		h.config.Claude.MaxTokens,
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to process message")
		respondWithError(w, http.StatusInternalServerError, "failed to process message: "+err.Error())
		return
	}

	// Return successful response
	respondWithJSON(w, http.StatusOK, response)
}

// respondWithJSON writes a JSON response
func respondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error().Err(err).Msg("failed to encode response")
	}
}

// respondWithError writes an error response
func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	respondWithJSON(w, statusCode, map[string]string{
		"error": message,
	})
}
