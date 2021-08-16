package service

type TimelineItem struct {
	ID     int64 `json:"id,omitempty"`
	UserID int64 `json:"-"`
	PostID int64 `json:"-"`
	Post   Post  `json:"post,omitempty"`
}
