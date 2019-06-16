package service

import (
	"database/sql"
	"encoding/json"
	"envelop/constant"
	"envelop/dao"
	"envelop/infra/algo"
	"envelop/infra/kafka"
	"envelop/infra/util"
	"envelop/models"
	redisClient "envelop/redis"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/validation"
	"github.com/go-redis/redis"
	"strconv"
	"time"
)

const (
	ENVELOPTAKEPENDING = "pending"
	ENVELOPTAKEFAILED = "take failed"
)

type EnvelopService interface {
	CreateEnvelop (* models.Envelop) (*models.EnvelopCreateVo, error) //发红包
	TakeEnvelop (models.TakeEnvelopVo) (*models.EnvelopDto, error) //抢红包
	TakeEnvelopNew (*models.TakeEnvelopVo) (*models.EnvelopItem, error) //抢红包, 新版本
	QueryEnvelop (*models.QueryEnvelopVo) (*models.EnvelopItem, error) //查询红包
}

var (
	envelopRandomDoubleStrategy = algo.EnvelopDoubleAvgStrategy{}
)

type EnvelopServiceImpl struct {
	AccountService *AccountServiceImpl `inject:""`
	EnvelopDao *dao.EnvelopDaoImpl `inject:""`
	EnvelopItemDao *dao.EnvelopItemDaoImpl `inject:""`
	AsyncKafkaProducer *kafka.AsyncKafkaProducer `inject:""`
}

func (this *EnvelopServiceImpl) CreateEnvelop(envelop * models.Envelop) (*models.EnvelopCreateVo ,error) {

	/**
	(1)如果是支付方式是平台，则扣款
		否则调用银行接口扣款
	(2)写入红包记录
	(3)把红包子订单写入redis cluster
	(4)
	*/

	valid := validation.Validation{}
	b, err:= valid.Valid(envelop)

	if err != nil || !b {
		logs.Error("validate error ,", err, "bool ,", b)
		return &models.EnvelopCreateVo{
			0,
			"",
		}, constant.ParamError
	}

	var key string

	var outAccountHistory models.AccountHistory

	err = this.EnvelopDao.Tx(func(tx *sql.Tx) error {

		if envelop.PayChannel == models.AccountHistoryChannelPlat {
			outAccountHistory.Amount = (int64)(0 - envelop.Amount)
			outAccountHistory.UserId = envelop.UserId
			outAccountHistory.GenCreateTime()
			outAccountHistory.Type = models.AccountHistoryTypeOut
			outAccountHistory.Channel = models.AccountHistoryChannelPlat
			outAccountHistory.Pattern = models.AccountHistoryPatternEnvelop
			outAccountHistory.Currency = envelop.Currency
			tradeNo := outAccountHistory.GenTradeNo()
			outAccountHistory.TradeNo = tradeNo
			outAccountHistory.Description = outAccountHistory.GenDescription()
			err := this.AccountService.UpdateBalance(tx, &outAccountHistory)
			if err != nil {
				tx.Rollback()
				return err
			}

			key = tradeNo
			envelop.AccountId = outAccountHistory.AccountId
			envelop.CreateTime = outAccountHistory.CreateTime
			t, _ := strconv.ParseInt(key, 10, 64)
			envelop.TradeNo = t
			envelop.RemainingAmount = uint64(envelop.Amount)
		}

		err:= this.EnvelopDao.Create(tx, envelop)

		if err != nil{
			tx.Rollback()
			return err
		}

		seeds := make([]interface{}, 0)

		envelopRandomDoubleStrategy.Generate(int64(envelop.Quantity), envelop.Amount, &seeds)

		oneDay := 24 * time.Hour

		b, err := json.Marshal(outAccountHistory)

		if err !=nil {
			return constant.ServerError
		}
		k := this.envelopOrderkey(key, envelop.Id)

		flag, err := redisClient.Client.SetNX(k, string(b), oneDay).Result()

		if err != nil {
			return constant.EnvelopCreateError
		}

		if flag == false {
			return constant.EnvelopExistsError
		}

		res, err:= redisClient.Client.SAdd(this.envelopSetKeyShard(key), envelop.Id).Result()

		if err != nil || res == 0 {
			return constant.EnvelopCreateError
		}

		res, err = redisClient.Client.LPush(key, seeds...).Result()


		if err != nil || res == 0{
			tx.Rollback()
			return err
		}

		tx.Commit()

		return nil

	})


	if err != nil {
		return&models.EnvelopCreateVo{
			0,
			"",
		}, err
	}


	return &models.EnvelopCreateVo{
		envelop.Id,
		key,
	}, err


}


