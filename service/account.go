package service

import (
	"database/sql"
	"encoding/json"
	"envelop/constant"
	"envelop/dao"
	"envelop/models"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/validation"
	"sync"
	"time"
)

var (
	once = sync.Once{}
	transferFactory *TransferFactory
)

type AccountService interface {
	CreateAccount (account *models.Account) (int64, error)
	UpdateBalanceByRecharge (accountHistory * models.AccountHistoryVO) (error) //充值->账户
	UpdateBalanceByWithdraw (accountHistory * models.AccountHistoryVO) (error) //账户->提现
	UpdateBalanceByTransfer (accountTransferVO * models.AccountTransferVO) (error) //转账->账户
	UpdateBalance (tx *sql.Tx ,accountHistory * models.AccountHistory) error //修改账户余额
	UpdateBalanceCallBankRpc (accountHistory * models.AccountHistoryVO) error //调银行rpc接口扣钱
	UpdateBankBalance (tx *sql.Tx, accountHistory *models.AccountHistoryVO, inAccountId uint64, outAccountId uint64) error
}

type AccountServiceImpl struct {
	UserDao *dao.UserDaoImpl `inject:""`
	AccountDao *dao.AccountDaoImpl `inject:""`
	AccountHistoryDao *dao.AccountHistoryDaoImpl `inject:""`
	AccountBankTransferHistoryDao *dao.AccountBankTransferHistoryDaoImpl `inject:""`
	TransferStrategyPlat2Plat *TransferStrategyPlat2Plat `inject:""`
	TransferStrategyPlat2UnionPay *TransferStrategyPlat2UnionPay `inject:""`
	TransferStrategyUnionPay2UnionPay *TransferStrategyUnionPay2UnionPay `inject:""`
	TransferStrategyUnionPay2Plat * TransferStrategyUnionPay2Plat `inject:""`
}

func (this *AccountServiceImpl) CreateAccount (account *models.Account) (int64, error) {
	json, err := json.Marshal(account)
	if err!=nil {
		logs.Error("json marshal err,", err)
		return 0, &constant.RuntimeError{constant.ConstantErrorCode, err.Error()}
	}
	logs.Info("create account ...", string(json))
	now := time.Now()
	account.CreateTime = now.Unix()
	account.UpdateTime = now.Unix()


	user, err := this.UserDao.FindUser(account.UserId)
	if err != nil {
		return 0, err
	}
	account.NickName = user.NickName
	rows, err := this.AccountDao.CreateAccount(account)
	if rows == 0 || err != nil {
		return rows, &constant.RuntimeError{constant.AccountCreateErrorCode, "account create failed ..."}
	}
	return rows, nil
}
//充值

func (this *AccountServiceImpl) UpdateBalanceByRecharge (accountHistoryVO * models.AccountHistoryVO) (error) {

	accountHistory := &accountHistoryVO.AccountHistory
	accountHistory.Type = models.AccountHistoryTypeIn
	accountHistory.Pattern = models.AccountHistoryPatternRecharge
	accountHistory.Channel = models.AccountHistoryChannelUnionPay
	accountHistory.CreateTime = time.Now().Unix()
	accountHistory.TradeNo = accountHistory.GenTradeNo()
	accountHistory.Description = accountHistory.GenDescription()


	err:= this.valid(accountHistoryVO)

	if err != nil {
		return err
	}

	return this.AccountDao.Tx(func(tx *sql.Tx) error {
		err:=this.UpdateBalance(tx, accountHistory)
		if err != nil {
			tx.Rollback()
			return err
		}

		err = this.updateBankBalanceByRecharge(tx, accountHistoryVO)
		if err != nil {
			tx.Rollback()
			return err
		}
		tx.Commit()
		return nil
	})
}

func (this *AccountServiceImpl) UpdateBalanceByWithdraw (accountHistoryVO * models.AccountHistoryVO) (error) {

	accountHistory := &accountHistoryVO.AccountHistory
	accountHistory.Type = models.AccountHistoryTypeOut
	accountHistory.Pattern = models.AccountHistoryPatternWithdraw
	accountHistory.Channel = models.AccountHistoryChannelPlat
	accountHistory.CreateTime = time.Now().Unix()
	accountHistory.TradeNo = accountHistory.GenTradeNo()
	accountHistory.Description = accountHistory.GenDescription()

	err:= this.valid(accountHistoryVO)

	if err != nil {
		return err
	}
	accountHistory.Amount = 0 - accountHistory.Amount
	return this.AccountDao.Tx(func(tx *sql.Tx) error {
		err:=this.UpdateBalance(tx, accountHistory)
		if err != nil {
			tx.Rollback()
			return err
		}

		err = this.updateBankBalanceByWithdraw(tx, accountHistoryVO)
		if err != nil {
			tx.Rollback()
			return err
		}
		tx.Commit()
		return nil
	})


	return nil
}

