package service_test

import (
	"envelop/dao"
	"envelop/service"
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

	g.Populate()

	os.Exit(m.Run())

}

func BenchmarkEnvelopServiceImpl_TakeEnvelopNew(b *testing.B) {

}

