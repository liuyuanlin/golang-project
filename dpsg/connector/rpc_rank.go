package connector

import (
	//gp "github.com/golang/protobuf/proto"
	"golang-project/dpsg/logger"
	"golang-project/dpsg/proto"
	"golang-project/dpsg/rpc"
	"strconv"
	//"fmt"
)

func (self *CNServer) GetRankPlayers(conn rpc.RpcConn, msg rpc.TryGetRankPlayers) error {
	ts("CNServer:GetRankPlayers", conn.GetId())
	defer te("CNServer:GetRankPlayers", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()
	if !exist {
		return nil
	}

	p.myselfGlobal = &rpc.RankPlayers{}
	result := &rpc.RankPlayers{}

	for _, value := range self.rankMgr.SaveRankGlobalPlayers.RpsTop {
		if value.GetUid() == p.GetUid() {
			result.RpsTop = append(result.RpsTop, value)
			p.myselfGlobal = result
		}
	}

	if p.myselfGlobal == nil {
		//玩家第一次请求的时候，去数据库查一下自己的排名
		mytry := &proto.GetMyself{Uid: p.GetUid()}
		myret := &proto.GetMyselfResult{}
		err := cns.center.Call("Center.GetRankMyselfGlobal", mytry, myret)
		if err != nil {
			logger.Error("Call Center.GetRankMyselfGlobal error", err.Error())
			//return nil
		} else {
			rp := rpc.Player{}
			rp.SetType(rpc.Player_Rank)
			rp.SetRank(uint32(myret.Rank))
			rp.SetName(p.GetName())
			rp.SetUid(p.GetUid())
			rp.SetTrophy(p.GetTrophy())
			rp.SetLevel(p.GetLevel())
			rp.SetClanName(p.GetClan())
			rp.SetClanSymbol(p.GetClanSymbol())

			result.RpsTop = append(result.RpsTop, &rp)
			p.myselfGlobal = result
		}
	}

	result.RpsTop = append(result.RpsTop, self.rankMgr.SaveRankGlobalPlayers.RpsTop...)

	self.rankMgr.GlobalPlayersLock.RLock()
	WriteResult(conn, result)
	self.rankMgr.GlobalPlayersLock.RUnlock()

	return nil
}

//查询区域玩家

func (self *CNServer) GetRankPlayersLocation(conn rpc.RpcConn, msg rpc.TryGetRankPlayers) error {
	ts("CNServer:GetRankPlayersLocation", conn.GetId())
	defer te("CNServer:GetRankPlayersLocation", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()
	if !exist {
		return nil
	}

	p.myselfLocation = &rpc.RankPlayers{}
	result := &rpc.RankPlayers{}

	if _, ok := self.rankMgr.SaveRankPlayersLocation[int64(p.GetGamelocation())]; ok {
		for _, myvalue := range self.rankMgr.SaveRankPlayersLocation[int64(p.GetGamelocation())].RpsTop {
			if myvalue.GetUid() == p.GetUid() {
				//myvalue.SetRank(uint32(index))
				result.RpsTop = append(result.RpsTop, myvalue)
				p.myselfLocation = result
			}
		}
	}

	if p.myselfLocation == nil {
		//玩家第一次请求的时候，去数据库查一下自己的排名
		mytry := &proto.GetMyself{Uid: p.GetUid(), Location: strconv.Itoa(int(p.GetGamelocation()))}
		myret := &proto.GetMyselfResult{}

		err := cns.center.Call("Center.GetRankMyselfLocation", mytry, myret)
		if err != nil {
			logger.Error("Call Center.GetRankMyselfLocation error", err.Error())
			//return nil
		} else {
			rp := rpc.Player{}
			rp.SetType(rpc.Player_Rank)
			rp.SetRank(uint32(myret.Rank))
			rp.SetName(p.GetName())
			rp.SetUid(p.GetUid())
			rp.SetTrophy(p.GetTrophy())
			rp.SetLevel(p.GetLevel())
			rp.SetClanName(p.GetClan())
			rp.SetClanSymbol(p.GetClanSymbol())

			result.RpsTop = append(result.RpsTop, &rp)
			p.myselfLocation = result
		}
	}

	self.rankMgr.PlayersLocationLock.RLock()
	if _, ok := self.rankMgr.SaveRankPlayersLocation[int64(p.GetGamelocation())]; ok {
		result.RpsTop = append(result.RpsTop, self.rankMgr.SaveRankPlayersLocation[int64(p.GetGamelocation())].RpsTop...)
	}
	WriteResult(conn, result)
	self.rankMgr.PlayersLocationLock.RUnlock()

	return nil
}

func (self *CNServer) GetRankClans(conn rpc.RpcConn, msg rpc.TryGetClans) error {
	ts("CNServer:GetRankClans", conn.GetId())
	defer te("CNServer:GetRankClans", conn.GetId())

	self.l.RLock()
	_, exist := self.players[conn.GetId()]
	self.l.RUnlock()
	if !exist {
		return nil
	}

	WriteResult(conn, &self.rankMgr.SaveRankClans)
	self.rankMgr.RankClansLock.RUnlock()
	return nil
}

//通天塔

func (self *CNServer) GetTTTRankPlayers(conn rpc.RpcConn, msg rpc.TryGetClans) error {
	ts("CNServer:GetTTTPlayers", conn.GetId())
	defer te("CNServer:GetTTTPlayers", conn.GetId())

	self.l.RLock()
	_, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	self.rankMgr.TTTPlayersLock.RLock()
	WriteResult(conn, &self.rankMgr.SaveTTTPlayers)
	self.rankMgr.TTTPlayersLock.RUnlock()
	return nil
}
