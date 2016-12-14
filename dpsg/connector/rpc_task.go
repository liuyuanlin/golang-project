package connector

import (
	"golang-project/dpsg/rpc"
	//"logger"
)

func (self *CNServer) UpdateTaskInfo(conn rpc.RpcConn, info rpc.UpdateTaskInfo) error {
	//ts("CNServer:UpdateTaskInfo", conn.GetId())
	//defer te("CNServer:UpdateTaskInfo", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	p.UpdateTaskInfo(&info)

	return nil
}

func (self *CNServer) GetTaskReward(conn rpc.RpcConn, tryget rpc.TryGetTaskReward) error {
	//ts("CNServer:GetTaskReward", conn.GetId())
	//defer te("CNServer:GetTaskReward", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	p.GetTaskReward(&tryget)

	return nil
}

func (self *CNServer) UserguideFinish(conn rpc.RpcConn, cmd rpc.Ping) error {
	ts("CNServer:UserguideFinish", conn.GetId())
	defer te("CNServer:UserguideFinish", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	p.SetIsUserguideFinish(true)
	p.Save()

	return nil
}

//分享
func (self *CNServer) ShareFinish(conn rpc.RpcConn, cmd rpc.ShareFinish) error {
	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	p.shareFinish(conn, &cmd)
	p.Save()

	return nil
}

//连续登陆奖励
func (self *CNServer) LandedReceiveAward(conn rpc.RpcConn, cmd rpc.LandedReceiveAward) error {
	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	p.LandedReceiveAward(conn, &cmd)
	p.Save()

	return nil
}
