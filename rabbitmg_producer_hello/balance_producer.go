package main

import (
	"errors"
	"golang-project/proto/example"
	"log"

	"github.com/golang/protobuf/proto"
	"github.com/streadway/amqp"
)

type BalanceProducer struct {
	Uri          string
	ExchangeName string
	QueueName    string
	Players      map[uint64]*example.PlayerInfo
	AmqpConnect  *amqp.Connection
	AmqpQueue    amqp.Queue
}

func (self *BalanceProducer) Run(uri string, exchangeName string, queueName string) error {
	self.Players = make(map[uint64]*example.PlayerInfo)
	self.Uri = uri
	self.ExchangeName = exchangeName
	self.QueueName = queueName

	player := &example.PlayerInfo{}
	player.Uid = proto.Uint64(1001)
	player.Age = proto.Int32(28)
	player.Name = proto.String("robot1")
	player.PhoneNum = proto.String("18030761111")

	self.Players[player.GetUid()] = player
	log.Println("start connect broker")
	var err error
	self.AmqpConnect, err = amqp.Dial(self.Uri)
	if err != nil {
		log.Println("amqp dial error:%v", err)
		return err
	}
	defer self.AmqpConnect.Close()
	log.Println("connect broker success")
	self.DBWrite(player.GetUid())
	return nil

}

func (self *BalanceProducer) DBWrite(uid uint64) error {

	log.Println("DBWrite enter")
	playerinfo, ok := self.Players[uid]
	if !ok {
		log.Println("%d not a key in players map", uid)
		err := errors.New("no palyer info")
		return err
	}

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

	if self.AmqpQueue.Name == "" {
		log.Println("AmqpQueue  create")
		self.AmqpQueue, err = channel.QueueDeclare(self.QueueName, false, false, false, false, nil)
		if err != nil {
			log.Println("amqp Queue Declare error:%v", err)
			return err
		}
	}

	data, err := proto.Marshal(playerinfo)
	if err != nil {
		log.Fatal("marshaling error: ", err)
		return err
	}
	log.Printf("declared queue, publishing %dB body (%v)", len(data), data)

	err = channel.Publish(
		self.ExchangeName,   // exchange
		self.AmqpQueue.Name, // routing key
		false,               // mandatory
		false,               // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
			ContentEncoding: "",
			Body:            data,
		})
	if err != nil {
		log.Fatal("Publish error: ", err)
		return err
	}

	log.Println("DBWrite success")
	return nil
}
