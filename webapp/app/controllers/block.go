package controllers

import (
	"github.com/robfig/revel"
	"github.com/jgallagher/dbproject/webapp/app/models"
)

type Block struct {
	*rev.Controller
}

func (c Block) BlockJson(id int) rev.Result {
	block := models.LoadBlock(id)
	if block.UserId != c.Session["uname"] {
		return c.NotFound("");
	}
	return c.RenderJson(block)
}
