package models

import (
	"database/sql"
	_ "github.com/bmizerany/pq"
)

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("postgres", "sslmode=disable")
	if err != nil {
		panic(err)
	}
}
