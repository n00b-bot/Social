package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
)

type Notification struct {
	ID       int
	UserID   int
	Actors   []string
	Type     string
	IsRead   bool
	IssuedAt time.Time
}

func (s *Service) nofityFollow(followerID, followeeID int) {
	tx, err := s.db.Begin()
	if err != nil {
		log.Println(err.Error())
		return
	}
	defer tx.Rollback()
	var actor string
	query := "SELECT username from users WHERE id = ?"
	if err = tx.QueryRow(query, followerID).Scan(&actor); err != nil {
		log.Println(err.Error())
		fmt.Println("actor")
		return
	}

	fmt.Println(actor)
	var notified bool
	query = `SELECT exists 
		(Select 1 from notifications where user_id = ? 
			and ?  member of (actors) 
			and type='follow')`
	if err = tx.QueryRow(query, followeeID, actor).Scan(&notified); err != nil {
		fmt.Println("notified")
		log.Println(err.Error())
	}

	if notified {
		fmt.Println("nothing")
		return
	}
	fmt.Println(followeeID)
	var nid int
	query = `SELECT id from notifications where user_id = ?  and type ='follow' and isread=false`
	err = tx.QueryRow(query, followeeID).Scan(&nid)
	if err != nil && err != sql.ErrNoRows {
		fmt.Println("nid")
		log.Println(err.Error())
		return
	}
	fmt.Println(nid)
	var n Notification
	//actors := []string{actor}
	if err == sql.ErrNoRows {
		fmt.Println("go insert")
		query = `INSERT INTO notifications (user_id,actors,type) values (?,json_array(?),'follow')`
		if _, err := tx.Exec(query, followeeID, actor); err != nil {
			fmt.Println("insert")
			log.Println(err.Error())
			return
		}
		query = "SELECT id,issued_at from notifications WHERE id =last_insert_id()"
		if err := tx.QueryRow(query).Scan(&n.ID, &n.IssuedAt); err != nil {
			log.Println(err.Error())
			fmt.Println("select")
			return
		}
		//n.Actors = actors
	} else {
		query = `
			update notifications set actors=JSON_ARRAY_APPEND(actors,'$',?),issued_at = now() where id = ? 
		`
		if _, err := tx.Exec(query, actor, nid); err != nil {
			fmt.Println("update")
			log.Println(err.Error())
			return
		}
		query = `
			SELECT actors,issued_at from notifications where id = ?
		`
		var test string
		fmt.Println(nid)
		if err := tx.QueryRow(query, nid).Scan(&test, &n.IssuedAt); err != nil {
			log.Println(err.Error())
			fmt.Println("actors")
			return
		}
		fmt.Println(test)
		n.ID = nid

	}
	n.UserID = followeeID
	n.Type = "follow"
	if err = tx.Commit(); err != nil {
		log.Println(err.Error())
		return
	}
}

func (s *Service) Notifications(ctx context.Context, last int, before int) ([]Notification, error) {
	uid, ok := ctx.Value(KeyAuthUserID).(int)
	if !ok {
		return nil, ErrUnauthorized
	}
	fmt.Println(uid)
	last = normalizePageSize(last)
	query, args, err := buildQuery(`
		SELECT id,actors,type,isread,issued_at
		FROM notifications
		WHERE user_id = @a1
		{{if .a2}} AND id < @a2 {{end}}
		ORDER BY issued_at DESC
		LIMIT @a3`,
		map[string]interface{}{
			"a1": uid,
			"a2": before,
			"a3": last,
		})
	if err != nil {
		fmt.Println("error")
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		fmt.Println("query")
		return nil, err
	}
	defer rows.Close()
	var nn []Notification
	for rows.Next() {
		var actors string
		var n Notification
		if err = rows.Scan(
			&n.ID,
			&actors,
			&n.Type,
			&n.IsRead,
			&n.IssuedAt,
		); err != nil {
			return nil, err
		}
		fmt.Println(actors)
		s := actors[1 : len(actors)-1]
		a := strings.Split(s, ",")
		for _, v := range a {
			v = strings.ReplaceAll(v, " ", "")
			n.Actors = append(n.Actors, v[1:len(v)-1])
		}

		nn = append(nn, n)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return nn, nil
}

func (s *Service) MarkNotificationAsRead(ctx context.Context, nid int) error {
	uid, ok := ctx.Value(KeyAuthUserID).(int)
	if !ok {
		return ErrUnauthorized
	}
	query := `
	UPDATE notifications set isread =true 
	WHERE id = ? and user_id =?
	`
	if _, err := s.db.Exec(query, nid, uid); err != nil {
		return err
	}
	return nil
}

func (s *Service) MarkNotificationsAsRead(ctx context.Context) error {
	uid, ok := ctx.Value(KeyAuthUserID).(int)
	if !ok {
		return ErrUnauthorized
	}
	query := `
	UPDATE notifications set isread =true 
	WHERE user_id =?
	`
	if _, err := s.db.Exec(query, uid); err != nil {
		return err
	}
	return nil
}
