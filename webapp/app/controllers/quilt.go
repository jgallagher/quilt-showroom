package controllers

import (
	"github.com/robfig/revel"
	"github.com/jgallagher/dbproject/webapp/app/models"
	"log"
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
	return c.Render()
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
