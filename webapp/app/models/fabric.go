package models

type FabricColor struct {
	Id    int
	Color string
}

type FabricImage struct {
	Id   int
	Url  string
	Name string
}

func LoadFabrics(username string) (color []FabricColor, image []FabricImage) {
	rows, err := db.Query(`
		SELECT fabric_id,color FROM fabric_colors NATURAL JOIN user_fabrics
		WHERE user_id = $1 ORDER BY color`, username)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var c FabricColor
		if err = rows.Scan(&c.Id, &c.Color); err != nil {
			panic(err)
		}
		color = append(color, c)
	}
	rows, err = db.Query(`
		SELECT fabric_id,url,name
		FROM fabric_images NATURAL JOIN user_fabrics NATURAL JOIN images
		WHERE user_id = $1 ORDER BY name`, username)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var i FabricImage
		if err = rows.Scan(&i.Id, &i.Url, &i.Name); err != nil {
			panic(err)
		}
		image = append(image, i)
	}
	return
}
