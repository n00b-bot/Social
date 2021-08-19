package handler

import (
	"mime"
	"net/http"
	"strconv"

	"github.com/matryer/way"
)

func (h *handler) notifications(w http.ResponseWriter, r *http.Request) {
	if a, _, err := mime.ParseMediaType(r.Header.Get("Accept")); err == nil && a == "text/event-stream" {
		h.subcribeToNotification(w, r)
		return
	}
	last, _ := strconv.Atoi(r.URL.Query().Get("last"))
	before, _ := strconv.Atoi(r.URL.Query().Get("before"))
	nn, err := h.Notifications(r.Context(), last, before)
	if err != nil {
		respondError(w, err)
		return
	}
	respond(w, nn, 200)

}
func (h *handler) markNotificationAsRead(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	nid, _ := strconv.Atoi(way.Param(ctx, "notification_id"))
	err := h.MarkNotificationAsRead(ctx, nid)
	if err != nil {
		respondError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (h *handler) markNotificationsAsRead(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := h.MarkNotificationsAsRead(ctx)
	if err != nil {
		respondError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) subcribeToNotification(w http.ResponseWriter, r *http.Request) {
	f, ok := w.(http.Flusher)
	if !ok {
		respondError(w, errStreaming)
		return
	}
	nn, err := h.SubcribeToNotification(r.Context())
	if err != nil {
		respondError(w, err)
		return
	}
	header := r.Header
	header.Set("Cache-Control", "no-cache")
	header.Set("Connection", "keep-alive")
	header.Set("Context-Type", "text/event-stream")

	for n := range nn {
		writeSSE(w, n)
		f.Flush()

	}
}
