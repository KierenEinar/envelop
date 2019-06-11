package rocketmq

import (
	"envelop/infra/util"
	"github.com/apache/rocketmq-client-go/core"
	"github.com/astaxie/beego/logs"
	"sync"
)

func init () {
	producer := GetProducerInstance()
	producer.Start(&MqConfig{
		"test",
		"localhost",
	})
}

var (
	producerService ProducerService
)


type MqConfig struct {
	GroupName string
	NameServer string `json: name_server`
}


type Message struct {
	Tag 	string
	Topic   string
	Keys 	string
	DelayTimeLevel int
	Body    string
}


type ProducerService interface {
	Start(*MqConfig) error
	//SendMessageAsync(message *Message, fn func (result rocketmq.SendResult, err error))
	SendMessageSync (message *Message) (*rocketmq.SendResult, error)
	SendMessageOrderly (message *Message, key string, retryTime int) (*rocketmq.SendResult, error)
	Stop() error
}

type ProducerServiceImpl struct {
	producer rocketmq.Producer
}

func GetProducerInstance () ProducerService {

	once:=sync.Once{}
	once.Do(func() {
		producerService = new(ProducerServiceImpl)
	})
	return producerService
}

func (this *ProducerServiceImpl) Start(mqConfig *MqConfig) error {
	producer, error := rocketmq.NewProducer(&rocketmq.ProducerConfig{
		ClientConfig: rocketmq.ClientConfig{
			NameServer: mqConfig.NameServer,
			GroupID: mqConfig.GroupName,
		},
	})

	if error != nil {
		return error
	}

	error = producer.Start()

	if error != nil {
		return error
	}
	logs.Info("producer start success .... ")

	this.producer = producer

	return nil
}


func (this *ProducerServiceImpl) Stop ()  error {
	err := this.producer.Shutdown()
	logs.Info("stop producer ... error ", err)
	return err
}

func (this * ProducerServiceImpl) SendMessageSync (message *Message) (*rocketmq.SendResult, error) {
	msg:=this.convertMessage(message)
	return this.producer.SendMessageSync(msg)
}


func (this *ProducerServiceImpl) convertMessage (message *Message) *rocketmq.Message {
	msg:=&rocketmq.Message{
		Topic: message.Topic,
		Body:  message.Body,
		Tags:  message.Tag,
		Keys:  message.Keys,
		DelayTimeLevel:message.DelayTimeLevel,
	}
	return msg
}

type MessageQueueSelectorById struct {}

func (this *MessageQueueSelectorById) Select (size int, m* rocketmq.Message, arg interface{} ) int {
	queue := arg.(int64) % int64(size)
	return int(queue)
}


func (this* ProducerServiceImpl) SendMessageOrderly(message *Message, key string, retryTimes int)(*rocketmq.SendResult, error) {
	msg:=this.convertMessage(message)
	return this.producer.SendMessageOrderly(msg, &MessageQueueSelectorById{}, util.HashCode(&key), retryTimes)
}

