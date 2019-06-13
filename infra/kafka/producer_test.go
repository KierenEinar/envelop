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
	kafkaConfig:= &kafka.KafkaConfig{
		Address: []string{"localhost:9092"},
	}

	producer.Start(kafkaConfig)
	message:= &kafka.KafkaMessage{Topic:"envelop-take",Body:"hello world"}
	defer producer.Shutdown()
	producer.SendMessage(message, func(result *kafka.Result, e error) {
		if e != nil {
			logs.Error("message send err %v", e.Error())
		}
		logs.Info("message send success, partition %d, offset %d, topic %s, value %s", result.Partition, result.Offset, result.Topic, result.Value)
	})
	duration,_:=time.ParseDuration("10s")

	time.Sleep(duration)
}

//func TestEnvelopTakeListener_OnListeningst(t *testing.T) {
//	container:= kafka.ConcumerContainer{
//		ConsumerConfig:kafka.ConsumerConfig{
//			Address: []string{"localhost:9092"},
//			GroupId: "group1",
//			Topic: "envelop-take",
//		},
//		MessageListener: new(kafka.EnvelopTakeListener),
//	}
//	containers:=make([]kafka.ConcumerContainer, 0)
//	containers = append(containers, container)
//	err := kafka.RegisterContainer(containers)
//	if err != nil {
//		logs.Error("register consumer failed ..., %v", err)
//		panic(err)
//	}
//
//	logs.Info("kafka all consumer start success ...")
//
//	duration,_:=time.ParseDuration("10s")
//
//	time.Sleep(duration)
//}

func TestMain(m *testing.M) {

	//kafkaConfig:= &kafka.KafkaConfig{
	//	Address: []string{"localhost:9092"},
	//}
	//
	//producer := kafka.GetProducerInstance()
	//err:= producer.Start(kafkaConfig)
	//if err != nil {
	//	logs.Error("producer start failed %v", err)
	//	panic(err)
	//}

	os.Exit(m.Run())

}


