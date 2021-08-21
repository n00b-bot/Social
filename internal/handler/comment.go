package handler

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"strconv"
	"time"

	"github.com/matryer/way"
)

type createCommentInput struct {
	Content string
}

func (h *handler) createComment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	defer r.Body.Close()
	var content createCommentInput
	if err := json.NewDecoder(r.Body).Decode(&content); err != nil {
		respondError(w, err)
		return
	}
	postID, _ := strconv.Atoi(way.Param(ctx, "post_id"))
	c, err := h.CreateComment(ctx, postID, content.Content)
	if err != nil {
		respondError(w, err)
		return
	}
	respond(w, c, 200)

}
func (h *handler) comments(w http.ResponseWriter, r *http.Request) {
	if a, _, err := mime.ParseMediaType(r.Header.Get("Accept")); err == nil && a == "text/event-stream" {
		h.subcribeToComment(w, r)
		return
	}
	ctx := r.Context()
	postID, _ := strconv.Atoi(way.Param(ctx, "post_id"))
	last, _ := strconv.Atoi(r.URL.Query().Get("last"))
	before, _ := strconv.Atoi(r.URL.Query().Get("before"))
	cc, err := h.Comments(ctx, postID, last, before)
	if err != nil {
		respondError(w, err)
		return
	}
	respond(w, cc, 200)

}
func (h *handler) toggleLikeComment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	comment, _ := strconv.Atoi(way.Param(ctx, "comment_id"))
	cc, err := h.ToggleCommentLike(ctx, comment)
	if err != nil {
		respondError(w, err)
		return
	}
	respond(w, cc, 200)
}

func (h *handler) subcribeToComment(w http.ResponseWriter, r *http.Request) {
	f, ok := w.(http.Flusher)
	if !ok {
		respondError(w, errStreaming)
		return
	}
	ctx := r.Context()
	postID, _ := strconv.Atoi(way.Param(ctx, "post_id"))
	cc := h.SubcribeToComment(ctx, postID)

	header := r.Header
	header.Set("Cache-Control", "no-cache")
	header.Set("Connection", "keep-alive")
	header.Set("Context-Type", "text/event-stream")

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(h.ping):
			fmt.Fprintf(w, "ping \n\n")
			f.Flush()
		case c := <-cc:
			writeSSE(w, c)
			f.Flush()
		}
	}
}
