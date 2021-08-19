package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/sanity-io/litter"
)

var (
	ErrInvalidContent = errors.New("Invalid content")
	ErrInvalidSpoiler = errors.New("Invalid spoler")
)

type ToggleLikeOutput struct {
	Liked      bool
	LikesCount int
}

type Post struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"-"`
	Content       string    `json:"content"`
	SpoilerOf     *string   `json:"spoiler_of"`
	LikesCount    int       `json:"likes_count"`
	CommentsCount int       `json:"comments_count"`
	NSFW          bool      `json:"nsfw"`
	CreateAt      time.Time `json:"create_at"`
	User          *User     `json:"user,omitempty"`
	Mine          bool      `json:"mine"`
	Liked         bool      `json:"liked"`
	Subscribed    bool
}

func (s *Service) CreatePost(ctx context.Context, content string, spoilerOf *string, nsfw bool) (TimelineItem, error) {
	var ti TimelineItem
	uid, ok := ctx.Value(KeyAuthUserID).(int)
	if !ok {
		return ti, ErrUnauthorized
	}
	content = strings.TrimSpace(content)
	if content == "" || len([]rune(content)) > 480 {
		return ti, ErrInvalidContent
	}
	if spoilerOf != nil {
		*spoilerOf = strings.TrimSpace(*spoilerOf)
		if *spoilerOf == "" || len([]rune(*spoilerOf)) > 64 {
			return ti, ErrInvalidSpoiler
		}
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ti, fmt.Errorf("cannot begin tx")
	}
	defer tx.Rollback()
	query := "INSERT INTO posts (user_id,content,spoiler_of,nsfw) values ($1,$2,$3,$4) returning id,create_at"
	if err = tx.QueryRowContext(ctx, query, uid, content, spoilerOf, nsfw).Scan(&ti.Post.ID, &ti.Post.CreateAt); err != nil {
		return ti, err
	}
	fmt.Print(ti.Post.ID)
	ti.Post.UserID = int64(uid)
	ti.Post.Content = content
	ti.Post.SpoilerOf = spoilerOf
	ti.Post.NSFW = nsfw
	ti.Post.Mine = true
	query = "insert into post_subcriptions (user_id,post_id) values ($1,$2)"
	if _, err = tx.ExecContext(ctx, query, uid, ti.Post.ID); err != nil {
		return ti, err
	}
	ti.Post.Subscribed = true
	query = " insert into timeline (user_id,post_id) values ($1,$2) returning id"
	if err = tx.QueryRowContext(ctx, query, uid, ti.Post.ID).Scan(&ti.ID); err != nil {
		return ti, err
	}
	ti.UserID = int64(uid)
	ti.PostID = ti.Post.ID
	if err = tx.Commit(); err != nil {
		return ti, err

	}
	go func(p Post) {
		uid, err := s.userByID(context.Background(), int(p.UserID))
		if err != nil {
			log.Println(err)
			return
		}
		p.User = &uid
		p.Mine = false
		p.Subscribed = false
		tt, err := s.fanoutPost(p)
		if err != nil {
			log.Println(err)
			return
		}
		for _, ti := range tt {
			log.Println(litter.Sdump(ti))
		}
	}(ti.Post)

	return ti, nil
}

func (s *Service) fanoutPost(p Post) ([]TimelineItem, error) {
	query := "INSERT INTO timeline (user_id,post_id) select follower_id,$1 FROM follows where followee_id = $2 returning id,user_id"
	rows, err := s.db.Query(query, p.ID, p.UserID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	tt := []TimelineItem{}
	for rows.Next() {
		var ti TimelineItem
		if err = rows.Scan(&ti.ID, &ti.UserID); err != nil {
			return nil, err
		}
		ti.PostID = p.ID
		ti.Post = p
		tt = append(tt, ti)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return tt, nil
}

func (s *Service) TogglePostLike(ctx context.Context, postID int) (ToggleLikeOutput, error) {
	var output ToggleLikeOutput
	uid, ok := ctx.Value(KeyAuthUserID).(int)
	if !ok {
		return output, ErrUnauthorized
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return output, err
	}
	defer tx.Rollback()
	query := "SELECT EXISTS (SELECT 1 FROM post_likes WHERE user_id= $1 and post_id =$2)"
	if err := tx.QueryRowContext(ctx, query, uid, postID).Scan(&output.Liked); err != nil {
		return output, nil
	}
	if output.Liked {
		query = " DELETE FROM post_likes WHERE user_id= $1 AND post_id= $2"
		if _, err = tx.ExecContext(ctx, query, uid, postID); err != nil {
			return output, err
		}
		query = "UPDATE posts SET likes_count=likes_count-1 where id = $1"
		if _, err = tx.ExecContext(ctx, query, postID); err != nil {
			return output, err
		}
		query = "select likes_count from posts where id= $1"
		if err = tx.QueryRowContext(ctx, query, postID).Scan(&output.LikesCount); err != nil {
			return output, err
		}
	} else {
		query = "INSERT INTO post_likes (user_id,post_id) values ($1,$2)"
		if _, err = tx.ExecContext(ctx, query, uid, postID); err != nil {
			return output, err
		}
		query = "UPDATE posts SET likes_count=likes_count+1 where id = $1"
		if _, err = tx.ExecContext(ctx, query, postID); err != nil {
			return output, err
		}
		query = "select likes_count from posts where id= $1"
		if err = tx.QueryRowContext(ctx, query, postID).Scan(&output.LikesCount); err != nil {
			return output, err
		}
	}
	if err = tx.Commit(); err != nil {
		return output, err
	}
	output.Liked = !output.Liked
	return output, nil
}

func (s *Service) Posts(ctx context.Context, username string, last int, before int) ([]Post, error) {
	var pp []Post
	uid, auth := ctx.Value(KeyAuthUserID).(int)
	last = normalizePageSize(last)

	query, args, err := buildQuery(`
		SELECT id,content,spoiler_of,nsfw,likes_count,comments_count,create_at
		{{ if .auth }}
		,posts.user_id = @uid as mine
		,likes.user_id IS NOT NULL as liked
		,subcriptions.user_id is not NULL as subscribed
		{{end}}
		FROM posts
		{{ if .auth }}
		LEFT JOIN post_likes AS likes
		 ON likes.user_id = @uid AND likes.post_id=posts.id
		 LEFT JOIN post_subcriptions AS subcriptions
		 ON subcriptions.user_id = @uid AND subcriptions.post_id=posts.i
		{{end}}
		WHERE posts.user_id = (SELECT id from users where username =@username)
		{{ if .before}}
		AND posts.id < @before
		{{end}}
		ORDER BY create_at	DESC 
		LIMIT @last
	`, map[string]interface{}{
		"auth":     auth,
		"uid":      uid,
		"username": username,
		"before":   before,
		"last":     last,
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
		var p Post
		dest := []interface{}{&p.ID, &p.Content, &p.SpoilerOf, &p.NSFW, &p.LikesCount, &p.CommentsCount, &p.CreateAt}
		if auth {
			dest = append(dest, &p.Mine, &p.Liked, &p.Subscribed)
		}
		if err := rows.Scan(dest...); err != nil {
			return nil, err
		}
		pp = append(pp, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return pp, nil
}

func (s *Service) Post(ctx context.Context, postID int) (Post, error) {
	var p Post
	uid, auth := ctx.Value(KeyAuthUserID).(int)
	query, args, err := buildQuery(`
		SELECT posts.id,content,spoiler_of,nsfw,likes_count,comments_count,create_at 
		,users.username,users.avatar
		{{ if .auth }}
		,posts.user_id = @uid as mine
		,likes.user_id IS NOT NULL as liked
		,subcriptions.user_id is not null as Subscribed
		{{end}}
		FROM posts
		INNER JOIN users ON  posts.user_id = users.id
		{{ if .auth }}
		LEFT JOIN post_likes AS likes
		 ON likes.user_id = @uid AND likes.post_id=posts.id
		 LEFT JOIN post_subcriptions AS subcriptions
		 ON subcriptions.user_id = @uid AND subcriptions.post_id=posts.id
		{{end}}
		WHERE posts.id = @post_id	
		
	`, map[string]interface{}{
		"auth":    auth,
		"uid":     uid,
		"post_id": postID,
	})
	if err != nil {
		return p, err
	}
	var u User
	var avatar sql.NullString
	dest := []interface{}{&p.ID, &p.Content, &p.SpoilerOf, &p.NSFW, &p.LikesCount, &p.CommentsCount, &p.CreateAt, &u.Username, &avatar}
	if auth {
		dest = append(dest, &p.Mine, &p.Liked, &p.Subscribed)
	}
	if err = s.db.QueryRowContext(ctx, query, args...).Scan(dest...); err != nil {
		return p, err
	}
	if avatar.Valid {
		avatarURL := "http://localhost:3000" + "/img/avatars" + avatar.String
		u.AvatarURL = &avatarURL
	}
	p.User = &u
	return p, nil
}
