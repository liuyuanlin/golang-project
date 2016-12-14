package gateserver

import (
	"fmt"
	"golang-project/dpsg/logger"
	//	"math/rand"
	"golang-project/dpsg/proto"
	"golang-project/dpsg/rpc"
	"golang-project/dpsg/rpcplus"
	"net"
	//	"strconv"
	"golang-project/dpsg/common"
	"sync"
	"time"
)

type serverInfo struct {
	PlayerCount uint16
	ServerIp    string
}

type GateServices struct {
	l            sync.RWMutex
	m            map[uint32]serverInfo
	stableServer string
}

var pGateServices *GateServices

func CreateGateServicesForCnserver(listener net.Listener) *GateServices {
	pGateServices = &GateServices{m: make(map[uint32]serverInfo)}
	rpcServer := rpcplus.NewServer()

	rpcServer.Register(pGateServices)

	//rpcServer.HandleHTTP("/center/rpc", "/debug/rpcdebug/rpc")

	var uConnId uint32 = 0
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("gateserver StartServices %s", err.Error())
			break
		}

		uConnId++
		go func(uConnId uint32) {

			pGateServices.l.Lock()
			pGateServices.m[uConnId] = serverInfo{0, ""}
			pGateServices.l.Unlock()

			rpcServer.ServeConnWithContext(conn, uConnId)

			pGateServices.l.Lock()
			delete(pGateServices.m, uConnId)
			pGateServices.l.Unlock()

		}(uConnId)
	}

	return pGateServices
}

func (self *GateServices) UpdateCnsPlayerCount(uConnId uint32, info *proto.SendCnsInfo, result *proto.SendCnsInfoResult) error {
	self.l.Lock()
	self.m[uConnId] = serverInfo{info.PlayerCount, info.ServerIp}

	playerCountMax := uint16(0xffff) //不会有哪个服务器更大吧
	self.stableServer = ""
	for _, v := range self.m {
		if len(v.ServerIp) > 0 && v.PlayerCount < playerCountMax {
			playerCountMax = v.PlayerCount
			self.stableServer = v.ServerIp
		}
	}

	self.l.Unlock()

	//fmt.Printf("recv cns msg : server %d , player count %d, player ip = %s \n", info.ServerId, info.PlayerCount, info.ServerIp)
	return nil
}

func (self *GateServices) getStableCns() (cnsIp string) {
	self.l.RLock()
	defer self.l.RUnlock()
	return self.stableServer
}

type GateServicesForClient struct {
	m string
}

var gateServicesForClient *GateServicesForClient

func CreateGateServicesForClient(listener net.Listener) *GateServicesForClient {

	gateServicesForClient = &GateServicesForClient{}
	rpcServer := rpc.NewServer()
	rpcServer.Register(gateServicesForClient)

	rpcServer.RegCallBackOnConn(
		func(conn rpc.RpcConn) {
			gateServicesForClient.onConn(conn)
		},
	)

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("gateserver StartServices %s", err.Error())
			break
		}
		go func() {
			rpcConn := rpc.NewProtoBufConn(rpcServer, conn, 4, 0)
			rpcServer.ServeConn(rpcConn)
		}()
	}

	return gateServicesForClient
}

func WriteResult(conn rpc.RpcConn, value interface{}) bool {
	err := conn.WriteObj(value)
	if err != nil {
		logger.Info("WriteResult Error %s", err.Error())
		return false
	}
	return true
}

func (c *GateServicesForClient) onConn(conn rpc.RpcConn) {
	rep := rpc.LoginCnsInfo{}

	cnsIp := pGateServices.getStableCns()
	rep.CnsIp = &cnsIp
	gasinfo := fmt.Sprintf("%s;%d", conn.GetRemoteIp(), time.Now().Unix())
	logger.Info("Client(%s) -> CnServer(%s)", conn.GetRemoteIp(), cnsIp)
	// encode
	encodeInfo := common.Base64Encode([]byte(gasinfo))

	gasinfo = fmt.Sprintf("%s;%s", gasinfo, encodeInfo)

	//fmt.Printf("%s \n", gasinfo)

	rep.GsInfo = &gasinfo

	WriteResult(conn, &rep)

	time.Sleep(10 * time.Second)
	conn.Close()
}
