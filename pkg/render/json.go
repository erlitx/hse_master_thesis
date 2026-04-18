package render

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

func JSON(w http.ResponseWriter, body any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(body); err != nil {
		log.Error().Err(err).Msg("json encode error")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(statusCode)

	if _, err := w.Write(buf.Bytes()); err != nil {
		log.Error().Err(err).Msg("write response error")
	}
}
