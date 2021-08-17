package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

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
