package v1

import "net/http"

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("ok"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
