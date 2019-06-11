package kafka_test

import (
	"envelop/infra/kafka"
	"github.com/astaxie/beego/logs"
	"os"
	"testing"
	"time"
)

func TestAsyncKafkaProducer_SendMessage(t *testing.T) {

	producer:=kafka.GetProducerInstance()
	message:= &kafka.KafkaMessage{Topic:"envelop-take",Body:"hello world"}
	//defer producer.Shutdown()
	producer.SendMessage(message, func(result *kafka.Result, e error) {
		if e != nil {
			logs.Error("message send err %v", e.Error())
		}
		logs.Info("message send success, %d", result.Partition)
	})
	duration,_:=time.ParseDuration("10s")

	time.Sleep(duration)
}

func TestMain(m *testing.M) {

	kafkaConfig:= &kafka.KafkaConfig{
		Address: []string{"localhost:9092"},
	}

	producer := kafka.GetProducerInstance()
	err:= producer.Start(kafkaConfig)
	if err != nil {
		logs.Error("producer start failed %v", err)
		panic(err)
	}

	os.Exit(m.Run())

}


