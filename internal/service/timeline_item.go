package service

import (
	"context"
	"database/sql"
)

type TimelineItem struct {
	ID     int64 `json:"id,omitempty"`
	UserID int64 `json:"-"`
	PostID int64 `json:"-"`
	Post   Post  `json:"post,omitempty"`
}

func (s *Service) Timeline(ctx context.Context, last int, before int) ([]TimelineItem, error) {
	var tt []TimelineItem
	uid, auth := ctx.Value(KeyAuthUserID).(int)
	if !auth {
		return nil, ErrUnauthorized
	}
	last = normalizePageSize(last)

	query, args, err := buildQuery(`
		SELECT timeline.id,posts.id,content,spoiler_of,nsfw,likes_count,comments_count,create_at
		,posts.user_id = @uid as mine
		,likes.user_id IS NOT NULL as liked	
		,subcriptions.user_id is not null as Subscribed
		,users.username,users.avatar
		FROM timeline 
		INNER JOIN posts ON timeline.post_id = posts.id
		INNER JOIN users ON posts.user_id = users.id 
		LEFT JOIN post_likes AS likes
		 ON likes.user_id = @uid AND likes.post_id=posts.id
		 LEFT JOIN post_subcriptions AS subcriptions
		 ON subcriptions.user_id = @uid AND subcriptions.post_id=posts.id
		WHERE timeline.user_id=@uid
		{{ if .before}}
		AND timeline.id < @before
		{{end}}
		ORDER BY create_at	DESC 
		LIMIT @last
	`, map[string]interface{}{
		"uid":    uid,
		"before": before,
		"last":   last,
	})
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var t TimelineItem
		var u User
		var avatar sql.NullString
		dest := []interface{}{
			&t.ID,
			&t.Post.ID,
			&t.Post.Content,
			&t.Post.SpoilerOf,
			&t.Post.NSFW,
			&t.Post.LikesCount,
			&t.Post.CommentsCount,
			&t.Post.CreateAt,
			&t.Post.Mine,
			&t.Post.Liked,
			&t.Post.Subscribed,
			&u.Username,
			&avatar,
		}
		if err := rows.Scan(dest...); err != nil {
			return nil, err
		}
		if avatar.Valid {
			avatarURL := "http://localhost:3000" + "/img/avatars" + avatar.String
			u.AvatarURL = &avatarURL
		}
		t.Post.User = &u
		tt = append(tt, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tt, nil

}
