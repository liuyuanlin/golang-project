package main

import (
	"errors"
	"log"
	"net/rpc"

	"rmqrpc"

	msg "rmqrpc/examples/message.pb"
)

type Arith int

func (t *Arith) Add(args *msg.ArithRequest, reply *msg.ArithResponse) error {
	reply.C = args.A + args.B
	log.Printf("Arith.Add(%v, %v): %v", args.A, args.B, reply.C)
	return nil
}

func (t *Arith) Mul(args *msg.ArithRequest, reply *msg.ArithResponse) error {
	reply.C = args.A * args.B
	log.Printf("Arith.Mul(%v, %v): %v", args.A, args.B, reply.C)
	return nil
}

func (t *Arith) Div(args *msg.ArithRequest, reply *msg.ArithResponse) error {
	if args.B == 0 {
		return errors.New("divide by zero")
	}
	reply.C = args.A / args.B
	log.Printf("Arith.Div(%v, %v): %v", args.A, args.B, reply.C)
	return nil
}

func (t *Arith) Error(args *msg.ArithRequest, reply *msg.ArithResponse) error {
	log.Printf("Arith.Error(%v, %v): %v", args.A, args.B, reply.C)
	return errors.New("ArithError")

}

func main() {
	srv := rpc.NewServer()
	if err := srv.RegisterName("ArithService", new(Arith)); err != nil {
		return
	}

	go srv.ServeCodec(rmqrpc.NewServerCodec("amqp://guest:guest@localhost:5672/", "", "rpc_queue"))

	f := make(chan int)
	<-f
}
