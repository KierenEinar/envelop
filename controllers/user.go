package controllers

import (
	"encoding/json"
	"envelop/constant"
	"envelop/models"
	"envelop/service"
	"github.com/astaxie/beego"
	"reflect"
)

var (
	userServie = new (service.UserServiceImpl)
)


type BaseController struct {
	beego.Controller
}

type APIResponse struct {
	Code int
	Error  string
	Data interface{}
}


func (this *BaseController) apiResponse(code int, err error, data interface{} )  {
	var errStr string
	if err != nil {
		errStr = err.Error()
		code = constant.ConstantErrorCode
		errType:= reflect.TypeOf(err).Elem()
		if errType.String() == "constant.RuntimeError" {
			code = err.(*constant.RuntimeError).Code
		}
	}

	response := &APIResponse{code,  errStr , data}
	this.Data["json"] = response
	this.ServeJSON()
}


type UserController struct {
	BaseController
}

// @router / [post]
func (this *UserController) CreateOne () {
	var user* models.User
	json.Unmarshal(this.Ctx.Input.RequestBody, &user)
	result, err := userServie.CreateUser(user)
	this.apiResponse(0, err, result)
}

// @router /:id [get]
func (this *UserController) FindOne () {
	id, err := this.GetInt(":id")
	data, err := userServie.FindOne (uint64(id))
	this.apiResponse(0, err, data)
}