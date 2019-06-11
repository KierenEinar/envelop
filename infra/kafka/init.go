package kafka

func init() {

	//kafkaConfig:= &KafkaConfig{
	//	Address: []string{"localhost:9092"},
	//}
	//
	//producer := GetProducerInstance()
	//err:= producer.Start(kafkaConfig)
	//if err != nil {
	//	logs.Error("producer start failed %v", err)
	//	panic(err)
	//}
	//container:= ConcumerContainer{
	//	ConsumerConfig:ConsumerConfig{
	//		Address: []string{"localhost:9092"},
	//		GroupId: "envelop-take",
	//		Topic: "envelop-take",
	//	},
	//	MessageListener: new(EnvelopTakeListener),
	//}
	//containers:=make([]ConcumerContainer, 0)
	//containers = append(containers, container)
	//err = RegisterContainer(containers)
	//if err != nil {
	//	logs.Error("register consumer failed ..., %v", err)
	//	panic(err)
	//}
	//
	//logs.Info("kafka all consumer start success ...")

}
