package main

import (
	"errors"
	"golang-project/proto/example"
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/golang/protobuf/proto"
	"github.com/streadway/amqp"
)

type BalanceComsumer struct {
	Uri          string
	ExchangeName string
	QueueName    string
	AmqpConnect  *amqp.Connection
	AmqpQueue    amqp.Queue
	RedisConn    redis.Conn
}

func (self *BalanceComsumer) Run(uri string, exchangeName string, queueName string) error {

	self.Uri = uri
	self.ExchangeName = exchangeName
	self.QueueName = queueName
	log.Println("start connect redis")
	var err error
	self.RedisConn, err = redis.Dial("tcp", "127.0.0.1:6379")
	failOnError(err, "connect redis")

	defer self.RedisConn.Close()

	log.Println("start connect broker")
	self.AmqpConnect, err = amqp.Dial(self.Uri)
	if err != nil {
		log.Println("amqp dial error:%v", err)
		return err
	}
	defer self.AmqpConnect.Close()
	log.Println("connect broker success")
	self.DBComsumer()
	return nil

}

func (self *BalanceComsumer) DBComsumer() error {

	log.Println("DBComsumer enter")

	if self.AmqpConnect == nil {
		err := errors.New("amqp not connect")
		return err
	}

	channel, err := self.AmqpConnect.Channel()
	if err != nil {
		log.Println("amqp channel error:%v", err)
		return err
	}
	defer channel.Close()
	channel.Qos(1, 0, false)

	if self.AmqpQueue.Name == "" {
		log.Println("AmqpQueue  create")
		self.AmqpQueue, err = channel.QueueDeclare(self.QueueName, false, false, false, false, nil)
		if err != nil {
			log.Println("amqp Queue Declare error:%v", err)
			return err
		}
	}
	log.Println("AmqpQueue  create success")
	//订阅消息
	msgs, err := channel.Consume(
		self.AmqpQueue.Name, // queue
		"",                  // consumer
		true,                // auto-ack
		false,               // exclusive
		false,               // no-local
		false,               // no-wait
		nil,                 // args
	)
	failOnError(err, "Failed to register a consumer")
	log.Println("ready get msg")
	for {
		log.Println("ready get msg 2")
		select {
		case <-time.After(time.Second * time.Duration(2)):
			{
				log.Printf("get msg timer out")
				goto ForEnd

			}
		case d := <-msgs:
			{
				/*
									// Properties
					ContentType     string    // MIME content type
					ContentEncoding string    // MIME content encoding
					DeliveryMode    uint8     // queue implementation use - non-persistent (1) or persistent (2)
					Priority        uint8     // queue implementation use - 0 to 9
					CorrelationId   string    // application use - correlation identifier
					ReplyTo         string    // application use - address to to reply to (ex: RPC)
					Expiration      string    // implementation use - message expiration spec
					MessageId       string    // application use - message identifier
					Timestamp       time.Time // application use - message timestamp
					Type            string    // application use - message type name
					UserId          string    // application use - creating user - should be authenticated user
					AppId           string    // application use - creating application id

				*/
				log.Printf("ContentType: %v, ContentEncoding: %v,DeliveryMode: %v, Priority: %v, CorrelationId: %v", d.ContentType, d.ContentEncoding, d.DeliveryMode, d.Priority, d.CorrelationId)
				log.Printf("ReplyTo: %s, Expiration: %s,MessageId: %s, Timestamp: %s, Type: %s, UserId: %s, AppId: %s", d.ReplyTo, d.Expiration, d.MessageId, d.Timestamp, d.Type, d.UserId, d.AppId)
				log.Println("ready get msg 1, %v", d.Body)
				log.Println("Received a message: %v", d.Body)
				log.Println("Received a message: %v, %v", d.ContentType, d.MessageCount)
				player := &example.PlayerInfo{}
				err := proto.Unmarshal(d.Body, player)
				if err != nil {
					log.Printf("proto unmarshal: %v", err)
					continue
				}
				log.Printf("Received a message: %v", player)
				v, err := self.RedisConn.Do("SET", player.GetUid(), d.Body)
				failOnError(err, "set error")
				log.Printf("set a message:key:%v  %v", player.GetUid(), v)

				v, err = redis.Bytes(self.RedisConn.Do("GET", player.GetUid()))
				failOnError(err, "get error")
				log.Printf("Get a message: %v", v)
			}

		}
	}
ForEnd:

	log.Printf("hello cousumer over")
	log.Println("DBcousumer success")
	return nil
}
