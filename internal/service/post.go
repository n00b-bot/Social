package service

import (
	"context"
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
	ID        int64     `json:"id"`
	UserID    int64     `json:"-"`
	Content   string    `json:"content"`
	SpoilerOf *string   `json:"spoiler_of"`
	NSFW      bool      `json:"nsfw"`
	CreateAt  time.Time `json:"create_at"`
	User      *User     `json:"user,omitempty"`
	Mine      bool      `json:"mine"`
}

func (s *Service) CreatePost(ctx context.Context, content string, spoilerOf *string, nsfw bool) (TimelineItem, error) {
	var ti TimelineItem
	uid, ok := ctx.Value(KeyAuthUserID).(int)
	if !ok {
		return ti, ErrUnauthorized
	}
	fmt.Println(content)
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
	query := "INSERT INTO posts (user_id,content,spoiler_of,nsfw) values (?,?,?,?)"
	if _, err = tx.ExecContext(ctx, query, uid, content, spoilerOf, nsfw); err != nil {
		return ti, err
	}
	query = "select id,create_at from posts where id = (select last_insert_id())"
	if err = tx.QueryRowContext(ctx, query).Scan(&ti.Post.ID, &ti.Post.CreateAt); err != nil {
		return ti, err
	}
	ti.Post.UserID = int64(uid)
	ti.Post.Content = content
	ti.Post.SpoilerOf = spoilerOf
	ti.Post.NSFW = nsfw
	ti.Post.Mine = true
	query = " insert into timeline (user_id,post_id) values (?,?)"
	if _, err = tx.ExecContext(ctx, query, uid, ti.Post.ID); err != nil {
		return ti, err
	}
	query = "select id from timeline where id = (select last_insert_id())"
	if err = tx.QueryRowContext(ctx, query).Scan(&ti.ID); err != nil {
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
	query := "INSERT INTO timeline (user_id,post_id) select follower_id,? FROM follows where followee_id = ?"
	if _, err := s.db.Exec(query, p.ID, p.UserID); err != nil {
		fmt.Println("1")
		return nil, err
	}
	query = "select id , user_id from timeline where post_id = (select post_id from timeline where id = last_insert_id()) and user_id != ?"
	rows, err := s.db.Query(query, p.UserID)
	if err != nil {
		fmt.Println("2")
		return nil, err
	}
	defer rows.Close()
	tt := []TimelineItem{}
	for rows.Next() {
		var ti TimelineItem
		if err = rows.Scan(&ti.ID, &ti.UserID); err != nil {
			fmt.Println("3")
			return nil, err
		}
		ti.PostID = p.ID
		ti.Post = p
		tt = append(tt, ti)
	}
	if err = rows.Err(); err != nil {
		fmt.Println("4")
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
	query := "SELECT EXISTS (SELECT 1 FROM post_likes WHERE user_id= ? and post_id =?)"
	if err := tx.QueryRowContext(ctx, query, uid, postID).Scan(&output.Liked); err != nil {
		return output, nil
	}
	if output.Liked {
		query = " DELETE FROM post_likes WHERE user_id= ? AND post_id= ?"
		if _, err = tx.ExecContext(ctx, query, uid, postID); err != nil {
			return output, err
		}
		query = "UPDATE posts SET likes_count=likes_count-1 where id = ?"
		if _, err = tx.ExecContext(ctx, query, postID); err != nil {
			return output, err
		}
		query = "select likes_count from posts where id= ?"
		if err = tx.QueryRowContext(ctx, query, postID).Scan(&output.LikesCount); err != nil {
			return output, err
		}
	} else {
		query = "INSERT INTO post_likes (user_id,post_id) values (?,?)"
		if _, err = tx.ExecContext(ctx, query, uid, postID); err != nil {
			return output, err
		}
		query = "UPDATE posts SET likes_count=likes_count+1 where id = ?"
		if _, err = tx.ExecContext(ctx, query, postID); err != nil {
			return output, err
		}
		query = "select likes_count from posts where id= ?"
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
