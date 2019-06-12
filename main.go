package main

import (
	"envelop/controllers"
	"envelop/dao"
	_ "envelop/infra/kafka"
	"envelop/routers"
	"envelop/service"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/facebookgo/inject"
	"log"
)

func main() {
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}

	logs.SetLogger(logs.AdapterConsole, `{"level":1, "color":true}`)

	beego.ErrorController(&controllers.ErrorController{})

	var g inject.Graph

	dao.MustInit(&g)

	service.MustInit(&g)

	controllers.MustInit(&g)

	r:= routers.MustInit(&g)

	if err:= g.Populate(); err!=nil {
		log.Fatal(err)
	}

	r.RegisterRouter()

	beego.Run()
}



