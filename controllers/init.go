package controllers

import "github.com/facebookarchive/inject"

func MustInit (g *inject.Graph) {
	g.Provide(
		&inject.Object{Value: &AccountController{}},
		&inject.Object{Value: &EnvelopController{}},
		&inject.Object{Value: &UserController{}},
		&inject.Object{Value: &ObjectController{}},
		)
}
