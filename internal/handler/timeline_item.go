package handler

import (
	"errors"
	"mime"
	"net/http"
	"strconv"
)

var errStreaming = errors.New("streaming unsupported")

func (h *handler) timeLine(w http.ResponseWriter, r *http.Request) {
	if a, _, err := mime.ParseMediaType(r.Header.Get("Accept")); err == nil && a == "text/event-stream" {
		h.subcribeToTimeline(w, r)
		return
	}
	ctx := r.Context()
	last, _ := strconv.Atoi(r.URL.Query().Get("last"))
	before, _ := strconv.Atoi(r.URL.Query().Get("before"))
	tt, err := h.Timeline(ctx, last, before)
	if err != nil {
		respondError(w, err)
		return
	}
	respond(w, tt, 200)
}

func (h *handler) subcribeToTimeline(w http.ResponseWriter, r *http.Request) {
	f, ok := w.(http.Flusher)
	if !ok {
		respondError(w, errStreaming)
		return
	}
	tt, err := h.SubcribeToTimeline(r.Context())
	if err != nil {
		respondError(w, err)
		return
	}
	//ctx := r.Context()
	header := w.Header()
	header.Set("Cache-Control", "no-cache")
	header.Set("Connection", "keep-alive")
	header.Set("Content-Type", "text/event-stream")

	for ti := range tt {
		writeSSE(w, ti)
		f.Flush()

	}

}

func (h *handler) unreadNotifications(w http.ResponseWriter, r *http.Request) {
	unread, err := h.HasUnreadNotifications(r.Context())
	if err != nil {
		respondError(w, err)
		return
	}
	respond(w, unread, 200)
}
