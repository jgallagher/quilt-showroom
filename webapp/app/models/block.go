package models

import (
	"encoding/json"
)

type Poly struct {
	Id       int
	Coords   [][]int
	FabricId int
	Url      *string `json:"Url,omitempty"`
	Color    *string `json:"Color,omitempty"`
}

type Block struct {
	UserId string
	Id     int
	Name   string
	Width  int
	Height int
	Polys  []*Poly
}

func LoadBlocks(userid string) (blocks []*Block) {
	rows, err := db.Query(`
		SELECT block_id, name, width, height
		FROM blocks WHERE user_id = $1`, userid)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var b Block
		if err := rows.Scan(&b.Id, &b.Name, &b.Width, &b.Height); err != nil {
			panic(err)
		}
		blocks = append(blocks, &b)
	}
	return
}

func LoadBlock(id int) *Block {
	var coords []byte

	b := &Block{
		Id:    id,
		Polys: make([]*Poly, 0),
	}

	row := db.QueryRow(`
		SELECT user_id, name, width, height FROM blocks WHERE block_id=$1`, id)
	if err := row.Scan(&b.UserId, &b.Name, &b.Width, &b.Height); err != nil {
		panic(err)
	}

	// load block polygons
	rows, err := db.Query(`
		SELECT block_poly_id,ST_AsGeoJSON(poly),fabric_id
		FROM block_polys WHERE block_id = $1`, id)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var p Poly
		var g geoJson
		if err := rows.Scan(&p.Id, &coords, &p.FabricId); err != nil {
			panic(err)
		}
		if err := json.Unmarshal(coords, &g); err != nil {
			panic(err)
		}
		p.Coords = g.Coordinates[0]
		b.Polys = append(b.Polys, &p)
	}

	return b
}
