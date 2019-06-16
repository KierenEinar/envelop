package kafka

import (
	"envelop/conf"
	"github.com/astaxie/beego/logs"
	"github.com/facebookarchive/inject"
)

func MustInit (g *inject.Graph) {

	config:= conf.GetInstance().KafkaConfig

	kafkaConfig:= &KafkaConfig{
		Address: config.Addr,
	}

	producer := GetProducerInstance()
	err:= producer.Start(kafkaConfig)
	if err != nil {
		logs.Error("producer start failed %v", err)
		panic(err)
	}

	g.Provide(
		&inject.Object{Value: producer},
	)
}