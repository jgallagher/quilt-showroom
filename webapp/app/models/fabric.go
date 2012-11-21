package models

type FabricColor struct {
	Color string
}

type FabricImage struct {
	Path string
	Name string
}

func LoadFabrics(username string) (color []FabricColor, image []FabricImage) {
	rows, err := db.Query(`
		SELECT color FROM fabric_colors NATURAL JOIN user_fabrics
		WHERE user_id = $1 ORDER BY color`, username)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var c FabricColor
		if err = rows.Scan(&c.Color); err != nil {
			panic(err)
		}
		color = append(color, c)
	}
	return
}
