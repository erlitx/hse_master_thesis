package mcpserver

import (
	"encoding/json"
	"io"
	"net/http"
)

// HTTPHandler returns an http.Handler that serves MCP JSON-RPC over HTTP.
//
// The endpoint expects POST with a JSON-RPC message body.
// It writes the JSON-RPC response (if any) as application/json.
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
		defer r.Body.Close()

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
		_, _ = w.Write(b)
	})
}
