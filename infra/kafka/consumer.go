package kafka

import (
	"errors"
	"github.com/Shopify/sarama"
	"github.com/astaxie/beego/logs"
	"github.com/bsm/sarama-cluster"
	"log"
)

type ConsumerConfig struct {
	Address []string `json:"address"`
	GroupId string  `json:"groupId"`
	Topic string `json:topic`
}

var (
	containers = make([]ConcumerContainer, 0)
)

type MessageListener interface {
	OnListening (string ,string, error)
}

func RegisterContainer (containers []ConcumerContainer) error {

	for _, container:= range containers{
		err:=container.Start()
		if err != nil {
			return err
		}
		go handlerMessageListener(container.consumer, container.MessageListener)
		containers = append(containers, container)
	}

	return nil
}




type ConcumerContainer struct {
	ConsumerConfig ConsumerConfig
	MessageListener MessageListener
	consumer cluster.Consumer
}

func (container*ConcumerContainer) Start ()  error {

	consumer, err:=cluster.NewConsumer(container.ConsumerConfig.Address, container.ConsumerConfig.GroupId, []string{container.ConsumerConfig.Topic}, container.config())

	if err != nil {
		return err
	}
	container.consumer = *consumer

	go func() {
		for err := range consumer.Errors() {
			log.Printf("%s:Error: %s\n", container.ConsumerConfig.GroupId, err.Error())
		}
	}()

	// consume notifications
	go func() {
		for ntf := range consumer.Notifications() {
			log.Printf("%s:Rebalanced: %+v \n", container.ConsumerConfig.GroupId, ntf)
		}
	}()

	return nil
}

func (this *ConcumerContainer) Shutdown () error {
	return this.consumer.Close()
}

func (this *ConcumerContainer) config () *cluster.Config {
	config:=cluster.NewConfig()
	config.Consumer.Return.Errors = true
	config.Group.Return.Notifications = true
	config.Group.Mode = cluster.ConsumerModePartitions
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	return config
}

func handlerMessageListener (consumer cluster.Consumer, messageListener MessageListener) error {

	for {
		select {
		case part, ok := <-consumer.Partitions():
			if !ok {
				return errors.New("partition not ok")
			}
			logs.Info("partirion", part)
			logs.Info("ok", ok)

			go func(pc cluster.PartitionConsumer) {
				for msg:= range pc.Messages() {
					logs.Info("msg->", msg.Topic)
					messageListener.OnListening(msg.Topic, string(msg.Value), nil)
					consumer.MarkOffset(msg, "")
				}

			}(part)
		}
	}


}