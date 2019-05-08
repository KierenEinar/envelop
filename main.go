package main

import (
	"envelop/controllers"
	_ "envelop/routers"
	"github.com/astaxie/beego"
	_ "envelop/dao"
	"github.com/astaxie/beego/logs"
)

func main() {
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}

	logs.SetLogger(logs.AdapterConsole, `{"level":1, "color":true}`)

	beego.ErrorController(&controllers.ErrorController{})

	beego.Run()
}



