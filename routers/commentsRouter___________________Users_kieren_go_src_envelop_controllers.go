package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {

    beego.GlobalControllerRouter["envelop/controllers:AccountController"] = append(beego.GlobalControllerRouter["envelop/controllers:AccountController"],
        beego.ControllerComments{
            Method: "CreateOne",
            Router: `/`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["envelop/controllers:AccountController"] = append(beego.GlobalControllerRouter["envelop/controllers:AccountController"],
        beego.ControllerComments{
            Method: "UpdateBalanceByRecharge",
            Router: `/recharge`,
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["envelop/controllers:AccountController"] = append(beego.GlobalControllerRouter["envelop/controllers:AccountController"],
        beego.ControllerComments{
            Method: "UpdateBalanceByTransfer",
            Router: `/transfer`,
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["envelop/controllers:AccountController"] = append(beego.GlobalControllerRouter["envelop/controllers:AccountController"],
        beego.ControllerComments{
            Method: "UpdateBalanceByWithdraw",
            Router: `/withdraw`,
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["envelop/controllers:ObjectController"] = append(beego.GlobalControllerRouter["envelop/controllers:ObjectController"],
        beego.ControllerComments{
            Method: "Post",
            Router: `/`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["envelop/controllers:ObjectController"] = append(beego.GlobalControllerRouter["envelop/controllers:ObjectController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: `/`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["envelop/controllers:ObjectController"] = append(beego.GlobalControllerRouter["envelop/controllers:ObjectController"],
        beego.ControllerComments{
            Method: "Get",
            Router: `/:objectId`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["envelop/controllers:ObjectController"] = append(beego.GlobalControllerRouter["envelop/controllers:ObjectController"],
        beego.ControllerComments{
            Method: "Put",
            Router: `/:objectId`,
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["envelop/controllers:ObjectController"] = append(beego.GlobalControllerRouter["envelop/controllers:ObjectController"],
        beego.ControllerComments{
            Method: "Delete",
            Router: `/:objectId`,
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["envelop/controllers:UserController"] = append(beego.GlobalControllerRouter["envelop/controllers:UserController"],
        beego.ControllerComments{
            Method: "CreateOne",
            Router: `/`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["envelop/controllers:UserController"] = append(beego.GlobalControllerRouter["envelop/controllers:UserController"],
        beego.ControllerComments{
            Method: "FindOne",
            Router: `/:id`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

}
