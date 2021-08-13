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
	api.HandleFunc("POST", "/users", h.createUser)
	// r := way.NewRouter()
	// r.Handle("POST", "/api", http.StripPrefix("/api", api))
	return api
}
