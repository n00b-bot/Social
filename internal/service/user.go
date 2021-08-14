package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type User struct {
	ID       int64
	Username string
}
var KeyAuthUserID="auth_user_id"
var (
	ErrUserNotFound = errors.New("User not found")
	ErrUserTaken = errors.New("User has been taken by another user")
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
		fmt.Println(err.Error())
		return fmt.Errorf("Can not create user: %v", username)
	}
	return nil
}
