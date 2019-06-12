package controllers

import (
	"encoding/json"
	"envelop/models"
	"envelop/service"
	"github.com/astaxie/beego/logs"
)

type AccountController struct {
	BaseController
	AccountService *service.AccountServiceImpl `inject:""`
}

// @router / [post]
func (this *AccountController) CreateOne () {
	var account* models.Account
	json.Unmarshal(this.Ctx.Input.RequestBody, &account)

	logs.Info("create account, ", string(this.Ctx.Input.RequestBody))
	res, error := this.AccountService.CreateAccount(account)
	this.apiResponse(0, error, res)
}

// @router /recharge [put]
//充值
func (this *AccountController) UpdateBalanceByRecharge () {
	var accountHistoryVO *models.AccountHistoryVO
	json.Unmarshal(this.Ctx.Input.RequestBody, &accountHistoryVO)
	logs.Info("create account_log, ", string(this.Ctx.Input.RequestBody))
	error := this.AccountService.UpdateBalanceByRecharge(accountHistoryVO)
	this.apiResponse(0, error, nil)
}



// @router /withdraw [put]
//提现
func (this *AccountController) UpdateBalanceByWithdraw () {
	var accountHistoryVO *models.AccountHistoryVO
	json.Unmarshal(this.Ctx.Input.RequestBody, &accountHistoryVO)
	logs.Info("create account_log, ", string(this.Ctx.Input.RequestBody))
	error :=this.AccountService.UpdateBalanceByWithdraw(accountHistoryVO)
	this.apiResponse(0, error, nil)
}


// @router /transfer [put]
//转账
func (this *AccountController) UpdateBalanceByTransfer () {
	var accountTransferVO *models.AccountTransferVO
	json.Unmarshal(this.Ctx.Input.RequestBody, &accountTransferVO)
	logs.Info("create account_log, ", string(this.Ctx.Input.RequestBody))
	error := this.AccountService.UpdateBalanceByTransfer(accountTransferVO)
	this.apiResponse(0, error, nil)
}