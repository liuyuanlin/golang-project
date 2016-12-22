// rabbitmq_consumer_hello project main.go
package main

import (
	"fmt"
	"log"
)

const (
	//AMQP URI
	uri = "amqp://guest:guest@localhost:5672/"
	//Durable AMQP exchange nam
	exchangeName = ""
	//Durable AMQP queue name
	queueName = "test-idoall-queues"
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
		var test HelloConsumer
		test.Run(uri, exchangeName, queueName)
	*/
	var test BalanceComsumer
	test.Run(uri, exchangeName, queueName)
}
