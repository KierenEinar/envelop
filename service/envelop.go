package service

import (
	"database/sql"
	"encoding/json"
	"envelop/constant"
	"envelop/dao"
	"envelop/infra/algo"
	"envelop/infra/kafka"
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
	ENVELOPTAKETOPIC = "envelop-take"
)

type EnvelopService interface {
	CreateEnvelop (envelop * models.Envelop) (*string, error) //发红包
	TakeEnvelop (models.TakeEnvelopVo) (*models.EnvelopDto, error) //抢红包
	TakeEnvelopNew (models.TakeEnvelopVo) (*models.EnvelopDto, error) //抢红包, 新版本
}

var (
	envelopRandomDoubleStrategy = algo.EnvelopDoubleAvgStrategy{}
)

type EnvelopServiceImpl struct {

	AccountService *AccountServiceImpl `inject:""`
	EnvelopDao *dao.EnvelopDaoImpl `inject:""`
}

func (this *EnvelopServiceImpl) CreateEnvelop(envelop * models.Envelop) (*string, error) {

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
		return nil, &constant.RuntimeError{
			constant.ParamErrorCode,
			"param error",
		}
	}

	var key string

	err = this.EnvelopDao.Tx(func(tx *sql.Tx) error {

		if envelop.PayChannel == models.AccountHistoryChannelPlat {
			outAccountHistory := new (models.AccountHistory)
			outAccountHistory.Amount = 0 - envelop.Amount
			outAccountHistory.UserId = envelop.UserId
			outAccountHistory.GenCreateTime()
			outAccountHistory.Type = models.AccountHistoryTypeOut
			outAccountHistory.Channel = models.AccountHistoryChannelPlat
			outAccountHistory.Pattern = models.AccountHistoryPatternEnvelop
			outAccountHistory.Currency = envelop.Currency
			tradeNo := outAccountHistory.GenTradeNo()
			outAccountHistory.TradeNo = tradeNo
			outAccountHistory.Description = outAccountHistory.GenDescription()
			err := this.AccountService.UpdateBalance(tx, outAccountHistory)
			if err != nil {
				tx.Rollback()
				return err
			}

			key = tradeNo
			envelop.AccountId = outAccountHistory.AccountId
			envelop.CreateTime = outAccountHistory.CreateTime
		}

		err:= this.EnvelopDao.Create(tx, envelop)

		if err != nil{
			tx.Rollback()
			return err
		}

		seeds := make([]interface{}, 0)

		envelopRandomDoubleStrategy.Generate(envelop.Quantity, envelop.Amount, &seeds)

		_, err = redisClient.TxPipeline(func (pipeliner redis.Pipeliner) error {

			oneDay,_ := time.ParseDuration("1d")

			flag, err:= pipeliner.SetNX(this.envelopOrderkey(key), "", oneDay).Result()
			if flag == false {
				return &constant.RuntimeError{
					constant.EnvelopCreateErrorCode,
					"envelop exists...",
				}
			}

			if err != nil {
				return &constant.RuntimeError{
					constant.EnvelopCreateErrorCode,
					err.Error(),
				}
			}

			cmd := pipeliner.LPush(key, seeds...)
			if cmd.Err() != nil {
				return cmd.Err()
			}

			return nil

		})

		if err != nil {
			tx.Rollback()
			return err
		}

		tx.Commit()

		return nil

	})


	return &key, err


}


func (this * EnvelopServiceImpl) envelopOrderkey (envelopTradeNo string) string {
	return fmt.Sprintf("envelop::%s", envelopTradeNo)
}

func (this *EnvelopServiceImpl) TakeEnvelopNew (takeEnvelopVo models.TakeEnvelopVo) (*models.EnvelopDto, error) {

	//	查找秒杀的key, envelop::${envelopId}::${uid}::take, 如果key存在的话并且 envelop::${envelopId}::${uid}::order存在, 返回订单内容
	//	如果没找到订单的key, 说明消费队列正在进行订单插入操作, 返回code为xxx, 让前端调用查询红包接口轮训


	err, order := this.isEnvelopTakeByUser(takeEnvelopVo)

	if err.Code == constant.EnvelopTakePendingErrorCode {
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

	if err != nil {
		return nil, err
	}


	return nil, nil
}

func (this *EnvelopServiceImpl) envelopTakeByUser(vo models.TakeEnvelopVo) *constant.RuntimeError {
	key:=vo.EnvelopTradeNo
	amount, err := redisClient.Client.LPop(key).Result()
	if err != nil {
		return &constant.RuntimeError{
			constant.EnvelopRunDownErrorCode,
			"envelop run down ... ",
		}
	}

	producer:=kafka.GetProducerInstance()

	message := &kafka.KafkaMessage{
		ENVELOPTAKETOPIC,
		amount,
	}

	producer.SendMessage(message, func(result *kafka.Result, e error) {
		if e != nil {
			logs.Error("发送消息失败 ====== 抢红包成功========  用户 %d, 红包 %s, 金额 %s ", result.Topic, vo.EnvelopTradeNo, result.Value)

		} else {
			logs.Info("发送消息成功 ====== 抢红包成功========  用户 %d, 红包 %s, 金额 %s ", result.Topic, vo.EnvelopTradeNo, result.Value)
		}
	})

	return nil
}


func (this *EnvelopServiceImpl) isEnvelopTakeByUser(takeEnvelopVo models.TakeEnvelopVo) (*constant.RuntimeError, *models.EnvelopDto) {
	key := fmt.Sprint("envelop::%s::%d::take", takeEnvelopVo.EnvelopTradeNo, takeEnvelopVo.UserId)
	order, err:= redisClient.Client.Get(key).Result()
	if err != nil {
		return &constant.RuntimeError{
			constant.ConstantErrorCode,
			err.Error(),
		}, nil
	}

	if len(order) == 0 {
		return nil , nil
	}

	if order == ENVELOPTAKEPENDING {
		return &constant.RuntimeError{
			constant.EnvelopTakePendingErrorCode,
			"envelop take pending",
		}, nil
	} else {
		orderJsonByte:= []byte(order)
		model:=&models.EnvelopDto{}
		json.Unmarshal(orderJsonByte, model)
		return nil, model
	}
}



//@deprecated
func (this *EnvelopServiceImpl) TakeEnvelop (takeEnvelopVo models.TakeEnvelopVo) (*models.EnvelopDto, error) {

	//判断是否已经抢过红包
	envelopOrdersKey := takeEnvelopVo.EnvelopTradeNo + "-" + "order-" + strconv.FormatUint(takeEnvelopVo.UserId, 10)

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


