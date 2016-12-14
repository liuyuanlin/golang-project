package main

import (
	"fmt"
	"golang-project/dpsg/common"
	"golang-project/dpsg/lockserver"
	"golang-project/dpsg/logger"
	"net"
	"os"
	"syscall"
)

func main() {
	var lsConfig common.LockServerCfg
	if err := common.ReadLockServerConfig(&lsConfig); err != nil {
		logger.Error("load config failed, error is: %v", err)
		return
	}

	common.DebugInit(lsConfig.GcTime, lsConfig.DebugHost)

	quitChan := make(chan int)

	listener, err := net.Listen("tcp", lsConfig.LockHost)
	defer listener.Close()
	if err != nil {
		println("Listening to: ", lsConfig.LockHost, " failed !!")
		return
	}
	println("Listening to: ", lsConfig.LockHost, "Success !!")

	//go func() { log.Println(http.ListenAndServe(lsConfig.LockHost, nil)) }()

	go lockserver.CreateServices(listener)

	handler := func(s os.Signal, arg interface{}) {
		fmt.Printf("handle signal: %v\n", s)
		println("logserver close")
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

		if nQuitCount == 2 {
			break
		}
	}

	println("lockserver close")

}
