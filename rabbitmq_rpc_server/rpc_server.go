package main

import (
	"errors"
	"golang-project/proto/example"
	"log"

	"github.com/garyburd/redigo/redis"
	"github.com/golang/protobuf/proto"
	"github.com/streadway/amqp"
)

type DBServer struct {
	Uri          string
	ExchangeName string
	QueueName    string
	AmqpConnect  *amqp.Connection
	AmqpChannel  *amqp.Channel
	AmqpQueue    amqp.Queue
	RedisConn    redis.Conn
}

func (self *DBServer) Init(uri string, exchangeName string, queueName string) error {
	self.Uri = uri
	self.ExchangeName = exchangeName
	self.QueueName = queueName

	log.Println("start connect broker")
	var err error
	self.AmqpConnect, err = amqp.Dial(self.Uri)
	if err != nil {
		log.Println("amqp dial error:%v", err)
		return err
	}
	log.Println("connect broker success")

	self.AmqpChannel, err = self.AmqpConnect.Channel()
	if err != nil {
		log.Println("amqp channel error:%v", err)
		return err
	}

	log.Println("AmqpQueue  create")
	self.AmqpQueue, err = self.AmqpChannel.QueueDeclare(self.QueueName, false, false, false, false, nil)
	if err != nil {
		log.Println("amqp Queue Declare error:%v", err)
		return err
	}

	self.RedisConn, err = redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		log.Println("redis dial error:%v", err)
		return err
	}
	return nil
}

func (self *DBServer) Close() {
	self.AmqpConnect.Close()
	self.AmqpChannel.Close()
	self.RedisConn.Close()
}
func (self *DBServer) DBServerStart() error {

	log.Println("DBServerStart enter")

	if self.AmqpConnect == nil {
		err := errors.New("amqp not connect")
		return err
	}

	msgs, err := self.AmqpChannel.Consume(
		self.AmqpQueue.Name, // queue
		"",                  // consumer
		true,                // auto-ack
		true,                // exclusive
		false,               // no-local
		false,               // no-wait
		nil,                 // args
	)
	if err != nil {
		log.Println("Consume fail :", err)
		return err
	}
	for {
		select {
		case d := <-msgs:
			{

				//log.Printf("Received a message: %v", d.Body)
				//log.Printf("Received a message: %v, %v", d.ContentType, d.MessageCount)

				go self.handleMsg(&d)

			}

		}
	}

	return nil
}

func (self *DBServer) handleMsg(msg *amqp.Delivery) {

	//log.Printf("ContentType: %v, ContentEncoding: %v,DeliveryMode: %v, Priority: %v, CorrelationId: %v", msg.ContentType, msg.ContentEncoding, msg.DeliveryMode, msg.Priority, msg.CorrelationId)
	//log.Printf("ReplyTo: %s, Expiration: %s,MessageId: %s, Timestamp: %s, Type: %s, UserId: %s, AppId: %s", msg.ReplyTo, msg.Expiration, msg.MessageId, msg.Timestamp, msg.Type, msg.UserId, msg.AppId)
	//log.Println("msg body:", msg.Body)
	if msg.MessageId == "Set" {
		request := &example.DBSetRequest{}
		err := proto.Unmarshal(msg.Body, request)
		if err != nil {
			log.Printf("proto unmarshal: %v", err)
			self.ResponseMsgErr(msg, err)
			return
		}

		//log.Printf("Set a message: %v", request)
		_, err = self.RedisConn.Do("SET", request.GetUid(), request.GetUidInfo())
		if err != nil {
			log.Printf("set msg result: %v", err)
			self.ResponseMsgErr(msg, err)
			return
		}
		//log.Printf("set msg result: %v", v)
		response := &example.DBSetResponse{}
		response.Uid = proto.Uint64(request.GetUid())
		response.Result = proto.Int32(0)

		data, err := proto.Marshal(response)
		if err != nil {
			log.Fatal("marshaling error: ", err)
			return
		}

		err = self.AmqpChannel.Publish(
			self.ExchangeName, // exchange
			msg.ReplyTo,       // routing key
			false,             // mandatory
			false,             // immediate
			amqp.Publishing{
				Headers:         amqp.Table{},
				ContentType:     "text/plain",
				ContentEncoding: "",
				MessageId:       msg.MessageId,
				Type:            "response",
				Body:            data,
			})
		if err != nil {
			log.Fatal("Publish error: ", err)
			return
		}
		//self.AmqpChannel.Ack(msg.DeliveryTag, false)
		//log.Printf("process over one message")
	}

}

func (self *DBServer) ResponseMsgErr(msg *amqp.Delivery, err error) {
	log.Printf("err, process over one message")
	response := &example.DBSetResponse{}
	response.Uid = proto.Uint64(0)
	response.Result = proto.Int32(1)

	data, err := proto.Marshal(response)
	if err != nil {
		log.Fatal("marshaling error: ", err)
		return
	}

	err = self.AmqpChannel.Publish(
		self.ExchangeName, // exchange
		msg.ReplyTo,       // routing key
		false,             // mandatory
		false,             // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
			ContentEncoding: "",
			MessageId:       msg.MessageId,
			Type:            "response",
			Body:            data,
		})
	if err != nil {
		log.Fatal("Publish error: ", err)
		return
	}
	//log.Printf("process over one message")
}

func main() {
	var err error
	var dbserver DBServer
	dbserver.Init("amqp://guest:guest@localhost:5672/", "", "rpc_queue")

	err = dbserver.DBServerStart()
	if err != nil {
		log.Println("call fail :", err)
	}
}
