package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"network/internal/service"
	"strings"
)

type loginInput struct {
	Email string
}

type sendMagicInput struct {
	Email       string
	RedirectURI string
}

func (h *handler) login(w http.ResponseWriter, r *http.Request) {
	var in loginInput
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	out, err := h.Login(r.Context(), in.Email)
	if err == service.ErrUserNotFound {
		http.Error(w, `{ "error": "user not found" }`, http.StatusNotFound)
		return
	}
	if err != nil {
		respondError(w, err)
		return
	}

	respond(w, out, 200)
}

func (h *handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimSpace(r.URL.Query().Get("token"))
		if token == "" {
			token = r.Header.Get("Authorization")
			if !strings.HasPrefix(token, "Bearer ") {
				next.ServeHTTP(w, r)
				return
			}
			token = token[7:]
		}
		if token == "" {
			next.ServeHTTP(w, r)
			return
		}
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

func (h *handler) sendMailLink(w http.ResponseWriter, r *http.Request) {
	var in sendMagicInput
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err := h.SendMagicLink(r.Context(), in.Email, in.RedirectURI)
	if err != nil {
		respondError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) authRedirect(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	uri, err := h.AuthURI(r.Context(), q.Get("verification_code"), q.Get("redirect_uri"))
	if err != nil {
		respondError(w, err)
		return
	}
	http.Redirect(w, r, uri, http.StatusFound)
}
