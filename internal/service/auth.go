package service

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/url"
	"strconv"
	"time"
)

const (
	TokenLifespan = time.Hour * 24 * 14
)

var ErrUnauthorized = errors.New("Unauthorized")
var MagicLinkTemplates *template.Template
var MinutesExpired = time.Minute * 15
var LinkExpired = errors.New("Link expired")

type LoginOutput struct {
	Token      string `json:"token"`
	Expiration time.Time
	AuthUser   User
}

func (s *Service) Login(ctx context.Context, email string) (LoginOutput, error) {
	var output LoginOutput
	query := "SELECT id,username,avatar FROM users WHERE email = $1"
	var avatar sql.NullString
	err := s.db.QueryRowContext(ctx, query, email).Scan(&output.AuthUser.ID, &output.AuthUser.Username, &avatar)
	if err == sql.ErrNoRows {
		return output, ErrUserNotFound
	}
	fmt.Println(output.AuthUser.ID)
	output.Token, err = s.codec.EncodeToString(strconv.FormatInt(output.AuthUser.ID, 10))
	if err != nil {
		return output, err

	}
	if avatar.Valid {
		avatarURL := "http://localhost:3000" + "/img/avatars" + avatar.String
		output.AuthUser.AvatarURL = &avatarURL
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
	if !ok {
		return user, ErrUnauthorized
	}
	return s.userByID(ctx, uid)

}

func (s *Service) SendMagicLink(ctx context.Context, email, redirectURL string) error {
	uri, err := url.ParseRequestURI(redirectURL)
	if err != nil {
		return err
	}
	var verification_code string
	query := "INSERT INTO verification_codes (user_id)  (select id from users where email =$1) returning id"
	if err := s.db.QueryRowContext(ctx, query, email).Scan(&verification_code); err != nil {
		return err
	}
	magicLink, _ := url.Parse("http://localhost:3000")
	magicLink.Path = "/api/auth_redirect"
	q := magicLink.Query()
	q.Set("verification_code", verification_code)
	q.Set("redirect_uri", uri.String())
	magicLink.RawQuery = q.Encode()
	if MagicLinkTemplates == nil {
		MagicLinkTemplates, err = template.ParseFiles("web/template/magic-link.html")
		if err != nil {
			return err
		}
	}
	var mail bytes.Buffer
	if err := MagicLinkTemplates.Execute(&mail, map[string]interface{}{
		"MagicLink": magicLink.String(),
		"Minutes":   int(MinutesExpired.Minutes()),
	}); err != nil {
		return err
	}
	go s.deleteExpriredCode(verification_code)
	return s.sendMail(email, "Magic Link", mail.String())
}

func (s *Service) AuthURI(ctx context.Context, verification_code, redirectURL string) (string, error) {
	uri, err := url.ParseRequestURI(redirectURL)
	if err != nil {
		return "", err
	}
	var uid int
	var ts time.Time
	query := "DELETE FROM verification_codes where id = $1 returning user_id,create_at"
	err = s.db.QueryRowContext(ctx, query, verification_code).Scan(&uid, &ts)
	if err != nil {
		return "", err
	}
	if ts.Add(MinutesExpired).Before(time.Now()) {
		return "", LinkExpired
	}
	token, err := s.codec.EncodeToString(strconv.Itoa(uid))
	if err != nil {
		return "", err
	}
	exp, err := time.Now().Add(MinutesExpired).MarshalText()
	if err != nil {
		return "", err
	}
	f := url.Values{}
	f.Set("token", token)
	f.Set("expired_at", string(exp))
	uri.Fragment = f.Encode()

	return uri.String(), nil
}

func (s *Service) deleteExpriredCode(code string) {
	<-time.After(time.Hour * 24)
	if _, err := s.db.Exec(`DELETE FROM verification_codes WHERE id = $1`, code); err != nil {
		log.Println(err)
	}
}
