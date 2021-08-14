package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"
)

const (
	TokenLifespan = time.Hour * 24 * 14
)

var ErrUnauthorized = errors.New("Unauthorized")

type LoginOutput struct {
	Token      string
	Expiration time.Time
	AuthUser   User
}

func (s *Service) Login(ctx context.Context, email string) (LoginOutput, error) {
	var output LoginOutput
	query := "SELECT id,username FROM users WHERE email = ?"
	err := s.db.QueryRowContext(ctx, query, email).Scan(&output.AuthUser.ID, &output.AuthUser.Username)
	if err == sql.ErrNoRows {
		fmt.Println("sql")
		return output, ErrUserNotFound
	}
	fmt.Println(output.AuthUser.ID)
	output.Token, err = s.codec.EncodeToString(strconv.FormatInt(output.AuthUser.ID, 10))
	if err != nil {
		fmt.Println(err.Error())
		return output, err

	}

	output.Expiration = time.Now().Add(TokenLifespan)
	return output, nil
}

func (s *Service) TokenDecode(token string) (int, error) {
	id, err := s.codec.DecodeToString(token)
	if err != nil {
		return 0, err
	}
	i, err := strconv.Atoi(id)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func (s *Service) AuthUser(ctx context.Context) (User, error) {
	var user User
	uid, ok := ctx.Value(KeyAuthUserID).(int)
	fmt.Println(uid)
	if !ok {
		return user, ErrUnauthorized
	}
	query := "SELECT id,username FROM users WHERE id=?"
	err := s.db.QueryRowContext(ctx, query, uid).Scan(&user.ID, &user.Username)
	if err != nil {
		return user, err
	}
	return user, nil
}
