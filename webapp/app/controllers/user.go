package controllers

import (
	"github.com/robfig/revel"
	"github.com/jgallagher/dbproject/webapp/app/models"
	"log"
)

type User struct {
	*rev.Controller
}

// User's public home page.
func (c User) PublicHome(username string) rev.Result {
	if username == c.Session["uname"] {
		c.RenderArgs["nav_home"] = true
	}
	user, err := models.LoadUser(username)
	if err != nil {
		return c.NotFound("Unknown user.")
	}

	quilts := user.PublicQuilts()
	return c.Render(user, quilts)
}

// User's quilts page.
func (c User) Quilts(username string) rev.Result {
	if invalid := c.checkUser(username); invalid != nil {
		return invalid
	}
	self := c.RenderArgs["self"].(models.User)
	quilts := self.Quilts()
	nav_quilts := true
	return c.Render(quilts, nav_quilts)
}

// User's fabrics page.
func (c User) Fabrics(username string) rev.Result {
	if invalid := c.checkUser(username); invalid != nil {
		return invalid
	}
	nav_fabrics := true
	return c.Render(nav_fabrics)
}

// User's blocks page.
func (c User) Blocks(username string) rev.Result {
	if invalid := c.checkUser(username); invalid != nil {
		return invalid
	}
	nav_blocks := true
	return c.Render(nav_blocks)
}

// Present form to create a new quilt.
func (c User) CreateQuilt(username string) rev.Result {
	if invalid := c.checkUser(username); invalid != nil {
		return invalid
	}
	for k, v := range c.Flash.Data {
		log.Printf("%s,%s", k, v)
	}
	return c.Render()
}

func (c User) HandleCreateQuilt(username, name, visibility string, width, height int) rev.Result {
	if invalid := c.checkUser(username); invalid != nil {
		return invalid
	}
	c.Validation.Required(name)
	c.Validation.Required(width)
	c.Validation.Min(width, 1)
	c.Validation.Required(height)
	c.Validation.Min(height, 1)

	if c.Validation.HasErrors() {
		c.Flash.Out["visibility" + visibility] = "selected"
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect("/users/%s/create-quilt", username)
	}

	self := c.RenderArgs["self"].(models.User)
	quilt, err := self.CreateQuilt(name, visibility, width, height)
	switch err {
	case nil:
		break
	case models.ErrQuiltName:
		c.Validation.Required(nil).Key("name").Message(err.Error())
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect("/users/%s/create-quilt", username)
	default:
		panic(err)
	}

	return c.Redirect("/quilts/%d", username, quilt.Id())
}

func (c User) checkUser(username string) rev.Result {
	if username != c.Session["uname"] {
		return c.Redirect(Accounts.Login)
	}
	return nil
}
