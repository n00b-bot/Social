package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type ToggleFollow struct {
	Following      bool
	FollowersCount int
}

type User struct {
	ID       int64
	Username string
}

var KeyAuthUserID = "auth_user_id"
var (
	ErrUserNotFound    = errors.New("User not found")
	ErrUserTaken       = errors.New("User has been taken by another user")
	ErrForbiddenFollow = errors.New("Forbidden follow yourself")
)

func (s *Service) CreateUser(ctx context.Context, username, email string) error {
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)
	query := "INSERT INTO users (username, email) VALUES (?, ?)"
	_, err := s.db.Exec(query, username, email)
	if isUniqueViolation(err) {
		return ErrUserTaken
	}
	if err != nil {

		return fmt.Errorf("Can not create user: %v", username)
	}
	return nil
}

func (s *Service) ToggleFollow(ctx context.Context, username string) (ToggleFollow, error) {
	var out ToggleFollow
	follower_id, ok := ctx.Value(KeyAuthUserID).(int)
	if !ok {
		return out, ErrUnauthorized
	}
	var followee_id int
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return out, err
	}
	defer tx.Rollback()
	query := "SELECT id FROM users WHERE username=?"
	err = tx.QueryRowContext(ctx, query, username).Scan(&followee_id)

	if err != nil {
		return out, err
	}
	if follower_id == followee_id {
		return out, ErrForbiddenFollow
	}
	query = "SELECT EXISTS (SELECT 1 FROM follows WHERE follower_id = ? AND followee_id = ?)"
	err = tx.QueryRowContext(ctx, query, follower_id, followee_id).Scan(&out.Following)
	if err != nil {
		return out, err
	}
	if out.Following {
		query = "DELETE FROM follows where follower_id = ? AND followee_id =?"
		if _, err = tx.ExecContext(ctx, query, follower_id, followee_id); err != nil {
			return out, err
		}
		query = "UPDATE users SET followees_count = followees_count-1 WHERE id=?"
		if _, err = tx.ExecContext(ctx, query, follower_id); err != nil {
			return out, err
		}
		query = "UPDATE users SET followers_count = followers_count-1 WHERE id=?"
		if _, err = tx.ExecContext(ctx, query, followee_id); err != nil {
			return out, err
		}
		query = "SELECT followers_count FROM users WHERE id = ?"
		if err = tx.QueryRowContext(ctx, query, followee_id).Scan(&out.FollowersCount); err != nil {
			return out, err
		}
	} else {
		query = "INSERT INTO follows (follower_id,followee_id) VALUES (?,?)"
		if _, err = tx.ExecContext(ctx, query, follower_id, followee_id); err != nil {
			return out, err
		}
		query = "UPDATE users SET followees_count = followees_count+1 WHERE id=?"
		if _, err = tx.ExecContext(ctx, query, follower_id); err != nil {
			return out, err
		}
		query = "UPDATE users SET followers_count = followers_count+1 WHERE id=?"
		if _, err = tx.ExecContext(ctx, query, followee_id); err != nil {
			return out, err
		}
		query = "SELECT followers_count FROM users WHERE id = ?"
		if err = tx.QueryRowContext(ctx, query, followee_id).Scan(&out.FollowersCount); err != nil {
			return out, err
		}
	}
	if err = tx.Commit(); err != nil {
		return out, err
	}
	out.Following = !out.Following
	if out.Following {

	}
	return out, nil

}