func (this *EnvelopServiceImpl) envelopOrderkey (envelopTradeNo string, envelopId uint64) string {
	return fmt.Sprintf("envelop::%s::%d", envelopTradeNo, envelopId)
}

func (this *EnvelopServiceImpl) envelopSetKeyShard(envelopTradeNo string) string {
	hash := util.HashCode(&envelopTradeNo)
	return fmt.Sprintf("envelop::exists::%d", hash % 100)
}


func (this *EnvelopServiceImpl) TakeEnvelopNew (takeEnvelopVo *models.TakeEnvelopVo) (*models.EnvelopItem, error) {

	//从set中判断红包是否存在

	exists, err := this.isEnvelopExists(takeEnvelopVo)

	if exists == false || err != nil {
		return nil, constant.EnvelopNotExists
	}

	outAmoutHistory, err := this.isEnvelopExpire(takeEnvelopVo)

	if err != nil {
		return nil, err
	}


	//	查找秒杀的key, envelop::${envelopId}::${uid}::take, 如果key存在的话并且 envelop::${envelopId}::${uid}::order存在, 返回订单内容
	//	如果没找到订单的key, 说明消费队列正在进行订单插入操作, 返回code为xxx, 让前端调用查询红包接口轮训


	takeEnvelopVo.OutAccountHistory = outAmoutHistory

	order, err := this.isEnvelopTakeByUser(takeEnvelopVo)

	if err != nil && err.Code == constant.EnvelopTakePendingErrorCode {
		return nil, err
	}

	if order != nil {
		return order, nil
	}

	if err != nil {
		return nil , err
	}


	//	如果没查到key, decr envelop::${envelopId}::size, 大于0放入消息队列
	//	否则直接返回红包抢完
	err = this.envelopTakeByUser(takeEnvelopVo)

	return nil, err
}

func (this *EnvelopServiceImpl) envelopTakeByUser(vo *models.TakeEnvelopVo) *constant.RuntimeError {

	key:=vo.EnvelopTradeNo

	mins_10:=10 * time.Minute

	res, err :=redisClient.Client.SetNX(this.envelopTakeKey(vo), ENVELOPTAKEPENDING, mins_10).Result()

	if res == false && err != nil {
		return constant.EnvelopTakeRetryError
	}

	if err != nil {
		return constant.ServerError
	}

	val, err := redisClient.Client.LPop(key).Result()
	if err != nil {
		return constant.EnvelopRunDown
	}


	logs.Info("redpacket pop key %s, amount %s", key, val)

	amount, err := strconv.ParseInt(val, 10, 64)

	if err != nil {
		return &constant.RuntimeError{
			constant.EnvelopTakeAmountParseErrorCode,
			"amount parse error",
		}
	}

	inAccountHistory := new (models.AccountHistory)
	inAccountHistory.Amount = amount
	inAccountHistory.UserId = vo.UserId
	inAccountHistory.GenCreateTime()
	inAccountHistory.Type = models.AccountHistoryTypeIn
	inAccountHistory.Channel = models.AccountHistoryChannelPlat
	inAccountHistory.Pattern = models.AccountHistoryPatternEnvelop
	inAccountHistory.Currency = vo.OutAccountHistory.Currency
	tradeNo := inAccountHistory.GenTradeNo()
	inAccountHistory.TradeNo = tradeNo
	inAccountHistory.Description = inAccountHistory.GenDescription()

	vo.InAccountHistory = inAccountHistory


	bytes, _ := json.Marshal(vo)

	message := &kafka.KafkaMessage{
		constant.ENVELOPTAKETOPIC,
		string(bytes),
	}

	this.AsyncKafkaProducer.SendMessage(message, func(result *kafka.Result, e error) {
		if e != nil {
			logs.Error("发送消息失败 ====== 抢红包成功========  用户 %d, 红包 %s, 金额 %s ", result.Topic, vo.EnvelopTradeNo, result.Value)

			var envelopItem models.EnvelopItem
			envelopItem.Amount = vo.InAccountHistory.Amount
			envelopItem.EnvelopId = vo.EnvelopId
			envelopItem.AccountId = vo.InAccountHistory.AccountId
			envelopItem.UserId = vo.InAccountHistory.UserId
			envelopItem.CreateTime = time.Now().Unix()
			envelopItem.Status = constant.ENVELOPITEMSTATUSTAKEFAILED
			tradeNo, _  := strconv.ParseInt(envelopItem.GenTradeNo(), 10, 64)
			envelopItem.TradeNo = tradeNo
			this.EnvelopItemDao.Tx(func(tx *sql.Tx) error {
				res, err := this.EnvelopItemDao.Create(tx, &envelopItem)
				if err != nil || res == 0 {
					logs.Error("insert envelop item for status = take-failed failed, accountId %d, envelopId %d  ", envelopItem.AccountId, envelopItem.EnvelopId)
				}else {
					logs.Info("insert envelop item for status = take-failed success, accountId %d, envelopId %d  ", envelopItem.AccountId, envelopItem.EnvelopId)
				}
				return err
			})

			//把红包金额重新放回去
		} else {
			logs.Info("发送消息成功 ====== 抢红包成功========  用户 %d, 红包 %s, 金额 %s ", result.Topic, vo.EnvelopTradeNo, result.Value)
		}
	})

	return constant.EnvelopTakePending

	return constant.EnvelopTakeRetryError
}

