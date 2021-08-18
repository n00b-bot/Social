package handler

import (
	"net/http"
	"strconv"

	"github.com/matryer/way"
)

func (h *handler) notifications(w http.ResponseWriter, r *http.Request) {
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
