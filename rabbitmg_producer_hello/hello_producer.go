// rabbitmg_producer_hello project main.go
package main

import (
	"golang-project/proto/example"
	"log"

	"github.com/golang/protobuf/proto"
	"github.com/streadway/amqp"
)

type HelloProducer struct {
	Uri          string
	ExchangeName string
	QueueName    string
}

func (self *HelloProducer) Run(uri string, exchangeName string, queueName string) {
	self.Uri = uri
	self.ExchangeName = exchangeName
	self.QueueName = queueName

	test := &example.Test{
		Label: proto.String("hello"),
		Type:  proto.Int32(17),
		Reps:  []int64{1, 2, 3},
		Optionalgroup: &example.Test_OptionalGroup{
			RequiredField: proto.String("good bye"),
		},
	}

	data, err := proto.Marshal(test)
	if err != nil {
		log.Fatal("marshaling error: ", err)
		return
	}
	//调用发布消息函数
	self.publish(self.Uri, self.ExchangeName, self.QueueName, data)
	log.Printf("published %dB OK", len(data))

}

//发布者的方法
//
//@amqpURI, amqp的地址
//@exchange, exchange的名称
//@queue, queue的名称
//@body, 主体内容
func (self *HelloProducer) publish(amqpURI string, exchange string, queue string, body []byte) {
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
	log.Printf("declared queue, publishing %dB body (%v)", len(body), body)

	// Producer只能发送到exchange，它是不能直接发送到queue的。
	// 现在我们使用默认的exchange（名字是空字符）。这个默认的exchange允许我们发送给指定的queue。
	// routing_key就是指定的queue名字。
	err = channel.Publish(
		exchange, // exchange
		q.Name,   // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
			ContentEncoding: "",
			Body:            body,
		})
	failOnError(err, "Failed to publish a message")
}
