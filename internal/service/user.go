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

var (
	ErrUserNotFound = errors.New("User not found")
)

func (s *Service) CreateUser(ctx context.Context, username, email string) error {
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)
	query := "INSERT INTO users (username, email) VALUES ($1, $2)"
	_, err := s.db.Exec(query, username, email)
	if err != nil {
		return fmt.Errorf("Can not create user: %v", username)
	}
	return nil
}
