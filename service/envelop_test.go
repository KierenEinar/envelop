package service_test

import (
	"envelop/controllers"
	"envelop/dao"
	"envelop/service"
	"github.com/astaxie/beego/logs"
	"github.com/facebookarchive/inject"
	"os"
	"testing"
)

var (
	g inject.Graph
)

func TestMain (m *testing.M) {

	dao.MustInit(&g)

	service.MustInit(&g)

	os.Exit(m.Run())

}

func TestEnvelopServiceImpl_QueryEnvelop(t *testing.T) {

	controller := controllers.EnvelopController{}

	g.Provide(&inject.Object{Value: &controller})

	g.Populate()

	service:= controller.EnvelopService

	logs.Info("service", service)

	item, err:= service.QueryEnvelop("1560425740444971000", 19 , 12)

	logs.Info("item", item, "err", err)

}

