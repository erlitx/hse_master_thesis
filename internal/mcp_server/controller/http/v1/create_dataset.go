package v1

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/erlitx/mcp_server/internal/mcp_server/domain"
	"github.com/erlitx/mcp_server/internal/mcp_server/dto"
	"github.com/rs/zerolog/log"
)

// Создаёт датасет в BI-инструменте
func (h *Handlers) CreateDataset(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req dto.CreateDatasetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.DatabaseID <= 0 || strings.TrimSpace(req.Schema) == "" || strings.TrimSpace(req.TableName) == "" {
		http.Error(w, "database_id, schema and table_name are required", http.StatusBadRequest)
		return
	}

	dataset, err := h.usecase.CreateDataset(ctx, domain.CreateDatasetInput{
		DatabaseID: req.DatabaseID,
		Schema:     req.Schema,
		TableName:  req.TableName,
		JWTToken:   req.JWTToken,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to create dataset")
		http.Error(w, "Failed to create dataset", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(dto.CreateDatasetResponse{
		Status: "success",
		ID:     dataset.ID,
	}); err != nil {
		log.Error().Err(err).Msg("failed to encode create dataset response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
