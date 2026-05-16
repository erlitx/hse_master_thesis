package mcpserver

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

// Возвращает HTTP-обработчик MCP
func (h *Handler) HTTPHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		defer func() {
			if err := r.Body.Close(); err != nil {
				// log it (don't panic in HTTP handler)
				log.Printf("failed to close request body: %v", err)
			}
		}()

		// Parse the JSON-RPC request, execute it, and get the response
		resp := h.srv.HandleMessage(r.Context(), body)
		if resp == nil {
			// JSON-RPC notifications have no response
			w.WriteHeader(http.StatusNoContent)
			return
		}

		b, err := json.Marshal(resp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(b)
		if err != nil {
			log.Info().Err(err).Msg("failed to write response")
		}
	})
}
