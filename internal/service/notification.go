package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/lib/pq"
)

type Notification struct {
	ID       int
	UserID   int
	Actors   []string
	Type     string
	PostID   *int
	Read     bool
	IssuedAt time.Time
}

type ToggleSubcriptionOutput struct {
	Subcribed bool
}

func (s *Service) nofityFollow(followerID, followeeID int) {
	tx, err := s.db.Begin()
	if err != nil {
		log.Println(err.Error())
		return
	}
	defer tx.Rollback()
	var actor string
	query := "SELECT username from users WHERE id = $1"
	if err = tx.QueryRow(query, followerID).Scan(&actor); err != nil {
		log.Println(err.Error())
		fmt.Println("actor")
		return
	}

	fmt.Println(actor)
	var notified bool
	query = `SELECT exists 
		(Select 1 from notifications where user_id = $1 
			and $2:::VARCHAR = any(actors)
			and type = 'follow')`
	if err = tx.QueryRow(query, followeeID, actor).Scan(&notified); err != nil {
		fmt.Println("notified")
		log.Println(err.Error())
	}

	if notified {
		fmt.Println("nothing")
		return
	}
	var nid int
	query = `SELECT id from notifications where user_id = $1  and type ='follow' and read=false`
	err = tx.QueryRow(query, followeeID).Scan(&nid)
	if err != nil && err != sql.ErrNoRows {
		fmt.Println("nid")
		log.Println(err.Error())
		return
	}
	fmt.Println(nid)
	var n Notification
	actors := []string{actor}
	if err == sql.ErrNoRows {
		fmt.Println("go insert")
		query = `INSERT INTO notifications (user_id,actors,type) values ($1,$2,'follow')
		returning id ,issued_at`
		if err := tx.QueryRow(query, followeeID, pq.Array(actors)).Scan(&n.ID, &n.IssuedAt); err != nil {
			fmt.Println("insert")
			log.Println(err.Error())
			return
		}
		n.Actors = actors
	} else {
		query = `
			update notifications set actors=array_prepend($1,notifications.actors),issued_at = now() where id = $2
			returning actors,issued_at
		`
		if err := tx.QueryRow(query, actor, nid).Scan(pq.Array(&n.Actors), &n.IssuedAt); err != nil {
			log.Println(err.Error())
			fmt.Println("actors")
			return
		}
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
		SELECT id,actors,type,post_id,read,issued_at
		FROM notifications
		WHERE user_id = @uid
		{{if .before}} AND id < @before {{end}}
		ORDER BY issued_at DESC
		LIMIT @last`,
		map[string]interface{}{
			"uid":    uid,
			"before": before,
			"last":   last,
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
		var n Notification
		if err = rows.Scan(
			&n.ID,
			pq.Array(&n.Actors),
			&n.Type,
			&n.PostID,
			&n.Read,
			&n.IssuedAt,
		); err != nil {
			return nil, err
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
	UPDATE notifications set read =true 
	WHERE id = $1 and user_id =$2
	`
	fmt.Println(nid)
	fmt.Println(uid)

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
	UPDATE notifications set read =true 
	WHERE user_id =$1
	`
	if _, err := s.db.Exec(query, uid); err != nil {
		return err
	}
	return nil
}

func (s *Service) notifyComment(c Comment) {
	actor := c.User.Username
	query := `
		INSERT INTO notifications (user_id, actors,type,post_id)
		SELECT user_id,$1,'comment',$2 from post_subcriptions
		WHERE post_subcriptions.user_id != $3
		AND post_subcriptions.post_id= $2
		ON CONFLICT (user_id,type,post_id,read) 
		DO UPDATE SET actors=array_prepend($4,array_remove(notifications.actors,$4)),
		issued_at=now()
		returning id,user_id,actors,issued_at
	`
	actors := []string{actor}
	rows, err := s.db.Query(query, pq.Array(actors), c.PostID, c.UserID, actor)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()
	var nn []Notification
	for rows.Next() {
		var n Notification
		if err = rows.Scan(&n.ID, n.UserID, pq.Array(n.Actors), n.IssuedAt); err != nil {
			log.Println(err)
			return
		}
		n.Type = "comment"
		n.PostID = &c.PostID
		nn = append(nn, n)

	}
	if err = rows.Err(); err != nil {
		log.Println(err)
		return
	}
}

func (s *Service) TogglePostSubcription(ctx context.Context, postID int) (ToggleSubcriptionOutput, error) {
	var o ToggleSubcriptionOutput
	uid, ok := ctx.Value(KeyAuthUserID).(int)
	if !ok {
		return o, ErrUnauthorized
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return o, err
	}
	defer tx.Rollback()
	query := "SELECT EXISTS (SELECT 1 FROM post_subcriptions where user_id = $1 and post_id = $2)"
	if err = tx.QueryRowContext(ctx, query, uid, postID).Scan(&o.Subcribed); err != nil {
		return o, err
	}
	if o.Subcribed {
		query = " DELETE FROM post_subcriptions where user_id = $1 and post_id = $2"
		if _, err = tx.ExecContext(ctx, query, uid, postID); err != nil {
			return o, nil
		}
	} else {
		query = "INSERT INTO post_subcriptions (user_id, post_id) values ($1,$2)"
		if _, err = tx.ExecContext(ctx, query, uid, postID); err != nil {
			return o, nil
		}
	}
	if err = tx.Commit(); err != nil {
		return o, err
	}
	o.Subcribed = !o.Subcribed
	return o, nil
}

func (s *Service) notifyPostMention(p Post) {
	mentions := collectionMentions(p.Content)
	if len(mentions) == 0 {
		return
	}
	actors := []string{p.User.Username}
	rows, err := s.db.Query(`
		INSERT INTO notifications (user_id,actors,type,post_id) 
		SELECT users.id,$1,'post_mention',$2 from Users
		where users.id != $3 and
		username = any($4)
		returning id ,user_id,issued_at
	`, pq.Array(actors), p.ID, p.UserID, pq.Array(mentions))
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()
	var nn []Notification
	for rows.Next() {
		var n Notification
		if err = rows.Scan(&n.ID, &n.UserID, &n.IssuedAt); err != nil {
			log.Println(err)
			return
		}
		n.Actors = actors
		n.Type = "post_mention"
		n.PostID = &p.ID
		nn = append(nn, n)
	}
	if err := rows.Err(); err != nil {
		log.Println(err)
		return
	}
	log.Printf("post notificatnion %v\n", nn)
}

func (s *Service) notifyCommentMention(c Comment) {
	mentions := collectionMentions(c.Content)
	if len(mentions) == 0 {
		return
	}
	actors := c.User.Username
	rows, err := s.db.Query(`
		INSERT INTO notifications (user_id,actors,type,post_id) 
		SELECT users.id,$1,'comment_mention',$2 from Users
		where users.id != $3 and
		username = any($4)	
		ON CONFLICT (user_id,type,post_id,read) do UPDATE SET
		actors=array_prepend($5,array_remove(notifications.actors,$5)),
		issued_at=now()
		returning id ,user_id,actors,issued_at
	`, pq.Array([]string{actors}), c.PostID, c.UserID, pq.Array(mentions), actors)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()
	var nn []Notification
	for rows.Next() {
		var n Notification
		if err = rows.Scan(&n.ID, &n.UserID, pq.Array(&n.Actors), &n.IssuedAt); err != nil {
			log.Println(err)
			return
		}
		n.Type = "comment_mention"
		n.PostID = &c.PostID
		nn = append(nn, n)
	}
	if err := rows.Err(); err != nil {
		log.Println(err)
		return
	}
	log.Printf("comment notificatnion %v\n", nn)
}