func (this *EnvelopServiceImpl) envelopTakeKey (takeEnvelopVo *models.TakeEnvelopVo) string {
	return fmt.Sprintf("envelop::%s::%d::take", takeEnvelopVo.EnvelopTradeNo, takeEnvelopVo.UserId)
}

func (this *EnvelopServiceImpl) PutEnvelopOrderRedis (takeEnvelopVo *models.TakeEnvelopVo, item *models.EnvelopItem)  error {
	key:= this.envelopTakeKey(takeEnvelopVo)
	bytes, _:= json.Marshal(item)
	day_7 := 7 * 24 * time.Hour
	_, err:= redisClient.Client.Set(key, string(bytes), day_7).Result()
	return err
}

func (this *EnvelopServiceImpl) isEnvelopTakeByUser(takeEnvelopVo *models.TakeEnvelopVo) ( *models.EnvelopItem, *constant.RuntimeError) {
	key := this.envelopTakeKey(takeEnvelopVo)
	order, err:= redisClient.Client.Get(key).Result()

	if err == redis.Nil {
		return nil, nil
	}


	if err != nil {
		return nil, constant.ServerError
	}

	if order == ENVELOPTAKEPENDING {
		return  nil, constant.EnvelopTakePending
	} else if order == ENVELOPTAKEFAILED {
		return nil, constant.EnvelopNotTakeByUserError
	} else {
		orderJsonByte:= []byte(order)
		model:=&models.EnvelopItem{}
		json.Unmarshal(orderJsonByte, model)
		return model, nil
	}
}