func (this *AccountServiceImpl) UpdateBalanceByTransfer (accountTransferVO * models.AccountTransferVO) (error) {

	accountHistory := &accountTransferVO.AccountHistory
	accountHistory.Type = models.AccountHistoryTypeOut
	accountHistory.Pattern = models.AccountHistoryPatternTransfer
	//accountHistory.Channel = models.AccountHistoryChannelPlat
	accountHistory.CreateTime = time.Now().Unix()
	accountHistory.TradeNo = accountHistory.GenTradeNo()
	accountHistory.Description = accountHistory.GenDescription()

	err:= this.validTransfer(accountTransferVO)

	if err != nil {
		return err
	}

	if (accountHistory.Channel == "") {
		return &constant.RuntimeError{
			constant.ParamErrorCode,
			"param error",
		}
	}

	key := accountTransferVO.Channel + "2" + accountTransferVO.InChannel

	instance:= this.transferFactoryInstance()

	strategy:= instance.GetStrategy(key)

	return strategy.Transfer(*accountTransferVO, this)


	//return accountDao.Tx(func(tx *sql.Tx) error {
	//	err:=this.updateBalance(tx, accountHistory)
	//	if err != nil {
	//		tx.Rollback()
	//		return err
	//	}
	//
	//	err = this.updateBankBalanceByTransfer(tx, accountHistoryVO)
	//	if err != nil {
	//		tx.Rollback()
	//		return err
	//	}
	//	tx.Commit()
	//	return nil
	//})
}

//修改账户余额
func (this *AccountServiceImpl) UpdateBalance (tx *sql.Tx ,accountHistory * models.AccountHistory) error {

	accountId, error:= this.AccountDao.FindIdByUserId (tx, accountHistory.UserId)

	if error != nil {
		return error
	}

	accountHistory.AccountId = accountId

	res, err:= this.AccountDao.UpdateAccountBalance(tx, accountHistory.UserId, accountHistory.Amount)

	logs.Info("update account, rows", res, ", err", err)
	if res == 0 || err!=nil {
		return &constant.RuntimeError{
			constant.AccountBalanceErrorCode, "update account balance failed ... "}
	}

	res, err = this.AccountHistoryDao.CreateAccountHistory(tx, accountHistory)
	logs.Info("insert account_history, rows", res, ", err", err)
	if res == 0 || err!=nil {
		return &constant.RuntimeError{
			constant.AccountBalanceErrorCode, "update account balance failed ... "}
	}
	return nil
}

/**发起银行转账rpc
*/
func (this *AccountServiceImpl) UpdateBalanceCallBankRpc (accountHistory * models.AccountHistoryVO) error {
	//假装转账成功
	return nil
}

func (this *AccountServiceImpl) UpdateBankBalance (tx *sql.Tx, accountHistory *models.AccountHistoryVO, inAccountId uint64, outAccountId uint64) error {

	accountBankTransferHistory := new (models.AccountBankTransferHistory)
	accountBankTransferHistory.Amount = accountHistory.Amount
	accountBankTransferHistory.CreateTime = accountHistory.CreateTime
	accountBankTransferHistory.BankCode = accountHistory.BankCode
	accountBankTransferHistory.TradeNo = accountHistory.TradeNo
	accountBankTransferHistory.BankNo = accountHistory.BankNo
	accountBankTransferHistory.InAccountId = inAccountId
	accountBankTransferHistory.OutAccountId = outAccountId
	accountBankTransferHistory.CreateTime = accountHistory.CreateTime
	accountBankTransferHistory.BankName = accountHistory.BankName
	rows, err := this.AccountBankTransferHistoryDao.Create(tx, accountBankTransferHistory)
	if rows == 0 || err != nil {
		return &constant.RuntimeError{
			constant.BankBalanceErrorCode,
			"bank balance not enough",
		}
	}

	err = this.UpdateBalanceCallBankRpc(accountHistory)

	if err != nil {
		return err
	}

	return nil


}


