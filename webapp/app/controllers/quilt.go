package controllers

import (
	"encoding/json"
	"github.com/robfig/revel"
	"github.com/jgallagher/dbproject/webapp/app/models"
	"log"
	"strconv"
)

type Quilt struct {
	*rev.Controller
}

func (c Quilt) PublicQuilt(id int) rev.Result {
	quilt, err := models.LoadQuilt(id)
	if err != nil {
		panic(err)
	}
	c.RenderArgs["quilt"] = quilt
	if quilt.UserId == c.Session["uname"] {
		return c.Quilt(quilt)
	}
	if quilt.Visibility == "private" {
		return c.NotFound("Quilt not visible.")
	}
	return c.Render()
}

func (c Quilt) QuiltJson(id int) rev.Result {
	quilt, err := models.LoadQuilt(id)
	if err != nil {
		panic(err)
	}
	if quilt.Visibility == "private" && quilt.UserId != c.Session["uname"] {
		return c.NotFound("Quilt not visible.")
	}
	return c.RenderJson(quilt)
}

func (c Quilt) Quilt(quilt *models.Quilt) rev.Result {
	color_fabrics, image_fabrics := models.LoadFabrics(quilt.UserId)
	blocks := models.LoadBlocks(quilt.UserId)
	return c.Render(color_fabrics, image_fabrics, blocks)
}

func (c Quilt) Comment(id int, comment string) rev.Result {
	username, ok := c.Session["uname"]
	if !ok {
		return c.Redirect(Accounts.Login)
	}
	quilt, err := models.LoadQuilt(id)
	if err != nil {
		panic(err)
	}
	if err := quilt.PostComment(username, comment); err != nil {
		panic(err)
	}
	return c.Redirect("/quilts/%d", id)
}

func (c Quilt) PolyDelete(id, polyid int) rev.Result {
	if !models.QuiltOwner(c.Session["uname"], id) {
		return c.NotFound("Action not allowed.")
	}
	log.Printf("deleting poly %d from quilt %d", polyid, id)
	models.DeletePoly(polyid)
	return c.RenderJson("ok")
}

func (c Quilt) PolyAdd(id, x, y int, polyjson string) rev.Result {
	if !models.QuiltOwner(c.Session["uname"], id) {
		return c.NotFound("Action not allowed.")
	}
	log.Printf("adding polys %s at %d,%d to quilt %d", polyjson, x, y, id)
	var polys []*models.ColorPoly
	if err := json.Unmarshal([]byte(polyjson), &polys); err != nil {
		panic(err)
	}
	log.Printf("unmarshalled into %q", polys)
	if err := models.AddPolys(id, x, y, polys); err != nil {
		panic(err)
	}
	log.Printf("returning polys %q", polys)
	return c.RenderJson(polys)
}

func (c Quilt) PolyAddWithFabric(id, x, y int, polyjson string) rev.Result {
	if !models.QuiltOwner(c.Session["uname"], id) {
		return c.NotFound("Action not allowed.")
	}
	log.Printf("adding polys with fabric %s at %d,%d to quilt %d", polyjson, x, y, id)
	var polys []*models.Poly
	if err := json.Unmarshal([]byte(polyjson), &polys); err != nil {
		panic(err)
	}
	log.Printf("unmarshalled into %q", polys)
	if err := models.AddPolysWithFabric(id, x, y, polys); err != nil {
		panic(err)
	}
	log.Printf("returning polys %q", polys)
	return c.RenderJson(polys)
}

func (c Quilt) PolySetFabric(id, polyid, fabricid int) rev.Result {
	if !models.QuiltOwner(c.Session["uname"], id) {
		return c.NotFound("Action not allowed.")
	}
	log.Printf("set poly %d fabric to %d", polyid, fabricid)
	models.SetPolyFabric(polyid, fabricid)
	return c.RenderJson("ok")
}

func (c Quilt) UploadImage(id int, comment, url string) rev.Result {
	if !models.QuiltOwner(c.Session["uname"], id) {
		return c.NotFound("Action not allowed.")
	}
	log.Printf("attach image (%s,%s) to quilt %d", comment, url, id)
	models.AddQuiltImage(id, comment, url)
	return c.RenderJson("ok")
}

func (c Quilt) CreateBlock(id int, name string) rev.Result {
	var polyid []int
	for _, p := range c.Params.Values["polyid"] {
		val, err := strconv.Atoi(p)
		if err != nil {
			panic(err)
		}
		polyid = append(polyid, val)
	}
	log.Printf("create block %s from quilt %d with poly ids %s", name, id, polyid)
	models.CreateBlockFromPolys(id, name, polyid)
	return c.Redirect("/quilts/%d", id)
}
