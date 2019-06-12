package service

import (
	"github.com/astaxie/beego/logs"
)

type EnvelopTakeListener struct {
	accountService *AccountServiceImpl	`inject:""`
	envelopService *EnvelopServiceImpl `inject:""`
}

func (this *EnvelopTakeListener) OnListening (topic string ,body string, err error) {

	logs.Info("OnListening... topic %s, body %s, err %v", topic, body, err)

	//抢红包入库

	//扣红包数量, 写入订单, 增加用户的金额





}