func (this *AccountServiceImpl) updateBankBalanceByRecharge(tx *sql.Tx, accountHistory *models.AccountHistoryVO) error {
	return this.UpdateBankBalance(tx, accountHistory, constant.SystemAdminAccountId,accountHistory.AccountId )
}

func (this *AccountServiceImpl) updateBankBalanceByWithdraw(tx *sql.Tx, accountHistory *models.AccountHistoryVO) error {
	accountHistory.Amount = 0 - accountHistory.Amount
	return this.UpdateBankBalance(tx, accountHistory, accountHistory.AccountId, constant.SystemAdminAccountId )
}

func (this *AccountServiceImpl) valid (accountHistoryVO *models.AccountHistoryVO) error {

	if accountHistoryVO.BankNo == "" || accountHistoryVO.BankCode == "" || accountHistoryVO.BankName == "" || accountHistoryVO.AccountHistory.Amount <=0 {
		return &constant.RuntimeError{
			constant.ParamErrorCode,
			"param error",
		}
	}


	valid := validation.Validation{}
	b, err:= valid.Valid(accountHistoryVO.AccountHistory)

	if err != nil || !b {
		logs.Error("validate error ,", err, "bool ,", b)
		return &constant.RuntimeError{
			constant.ParamErrorCode,
			"param error",
		}
	}

	return nil
}


func (this *AccountServiceImpl) validTransfer (vo *models.AccountTransferVO) error {

	valid := validation.Validation{}
	b, err:= valid.Valid(vo.AccountHistory)

	if err != nil || !b {
		logs.Error("validate error ,", err, "bool ,", b)
		return &constant.RuntimeError{
			constant.ParamErrorCode,
			"param error",
		}
	}

	if vo.Pattern != models.AccountHistoryPatternTransfer {
		return &constant.RuntimeError{
			constant.ParamErrorCode,
			"param error",
		}
	}

	if vo.Channel != models.AccountHistoryChannelPlat && vo.Channel != models.AccountHistoryChannelUnionPay {
		return &constant.RuntimeError{
			constant.ParamErrorCode,
			"param error",
		}
	}

	if vo.InChannel != models.AccountHistoryChannelPlat && vo.InChannel != models.AccountHistoryChannelUnionPay {
		return &constant.RuntimeError{
			constant.ParamErrorCode,
			"param error",
		}
	}


	key := vo.Channel + "2" + vo.InChannel

	instance:= this.transferFactoryInstance()

	strategy:= instance.GetStrategy(key)

	return strategy.Valid(*vo)

}

func (this *AccountServiceImpl) updateBankBalanceByTransfer(tx *sql.Tx, vo *models.AccountHistoryVO) error {

	accountHistory:=vo.AccountHistory
	accountHistory.Amount = 0 - vo.Amount
	accountHistory.Type = models.AccountHistoryTypeIn

	if (vo.Channel == models.AccountHistoryChannelPlat) {
		return this.UpdateBalance(tx, &accountHistory)
	} else {

		accountHistory.Amount = 0 - accountHistory.Amount
		return this.UpdateBankBalance(tx, vo, accountHistory.AccountId, constant.SystemAdminAccountId )

	}
}

type TransferStrategy interface {
	Valid(models.AccountTransferVO) error
	Transfer (vo models.AccountTransferVO, accountService AccountService) error
}

type TransferStrategyPlat2UnionPay struct {
	AccountDao *dao.AccountDaoImpl `inject:""`
}

func (this *TransferStrategyPlat2UnionPay) Valid(vo models.AccountTransferVO) error {
	//平台->银行转账
	if vo.Channel == models.AccountHistoryChannelPlat && vo.InChannel == models.AccountHistoryChannelUnionPay {
		if vo.InBankName == "" || vo.InBankCode == "" || vo.InBankNo == "" {
			return &constant.RuntimeError{
				constant.ParamErrorCode,
				"param error",
			}
		}
	}
	return nil
}

