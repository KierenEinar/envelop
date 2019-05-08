package constant

const (
	UserNotFoundErrorCode 	= 203 //user -> 200-300
	ConstantErrorCode     	= 500
	AccountCreateErrorCode 	= 701    //account -> 700-800
	AccountBalanceErrorCode = 705
	ParamErrorCode          = 101
	EnvelopCreateErrorCode  = 901 //红包创建失败
	BankBalanceErrorCode    = 1001 // 银行卡扣款失败

)


type RuntimeError struct {
	Code int
	Err string
}

func (this * RuntimeError) Error() string {
	return this.Err
}




