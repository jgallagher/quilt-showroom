package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"strings"
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

type geoJson struct {
	Type        string    `json:"type"`
	Coordinates [][][]int `json:"coordinates"`
}

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

func loadColorPoly(rows *sql.Rows) *ColorPoly {
	var coordsJson []byte
	var p ColorPoly
	var coords geoJson
	if err := rows.Scan(&p.Id, &coordsJson, &p.Color); err != nil {
		panic(err)
	}
	if err := json.Unmarshal(coordsJson, &coords); err != nil {
		panic(err)
	}
	p.Coords = coords.Coordinates[0]
	return &p
}

func loadImagePoly(rows *sql.Rows) *ImagePoly {
	var coordsJson []byte
	var p ImagePoly
	var coords geoJson
	if err := rows.Scan(&p.Id, &coordsJson, &p.Url); err != nil {
		panic(err)
	}
	if err := json.Unmarshal(coordsJson, &coords); err != nil {
		panic(err)
	}
	p.Coords = coords.Coordinates[0]
	return &p
}

func LoadQuilt(id int) (*Quilt, error) {
	q := &Quilt{
		Id:         id,
		ColorPolys: make([]*ColorPoly, 0),
		ImagePolys: make([]*ImagePoly, 0),
	}
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
		q.ColorPolys = append(q.ColorPolys, loadColorPoly(rows))
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
		q.ImagePolys = append(q.ImagePolys, loadImagePoly(rows))
	}

	return q, nil
}

func QuiltOwner(userid string, quiltid int) bool {
	rows, err := db.Query(`SELECT 1 FROM quilts WHERE user_id=$1 AND quilt_id=$2`,
		userid, quiltid)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	if rows.Next() {
		return true
	}

	return false
}

func DeletePoly(id int) {
	if _, err := db.Exec(`DELETE FROM quilt_polys WHERE quilt_poly_id=$1`, id); err != nil {
		panic(err)
	}
}

func AddPolysWithFabric(quiltid, x, y int, polys []*Poly) error {
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	insertStmt, err := tx.Prepare(`
		INSERT INTO quilt_polys(quilt_id, fabric_id, poly)
		VALUES ($1, $2, ST_Translate(ST_GeomFromGeoJSON($3), $4, $5))
		RETURNING quilt_poly_id, ST_AsGeoJSON(poly)`)
	if err != nil {
		tx.Rollback()
		panic(err)
	}
	fabricStmt, err := tx.Prepare(`
		SELECT color, url
		FROM
			fabrics AS f
			LEFT JOIN fabric_colors AS fc ON (f.fabric_id = fc.fabric_id)
			LEFT JOIN fabric_images AS fi ON (f.fabric_id = fi.fabric_id)
			LEFT JOIN images AS i         ON (fi.image_id = i.image_id)
		WHERE f.fabric_id = $1`)
	if err != nil {
		tx.Rollback()
		panic(err)
	}

	for _, p := range polys {
		var coordsJson []byte
		g := geoJson{"Polygon", make([][][]int, 1)}
		g.Coordinates[0] = p.Coords
		s, err := json.Marshal(g)
		if err != nil {
			tx.Rollback()
			panic(err)
		}

		row := insertStmt.QueryRow(quiltid, p.FabricId, string(s), x, y)
		if err := row.Scan(&p.Id, &coordsJson); err != nil {
			tx.Rollback()
			panic(err)
		}
		if err := json.Unmarshal(coordsJson, &g); err != nil {
			tx.Rollback()
			panic(err)
		}
		p.Coords = g.Coordinates[0]

		row = fabricStmt.QueryRow(p.FabricId)
		if err := row.Scan(&p.Color, &p.Url); err != nil {
			tx.Rollback()
			panic(err)
		}
	}

	return tx.Commit()
}

func AddPolys(quiltid, x, y int, polys []*ColorPoly) error {
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	stmt, err := tx.Prepare(`
		INSERT INTO quilt_polys(quilt_id, fabric_id, poly)
		VALUES ($1, fabric_color('ffffff'),
			ST_Translate(ST_GeomFromGeoJson($2), $3, $4))
		RETURNING quilt_poly_id, 'ffffff', ST_AsGeoJSON(poly)`)
	if err != nil {
		tx.Rollback()
		panic(err)
	}

	// convert polys into geoJson values
	for _, p := range polys {
		var coordsJson []byte
		g := geoJson{"Polygon", make([][][]int, 1)}
		g.Coordinates[0] = p.Coords
		s, err := json.Marshal(g)
		if err != nil {
			tx.Rollback()
			panic(err)
		}

		log.Printf("marshalled into %s", string(s))
		rows, err := stmt.Query(quiltid, string(s), x, y)
		log.Printf("finished issuing query")
		if err != nil {
			log.Printf("query failed: %s", err)
			tx.Rollback()
			panic(err)
		}
		log.Printf("checking rows.Next")
		if rows.Next() {
			log.Printf("scanning...")
			if err := rows.Scan(&p.Id, &p.Color, &coordsJson); err != nil {
				tx.Rollback()
				panic(err)
			}
			if err := json.Unmarshal(coordsJson, &g); err != nil {
				tx.Rollback()
				panic(err)
			}
			p.Coords = g.Coordinates[0]
			log.Printf("set coords = %v", p.Coords)
		} else {
			log.Printf("did not get row; err = %s", rows.Err())
			tx.Rollback()
			panic(rows.Err())
		}
		rows.Close()
	}

	return tx.Commit()
}

func SetPolyFabric(polyid, fabricid int) {
	if _, err := db.Exec(`UPDATE quilt_polys SET fabric_id=$1 WHERE quilt_poly_id=$2`,
		fabricid, polyid); err != nil {
		panic(err)
	}
}

func CreateBlockFromPolys(quiltid int, name string, polyid []int) {
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Commit()

	// build up string of poly ids used for IN clause below
	polyidStrings := make([]string, len(polyid))
	for i, id := range polyid {
		polyidStrings[i] = strconv.Itoa(id)
	}
	polyidString := strings.Join(polyidStrings, ",")
	log.Printf("polyidString = %s", polyidString)

	// compute bounding box of polys
	row := tx.QueryRow(`
		SELECT
			MIN(ST_XMin(poly)), MAX(ST_XMax(poly)),
			MIN(ST_YMin(poly)), MAX(ST_YMax(poly))
		FROM quilt_polys WHERE quilt_poly_id IN (` + polyidString + `)`)
	var xmin, xmax, ymin, ymax int
	if err := row.Scan(&xmin, &xmax, &ymin, &ymax); err != nil {
		tx.Rollback()
		panic(err)
	}

	log.Printf("computed bbox %d,%d  %d,%d", xmin, xmax, ymin, ymax)

	// create the block
	row = tx.QueryRow(`
		INSERT INTO blocks(user_id,name,width,height)
		VALUES (
			(SELECT user_id FROM quilts WHERE quilt_id=$1),
			$2, $3, $4)
		RETURNING block_id`, quiltid, name, xmax-xmin, ymax-ymin)
	var blockid int
	if err := row.Scan(&blockid); err != nil {
		tx.Rollback()
		panic(err)
	}

	// insert all the translated quilt polys into the block
	if _, err := tx.Exec(`
		WITH src AS (
			SELECT fabric_id,poly FROM quilt_polys
			WHERE quilt_poly_id IN (`+polyidString+`))
		INSERT INTO block_polys(block_id,fabric_id,poly)
		(SELECT $1, fabric_id, ST_Translate(poly, $2, $3) FROM src)`,
		blockid, -xmin, -ymin); err != nil {
		panic(err)
	}
}
