package kafka

import (
	"github.com/astaxie/beego/logs"
	"github.com/facebookarchive/inject"
)

func MustInit (g *inject.Graph) {
	kafkaConfig:= &KafkaConfig{
		Address: []string{"localhost:9092"},
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