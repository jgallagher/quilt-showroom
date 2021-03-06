package models

import (
	"errors"
)

type User interface {
	Name() string
	NumPublicQuilts() int
	PublicQuilts() []*Quilt
	Quilts() []*Quilt
	CreateQuilt(name, visibility string, width, height int) (*Quilt, error)
}

type user struct {
	name            string
	numPublicQuilts int
}

var (
	ErrNameTaken    = errors.New("This username is already taken.")
	ErrEmailTaken   = errors.New("This email address is already registered.")
	ErrBadName      = errors.New("This username is invalid (use only letters, numbers, -, and _).")
	ErrBadPass      = errors.New("Password must be at least 8 characters long.")
	ErrInvalidLogin = errors.New("Invalid email address or password.")
	ErrNoUser       = errors.New("User not found.")
)

func AllUsers() (users []User) {
	rows, err := db.Query(`
		SELECT user_id,COUNT(quilt_id) FROM quilts
		GROUP BY user_id ORDER BY user_id`)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var u user
		if err := rows.Scan(&u.name, &u.numPublicQuilts); err != nil {
			rows.Close()
			panic(err)
		}
		users = append(users, &u)
	}

	return users
}

func CreateUser(name, email, password string) error {
	var code string
	row := db.QueryRow(`SELECT * FROM users_create($1,$2,$3)`, name, email, password)
	if err := row.Scan(&code); err != nil {
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

func Login(email, password string) (name string, err error) {
	rows, err := db.Query(
		`SELECT user_id FROM users WHERE email = $1 AND
		 password = crypt($2, password)`, email, password)
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

func AddColorFabric(name, color string) {
	if color[0] == '#' {
		color = color[1:]
	}
	// ignore errors on insert for colors
	db.Exec(`INSERT INTO user_fabrics(user_id,fabric_id)
	         VALUES($1,fabric_color($2))`, name, color)
}

func AddImageFabric(username, fabricname, url string) int {
	var imageId, fabricId int
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Commit()

	row := tx.QueryRow(`INSERT INTO images(user_id,url)
	                    VALUES($1,$2) RETURNING image_id`, username, url)
	if err := row.Scan(&imageId); err != nil {
		tx.Rollback()
		panic(err)
	}
	row = tx.QueryRow(`SELECT * FROM fabric_image($1,$2)`, imageId, fabricname)
	if err := row.Scan(&fabricId); err != nil {
		tx.Rollback()
		panic(err)
	}
	return fabricId
}

func LoadUser(name string) (User, error) {
	rows, err := db.Query(`SELECT 1 FROM users WHERE user_id = $1`, name)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	if rows.Next() {
		return &user{name: name}, nil
	}

	return nil, ErrNoUser
}

func (u *user) Name() string {
	return u.name
}

func (u *user) NumPublicQuilts() int {
	return u.numPublicQuilts
}

func (u *user) PublicQuilts() (quilts []*Quilt) {
	rows, err := db.Query(`
		SELECT quilt_id,name FROM quilts WHERE
		user_id = $1 AND visibility = 'public'`, u.name)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var id int
	var name string
	for rows.Next() {
		if err = rows.Scan(&id, &name); err != nil {
			panic(err)
		}
		quilts = append(quilts, &Quilt{Id: id, Name: name})
	}

	return quilts
}

func (u *user) Quilts() (quilts []*Quilt) {
	rows, err := db.Query(`SELECT quilt_id,name FROM quilts WHERE user_id = $1`, u.name)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var id int
	var name string
	for rows.Next() {
		if err = rows.Scan(&id, &name); err != nil {
			panic(err)
		}
		quilts = append(quilts, &Quilt{Id: id, Name: name})
	}

	return quilts
}

func (u *user) CreateQuilt(name, visibility string, width, height int) (*Quilt, error) {
	return createQuilt(u.name, name, visibility, width, height)
}
