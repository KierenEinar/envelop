package service

import (
	"database/sql"
	"encoding/json"
	"envelop/models"
	"github.com/astaxie/beego/logs"
)

type EnvelopTakeListener struct {
	AccountService *AccountServiceImpl	`inject:""`
	EnvelopService *EnvelopServiceImpl `inject:""`
}

func (this *EnvelopTakeListener) OnListening (topic string ,body string, err error) {

	logs.Info("OnListening... topic %s, body %s, err %v, AccountService %p", topic, body, err, this.AccountService)

	//抢红包入库

	//扣红包数量, 写入订单, 增加用户的金额

	if err != nil {
		logs.Error("抢红包转换数值错误, err %v", err)
		panic(err)
	}

	var takeEnvelopVo models.TakeEnvelopVo

	err = json.Unmarshal([]byte(body), &takeEnvelopVo)

	if err != nil {
		logs.Error("EnvelopTakeListener OnListening ... ", err.Error())
		return
	}

	err = this.AccountService.AccountDao.Tx(func(tx *sql.Tx) error {

		if err = this.AccountService.UpdateBalance(tx, takeEnvelopVo.InAccountHistory); err != nil {
			tx.Rollback()
			return err
		}

		item, err := this.EnvelopService.TakeEnvelopByUser(tx, takeEnvelopVo)

		if err != nil {
			tx.Rollback()
			return err
		}

		if err = tx.Commit(); err != nil {
			return err
		}

		err = this.EnvelopService.PutEnvelopOrderRedis(&takeEnvelopVo, item)
		if err != nil {
			this.retryPutEnvelopOrderRedis(10, &takeEnvelopVo, item)
		}


		return nil

	})

	if err != nil {
		logs.Error("消费者消费失败", err)
	}

}


func (this *EnvelopTakeListener) retryPutEnvelopOrderRedis(retryTime uint16, vo *models.TakeEnvelopVo, item *models.EnvelopItem) {
	for i:=uint16(0); i<retryTime; i++{
		err:= this.EnvelopService.PutEnvelopOrderRedis(vo, item)
		logs.Info("重试将红包订单写入redis, 第%d次, envelop_item id -> %d", item.Id)
		if err == nil {
			break
		}
	}
}