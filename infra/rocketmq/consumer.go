package rocketmq

import (
	"fmt"
	"github.com/apache/rocketmq-client-go/core"
	"github.com/astaxie/beego/logs"
)

var (
	consumers = make(map[string]ConsumerContainer, 0)
)

type OnListening func (message *Message) error

type Listener struct {
	Config *MqConfig
	Topic string
	OnListening OnListening
}

func RegisterListener (listener *Listener ) {

	if consumers[listener.Topic] != nil {
		panic("重复注册相同的topic")
	}

	consumer := &DefaultConsumerContainer{
		Config: listener.Config,
	}

	err:= consumer.Start()

	if err != nil {
		panic(fmt.Sprintf("消费者启动不了, group: %v, topic: %v, nameserver : %v", listener.Config.GroupName, listener.Topic, listener.Config.NameServer))
	}

	logs.Info("消费者启动成功, group: %v, topic: %v", listener.Config.GroupName, listener.Topic)

	go consumer.Subscribe(listener.Topic, listener.OnListening)


}


type ConsumerContainer interface {
	Start() error
	Stop()  error
	Subscribe(topic string, listening OnListening)
}

type DefaultConsumerContainer struct {
	pullConsumer *rocketmq.PullConsumer
	Config *MqConfig
}

func (this *DefaultConsumerContainer) Start() error{

	pullConsumer, err:= rocketmq.NewPullConsumer(&rocketmq.PullConsumerConfig{
		ClientConfig: rocketmq.ClientConfig{
			GroupID: this.Config.GroupName,
			NameServer:this.Config.NameServer,
		},
	})

	if err != nil {
		return nil
	}

	this.pullConsumer = &pullConsumer

	return nil

}

func (this *DefaultConsumerContainer) Stop() error {

	if this.pullConsumer == nil {
		return &ProducerShutdownError{}
	}

	return  (*this.pullConsumer).Shutdown()
}


func (this *DefaultConsumerContainer) Subscribe(topic string, listening OnListening) error {
	consumer:= *this.pullConsumer
	messagequeues:= consumer.FetchSubscriptionMessageQueues(topic)
	offset := int64(0)
	for _, messagequeue:= range messagequeues {

		pullResult := consumer.Pull(messagequeue, "*", offset, 32)
		offset = pullResult.NextBeginOffset
		messages:= pullResult.Messages
		body := ""
		var msg *Message
		for _, message:= range messages {
			body = body + message.Body
			if msg == nil  {
				msg = &Message{}
				msg.Topic = message.Topic
				msg.Keys = message.Keys
				msg.Tag = message.Tags
				msg.DelayTimeLevel = message.DelayTimeLevel
			}
		}

		if msg != nil {
			msg.Body = body
		}

		listening(msg)

	}

}