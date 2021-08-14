package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"network/internal/service"
	"strings"
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
		return
	}
	if err == service.ErrUserNotFound {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	respond(w, out, 200)
}

func (h *handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if !strings.HasPrefix(token, "Bearer ") {
			next.ServeHTTP(w, r)
			return
		}
		fmt.Println(token)
		token = token[7:]
		id, err := h.TokenDecode(token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, service.KeyAuthUserID, id)
		next.ServeHTTP(w, r.WithContext(ctx))

	})
}

func (h *handler) authUser(w http.ResponseWriter, r *http.Request) {
	u, err := h.AuthUser(r.Context())
	if err == service.ErrUnauthorized {
		respondError(w, err)
		return
	}
	if err != nil {
		respondError(w, err)
		return
	}
	respond(w, u, 200)
}
