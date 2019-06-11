package rocketmq

type ConsumerShutdownError struct {}

func (this*ConsumerShutdownError) Error () string {
	return "ConsumerShutdownError"
}

type ProducerShutdownError struct {}

func (this*ProducerShutdownError) Error () string {
	return "ProducerShutdownError"
}