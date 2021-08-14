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

type UserProfile struct {
	User
	Email          string
	FollowersCount int
	FolloweesCount int
	Me             bool
	Following      bool
	Followeed      bool
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

func (s *Service) User(ctx context.Context, username string) (UserProfile, error) {
	var out UserProfile
	uid, ok := ctx.Value(KeyAuthUserID).(int)
	args := []interface{}{}
	dest := []interface{}{&out.ID, &out.Username, &out.Email, &out.FolloweesCount, &out.FollowersCount}
	query := "SELECT id,username, email,followees_count,followers_count "
	if ok {
		query += ", followers.follower_id IS NOT NULL as Following," +
			"followees.followee_id IS NOT NULL AS Followeed "
		dest = append(dest, &out.Following, &out.Followeed)
	}
	query += " FROM Users "
	if ok {
		query += "LEFT JOIN follows AS followers ON followers.follower_id = ? AND followers.followee_id = users.id " +
			" LEFT JOIN follows AS followees ON followees.follower_id =users.id AND followees.followee_id = ?"
		args = append(args, uid, uid, username)
	}
	query += " WHERE username = ?"
	if err := s.db.QueryRowContext(ctx, query, args...).Scan(dest...); err != nil {
		return out, err
	}
	out.Me = ok && int64(uid) == out.ID
	if !ok || int64(uid) != out.ID {
		out.ID = 0
		out.Email = ""
	}
	return out, nil

}
