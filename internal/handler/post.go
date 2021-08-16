package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/matryer/way"
)

type inputPost struct {
	Content   string  `json:"content"`
	SpoilerOf *string `json:"spoiler_of"`
	NSFW      bool    `json:"nsfw"`
}

func (h *handler) createPost(w http.ResponseWriter, r *http.Request) {
	var in inputPost
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondError(w, err)
		return
	}
	ti, err := h.CreatePost(r.Context(), in.Content, in.SpoilerOf, in.NSFW)
	if err != nil {
		respondError(w, err)
		return
	}
	respond(w, ti, 200)

}
func (h *handler) toggleLike(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	postID, _ := strconv.Atoi(way.Param(ctx, "post_id"))
	out, err := h.TogglePostLike(ctx, postID)
	if err != nil {
		respondError(w, err)
		return
	}
	respond(w, out, 200)
}