//@deprecated
func (this *EnvelopServiceImpl) TakeEnvelop (takeEnvelopVo models.TakeEnvelopVo) (*models.EnvelopDto, error) {

	//判断是否已经抢过红包
	envelopOrdersKey := takeEnvelopVo.EnvelopTradeNo + "::" + "order::" + strconv.FormatUint(takeEnvelopVo.UserId, 10)

	order, _ := redisClient.Client.Get(envelopOrdersKey).Result()

	if len(order) > 0{
		var result *models.EnvelopDto
		err := json.Unmarshal([]byte(order), &result)
		if err != nil {
			return nil, err
		}
		return result, nil
	}



	//判断红包是否存在
	intcmd:= redisClient.Client.Exists(takeEnvelopVo.EnvelopTradeNo)

	res, err:= intcmd.Result()

	if err != nil {
		return nil, &constant.RuntimeError{
			constant.ConstantErrorCode,
			err.Error(),
		}
	}

	if res == 0 {
		return nil , &constant.RuntimeError{
			constant.EnvelopNotExistsErrorCode,
			"envelop not exists",
		}
	}

	envelopUsersSetKey := takeEnvelopVo.EnvelopTradeNo + "-" + "users"

	intcmd = redisClient.Client.ZAddNX(envelopUsersSetKey, redis.Z{
		Score: 0,
		Member : takeEnvelopVo.UserId,
	})

	res, err = intcmd.Result()

	if err != nil {
		return nil , &constant.RuntimeError{
			constant.ConstantErrorCode,
			err.Error(),
		}
	}

	if res == 0 {
		return this.FindEnvelopFromRedis (takeEnvelopVo, envelopUsersSetKey)
	}

	//从list取出钱, 后面出错都需要把钱放回redis中，把用户从zset删掉
	stringCmd:= redisClient.Client.LPop(takeEnvelopVo.EnvelopTradeNo)

	result, err := stringCmd.Result()

	envelop:=new (models.EnvelopDto)
	envelop.Amount, _ = strconv.ParseInt(result, 10, 64)
	envelop.UserId = takeEnvelopVo.UserId
	envelop.EnvelopTradeNo = takeEnvelopVo.EnvelopTradeNo

	if err != nil {
		return nil, this.resetEnvelopMoney(envelopUsersSetKey ,envelop.Amount, envelop.UserId)
	}


	floatCmd:= redisClient.Client.ZIncrBy(envelopUsersSetKey, float64(envelop.Amount), strconv.FormatUint(takeEnvelopVo.UserId, 10))

	_, err = floatCmd.Result()

	if err != nil {
		return nil, this.resetEnvelopMoney(envelopUsersSetKey ,envelop.Amount, envelop.UserId)
	}

	envelop.CreateTime = time.Now().Unix()

	envelop.TradeNo = envelop.GenTradeNo()

	//把订单写入redis中

	orderByte, err := json.Marshal(envelop)

	if err != nil {
		return nil, this.resetEnvelopMoney(envelopUsersSetKey ,envelop.Amount, envelop.UserId)
	}

	resBool, err := redisClient.Client.SetNX(envelopOrdersKey, string(orderByte), 36 * time.Hour).Result()

	if err != nil  || !resBool{
		return nil, this.resetEnvelopMoney(envelopUsersSetKey ,envelop.Amount, envelop.UserId)
	}

	return envelop, nil

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


func (this *EnvelopServiceImpl) FindEnvelopFromRedis(vo models.TakeEnvelopVo, envelopUsersSetKey string) (*models.EnvelopDto, error) {
	cmd := redisClient.Client.ZScore(envelopUsersSetKey, strconv.FormatUint(vo.UserId, 10))
	res, err := cmd.Result()
	if err != nil {
		return nil, &constant.RuntimeError{
			constant.ConstantErrorCode,
			err.Error(),
		}
	}

	envelopDto:=new (models.EnvelopDto)
	envelopDto.UserId = vo.UserId
	envelopDto.EnvelopTradeNo = vo.EnvelopTradeNo
	envelopDto.Amount = int64(res)
	return envelopDto, nil
}
func (this *EnvelopServiceImpl) resetEnvelopMoney(EnvelopTradeNo string, Amount int64, UserId uint64)  error {

	_, err:= redisClient.TxPipeline(func(pipeliner redis.Pipeliner) error {

		cmd := pipeliner.LPush(EnvelopTradeNo, Amount)

		if cmd.Err() != nil  {
			return cmd.Err()
		}

		cmd = pipeliner.ZRem(EnvelopTradeNo, redis.Z{
			Score:0,
			Member:UserId,
		})


		if cmd.Err() != nil  {
			return cmd.Err()
		}

		return nil

	})

	return err
}
func (this *EnvelopServiceImpl) isEnvelopExists(vo *models.TakeEnvelopVo) (bool, *constant.RuntimeError) {
	key := this.envelopSetKeyShard(vo.EnvelopTradeNo)
	res, err:= redisClient.Client.SIsMember(key, vo.EnvelopId).Result()
	if err != nil {
		return false, &constant.RuntimeError{
			constant.ConstantErrorCode,
			"server redis error..",
		}
	}
	return res, nil
}
func (this *EnvelopServiceImpl) isEnvelopExpire(vo *models.TakeEnvelopVo) (*models.AccountHistory, *constant.RuntimeError) {
	//从redis中拿出红包, 如果红包不存在, 说明红包过期了, 需要前端调查询红包接口详情
	amountStr, err := redisClient.Client.Get(this.envelopOrderkey(vo.EnvelopTradeNo, vo.EnvelopId)).Result()

	if err == redis.Nil {
		return nil, &constant.RuntimeError{
			constant.EnvelopExpireErrorCode,
			"envelop expire ... ",
		}
	}

	if err != nil {
		return nil, &constant.RuntimeError{
			constant.ConstantErrorCode,
			"server redis error..",
		}
	}

	var history models.AccountHistory


	err = json.Unmarshal([]byte(amountStr), &history)

	if err != nil {
		return nil, &constant.RuntimeError{
			constant.ConstantErrorCode,
			"json parse error..",
		}
	}

	return &history, nil

}


func (this *EnvelopServiceImpl) TakeEnvelopByUser(tx *sql.Tx, vo models.TakeEnvelopVo) (*models.EnvelopItem, error) {

	var envelopItem models.EnvelopItem

	envelopItem.Amount = vo.InAccountHistory.Amount
	envelopItem.EnvelopId = vo.EnvelopId
	envelopItem.AccountId = vo.InAccountHistory.AccountId
	envelopItem.UserId = vo.InAccountHistory.UserId
	envelopItem.CreateTime = time.Now().Unix()
	envelopItem.Status = constant.ENVELOPITEMSTATUSTAKE
	tradeNo, _  := strconv.ParseInt(envelopItem.GenTradeNo(), 10, 64)
	envelopItem.TradeNo = tradeNo
	res, err := this.EnvelopItemDao.Create(tx, &envelopItem)

	logs.Info("insert envelop item, rows %d, err %v", res, err)

	if err != nil {
		return nil, err
	}
	if res == 0 {
		return nil, constant.EnvelopItemCreateError
	}


	res, err = this.EnvelopDao.ReduceQuantityAndRemainingAmount(tx,	envelopItem.EnvelopId, envelopItem.Amount)

	logs.Info("update envelop quantity, rows %d, err %v", res, err)

	if err != nil {
		return nil, err
	}
	if res == 0 {
		return nil, constant.EnvelopItemCreateError
	}

	return &envelopItem ,nil

}


func (this *EnvelopServiceImpl) QueryEnvelop (vo *models.QueryEnvelopVo) (*models.EnvelopItem, error) {

	takeEnvelopVo := &models.TakeEnvelopVo{
		UserId:	vo.UserId,
		EnvelopId:	vo.EnvelopId,
		EnvelopTradeNo:	vo.EnvelopTradeNo,
	}

	exists, err := this.isEnvelopExists(takeEnvelopVo)

	if err != nil {
		return nil, err
	}

	if exists == false {
		return nil, constant.EnvelopNotExists
	}

	//在redis查询红包的订单信息

	//如果红包没有过期, 说明用户没有抢到红包

	//如果红包已经过期, 则从mysql中查询并放回redis

	order, err := this.isEnvelopTakeByUser(takeEnvelopVo)

	if err != nil && err.Code == constant.EnvelopTakePendingErrorCode {
		return nil, err
	}

	if err != nil && err.Code == constant.EnvelopNotTakeByUserErrorCode {
		return nil ,err
	}

	if order != nil {
		return order, nil
	}

	const retryCount int = 50

	for i:=0; i<retryCount; i++ {

		res, err:= this.tryQueryEnvelopItemByMysql(takeEnvelopVo)

		if err != nil {
			return nil ,err
		}

		if res != nil {
			return res, nil
		}

		sleep:= time.Millisecond * 500

		time.Sleep(sleep)

	}

	return nil, constant.EnvelopNotTakeByUserError
}

func (this *EnvelopServiceImpl) tryQueryEnvelopItemByMysql (vo *models.TakeEnvelopVo) (*models.EnvelopItem, error) {

	res, err := this.tryLockQueryEnvelopItem(vo)

	if res == true {
		item, err := this.EnvelopItemDao.SelectByEnvelopIdAndUserId(vo)
		if item == nil {
			key := this.envelopTakeKey(vo)
			duration := 10 * time.Second
			redisClient.Client.SetNX(key, ENVELOPTAKEFAILED, duration)
			return nil, constant.EnvelopNotTakeByUserError
		}
		err = this.PutEnvelopOrderRedis(vo, item)

		if err != nil {
			return item, err
		}
	}

	return nil, err

}

func (this * EnvelopServiceImpl) tryLockQueryEnvelopItem(vo *models.TakeEnvelopVo) (bool, error) {

	key := fmt.Sprintf("lock::envelop::item::%d::uid::%d", vo.EnvelopId, vo.UserId)

	secondsOfTwo := 5 * time.Second

	res, err:= redisClient.Client.SetNX(key, time.Now().Unix(), secondsOfTwo).Result()

	if err != nil {
		logs.Info("redis set nx failed, key %s, err %v", key, err)
		return false, constant.ServerError
	}

	if res == false {
		return res, nil
	}

	return true, nil

}
