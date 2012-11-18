package controllers

import (
	"github.com/robfig/revel"
)

type Accounts struct {
	*rev.Controller
}

type RenderUserInfoPlugin struct {
	rev.EmptyPlugin
}

func init() {
	rev.RegisterPlugin(RenderUserInfoPlugin{})
}

func (p RenderUserInfoPlugin) BeforeRequest(c *rev.Controller) {
	uname, ok := c.Session["uname"]
	if ok {
		c.RenderArgs["uname"] = uname
	}
}

func (c Accounts) Login() rev.Result {
	return c.Render()
}

func (c Accounts) HandleLogin(email, password string) rev.Result {
	name, err := db.Login(email, password)

	if err != nil {
		c.Validation.Required(nil)
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(Accounts.Login)
	}

	c.Session["uname"] = name
	return c.Redirect(Application.Index)
}

func (c Accounts) Logout() rev.Result {
	delete(c.Session, "uname")
	return c.Redirect(Application.Index)
}

func (c Accounts) Create() rev.Result {
	return c.Render()
}

func (c Accounts) HandleCreate(username, email,
	password, password_confirm string) rev.Result {

	c.Validation.Required(username)
	c.Validation.Required(email)
	c.Validation.Required(password)
	c.Validation.MinSize(password, 8)
	c.Validation.Required(password == password_confirm).Message("Your passwords do not match.")

	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(Accounts.Create)
	}

	err := db.CreateUser(username, email, password)

	switch err {
	case nil:
		break
	case ErrNameTaken, ErrBadName:
		c.Validation.Required(nil).Key("username").Message(err.Error())
	case ErrEmailTaken:
		c.Validation.Required(nil).Key("email").Message(err.Error())
	case ErrBadPass:
		c.Validation.Required(nil).Key("password").Message(err.Error())
	default:
		c.Validation.Required(nil).Key("general").Message(err.Error())
	}

	if err != nil {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(Accounts.Create)
	}

	c.Session["uname"] = username
	return c.Redirect(Application.Index)
}
