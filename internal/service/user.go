package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path"
	"strings"

	"github.com/disintegration/imaging"
	gonanoid "github.com/matoous/go-nanoid"
	"github.com/sanity-io/litter"
)

type ToggleFollow struct {
	Following      bool
	FollowersCount int
}

type User struct {
	ID        int64   `json:"id,omitempty"`
	Username  string  `json:"username,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
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

var (
	MaxAvatarSize = 5 << 20
	KeyAuthUserID = "auth_user_id"
	AvatarDir     = path.Join("web", "static", "img", "avatar")
)

var (
	ErrUserNotFound    = errors.New("User not found")
	ErrUserTaken       = errors.New("User has been taken by another user")
	ErrForbiddenFollow = errors.New("Forbidden follow yourself")
	ErrAvatarType      = errors.New("Avatar type is unsupported")
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
	var avatar sql.NullString
	args := []interface{}{}
	dest := []interface{}{&out.ID, &out.Username, &out.Email, &avatar, &out.FolloweesCount, &out.FollowersCount}
	query := "SELECT id,username, email,avatar,followees_count,followers_count "
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
	if out.Me {
		out.ID = 0
		out.Email = ""
	}
	if avatar.Valid {
		avatarURL := "http://localhost:3000" + "/img/avatars" + avatar.String
		out.AvatarURL = &avatarURL
	}
	return out, nil

}

func (s *Service) Users(ctx context.Context, search string, first int, after string) ([]UserProfile, error) {
	uid, ok := ctx.Value(KeyAuthUserID).(int)
	first = normalizePageSize(first)
	ints := map[string]interface{}{
		"auth": ok,
		"1":    uid,
		"2":    uid,
		"3":    search,
		"4":    after,
		"5":    first}

	query, args, err := buildQuery(`
		SELECT id, email, username,avatar, followers_count, followees_count
		{{ if .auth }}
		, followers.follower_id IS NOT NULL AS following
		, followees.followee_id IS NOT NULL AS followeed
		{{ end }}
		FROM users
		{{ if .auth }}
		LEFT JOIN follows AS followers
			ON followers.follower_id = @1 AND followers.followee_id = users.id
		LEFT JOIN follows AS followees
			ON followees.follower_id = users.id AND followees.followee_id = @2
		{{ end }}
		{{ if or .3 .4 }}WHERE{{ end }}
		{{ if .3 }} username LIKE concat('%', @3 ,'%'){{ end }}
		{{ if and .3 .4 }}AND{{ end }}
		{{ if .3 }}username > @4 {{ end }}
		ORDER BY username ASC
		LIMIT @5`, ints)

	row, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer row.Close()
	uu := make([]UserProfile, 0, first)
	for row.Next() {
		var avatar sql.NullString
		var u UserProfile
		dest := []interface{}{&u.ID, &u.Email, &u.Username, &avatar, &u.FollowersCount, &u.FolloweesCount}
		if ok {
			dest = append(dest, &u.Following, &u.Followeed)
		}
		if err := row.Scan(dest...); err != nil {
			return uu, err
		}
		u.Me = ok && int64(uid) == u.ID
		if !u.Me {
			u.ID = 0
			u.Email = ""
		}
		if avatar.Valid {
			avatarURL := "http://localhost:3000" + "/img/avatars" + avatar.String
			u.AvatarURL = &avatarURL
		}
		uu = append(uu, u)
	}
	if err := row.Err(); err != nil {
		return nil, err
	}
	return uu, nil
}

func (s *Service) Follwers(ctx context.Context, username string, first int, after string) ([]UserProfile, error) {
	uid, ok := ctx.Value(KeyAuthUserID).(int)
	first = normalizePageSize(first)
	ints := map[string]interface{}{
		"auth": ok,
		"a1":   uid,
		"a2":   uid,
		"a3":   username,
		"a4":   after,
		"a5":   first}

	query, args, err := buildQuery(`
		SELECT id, email, username,avatar, followers_count, followees_count
		{{ if .auth }}
		, followers.follower_id IS NOT NULL AS following
		, followees.followee_id IS NOT NULL AS followeed
		{{ end }}
		FROM follows
		INNER JOIN users on follower_id=users.id
		{{ if .auth }}
		LEFT JOIN follows AS followers
			ON followers.follower_id = @a1 AND followers.followee_id = users.id
		LEFT JOIN follows AS followees
			ON followees.follower_id = users.id AND followees.followee_id = @a2
		{{ end }}
		WHERE follows.followee_id =(SELECT id from users where username = @a3)
		{{ if  .a4 }}AND username > @a4 {{ end }}
		ORDER BY username ASC
		LIMIT @a5`, ints)
	fmt.Println(litter.Sdump(args))
	row, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer row.Close()
	var avatar sql.NullString
	uu := make([]UserProfile, 0, first)
	for row.Next() {
		var u UserProfile
		dest := []interface{}{&u.ID, &u.Email, &u.Username, &avatar, &u.FollowersCount, &u.FolloweesCount}
		if ok {
			dest = append(dest, &u.Following, &u.Followeed)
		}
		if err := row.Scan(dest...); err != nil {
			return uu, err
		}
		u.Me = ok && int64(uid) == u.ID
		if !u.Me {
			u.ID = 0
			u.Email = ""
		}
		if avatar.Valid {
			avatarURL := "http://localhost:3000" + "/img/avatars" + avatar.String
			u.AvatarURL = &avatarURL
		}
		uu = append(uu, u)
	}
	if err := row.Err(); err != nil {
		return nil, err
	}
	return uu, nil
}

func (s *Service) Follwees(ctx context.Context, username string, first int, after string) ([]UserProfile, error) {
	uid, ok := ctx.Value(KeyAuthUserID).(int)
	first = normalizePageSize(first)
	ints := map[string]interface{}{
		"auth": ok,
		"a1":   uid,
		"a2":   uid,
		"a3":   username,
		"a5":   first,
	}
	if after != "" {
		ints["a4"] = after
	}

	query, args, err := buildQuery(`
		SELECT id, email, username, avatar,followers_count, followees_count
		{{ if .auth }}
		, followers.follower_id IS NOT NULL AS following
		, followees.followee_id IS NOT NULL AS followeed
		{{ end }}
		FROM follows
		INNER JOIN users on followee_id=users.id
		{{ if .auth }}
		LEFT JOIN follows AS followers
			ON followers.follower_id = @a1 AND followers.followee_id = users.id
		LEFT JOIN follows AS followees
			ON followees.follower_id = users.id AND followees.followee_id = @a2
		{{ end }}
		WHERE follows.follower_id =(SELECT id from users where username = @a3)
		{{ if  .a4 }}
		AND username > @a4 
		{{ end }}
		ORDER BY username ASC
		LIMIT @a5`, ints)
	row, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer row.Close()
	uu := make([]UserProfile, 0, first)
	for row.Next() {
		var u UserProfile
		var avatar sql.NullString
		dest := []interface{}{&u.ID, &u.Email, &u.Username, &avatar, &u.FollowersCount, &u.FolloweesCount}
		if ok {
			dest = append(dest, &u.Following, &u.Followeed)
		}
		if err := row.Scan(dest...); err != nil {
			return uu, err
		}
		u.Me = ok && int64(uid) == u.ID
		if !u.Me {
			u.ID = 0
			u.Email = ""
		}
		if avatar.Valid {
			avatarURL := "http://localhost:3000" + "/img/avatars" + avatar.String
			u.AvatarURL = &avatarURL
		}
		uu = append(uu, u)
	}
	if err := row.Err(); err != nil {
		return nil, err
	}
	return uu, nil
}

func (s *Service) UpdateAvatar(ctx context.Context, r io.Reader) (string, error) {
	uid, ok := ctx.Value(KeyAuthUserID).(int)
	if !ok {
		return "", ErrUnauthorized
	}
	img, format, err := image.Decode(r)
	if err != nil {
		return "", err
	}

	if format != "png" && format != "jpeg" {
		return "", ErrAvatarType
	}
	avatar, err := gonanoid.Nanoid()
	if err != nil {
		return "", err
	}
	if format == "png" {
		avatar += ".png"
	} else {
		avatar += ".jpg"
	}
	avatarPath := path.Join(AvatarDir, avatar)
	f, err := os.Create(avatarPath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	img = imaging.Fill(img, 400, 400, imaging.Center, imaging.CatmullRom)
	if format == "png" {
		err = png.Encode(f, img)
	} else {
		err = jpeg.Encode(f, img, nil)
	}
	if err != nil {
		return "", err
	}
	var oldAvartar sql.NullString
	if err = s.db.QueryRowContext(ctx, `SELECT avatar FROM users WHERE id = ?`, uid).Scan(&oldAvartar); err != nil {
		defer os.Remove(avatarPath)
		return "", err
	}
	if oldAvartar.Valid {
		if err = os.Remove(path.Join(AvatarDir, oldAvartar.String)); err != nil {
			fmt.Print(err)
		}
	}
	s.db.ExecContext(ctx, `UPDATE users SET avatar=? WHERE id=?`, avatar, uid)

	return "http://localhost:3000/web/static/img/avatar" + avatar, nil
}
func (s *Service) userByID(ctx context.Context, uid int) (User, error) {
	var user User
	query := "SELECT username,avatar FROM users WHERE id=?"
	var avatar sql.NullString
	err := s.db.QueryRowContext(ctx, query, uid).Scan(&user.Username, &avatar)
	user.ID = int64(uid)
	if err != nil {
		return user, err
	}
	if avatar.Valid {
		avatarURL := "http://localhost:3000" + "/img/avatars" + avatar.String
		user.AvatarURL = &avatarURL
	}
	return user, nil

}
