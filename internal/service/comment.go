package service

import (
	"context"
	"database/sql"
	"log"
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
type commentClient struct {
	comments chan Comment
	postID   int
	userID   *int
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
	query := "INSERT INTO comments (user_id,post_id,content) VALUES ($1,$2,$3) returning id,create_at"
	if err = tx.QueryRowContext(ctx, query, uid, postID, content).Scan(&c.ID, &c.CreateAt); err != nil {
		return c, err
	}
	c.UserID = uid
	c.PostID = postID
	c.Content = content
	c.Mine = true
	query = "insert into post_subcriptions (user_id,post_id) values ($1,$2) on conflict(user_id,post_id) DO NOTHING"
	if _, err = tx.ExecContext(ctx, query, uid, postID); err != nil {
		return c, err
	}
	query = "UPDATE posts SET comments_count= comments_count+1 where id = $1"
	if _, err := tx.ExecContext(ctx, query, postID); err != nil {
		return c, err
	}
	if err = tx.Commit(); err != nil {
		return c, err
	}
	go s.commentCreated(c)
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
		,comments.user_id =@uid as mine
		,likes.user_id is not null as liked
		{{end}}
		FROM comments
		INNER JOIN users ON comments.user_id = users.id
		{{if .auth }}
		LEFT JOIN comment_likes AS likes 
		ON likes.comment_id = comments.id AND likes.user_id =@uid 
		{{end}}
		WHERE comments.post_id = @post_id
		{{if .before}}
		AND comments.id < @before
		{{end}}
		ORDER BY create_at DESC
		LIMIT @last
	`, map[string]interface{}{
		"uid":     uid,
		"post_id": postID,
		"before":  before,
		"last":    last,
		"auth":    ok,
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
		SELECT EXISTS (SELECT 1 FROM comment_likes where user_id =$1 and comment_id =$2)
	`
	if err = tx.QueryRowContext(ctx, query, uid, commentId).Scan(&output.Liked); err != nil {
		return output, err
	}
	if output.Liked {
		query = " DELETE FROM comment_likes where user_id = $1 and comment_id = $2"
		if _, err := tx.ExecContext(ctx, query, uid, commentId); err != nil {
			return output, err
		}
		query = "UPDATE comments SET  likes_count =likes_count - 1 where id=$1"
		if _, err := tx.ExecContext(ctx, query, commentId); err != nil {
			return output, err
		}
		query = "SELECT likes_count FROM comments where id = $1"
		if err := tx.QueryRowContext(ctx, query, commentId).Scan(&output.LikesCount); err != nil {
			return output, err
		}
	} else {
		query = "INSERT INTO  comment_likes (user_id, comment_id) VALUES ($1,$2)"
		if _, err := tx.ExecContext(ctx, query, uid, commentId); err != nil {
			return output, err
		}
		query = "UPDATE comments SET  likes_count =likes_count + 1 where id=$1"
		if _, err := tx.ExecContext(ctx, query, commentId); err != nil {
			return output, err
		}
		query = "SELECT likes_count FROM comments where id = $1"
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

func (s *Service) commentCreated(c Comment) {
	u, err := s.userByID(context.Background(), c.UserID)
	if err != nil {
		log.Println(err)
		return
	}
	c.User = &u
	c.Mine = false
	go s.notifyComment(c)
	go s.notifyCommentMention(c)
	go s.broadcastComment(c)
}

func (s *Service) SubcribeToComment(ctx context.Context, postID int) chan Comment {

	cc := make(chan Comment)
	c := &commentClient{
		comments: cc,
		postID:   postID,
	}
	if uid, ok := ctx.Value(KeyAuthUserID).(int); ok {
		c.userID = &uid
	}
	s.commentClient.Store(c, struct{}{})
	go func() {
		<-ctx.Done()
		s.commentClient.Delete(c)
		close(cc)
	}()
	return cc
}

func (s *Service) broadcastComment(c Comment) {

	s.commentClient.Range(func(key, _ interface{}) bool {
		client := key.(*commentClient)
		if client.postID == c.PostID && !(client.userID != nil && *client.userID == c.UserID) {
			client.comments <- c
		}
		return true
	})
}
