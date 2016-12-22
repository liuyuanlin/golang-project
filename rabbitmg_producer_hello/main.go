// rabbitmg_producer_hello project main.go
package main

import (
	"fmt"
	"log"
)

const (
	//AMQP URI
	uri = "amqp://guest:guest@localhost:5672/"
	//Durable AMQP exchange name
	exchangeName = ""
	//Durable AMQP queue name
	queueName = "test-idoall-queues"
	//Body of message
	bodyMsg string = "hello idoall.org"
)

//如果存在错误，则输出
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func main() {

	/*
		var test HelloProducer
		test.Run(uri, exchangeName, queueName)
	*/
	var test BalanceProducer
	test.Run(uri, exchangeName, queueName)
}
