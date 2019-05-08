package controllers

type ErrorController struct {
	BaseController
}

func (this * ErrorController) Error500 () {
	this.Data["json"] = "error"
	this.ServeJSON()
}
