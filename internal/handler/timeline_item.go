package handler

import (
	"net/http"
	"strconv"
)

func (h *handler) timeLine(w http.ResponseWriter, r *http.Request) {
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
