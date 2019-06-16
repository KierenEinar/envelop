package dao_test

import (
	"envelop/dao"
	"envelop/models"
	"os"
	"testing"
)



func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestEnvelopItemDaoImpl_SelectByEnvelopIdAndUserId(t *testing.T) {

	envelopItemDao := &dao.EnvelopItemDaoImpl{}

	envelopItemDao.SelectByEnvelopIdAndUserId(&models.TakeEnvelopVo{
		UserId:11,
		EnvelopId:19,
	})



}