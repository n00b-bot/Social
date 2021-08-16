package handler

import (
	"encoding/json"
	"net/http"
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
