package models

import (
	"strconv"
	"time"
)

type AccountHistoryVO struct {
	AccountHistory
	BankNo string
	BankName string
	BankCode string
}

func (this *AccountHistoryVO) Build() AccountHistory {
	return this.AccountHistory
}


type AccountTransferVO struct {
	AccountHistoryVO
	InChannel string `valid:Required`
	InUserId uint64 `valid:"Required"`
	InAccountId uint64
	InBankCode string
	InBankName string
	InBankNo string
}


type TakeEnvelopVo struct {
	UserId uint64 `valid:"Required"`
	EnvelopTradeNo string `valid:"Required"`
	EnvelopId uint64 `valid:"Required"`
	OutAccountHistory *AccountHistory
	InAccountHistory *AccountHistory
}

type QueryEnvelopVo struct {
	EnvelopTradeNo string `valid:"Required"`
	EnvelopId uint64  `valid:"Required"`
	UserId uint64  `valid:"Required"`
}

type EnvelopDto struct {
	TakeEnvelopVo
	Amount int64
	CreateTime int64
	TradeNo string
}

func (*EnvelopDto) GenTradeNo () string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}


type EnvelopCreateVo struct {
	 EnvelopId uint64 `json:"envelop_id"`
	 TradeNo   string `json:"trade_no"`
}