package handler

import (
	"net/http"
	"network/internal/service"

	"github.com/matryer/way"
)

type handler struct {
	*service.Service
}

func New(s *service.Service) http.Handler {
	h := handler{Service: s}
	api := way.NewRouter()
	api.HandleFunc("GET", "/users/:username/profile", h.user)
	api.HandleFunc("POST", "/login", h.login)
	api.HandleFunc("GET", "/auth_user", h.authUser)
	api.HandleFunc("PUT", "/auth_user/avatar", h.updateAvatar)
	api.HandleFunc("GET", "/users/all", h.users)
	api.HandleFunc("POST", "/users", h.createUser)
	api.HandleFunc("POST", "/users/:username/followers", h.followers)
	api.HandleFunc("POST", "/users/:username/followees", h.followees)
	api.HandleFunc("POST", "/users/:username/toggle_follow", h.toggleFollow)

	r := way.NewRouter()
	r.Handle("*", "/api...", http.StripPrefix("/api", h.AuthMiddleware(api)))
	return r
}
