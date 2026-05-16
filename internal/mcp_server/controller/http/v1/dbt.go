package v1

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

// Возвращает manifest DBT по HTTP
func (h *Handlers) GetManifest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	manifest, err := h.usecase.GetDBTManifest(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get DBT manifest")
		http.Error(w, "Failed to retrieve DBT manifest", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(manifest); err != nil {
		log.Error().Err(err).Msg("failed to encode manifest response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
