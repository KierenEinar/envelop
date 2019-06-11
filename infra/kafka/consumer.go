package kafka

import (
	"github.com/bsm/sarama-cluster"
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

	consumer, err:=cluster.NewConsumer(container.ConsumerConfig.Address, container.ConsumerConfig.GroupId, container.ConsumerConfig.Address, container.config())

	if err != nil {
		return err
	}
	container.consumer = *consumer
	return nil
}

func (this *ConcumerContainer) Shutdown () error {
	return this.consumer.Close()
}

func (this *ConcumerContainer) config () *cluster.Config {
	config:=cluster.NewConfig()
	config.Group.Mode = cluster.ConsumerModePartitions
	return config
}

func handlerMessageListener (consumer cluster.Consumer, messageListener MessageListener) error {

	for {

		select {
		case part, _ := <-consumer.Partitions():
			go func(pc cluster.PartitionConsumer) {

				for {
					select {
						case message:= <-pc.Messages():
							messageListener.OnListening(message.Topic ,string(message.Value), nil)
						case err:= <-pc.Errors():
							messageListener.OnListening(err.Topic, "", err.Err)
					}
				}

			}(part)
		}

	}

}