func (this *TransferStrategyPlat2UnionPay) Transfer(accountTransferVO models.AccountTransferVO, accountService AccountService) error {


	//平台->银行转账

	outAccountHistory:=new (models.AccountHistory)
	outAccountHistory.UserId = accountTransferVO.UserId
	outAccountHistory.Currency = accountTransferVO.Currency
	tradeNo := outAccountHistory.GenTradeNo()

	outAccountHistory.TradeNo = tradeNo
	createTime := time.Now().Unix()
	outAccountHistory.CreateTime = createTime
	outAccountHistory.Amount = 0 - accountTransferVO.Amount
	outAccountHistory.Channel = models.AccountHistoryChannelPlat
	outAccountHistory.Pattern = models.AccountHistoryPatternTransfer
	outAccountHistory.Type = models.AccountHistoryTypeOut
	outAccountHistory.Description = outAccountHistory.GenDescription()



	inAccountHistoery:=new (models.AccountHistoryVO)
	inAccountHistoery.Type = models.AccountHistoryTypeIn
	inAccountHistoery.Pattern = models.AccountHistoryPatternTransfer
	inAccountHistoery.Amount = accountTransferVO.Amount
	inAccountHistoery.Channel = models.AccountHistoryChannelUnionPay
	inAccountHistoery.CreateTime = createTime
	inAccountHistoery.TradeNo = tradeNo
	inAccountHistoery.Currency = accountTransferVO.Currency
	inAccountHistoery.UserId = accountTransferVO.InUserId
	inAccountHistoery.Description = inAccountHistoery.GenDescription()
	inAccountHistoery.BankName = accountTransferVO.InBankName
	inAccountHistoery.BankCode = accountTransferVO.InBankCode
	inAccountHistoery.BankNo = accountTransferVO.InBankNo


	return this.AccountDao.Tx(func(tx *sql.Tx) error {
		err := accountService.UpdateBalance(tx, outAccountHistory)
		if err != nil {
			tx.Rollback()
			return nil
		}

		inAccountId, err := this.AccountDao.FindIdByUserId(tx, inAccountHistoery.UserId)

		if err != nil {
			tx.Rollback()
			return nil
		}

		err = accountService.UpdateBankBalance(tx, inAccountHistoery, inAccountId, constant.SystemAdminAccountId)

		if err != nil {
			tx.Rollback()
			return nil
		}

		tx.Commit()
		return nil

	})

}

type TransferStrategyPlat2Plat struct {
	AccountDao *dao.AccountDaoImpl `inject:""`
}

func (this *TransferStrategyPlat2Plat) Valid (vo models.AccountTransferVO) error {
	//平台->平台转账
	return nil
}

//平台间的转账
func (this *TransferStrategyPlat2Plat) Transfer(accountTransferVO models.AccountTransferVO, accountService AccountService) error {

	outAccountHistory := new (models.AccountHistory)
	outAccountHistory.Channel = models.AccountHistoryChannelPlat
	outAccountHistory.Pattern = models.AccountHistoryPatternTransfer
	outAccountHistory.Type = models.AccountHistoryTypeOut
	outAccountHistory.Amount = 0 - accountTransferVO.Amount
	now := time.Now().Unix()
	outAccountHistory.CreateTime = now
	tradeNo:=outAccountHistory.GenTradeNo()
	outAccountHistory.Description = outAccountHistory.GenDescription()
	outAccountHistory.TradeNo = tradeNo
	outAccountHistory.Currency = accountTransferVO.Currency
	outAccountHistory.UserId = accountTransferVO.UserId


	inAccountHistory := new (models.AccountHistory)
	inAccountHistory.Channel = models.AccountHistoryChannelPlat
	inAccountHistory.Pattern = models.AccountHistoryPatternTransfer
	inAccountHistory.Type = models.AccountHistoryTypeIn
	inAccountHistory.Amount = accountTransferVO.Amount
	inAccountHistory.CreateTime = now
	inAccountHistory.Description = inAccountHistory.GenDescription()
	inAccountHistory.TradeNo = tradeNo
	inAccountHistory.Currency = accountTransferVO.Currency
	inAccountHistory.UserId = accountTransferVO.InUserId


	return this.AccountDao.Tx(func(tx *sql.Tx) error {
		err:= accountService.UpdateBalance(tx, outAccountHistory)
		if err != nil {
			tx.Rollback()
			return err
		}

		err = accountService.UpdateBalance(tx, inAccountHistory)
		if err != nil {
			tx.Rollback()
			return err
		}

		err = tx.Commit()

		return err
	})

}

type TransferStrategyUnionPay2Plat struct {
	AccountDao *dao.AccountDaoImpl `inject:""`
}

