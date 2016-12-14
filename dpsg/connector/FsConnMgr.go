package connector

import (
	"fmt"
	"golang-project/dpsg/common"
	"golang-project/dpsg/logger"
	"golang-project/dpsg/rpc"
	"net"
	"os"
	"runtime/debug"
	"strings"
	"sync"
)

type FServerConnMgr struct {
	poollock  sync.RWMutex
	connpool  []rpc.RpcConn
	poolsize  uint8
	workindex int8
	poolid    uint8
	fightjob  chan interface{}
	quit      chan bool
}

func (self *FServerConnMgr) GetConn() rpc.RpcConn {
	return self.connpool[self.workindex]
}

func (self *FServerConnMgr) Open(poolsize uint8) {
	self.poolid = 0
	self.workindex = -1
	self.poolsize = poolsize
	self.connpool = make([]rpc.RpcConn, poolsize)
	self.quit = make(chan bool)

}

func (self *FServerConnMgr) GetNewConnId() uint8 {
	self.poollock.Lock()
	defer self.poollock.Unlock()

	self.poolid++
	return self.poolid - 1
}

func (self *FServerConnMgr) NewConn(conn rpc.RpcConn, uConnId uint8) {
	self.poollock.Lock()
	defer self.poollock.Unlock()

	self.connpool[uConnId] = conn
}

func (self *FServerConnMgr) GetWorkConn() rpc.RpcConn {
	self.poollock.Lock()
	defer self.poollock.Unlock()

	self.workindex++
	if uint8(self.workindex) >= self.poolsize {
		self.workindex = 0
	}

	return self.connpool[self.workindex]
}

func (self *FServerConnMgr) SendFightJob(arg interface{}) error {
	return self.GetWorkConn().WriteObj(arg)
}

func (self *FServerConnMgr) Quit() {
	self.poollock.Lock()
	defer self.poollock.Unlock()
	for i, v := range self.connpool {

		logger.Info("ShutDown FServerConnMgr -----> %d", i)
		v.Close()
		self.quit <- true

	}
}

func (self *FServerConnMgr) Init(server *rpc.Server, cfg *common.CnsConfig) {

	fsCount := len(cfg.FsHost)
	self.Open(uint8(fsCount))

	for i := 0; i < fsCount; i++ {
		go func() {

			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("FServerConnMgr runtime error:", r)

					debug.PrintStack()
				}
			}()

			connId := self.GetNewConnId()
			host := cfg.FsHost[connId]

			param := strings.Split(host, ":")

			for {
				select {
				case <-self.quit:
					{
						logger.Info("FServerConnMgr Goroutine Quit ----->")
						return
					}
				default:
					{
						args := make([]string, 2)
						args[0] = param[0]
						args[1] = param[1]

						//fmt.Println("send data : ", param[0], param[1])

						_, ret := os.StartProcess("./FightServer", args, &os.ProcAttr{Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}})
						if ret != nil {
							logger.Fatal("Start FightServer Error :%s", ret.Error())
						}

						var err error
						var fsConn net.Conn

						for {
							fsConn, err = net.Dial("tcp", host)
							if err != nil {
								//logger.Fatal("Connect FightServer Error :%s", err.Error())
							} else {
								break
							}
						}

						logger.Info("Connect to FightServer : %s ok!!!!", host)

						fsRpcConn := rpc.NewProtoBufConn(server, fsConn, 1000, 0)
						self.NewConn(fsRpcConn, connId)

						server.ServeConn(fsRpcConn)
					}
				}
			}
		}()

	}
}
