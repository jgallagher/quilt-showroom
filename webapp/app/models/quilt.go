package models

import (
	"database/sql"
	"errors"
	"time"
)

var (
	ErrQuiltName = errors.New("You already have a quilt with this name!")
)

type Comment struct {
	User      string
	Comment   string
	Timestamp time.Time
}

type Quilt interface {
	Id() int
	Name() string
	UserId() string
	Visibility() string
	PostComment(username, comment string) error
	Comments() []Comment
}

type quilt struct {
	id         int
	name       string
	userId     string
	visibility string
}

func (q *quilt) Name() string       { return q.name }
func (q *quilt) Id() int            { return q.id }
func (q *quilt) UserId() string     { return q.userId }
func (q *quilt) Visibility() string { return q.visibility }

func (q *quilt) PostComment(username, comment string) error {
	_, err := db.Exec(
		`INSERT INTO quilt_comments(user_id,quilt_id,comment) VALUES($1,$2,$3)`,
		username, q.id, comment)
	return err
}

func (q *quilt) Comments() (comments []Comment) {
	rows, err := db.Query(`
		SELECT user_id,comment,created FROM quilt_comments WHERE quilt_id=$1
		ORDER BY created DESC`, q.id)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var c Comment
		if err = rows.Scan(&c.User, &c.Comment, &c.Timestamp); err != nil {
			panic(err)
		}
		comments = append(comments, c)
	}
	return
}

func createQuilt(username, name, visibility string, width, height int) (Quilt, error) {
	/*
		row := db.QueryRow(`
			INSERT INTO quilts(user_id, name, visibility, width, height)
			VALUES ($1, $2, $3, $4, $5) RETURNING quilt_id`,
			username, name, visibility, width, height)
	*/
	var code string
	var id sql.NullInt64
	row := db.QueryRow(`SELECT * FROM quilt_create($1, $2, $3, $4, $5)`,
		username, name, visibility, width, height)
	if err := row.Scan(&code, &id); err != nil {
		return nil, err
	}
	switch code {
	case "success":
		return &quilt{id: int(id.Int64), name: name}, nil
	case "dup_name":
		return nil, ErrQuiltName
	}
	panic("unexpected code from quilt_create")
}

func LoadQuilt(id int) (Quilt, error) {
	q := &quilt{id: id}
	row := db.QueryRow(`SELECT user_id,name,visibility FROM quilts WHERE quilt_id=$1`, id)
	if err := row.Scan(&q.userId, &q.name, &q.visibility); err != nil {
		return nil, err
	}
	return q, nil
}
