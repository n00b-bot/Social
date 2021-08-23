package handler

import (
	"mime"
	"net/http"
	"network/internal/service"
	"time"

	"github.com/matryer/way"
)

type handler struct {
	*service.Service
	ping time.Duration
}

func New(s *service.Service, time time.Duration) http.Handler {
	h := handler{Service: s, ping: time}
	api := way.NewRouter()
	api.HandleFunc("POST", "/send_magic_link", h.sendMailLink)
	api.HandleFunc("GET", "/auth_redirect", h.authRedirect)
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
	api.HandleFunc("GET", "/has_unread_notifications", h.unreadNotifications)

	fs := http.FileServer(&spaFileSystem{http.Dir("web/static")})

	mime.AddExtensionType(".js", "application/javascript; charset=utf-8")

	r := way.NewRouter()
	r.Handle("*", "/api...", http.StripPrefix("/api", h.AuthMiddleware(api)))
	r.Handle("GET", "/...", NoCache(fs))
	return r
}
