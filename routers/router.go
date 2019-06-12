// @APIVersion 1.0.0
// @Title beego Test API
// @Description beego has a very cool tools to autogenerate documents for your API
// @Contact astaxie@gmail.com
// @TermsOfServiceUrl http://beego.me/
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	"envelop/controllers"

	"github.com/astaxie/beego"
)

type Router struct {
	UserController *controllers.UserController `inject:""`
	AccountController *controllers.AccountController `inject:""`
	EnvelopController *controllers.EnvelopController `inject:""`
	ObjectController *controllers.ObjectController `inject:""`
}

func (this *Router) RegisterRouter() {
	ns := beego.NewNamespace("/api/v1",
		beego.NSNamespace("/object",
			beego.NSInclude(
				this.ObjectController,
			),
		),

		beego.NSNamespace("/users",
			beego.NSInclude(
				this.UserController,
			),
		),

		beego.NSNamespace("/accounts",
			beego.NSInclude(
				this.AccountController,
			),
		),


		beego.NSNamespace("/envelops",
			beego.NSInclude(
				this.EnvelopController,
			),
		),
	)
	beego.AddNamespace(ns)
}



