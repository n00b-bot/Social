package service

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"
)

const (
	TokenLifespan = time.Hour * 24 * 14
)

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
