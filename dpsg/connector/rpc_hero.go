package connector

import (
	"golang-project/dpsg/logger"
	"golang-project/dpsg/rpc"
)

func (self *CNServer) HeroCreate(conn rpc.RpcConn, msg rpc.HeroChoose) error {
	ts("CNServer:HeroCreate", conn.GetId())
	defer te("CNServer:HeroCreate", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()
	if !exist {
		return nil
	}

	if !p.v.hero_create(msg.GetIdx(), msg.GetType()) {
		logger.Error("!!!!!!!HeroCreate Failed!!!!!")
	}

	return nil
}

func (self *CNServer) HeroUpgrade(conn rpc.RpcConn, msg rpc.HeroChoose) error {
	ts("CNServer:HeroUpgrade", conn.GetId())
	defer te("CNServer:HeroUpgrade", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()
	if !exist {
		return nil
	}

	if !p.v.hero_upgrade(msg.GetIdx(), msg.GetType()) {
		logger.Error("!!!!!!!HeroUpgrade Failed!!!!!")
	}

	return nil
}

func (self *CNServer) HeroFinishNow(conn rpc.RpcConn, msg rpc.HeroChoose) error {
	ts("CNServer:HeroFinishNow", conn.GetId())
	defer te("CNServer:HeroFinishNow", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()
	if !exist {
		return nil
	}

	if !p.v.hero_FinishNow(msg.GetIdx(), msg.GetType()) {
		logger.Error("!!!!!!!HeroFinishNow Failed!!!!!")
	}

	return nil
}

func (self *CNServer) HeroChoose(conn rpc.RpcConn, msg rpc.HeroChoose) error {
	ts("CNServer:GetRankPlayers", conn.GetId())
	defer te("CNServer:GetRankPlayers", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()
	if !exist {
		return nil
	}

	if !p.v.hero_choose(msg.GetIdx(), msg.GetType()) {
		logger.Error("!!!!!!!HeroChoose Failed!!!!!")
	}

	return nil
}
