package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/astaxie/beego/logs"
	"sync"
	"time"
)

var (
	producer *AsyncKafkaProducer
	once = sync.Once{}
)

type KafkaConfig struct {
	Address[] string `json:"address"`
}

type KafkaMessage struct {
	Topic string
	Body  string
}

type KafkaProducer interface {
	Start(*KafkaConfig) error
	//SendMessage(*KafkaMessage) error
	Shutdown() error
	Config () *sarama.Config
}

type Result struct {
	Value string
	Topic string
	Offset int64
	Partition int32
	TimeStamp time.Time
}

type OnSend func (*Result, error)

type KafkaProducerSendMessage interface {
	SendMessage(*KafkaMessage, OnSend)
}

type AsyncKafkaProducer struct {
	KafkaConfig *KafkaConfig
	AsyncKafkaProducer sarama.AsyncProducer
}


func (this *AsyncKafkaProducer) Start(kafkaConfig *KafkaConfig) error {
	logs.Info("kafka producer starting ...")
	this.KafkaConfig = kafkaConfig
	config := this.Config()
	producer, error := sarama.NewAsyncProducer(kafkaConfig.Address, config)
	if error != nil {
		return error
	}
	this.AsyncKafkaProducer = producer
	logs.Info("kafka producer start success ... ")
	return nil
}

func (this *AsyncKafkaProducer) SendMessage(kafkaMessage *KafkaMessage, onSend OnSend) {

	msg:= &sarama.ProducerMessage{
		Topic:kafkaMessage.Topic,
		Value:sarama.ByteEncoder(kafkaMessage.Body),
	}

	this.AsyncKafkaProducer.Input() <- msg

	go this.OnSend(onSend)

}


func (this *AsyncKafkaProducer) OnSend(onSend OnSend) {

	producer:= this.AsyncKafkaProducer

	for {
		select {
		case success := <- producer.Successes():
			 b, _:=success.Value.Encode()
			 v := string(b)
			 result := Result{
			 	Value: v,
			 	Topic: success.Topic,
			 	TimeStamp:success.Timestamp,
			 	Offset: success.Offset,
			 	Partition:success.Partition,
			 }
			 onSend(&result, nil)
		case fail := <- producer.Errors():
			onSend(nil, fail.Err)
		}

	}

}

func (this *AsyncKafkaProducer) Shutdown() error {
	logs.Info("kafka producer shutdown ...")
	this.AsyncKafkaProducer.AsyncClose()
	return nil
}

func (this *AsyncKafkaProducer) Config () *sarama.Config {
	config:=sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Version = sarama.V0_10_0_1
	return config
}

func GetProducerInstance () *AsyncKafkaProducer {
	once.Do(func() {
		producer = &AsyncKafkaProducer{}
		logs.Info("hello world")
	})
	logs.Info("producer", producer)

	return producer
}