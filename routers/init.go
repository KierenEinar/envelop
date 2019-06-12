package routers

import (
	"github.com/facebookgo/inject"
)

func MustInit(g *inject.Graph) *Router {

	var r Router

	g.Provide(
		&inject.Object{Value: &r},
	)

	return &r
}
