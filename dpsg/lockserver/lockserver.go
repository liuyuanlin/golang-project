package lockserver

import (
	"fmt"
	"golang-project/dpsg/rpcplus"
	"net"
	//"time"
)

type LockServerServices struct {
	Services *LockServices
}

var pLockServices *LockServerServices

func CreateServices(listener net.Listener) *LockServerServices {

	pLockServices = &LockServerServices{Services: NewLockServices()}
	rpcServer := rpcplus.NewServer()
	rpcServer.Register(pLockServices.Services)

	for {
		conn, err := listener.Accept()

		if err != nil {
			fmt.Printf("StartServices %s \n", err.Error())
			break
		}

		//开始对cns的RPC服务
		go func() {
			rpcServer.ServeConn(conn)
		}()
	}

	return pLockServices
}
