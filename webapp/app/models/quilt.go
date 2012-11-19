package models

import (
	"database/sql"
	"errors"
)

var (
	ErrQuiltName = errors.New("You already have a quilt with this name!")
)

type Quilt interface {
	Id() int
	Name() string
}

type quilt struct {
	id   int
	name string
}

func (q *quilt) Name() string { return q.name }
func (q *quilt) Id() int      { return q.id }

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
		return &quilt{int(id.Int64), name}, nil
	case "dup_name":
		return nil, ErrQuiltName
	}
	panic("unexpected code from quilt_create")
}
