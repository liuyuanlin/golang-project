package main

import (
	"golang-project/proto/example"
	"log"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/streadway/amqp"
)

type HelloConsumer struct {
	Uri          string
	ExchangeName string
	QueueName    string
}

func (self *HelloConsumer) Run(uri string, exchangeName string, queueName string) {
	self.Uri = uri
	self.ExchangeName = exchangeName
	self.QueueName = queueName
	//调用消息接收者
	self.consumer(self.Uri, self.ExchangeName, self.QueueName)
}

//接收者方法
//
//@amqpURI, amqp的地址
//@exchange, exchange的名称
//@queue, queue的名称
func (self *HelloConsumer) consumer(amqpURI string, exchange string, queue string) {
	//建立连接
	log.Printf("dialing %q", amqpURI)
	connection, err := amqp.Dial(amqpURI)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer connection.Close()

	//创建一个Channel
	log.Printf("got Connection, getting Channel")
	channel, err := connection.Channel()
	failOnError(err, "Failed to open a channel")
	defer channel.Close()

	log.Printf("got queue, declaring %q", queue)

	//创建一个queue
	q, err := channel.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, "Failed to declare a queue")

	log.Printf("Queue bound to Exchange, starting Consume")
	//订阅消息
	msgs, err := channel.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	for {
		select {
		case <-time.After(time.Second * time.Duration(2)):
			{
				log.Printf("get msg timer out")
				goto ForEnd

			}
		case <-msgs:
			{
				d := <-msgs
				log.Printf("Received a message: %v", d.Body)
				log.Printf("Received a message: %v, %v", d.ContentType, d.MessageCount)
				test := &example.Test{}
				err := proto.Unmarshal(d.Body, test)
				if err != nil {
					log.Printf("proto unmarshal: %v", err)
					continue
				}
				log.Printf("Received a message: %v", test)
			}

		}
	}
ForEnd:

	log.Printf("hello cousumer over")

}
