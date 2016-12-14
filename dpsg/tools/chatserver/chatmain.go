package main

import (
	"golang-project/dpsg/chatserver"
	"golang-project/dpsg/common"
	//	"log"
	"net"
	"os"
	//"strconv"
	"fmt"
	"syscall"
)

var ipcfg common.ChatServerCfg

func main() {

	if err := common.ReadChatConfig(&ipcfg); err != nil {
		return
	}

	common.DebugInit(ipcfg.GcTime, ipcfg.DebugHost)

	quitChan := make(chan int)

	listenerForClient, err := net.Listen("tcp", ipcfg.ListenForClient)
	defer listenerForClient.Close()
	if err != nil {
		println("Listening to: ", ipcfg.ListenForClient, " failed !!")
		return
	}

	println("ChatServer Listening to: ", ipcfg.ListenForClient)

	listenerForServer, err := net.Listen("tcp", ipcfg.ListenForServer)
	defer listenerForServer.Close()
	if err != nil {
		println("Listening to: ", ipcfg.ListenForServer, " failed !!")
		return
	}
	println("ChatServer Listening to: ", ipcfg.ListenForServer)

	listenerForGm, err := net.Listen("tcp", ipcfg.ListenForGm)
	defer listenerForGm.Close()
	if err != nil {
		println("Listening to: ", ipcfg.ListenForGm, " failed !!")
		return
	}
	println("ChatServer Listening to: ", ipcfg.ListenForGm)

	go chatserver.CreateChatServicesForCnserver(listenerForServer)
	go chatserver.CreateChatServicesForClient(listenerForClient)
	go chatserver.CreateChatServicesForGm(listenerForGm)
	go chatserver.CreateMailServices(ipcfg)

	handler := func(s os.Signal, arg interface{}) {
		fmt.Printf("handle signal: %v\n", s)
		println("chatserver close")
		os.Exit(0)
	}

	handlerArray := []os.Signal{syscall.SIGINT,
		syscall.SIGILL,
		syscall.SIGFPE,
		syscall.SIGSEGV,
		syscall.SIGTERM,
		syscall.SIGABRT}

	common.WatchSystemSignal(&handlerArray, handler)

	nQuitCount := 0
	for {
		select {
		case <-quitChan:
			nQuitCount = nQuitCount + 1
		}

		//println("nQuitCount = %s", strconv.Itoa(nQuitCount))
		if nQuitCount == 2 {
			break
		}
	}

	println("chatserver close")

}
