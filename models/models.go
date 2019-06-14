package models

import (
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/validation"
	"strconv"
	"time"
)

type Model struct {
	Id uint64 `gorm:"Column:id;PRIMARY_KEY;AUTO_INCREMENT"`
	CreateTime int64 `gorm:"Column:create_time"`
}


type User struct {
	Model
	NickName string `gorm:"Column:nick_name;"`
	Avater string `gorm:"Column:avater"`
	Sex string `gorm:"Column:sex; default:'male'"`
	LastName string `gorm:"Column:last_name; not_null"`
	FirstName string `gorm:"Column:first_name; not_null"`
	Age uint8 `gorm:"Column:age"`
}

func (*User) TableName() string {
	return "user"
}

type Account struct {
	Model
	UserId uint64
	Currency string `gorm:"not null;" valid:"Required"`
	Balance uint64 `gorm: "not null"`
	UpdateTime int64
	NickName string `gorm: "not null"`
	Version uint64
}

func (*Account) TableName() string {
	return "account"
}


type AccountHistory struct {
	Model
	UserId uint64 `valid:"Required"`
	AccountId uint64
	TradeNo string `gorm:"not null"`
	Type string `gorm: "not null" valid:"Required"`
	Description string `gorm: "not null"`
	Channel string `gorm: "not null" valid:"Required"`
	Currency string `gorm:"not null;" valid:"Required"`
	Amount int64 `valid:"Required;"`
	Pattern string `gorm: "not null"`
}



func (this * AccountHistory) Valid (v *validation.Validation) {

	if AccountHistoryTypeIn != this.Type && AccountHistoryTypeOut != this.Type {
		logs.Error("Type error")
		v.SetError("Type", "Type error")
	}

	if AccountHistoryCurrencyCNY != this.Currency {
		logs.Error("Currency error")
		v.SetError("Currency", "Currency error")
	}

	if AccountHistoryChannelUnionPay != this.Channel && AccountHistoryChannelPlat != this.Channel {
		logs.Error("Channel error")
		v.SetError("Channel", "Channel error")
	}

	if AccountHistoryPatternTransfer != this.Pattern && AccountHistoryPatternEnvelop != this.Pattern && AccountHistoryPatternRecharge != this.Pattern && AccountHistoryPatternWithdraw != this.Pattern {
		logs.Error("Pattern error")
		v.SetError("Pattern", "Pattern error")
	}

}

type Envelop struct {
	Model
	UserId uint64 `valid:"Required"`
	AccountId uint64
	Amount int64 `valid:"Required;"`
	Type string `enum: "AVG, RAND" valid:"Required"`
	Quantity uint8 `valid: "Required"`
	Version uint64
	PayChannel string `valid: "Required"`
	Currency string `valid:"Required"`
	TradeNo int64
	RemainingAmount uint64
}

type EnvelopItem struct {
	Model
	UserId uint64 `gorm:"column:user_id"`
	AccountId uint64 `gorm:"column:account_id"`
	EnvelopId uint64 `gorm:"column:envelop_id"`
	Amount int64 	`gorm:"column:amount"`
	Status string 	`gorm:"column:status"`
	TradeNo int64   `gorm:"column:trade_no"`
}



func (this *Envelop) Valid (v *validation.Validation) {
	if this.Amount <= 0 {
		logs.Error("amount error")
		v.SetError("Amount", "Amount error")
	}

	if this.PayChannel != AccountHistoryChannelUnionPay && this.PayChannel != AccountHistoryChannelPlat {
		logs.Error("PayChannel error")
		v.SetError("PayChannel", "PayChannel error")
	}

	if this.Quantity <= 0 {
		logs.Error("Quantity error")
		v.SetError("Quantity", "Quantity error")
	}

	if uint64(this.Amount) / uint64(this.Quantity) < 1 {
		logs.Error("Amount divide Quantity error")
		v.SetError("Amount / Quantity", "Amount / Quantity error")
	}

	if this.Type != EnvelopTypeAvg && this.Type != EnvelopTypeRand {
		logs.Error("Type error")
		v.SetError("Type", "Type error")
	}

}


type AccountBankTransferHistory struct {
	Model
	TradeNo string `gorm:"not null"`
	InAccountId uint64 `gorm: "not null"`
	OutAccountId uint64 `gorm: "not null"`
	BankNo string `gorm: "not null"`
	BankCode string `gorm: "not null"`
	BankName string `gorm: "not null"`
	Amount int64 `gorm: "not null"`
}



//使用货币
var (
	AccountHistoryCurrencyCNY = "CNY"
	AccountHistoryCurrencyUSD = "USD"
	AccountHistoryCurrencyHKD = "HKD"
	AccountHistoryCurrencyEUR = "EUR"
)

//支出还是收入
var (
	AccountHistoryTypeIn = "In"
	AccountHistoryTypeOut = "Out"
)

//支付方式
var (
	AccountHistoryChannelUnionPay = "UnionPay"
	AccountHistoryChannelPlat = "Plat"
)

//方式
var (
	AccountHistoryPatternTransfer = "Transfer" //转账
	AccountHistoryPatternEnvelop  = "Envelop"  //红包
	AccountHistoryPatternRecharge = "Recharge" //充值
	AccountHistoryPatternWithdraw = "Withdraw" //提现
)

//type accountHistoryDesc map[string]string


//转账描述
var (
	 accountHistoryDesc = map[string]string {
		 "AccountHistoryDescPlatInTransfer" : "平台-转账收入",
		 "AccountHistoryDescUnionPayInTransfer" : "银联-转账收入",
		 "AccountHistoryDescPlatOutTransfer" : "平台-转账支出",
		 "AccountHistoryDescUnionPayOutTransfer" : "银联-转账支出",
		 "AccountHistoryDescPlatInEnvelop" : "平台-红包收入",
		 "AccountHistoryDescUnionPayInEnvelop" : "银联-红包收入",
		 "AccountHistoryDescPlatOutEnvelop" : "平台-红包支出",
		 "AccountHistoryDescUnionPayOutEnvelop" : "银联-红包支出",
		 "AccountHistoryDescUnionPayInRecharge" : "银联-充值收入",
		 "AccountHistoryDescPlatOutWithdraw" : "平台-余额提现",
	 }
)

//红包类型
var (
	EnvelopTypeAvg = "AVG"
	EnvelopTypeRand = "RAND"
)


func (*AccountHistory) TableName() string {
	return "account_history"
}


func (*AccountHistory) GenTradeNo () string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}

func (this *AccountHistory) GenDescription () string {
	prefix := "AccountHistoryDesc"
	return accountHistoryDesc[prefix + this.Channel + this.Type + this.Pattern]
}
func (this *AccountHistory) GenCreateTime() {
	this.CreateTime = time.Now().Unix()
}

func (*Model) GenTradeNo () string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}