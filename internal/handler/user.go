package handler

import (
	"encoding/json"
	"net/http"
)

type createUserInput struct {
	Email    string
	Username string
}

func (h *handler) createUser(w http.ResponseWriter, r *http.Request) {
	var input createUserInput
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, err)
	}
	err := h.CreateUser(r.Context(), input.Email, input.Username)
	if err != nil {
		respondError(w, err)
	}
	w.WriteHeader(http.StatusNoContent)

}
