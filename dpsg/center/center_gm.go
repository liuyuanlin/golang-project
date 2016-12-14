package center

import (
	"golang-project/dpsg/lockclient"
	"golang-project/dpsg/logger"
	"golang-project/dpsg/proto"
	"golang-project/dpsg/rpcplus"
	"net"
)

type CenterGmServices struct {
}

var pCenterGmServices *CenterGmServices

func CreateCenterServiceForGM(listener net.Listener) {
	pCenterGmServices = &CenterGmServices{}

	rpcServer := rpcplus.NewServer()
	rpcServer.Register(pCenterGmServices)

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("centerservergm StartServices %s", err.Error())
			break
		}
		go func() {
			rpcServer.ServeConn(conn)
			conn.Close()
		}()
	}
}

//锁定
func (self *CenterGmServices) GmLockPlayer(req *proto.TryGetLock, rep *proto.GetLockResult) (err error) {
	logger.Info("GmLockPlayer:%s", req.Name)

	rep.Result, rep.OldValue, err = lockclient.TryLock(req.Service, req.Name, req.Value)

	return
}

//解锁
func (self *CenterGmServices) GmUnLockPlayer(req *proto.FreeLock, rep *proto.FreeLockResult) (err error) {
	logger.Info("GmUnLockPlayer:%s", req.Name)

	rep.Result, err = lockclient.ForceUnLock(req.Service, req.Name)

	return
}

//取在线玩家数量，现在未区分地区
func (self *CenterGmServices) GmGetOnlineNum(req *proto.GmGetOnlineNum, rst *proto.GmGetOnlineNumResult) error {
	rst.Value = uint32(len(centerServer.POnline))

	return nil
}
