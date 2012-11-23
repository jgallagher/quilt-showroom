package models

import (
	"database/sql"
	"encoding/json"
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

type Poly interface {
	Coords() string
	Color() string
	Url() string
}

type ColorPoly struct {
	Id     int
	Coords [][]int
	Color  string
}

type ImagePoly struct {
	Id     int
	Coords [][]int
	Url    string
}

type Quilt struct {
	Id         int
	Name       string
	UserId     string
	Visibility string
	Width      int
	Height     int
	ColorPolys []*ColorPoly
	ImagePolys []*ImagePoly
}

type poly struct {
	coords string
	color  string
	url    string
}

type geoJson struct {
	Coordinates [][][]int `json:"coordinates"`
}

func (p *poly) Coords() string { return p.coords }
func (p *poly) Color() string  { return "#" + p.color }
func (p *poly) Url() string    { return p.url }

func (q *Quilt) PostComment(username, comment string) error {
	_, err := db.Exec(
		`INSERT INTO quilt_comments(user_id,quilt_id,comment) VALUES($1,$2,$3)`,
		username, q.Id, comment)
	return err
}

func (q *Quilt) Comments() (comments []Comment) {
	rows, err := db.Query(`
		SELECT user_id,comment,created FROM quilt_comments WHERE quilt_id=$1
		ORDER BY created DESC`, q.Id)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var c Comment
		if err = rows.Scan(&c.User, &c.Comment, &c.Timestamp); err != nil {
			panic(err)
		}
		comments = append(comments, c)
	}
	if rows.Err() != nil {
		panic(rows.Err())
	}
	return
}

func createQuilt(username, name, visibility string, width, height int) (*Quilt, error) {
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
		return &Quilt{Id: int(id.Int64), Name: name}, nil
	case "dup_name":
		return nil, ErrQuiltName
	}
	panic("unexpected code from quilt_create")
}

func LoadQuilt(id int) (*Quilt, error) {
	var coordsJson []byte

	q := &Quilt{Id: id}
	row := db.QueryRow(`
		SELECT user_id,name,visibility,width,height
		FROM quilts WHERE quilt_id=$1`, id)
	if err := row.Scan(&q.UserId, &q.Name, &q.Visibility, &q.Width, &q.Height); err != nil {
		return nil, err
	}

	// load quilt polygons that have color fabrics
	rows, err := db.Query(`
		SELECT quilt_poly_id,ST_AsGeoJSON(poly),color
		FROM quilt_polys NATURAL JOIN fabric_colors
		WHERE quilt_id = $1`, q.Id)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var p ColorPoly
		var coords geoJson
		if err := rows.Scan(&p.Id, &coordsJson, &p.Color); err != nil {
			panic(err)
		}
		if err := json.Unmarshal(coordsJson, &coords); err != nil {
			panic(err)
		}
		p.Coords = coords.Coordinates[0]
		q.ColorPolys = append(q.ColorPolys, &p)
	}

	// load quilt polygons that have image fabrics
	rows, err = db.Query(`
		SELECT quilt_poly_id,ST_AsGeoJSON(poly),url
		FROM quilt_polys NATURAL JOIN fabric_images NATURAL JOIN images
		WHERE quilt_id = $1`, q.Id)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var p ImagePoly
		var coords geoJson
		if err := rows.Scan(&p.Id, &coordsJson, &p.Url); err != nil {
			panic(err)
		}
		if err := json.Unmarshal(coordsJson, &coords); err != nil {
			panic(err)
		}
		p.Coords = coords.Coordinates[0]
		q.ImagePolys = append(q.ImagePolys, &p)
	}

	return q, nil
}