func (this *TransferStrategyUnionPay2Plat) Valid (vo models.AccountTransferVO) error {

	//银行->平台转账
	if vo.Channel == models.AccountHistoryChannelUnionPay && vo.InChannel == models.AccountHistoryChannelPlat {
		if vo.BankName == "" || vo.BankCode == "" || vo.BankNo == "" {
			return &constant.RuntimeError{
				constant.ParamErrorCode,
				"param error",
			}
		}
	}
	return nil
}

func (this *TransferStrategyUnionPay2Plat) Transfer(accountTransferVO models.AccountTransferVO, accountService AccountService) error {

	inAccountHistory := new (models.AccountHistory)
	inAccountHistory.UserId = accountTransferVO.InUserId
	now := time.Now().Unix()
	inAccountHistory.CreateTime = now
	inAccountHistory.Amount = accountTransferVO.Amount
	inAccountHistory.Channel = models.AccountHistoryChannelPlat
	inAccountHistory.Type = models.AccountHistoryTypeIn
	inAccountHistory.Currency = accountTransferVO.Currency
	inAccountHistory.Pattern = models.AccountHistoryPatternTransfer
	tradeNo:=inAccountHistory.GenTradeNo()
	inAccountHistory.TradeNo = tradeNo
	inAccountHistory.Description = inAccountHistory.GenDescription()



	outAccountHistory:= new (models.AccountHistoryVO)
	outAccountHistory.UserId = accountTransferVO.UserId

	outAccountHistory.TradeNo = tradeNo
	outAccountHistory.Currency = accountTransferVO.Currency
	outAccountHistory.Pattern = models.AccountHistoryPatternTransfer
	outAccountHistory.Type = models.AccountHistoryTypeOut
	outAccountHistory.Channel = models.AccountHistoryChannelUnionPay
	outAccountHistory.CreateTime = now
	outAccountHistory.Amount = accountTransferVO.Amount
	outAccountHistory.BankNo = accountTransferVO.BankNo
	outAccountHistory.BankCode = accountTransferVO.BankCode
	outAccountHistory.BankName = accountTransferVO.BankName

	outAccountHistory.Description = outAccountHistory.GenDescription()
	return this.AccountDao.Tx(func(tx *sql.Tx) error {
		 err := accountService.UpdateBalance(tx, inAccountHistory)
		 if err != nil {
		 	tx.Rollback()
		 	return err
		 }


		outAccountId, err := this.AccountDao.FindIdByUserId(tx, outAccountHistory.UserId)

		if err != nil {
			tx.Rollback()
			return err
		}

		err = accountService.UpdateBankBalance(tx, outAccountHistory, constant.SystemAdminAccountId, outAccountId)

		if err != nil {
			tx.Rollback()
			return err
		}

		tx.Commit()
		return nil

	})

}


type TransferStrategyUnionPay2UnionPay struct {}

func (this *TransferStrategyUnionPay2UnionPay) Valid (vo models.AccountTransferVO) error {

	//银行->银行转账
	if vo.Channel == models.AccountHistoryChannelUnionPay && vo.InChannel == models.AccountHistoryChannelUnionPay {
		if vo.InBankName == "" || vo.InBankCode == "" {
			return &constant.RuntimeError{
				constant.ParamErrorCode,
				"param error",
			}
		}
	}
	return nil
}

func (this *TransferStrategyUnionPay2UnionPay) Transfer(accountTransferVO models.AccountTransferVO, accountService AccountService) error {
	return nil
}


type TransferFactory struct {
	strategy map[string]TransferStrategy
}

func (this *TransferFactory) GetStrategy (name string) TransferStrategy {
	return this.strategy[name]
}


func (this *AccountServiceImpl) transferFactoryInstance () *TransferFactory  {
	once.Do(func() {
		transferFactory = new(TransferFactory)
		transferFactory.strategy = make(map[string]TransferStrategy)
		transferFactory.strategy["Plat2Plat"] = this.TransferStrategyPlat2Plat
		transferFactory.strategy["Plat2UnionPay"] = this.TransferStrategyPlat2UnionPay
		transferFactory.strategy["UnionPay2Plat"] = this.TransferStrategyUnionPay2Plat
		transferFactory.strategy["UnionPay2UnionPay"] = this.TransferStrategyUnionPay2UnionPay

	})
	return transferFactory
}



























