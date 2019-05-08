package service

import (
	"database/sql"
	"envelop/dao"
	"envelop/models"
)

type EnvelopService interface {
	CreateEnvelop(envelop * models.Envelop) error //发红包
}

var (
	accountService = new (AccountServiceImpl)
	envelopDao	   = new (dao.EnvelopDaoImpl)
)

type EnvelopServiceImpl struct {

}

func (this *EnvelopServiceImpl) CreateEnvelop(envelop * models.Envelop) error {


	return envelopDao.Tx(func(tx *sql.Tx) error {

		//accountId, err:= accountDao.FindIdByUserId(tx, envelop.UserId)
		//
		//if err != nil {
		//	return err
		//}
		//
		//envelop.AccountId = accountId
		//
		//rows, err := envelopDao.Create(tx, envelop)
		//
		//if err != nil || rows == 0{
		//	tx.Rollback()
		//	return err
		//}
		//
		//
		//err = accountService.UpdateBalance(tx, nil)
		//
		//if err != nil {
		//	tx.Rollback()
		//	return err
		//}
		//
		//tx.Commit()

		return nil

	})


}


func (this *EnvelopServiceImpl) accountHistory (envelop * models.Envelop) *models.AccountHistory {

	accountHistory:= new (models.AccountHistory)
	accountHistory.AccountId = envelop.AccountId
	accountHistory.UserId = envelop.UserId
	accountHistory.Type = models.AccountHistoryTypeOut
	accountHistory.Channel = envelop.PayChannel
	accountHistory.Currency = envelop.Currency
	accountHistory.Amount = envelop.Amount
	//accountHistory.TradeNo =
	return accountHistory
}

