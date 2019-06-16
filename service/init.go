package service

import (
	"envelop/conf"
	"envelop/constant"
	"envelop/infra/kafka"
	"github.com/astaxie/beego/logs"
	"github.com/facebookarchive/inject"
)

func MustInit(g *inject.Graph) {

	initProducer(g)

	g.Provide(
		&inject.Object{Value: &TransferStrategyPlat2Plat{}},
		&inject.Object{Value: &TransferStrategyPlat2UnionPay{}},
		&inject.Object{Value: &TransferStrategyUnionPay2UnionPay{}},
		&inject.Object{Value: &TransferStrategyUnionPay2Plat{}},
		&inject.Object{Value: &AccountServiceImpl{}},
		&inject.Object{Value: &EnvelopServiceImpl{}},
		&inject.Object{Value: &UserServiceImpl{}},
	)

	initConsumer(g)

}

func initProducer (g *inject.Graph) {
	kafka.MustInit(g)
}


func initConsumer (g *inject.Graph) {

	var envelopTakeListener EnvelopTakeListener

	config:= conf.GetInstance().KafkaConfig

	container:= kafka.ConcumerContainer{
		ConsumerConfig:kafka.ConsumerConfig{
			Address: config.Addr,
			GroupId: config.ConsumerConfig.GroupId,
			Topic: constant.ENVELOPTAKETOPIC,
		},
		MessageListener: &envelopTakeListener,
	}
	containers:=make([]kafka.ConcumerContainer, 0)
	containers = append(containers, container)
	err := kafka.RegisterContainer(containers)
	if err != nil {
		logs.Error("register consumer failed ..., %v", err)
		panic(err)
	}

	g.Provide(
		&inject.Object{Value: &envelopTakeListener},
	)

	logs.Info("kafka all consumer start success ...")

}
