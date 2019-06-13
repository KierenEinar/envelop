package constant

const (
	UserNotFoundErrorCode 	= 203 //user -> 200-300
	ConstantErrorCode     	= 500
	AccountCreateErrorCode 	= 701    //account -> 700-800
	AccountBalanceErrorCode = 705
	ParamErrorCode          = 101
	EnvelopCreateErrorCode  = 901 //红包创建失败
	BankBalanceErrorCode    = 1001 // 银行卡扣款失败
	EnvelopNotExistsErrorCode = 902 //红包不存在
	EnvelopTakePendingErrorCode = 961 //红包抢到了, 但是正在入库, 需要前端调用查询红包接口
	EnvelopRunDownErrorCode = 971 //红包抢完了
	EnvelopTakeAmountParseErrorCode = 981 //红包数值转换出错
	EnvelopExpireErrorCode = 904 //红包过期
	EnvelopItemCreateErrorCode  = 942 //红包订单创建失败
	EnvelopExistsErrorCode = 977 //红包存在
	EnvelopTakeRetry = 999 //重试抢红包
)


type RuntimeError struct {
	Code int
	Err string
}

func (this * RuntimeError) Error() string {
	return this.Err
}




