package service

import (
	"context"
	"database/sql"
	"time"
)

type Comment struct {
	ID         int
	UserID     int
	PostID     int
	Content    string
	LikesCount int
	CreateAt   time.Time
	User       *User
	Mine       bool
	Liked      bool
}

func (s *Service) CreateComment(ctx context.Context, postID int, content string) (Comment, error) {
	var c Comment
	uid, auth := ctx.Value(KeyAuthUserID).(int)
	if !auth {
		return c, ErrUnauthorized
	}
	if content == "" || len([]rune(content)) > 480 {
		return c, ErrInvalidContent
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return c, err
	}
	defer tx.Rollback()
	query := "INSERT INTO comments (user_id,post_id,content) VALUES (?,?,?)"
	if _, err := tx.ExecContext(ctx, query, uid, postID, content); err != nil {
		return c, err
	}
	query = "SELECT id,create_at from comments WHERE id = (SELECT last_insert_id())"
	if err = tx.QueryRowContext(ctx, query).Scan(&c.ID, &c.CreateAt); err != nil {
		return c, err
	}
	c.UserID = uid
	c.PostID = postID
	c.Content = content
	c.Mine = true

	query = "UPDATE posts SET comments_count= comments_count+1 where id = ?"
	if _, err := tx.ExecContext(ctx, query, postID); err != nil {
		return c, err
	}
	if err = tx.Commit(); err != nil {
		return c, err
	}
	return c, nil
}

func (s *Service) Comments(ctx context.Context, postID int, last int, before int) ([]Comment, error) {
	var cc []Comment
	uid, ok := ctx.Value(KeyAuthUserID).(int)
	last = normalizePageSize(last)
	if !ok {
		return nil, ErrUnauthorized
	}
	query, args, err := buildQuery(`
		SELECT comments.id,content,likes_count,create_at,username,avatar
		{{ if .auth }}
		,comments.user_id =@a1 as mine
		,likes.user_id is not null as liked
		{{end}}
		FROM comments
		INNER JOIN users ON comments.user_id = users.id
		{{if .auth }}
		LEFT JOIN comment_likes AS likes 
		ON likes.comment_id = comments.id AND likes.user_id =@a2 
		{{end}}
		WHERE comments.post_id = @a3
		{{if .a4}}
		AND comments.id < @a4
		{{end}}
		ORDER BY create_at DESC
		LIMIT @a5
	`, map[string]interface{}{
		"a1":   uid,
		"a2":   uid,
		"a3":   postID,
		"a4":   before,
		"a5":   last,
		"auth": ok,
	})
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var c Comment
		var u User
		var avatar sql.NullString
		dest := []interface{}{
			&c.ID,
			&c.Content,
			&c.LikesCount,
			&c.CreateAt,
			&u.Username,
			&avatar,
		}
		if ok {
			dest = append(dest, &c.Mine, &c.Liked)
		}
		if err = rows.Scan(dest...); err != nil {
			return nil, err
		}
		if avatar.Valid {
			avatarURL := "http://localhost:3000" + "/img/avatars" + avatar.String
			u.AvatarURL = &avatarURL
		}
		c.User = &u
		cc = append(cc, c)

	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return cc, nil
}

func (s *Service) ToggleCommentLike(ctx context.Context, commentId int) (ToggleLikeOutput, error) {
	var output ToggleLikeOutput
	uid, auth := ctx.Value(KeyAuthUserID).(int)
	if !auth {
		return output, ErrUnauthorized
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return output, err
	}
	query := `
		SELECT EXISTS (SELECT 1 FROM comment_likes where user_id =? and comment_id =?)
	`
	if err = tx.QueryRowContext(ctx, query, uid, commentId).Scan(&output.Liked); err != nil {
		return output, err
	}
	if output.Liked {
		query = " DELETE FROM comment_likes where user_id = ? and comment_id = ?"
		if _, err := tx.ExecContext(ctx, query, uid, commentId); err != nil {
			return output, err
		}
		query = "UPDATE comments SET  likes_count =likes_count - 1 where id=?"
		if _, err := tx.ExecContext(ctx, query, commentId); err != nil {
			return output, err
		}
		query = "SELECT likes_count FROM comments where id = ?"
		if err := tx.QueryRowContext(ctx, query, commentId).Scan(&output.LikesCount); err != nil {
			return output, err
		}
	} else {
		query = "INSERT INTO  comment_likes (user_id, comment_id) VALUES (?,?)"
		if _, err := tx.ExecContext(ctx, query, uid, commentId); err != nil {
			return output, err
		}
		query = "UPDATE comments SET  likes_count =likes_count + 1 where id=?"
		if _, err := tx.ExecContext(ctx, query, commentId); err != nil {
			return output, err
		}
		query = "SELECT likes_count FROM comments where id = ?"
		if err := tx.QueryRowContext(ctx, query, commentId).Scan(&output.LikesCount); err != nil {
			return output, err
		}
	}
	if err = tx.Commit(); err != nil {
		return output, err
	}
	output.Liked = !output.Liked
	return output, nil
}
