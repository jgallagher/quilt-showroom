package controllers

import (
	"github.com/jgallagher/dbproject/webapp/app/models"
	"github.com/robfig/revel"
)

type Application struct {
	*rev.Controller
}

func (c Application) Index() rev.Result {
	users := models.AllUsers()
	return c.Render(users)
}
