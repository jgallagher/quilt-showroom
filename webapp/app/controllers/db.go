package controllers

import (
	"database/sql"
	"errors"
	_ "github.com/bmizerany/pq"
)

type Db struct {
	*sql.DB
}

var (
	db Db
	ErrNameTaken = errors.New("This username is already taken.")
	ErrEmailTaken = errors.New("This email address is already registered.")
	ErrBadName = errors.New("This username is invalid (use only letters, numbers, -, and _).")
	ErrBadPass = errors.New("Password must be at least 8 characters long.")
	ErrInvalidLogin = errors.New("Invalid email address or password.")
)

func init() {
	dbh, err := sql.Open("postgres", "sslmode=disable")
	if err != nil {
		panic(err)
	}
	db = Db{dbh}
}

func (db Db) CreateUser(name, email, password string) error {
	var code string
	var id sql.NullInt64
	row := db.QueryRow(`SELECT * FROM users_create($1,$2,$3)`, name, email, password)
	if err := row.Scan(&code, &id); err != nil {
		panic(err)
	}
	switch code {
	case "success":
		return nil
	case "dup_name":
		return ErrNameTaken
	case "dup_email":
		return ErrEmailTaken
	case "bad_name":
		return ErrBadName
	case "bad_pass":
		return ErrBadPass
	}
	panic("not reached")
}

func (db Db) Login(email, password string) (name string, err error) {
	rows, err := db.Query(`SELECT name FROM users WHERE email = $1 AND password = crypt($2, password)`, email, password)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	if rows.Next() {
		if err = rows.Scan(&name); err != nil {
			panic(err)
		}
		return
	}

	err = ErrInvalidLogin
	return
}
