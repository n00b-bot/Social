package handler

import (
	"encoding/json"
	"net/http"
	"network/internal/service"
)

type loginInput struct {
	Email string
}

func (h *handler) login(w http.ResponseWriter, r *http.Request) {
	var in loginInput
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	out, err := h.Login(r.Context(), in.Email)
	if err != nil {
		respondError(w, err)
	}
	if err == service.ErrUserNotFound {
		http.Error(w, err.Error(), http.StatusNotFound)
	}
	respond(w, out, 200)
}
