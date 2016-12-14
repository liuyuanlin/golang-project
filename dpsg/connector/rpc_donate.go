package connector

import (
	"golang-project/dpsg/logger"
	"golang-project/dpsg/proto"
	"golang-project/dpsg/rpc"
)

//发起捐兵申请
func (self *CNServer) RequestDonate(conn rpc.RpcConn, ping rpc.Ping) error {
	ts("CNServer:RequestDonate", conn.GetId())
	defer te("CNServer:RequestDonate", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	if p.GetClan() == "" {
		return nil
	}

	p.v.castle_RequestDonate()

	return nil
}

//捐兵
func (self *CNServer) Donate(conn rpc.RpcConn, c2sDonate rpc.C2SDonate) error {
	ts("CNServer:Donate", conn.GetId())
	defer te("CNServer:Donate", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	if p.GetClan() == "" {
		return nil
	}

	if p.GetUid() == c2sDonate.GetUid() {
		return nil
	}

	Donate(p, c2sDonate)

	return nil
}

//来自Center的通知
func (s *CenterService) NotifyGetDonate(req *proto.NotifyGetDonate, reply *proto.NotifyGetDonateResult) (err error) {
	logger.Info("CenterService.NotifyGetDonate:%s, %s", req.Uid, req.Name)

	cns.l.RLock()
	p, exist := cns.playersbyid[req.Uid]
	cns.l.RUnlock()

	if !exist {
		return nil
	}

	if p.conn != nil {
		p.conn.Lock()
	}

	defer func() {
		if p.conn != nil {
			p.conn.Unlock()
		}
	}()

	p.v.castle_GetDonate(req)

	return nil
}
