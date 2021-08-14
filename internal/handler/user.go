package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/matryer/way"
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
func (h *handler) toggleFollow(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	username := way.Param(ctx, "username")
	out, err := h.ToggleFollow(ctx, username)
	if err != nil {
		respondError(w, err)
		return
	}
	respond(w, out, 200)
}

func (h *handler) user(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	username := way.Param(ctx, "username")
	out, err := h.User(ctx, username)
	if err != nil {
		respondError(w, err)
		return
	}
	respond(w, out, 200)

}

func (h *handler) users(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	search := q.Get("search")
	first, _ := strconv.Atoi(q.Get("first"))
	after := q.Get("after")
	uu, err := h.Users(r.Context(), search, first, after)
	if err != nil {
		respondError(w, err)
		return
	}
	respond(w, uu, 200)
}
