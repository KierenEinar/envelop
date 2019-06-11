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