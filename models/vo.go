package models

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