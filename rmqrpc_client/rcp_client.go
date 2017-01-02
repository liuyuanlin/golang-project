package main

import (
	//	"errors"
	"log"
	"net/rpc"
	//	"time"

	"github.com/liuyuanlin/rmqrpc"

	msg "github.com/liuyuanlin/rmqrpc/examples/message.pb"
)

func main() {
	client := rpc.NewClientWithCodec(rmqrpc.NewClientCodec("amqp://guest:guest@localhost:5672/", "", "rpc_queue"))
	var args msg.ArithRequest
	var reply msg.ArithResponse
	var err error
	log.Println("main: step 1")
	// Add
	args.A = 1
	args.B = 2
	if err = client.Call("ArithService.Add", &args, &reply); err != nil {
		log.Printf(`arith.Add: %v`, err)
	}
	if reply.C != 3 {
		log.Printf(`arith.Add: expected = %d, got = %d`, 3, reply.C)
	}
	log.Println("main: step 2")
	// Mul
	args.A = 2
	args.B = 3
	if err = client.Call("ArithService.Mul", &args, &reply); err != nil {
		log.Printf(`arith.Mul: %v`, err)
	}
	if reply.C != 6 {
		log.Printf(`arith.Mul: expected = %d, got = %d`, 6, reply.C)
	}
	log.Println("main: step 3")
	// Div
	args.A = 13
	args.B = 5
	if err = client.Call("ArithService.Div", &args, &reply); err != nil {
		log.Printf(`arith.Div: %v`, err)
	}
	if reply.C != 2 {
		log.Printf(`arith.Div: expected = %d, got = %d`, 2, reply.C)
	}
	log.Println("main: step 4")
	// Div zero
	args.A = 1
	args.B = 0
	if err = client.Call("ArithService.Div", &args, &reply); err.Error() != "divide by zero" {
		log.Printf(`arith.Error: expected = "%s", got = "%s"`, "divide by zero", err.Error())
	}
	log.Println("main: step 5")
	// Error
	args.A = 1
	args.B = 2
	if err = client.Call("ArithService.Error", &args, &reply); err.Error() != "ArithError" {
		log.Printf(`arith.Error: expected = "%s", got = "%s"`, "ArithError", err.Error())
	}

}
