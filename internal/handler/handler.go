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
	api.HandleFunc("POST", "/posts", h.createPost)
	api.HandleFunc("POST", "/posts/:post_id/toggle_like", h.toggleLike)
	api.HandleFunc("POST", "/posts/:post_id/toggle_subcription", h.togglePostSubcription)
	api.HandleFunc("POST", "/users/:username/posts", h.posts)
	api.HandleFunc("GET", "/timeline", h.timeLine)
	api.HandleFunc("POST", "/posts/:post_id/comments", h.createComment)
	api.HandleFunc("GET", "/posts/:post_id/comments", h.comments)
	api.HandleFunc("POST", "/comments/:comment_id/toggle_like", h.toggleLikeComment)
	api.HandleFunc("GET", "/posts/:post_id/", h.post)
	api.HandleFunc("GET", "/notifications", h.notifications)
	api.HandleFunc("GET", "/notifications/:notification_id/mark_as_read", h.markNotificationAsRead)
	api.HandleFunc("GET", "/mark_notification_as_read", h.markNotificationsAsRead)

	r := way.NewRouter()
	r.Handle("*", "/api...", http.StripPrefix("/api", h.AuthMiddleware(api)))
	return r
}
