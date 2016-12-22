package main

import (
	"errors"
	"golang-project/proto/example"
	"log"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/streadway/amqp"
)

type DBClient struct {
	Uri          string
	ExchangeName string
	QueueName    string
	AmqpConnect  *amqp.Connection
	AmqpChannel  *amqp.Channel
	AmqpQueue    amqp.Queue
}

func (self *DBClient) Init(uri string, exchangeName string, queueName string) error {
	self.Uri = uri
	self.ExchangeName = exchangeName
	self.QueueName = queueName

	//log.Println("start connect broker")
	var err error
	self.AmqpConnect, err = amqp.Dial(self.Uri)
	if err != nil {
		log.Println("amqp dial error:%v", err)
		return err
	}
	//log.Println("connect broker success")

	self.AmqpChannel, err = self.AmqpConnect.Channel()
	if err != nil {
		log.Println("amqp channel error:%v", err)
		return err
	}

	//log.Println("AmqpQueue  create")
	self.AmqpQueue, err = self.AmqpChannel.QueueDeclare(self.QueueName, false, false, false, false, nil)
	if err != nil {
		log.Println("amqp Queue Declare error:%v", err)
		return err
	}

	return nil
}

func (self *DBClient) Close() {
	self.AmqpConnect.Close()
	self.AmqpChannel.Close()
}
func (self *DBClient) DBWrite(request *example.DBSetRequest) (*example.DBSetResponse, error) {

	//log.Println("DBWrite enter")

	if self.AmqpConnect == nil {
		err := errors.New("amqp not connect")
		return nil, err
	}

	data, err := proto.Marshal(request)
	if err != nil {
		log.Fatal("marshaling error: ", err)
		return nil, err
	}
	//log.Printf("declared queue, publishing %dB body (%v)", len(data), data)

	temQuene, err := self.AmqpChannel.QueueDeclare("", false, false, true, false, nil)
	//log.Printf("temQuene name:%s", temQuene.Name)

	err = self.AmqpChannel.Publish(
		self.ExchangeName,   // exchange
		self.AmqpQueue.Name, // routing key
		false,               // mandatory
		false,               // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
			ContentEncoding: "",
			MessageId:       "Set",
			Type:            "Request",
			ReplyTo:         temQuene.Name,
			Body:            data,
		})
	if err != nil {
		log.Fatal("Publish error: ", err)
		return nil, err
	}

	msgs, err := self.AmqpChannel.Consume(
		temQuene.Name, // queue
		"",            // consumer
		true,          // auto-ack
		true,          // exclusive
		false,         // no-local
		false,         // no-wait
		nil,           // args
	)
	if err != nil {
		log.Fatal("Consume error: ", err)
		return nil, err
	}
	//log.Printf("Consume temQuene name:%s", temQuene.Name)
	response := &example.DBSetResponse{}
	for {
		select {
		case <-time.After(time.Second * time.Duration(2)):
			{
				log.Printf("get msg timer out")
				response = nil
				goto ForEnd

			}
		case d := <-msgs:
			{

				//log.Printf("Received a message: %v", d.Body)
				//log.Printf("Received a message: %v, %v", d.ContentType, d.MessageCount)
				err := proto.Unmarshal(d.Body, response)
				if err != nil {
					log.Printf("proto unmarshal: %v", err)
					response = nil
				} else {
					//log.Printf("Unmarshal a message: %v", response)
					//log.Println("DBWrite success")
				}

				goto ForEnd

			}

		}
	}
ForEnd:
	self.AmqpChannel.QueueDelete(temQuene.Name, false, false, true)
	return response, nil
}

func WritPlayer(dbclient *DBClient, uid uint64) {
	var err error
	player := &example.PlayerInfo{}
	player.Uid = proto.Uint64(uid)
	player.Age = proto.Int32(28)
	player.Name = proto.String("robot1")
	player.PhoneNum = proto.String("18030761111")

	request := &example.DBSetRequest{}
	request.Uid = proto.Uint64(uid)
	request.UidInfo, err = proto.Marshal(player)
	if err != nil {
		log.Println("Marshal fail")
	}

	//response, lerr := dbclient.DBWrite(request)
	_, lerr := dbclient.DBWrite(request)
	if lerr != nil {
		log.Println("call fail :", lerr)
	} else {
		//log.Println("call reuslt:", response)
	}
}

func main() {
	var dbclient DBClient
	dbclient.Init("amqp://guest:guest@localhost:5672/", "", "rpc_queue")
	defer dbclient.Close()
	t := time.Now().UnixNano()
	log.Println(t)
	for i := 0; i < 1000; i++ {
		WritPlayer(&dbclient, uint64(i))
	}
	t2 := time.Now().UnixNano()
	log.Println(t2)
	log.Println(t2 - t)
	l := make(chan int)
	<-l
}
