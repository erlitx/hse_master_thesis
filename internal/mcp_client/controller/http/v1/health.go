package v1

import (
	"encoding/json"
	"net/http"
)

// Health handles GET /health
func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"service": "mcp_client",
	})
}
