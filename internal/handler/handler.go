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
	api.HandleFunc("POST", "/login", h.login)
	api.HandleFunc("GET", "/auth_user", h.authUser)
	api.HandleFunc("POST", "/users", h.createUser)
	r := way.NewRouter()
	r.Handle("*", "/api...", http.StripPrefix("/api", h.AuthMiddleware(api)))
	return r
}
