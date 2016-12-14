package center

import (
	"golang-project/dpsg/logger"
	"golang-project/dpsg/proto"
	"golang-project/dpsg/rpc"
	"strconv"

	gp "github.com/golang/protobuf/proto"
	//add for update rankplayers
	"golang-project/dpsg/common"
	"golang-project/dpsg/dbclient"
	"golang-project/dpsg/rpcplus"
	"golang-project/dpsg/timer"
	"time"
)

type SaveRankResult struct {
	RankPlayers         proto.SaveRankPlayer
	RankLocationPlayers proto.SaveRankLocationPlayers
	RankClans           proto.SaveRankClan
	RankTTTplayers      proto.SaveTTTRank
	RankMyself          proto.SaveMyself
}

var SaveAllRankPlayers SaveRankResult

//第一次推送，看看那个cns连接到我了,就推送一次
func (self *Center) theFirstUpdate(conn *rpcplus.Client) {

	RankPlayerReply := &proto.SaveRankPlayerResult{}
	RankLocationPlayersReply := &proto.SaveRankLocationPlayersResult{}
	RankClansReply := &proto.SaveRankClanResult{}
	RankTTTplayersReply := &proto.SaveTTTRankResult{}

	conn.Call("CenterService.SaveRankPlayers", SaveAllRankPlayers.RankPlayers, RankPlayerReply)
	if !RankPlayerReply.OK {
		logger.Error("CenterService.SaveRankPlayers error!")
	}

	conn.Call("CenterService.SaveRankLocationPlayers", SaveAllRankPlayers.RankLocationPlayers, RankLocationPlayersReply)
	if !RankLocationPlayersReply.OK {
		logger.Error("CenterService.SaveRankLocationPlayers error!")
	}

	conn.Call("CenterService.SaveRankClans", SaveAllRankPlayers.RankClans, RankClansReply)
	if !RankClansReply.OK {
		logger.Error("CenterService.SaveRankClans error!")
	}

	conn.Call("CenterService.SaveTTTPlayers", SaveAllRankPlayers.RankTTTplayers, RankTTTplayersReply)
	if !RankTTTplayersReply.OK {
		logger.Error("CenterService.SaveTTTPlayers error!")
	}
}

func (self *Center) initUpdateRankPlayers() {
	//开始定时器

	var centercfg common.CenterConfig
	if err := common.ReadCenterConfig(&centercfg); err != nil {
		return
	}

	updatetime, _ := strconv.Atoi(centercfg.UpdateTime)

	logger.Info("现在的排行榜自动推送时间是 : ", time.Duration(updatetime)*time.Hour)

	self.updatetime = timer.NewTimer(time.Duration(updatetime) * time.Hour)
	self.updatetime.Start(
		func() {
			self.updateAllRankplayers(false)
		},
	)

	//立即调用一次更新缓存
	self.updateAllRankplayers(true)
}

func (self *Center) updateAllRankplayers(bIsFirst bool) {

	self.updateRankplayers(bIsFirst)
	self.updateRankplayersLocation(bIsFirst)
	self.updateRankClans(bIsFirst)
	self.updateTTTplayers(bIsFirst)
}

func (self *Center) updateRankplayers(bIsFirst bool) {

	var req proto.GetRankPlayers
	var reply proto.GetRankPlayersResult
	req.Start = 0
	req.Stop = 99
	self.getRankPlayers(&req, &reply, bIsFirst)
}

func (self *Center) updateRankplayersLocation(bIsFirst bool) {

	var req proto.GetRankPlayersLocation
	var reply proto.GetRankPlayersLocationResult
	req.Start = 0
	req.Stop = 99

	for _, value := range rpc.GameLocation_value {

		req.Location = int64(value)
		self.getRankPlayersLocation(&req, &reply, bIsFirst)
	}
}

func (self *Center) updateRankClans(bIsFirst bool) {

	var req proto.GetRankClans
	var reply proto.GetRankClansResult
	req.Start = 0
	req.Stop = 99
	self.getRankClans(&req, &reply, bIsFirst)
}

func (self *Center) updateTTTplayers(bIsFirst bool) {

	var req proto.GetRankPlayerTTTScore
	var reply proto.GetRankPlayerTTTScoreResult
	req.Start = 0
	req.Stop = 99
	self.getRankPlayerTTTScore(&req, &reply, bIsFirst)
}

func (self *Center) getRankPlayers(req *proto.GetRankPlayers, reply *proto.GetRankPlayersResult, bIsFirst bool) (err error) {

	buf, err := self.zrevrange("rank", "player", req.Start, req.Stop)
	if err != nil {
		logger.Error("GetRankPlayers Error On zrevrange (%s, %v)", err.Error(), buf)
		return
	}

	reply.Code = proto.GetRankPlayerOk
	reply.Value = buf

	rps := &rpc.RankPlayers{}
	myreply := &proto.SaveRankPlayerResult{}

	for index, uid := range reply.Value {
		var p rpc.PlayerBaseInfo

		exist, err := dbclient.KVQueryBase(common.PlayerBase, uid, &p)
		if err != nil {
			continue
		}

		if exist {
			rp := rpc.Player{}
			rp.SetType(rpc.Player_Rank)
			rp.SetName(p.GetName())
			rp.SetRank(uint32(index))
			rp.SetUid(p.GetUid())
			rp.SetTrophy(p.GetTrophy())
			rp.SetLevel(p.GetLevel())
			rp.SetClanName(p.GetClan())
			rp.SetClanSymbol(p.GetClanSymbol())

			rps.RpsTop = append(rps.RpsTop, &rp)
		}
	}

	buff, err := gp.Marshal(rps)
	if err != nil {
		logger.Error("SearchClan Error On Marshal (%s, %v)", err.Error(), buff)
		return
	}

	globalPlayer := &proto.SaveRankPlayer{}
	globalPlayer.Value = buff

	if !bIsFirst {
		for _, conn := range centerServer.cnss {
			conn.Call("CenterService.SaveRankPlayers", globalPlayer, myreply)

			if !myreply.OK {
				logger.Error("CenterService.SaveRankPlayers error!")
			}
		}
	}

	//这里保存一下第一次查到的结果
	SaveAllRankPlayers.RankPlayers = *globalPlayer

	return nil
}

//add for location player 查询区域玩家杯数

func (self *Center) getRankPlayersLocation(req *proto.GetRankPlayersLocation, reply *proto.GetRankPlayersLocationResult, bIsFirst bool) (err error) {
	buf, err := self.zrevrange("rank", "PlayerLocation"+strconv.Itoa(int(req.Location)), req.Start, req.Stop)
	if err != nil {
		logger.Error("GetRankPlayersLocation Error On zrevrange (%s, %v)", err.Error(), buf)
		return
	}

	reply.Code = proto.GetRankPlayersLocationResultOK
	reply.Value = buf
	var location int64

	rps := &rpc.RankPlayers{}
	myreply := &proto.SaveRankLocationPlayersResult{}

	for index, uid := range reply.Value {
		var p rpc.PlayerBaseInfo

		exist, err := dbclient.KVQueryBase(common.PlayerBase, uid, &p)
		if err != nil {
			continue
		}

		if exist {
			rp := rpc.Player{}
			rp.SetType(rpc.Player_Rank)
			rp.SetName(p.GetName())
			rp.SetRank(uint32(index))
			rp.SetUid(p.GetUid())
			rp.SetTrophy(p.GetTrophy())
			rp.SetLevel(p.GetLevel())
			rp.SetClanName(p.GetClan())
			rp.SetClanSymbol(p.GetClanSymbol())

			rps.RpsTop = append(rps.RpsTop, &rp)
		}

		location = int64(p.GetGamelocation())
	}

	buff, err := gp.Marshal(rps)
	if err != nil {
		logger.Error("SearchClan Error On Marshal (%s, %v)", err.Error(), buff)
		return
	}

	locationPlayer := &proto.SaveRankLocationPlayers{}
	locationPlayer.Value = buff
	locationPlayer.Location = req.Location

	if !bIsFirst {
		for _, conn := range centerServer.cnss {
			conn.Call("CenterService.SaveRankLocationPlayers", locationPlayer, myreply)

			if !myreply.OK {
				logger.Error("CenterService.SaveRankPlayers error!")
			}
		}
	}

	//这里保存一下第一次查到的结果
	if locationPlayer.Location == location {

		SaveAllRankPlayers.RankLocationPlayers = *locationPlayer
	}

	return nil
}

func (self *Center) getRankClans(req *proto.GetRankClans, reply *proto.GetRankClansResult, bIsFirst bool) (err error) {
	cs, err := self.zrevrange("rank", "clan", req.Start, req.Stop)
	if err != nil {
		logger.Error("GetRankClans Error On zrevrange (%s, %v)", err.Error(), cs)
		return
	}

	rclans := &rpc.ClanInfos{}
	rclans.SetType(rpc.GetClanType_GetClan_RankClan)

	self.l.RLock()
	for _, cname := range cs {
		if clan, exist := self.clans[cname]; exist {
			rclans.Infos = append(rclans.Infos, clan.GetInfo())
		}
	}
	self.l.RUnlock()

	buf, err := gp.Marshal(rclans)
	if err != nil {
		logger.Error("SearchClan Error On Marshal (%s, %v)", err.Error(), rclans)
		return
	}

	myreply := &proto.SaveRankClanResult{}
	clanResult := &proto.SaveRankClan{}

	clanResult.Value = buf
	if !bIsFirst {
		for _, conn := range centerServer.cnss {
			conn.Call("CenterService.SaveRankClans", clanResult, myreply)

			if !myreply.OK {
				logger.Error("CenterService.SaveRankPlayers error")
			}
		}
	}
	//这里保存一下第一次查到的结果
	SaveAllRankPlayers.RankClans = *clanResult

	return nil
}

func (self *Center) getRankMyselfLocation(req *proto.GetMyself, reply *proto.GetMyselfResult) error {
	logger.Info("进入center上的GetRankMyself")
	location, err := self.zrank("rank", "PlayerLocation"+req.Location, req.Uid)

	if err != nil {
		logger.Error("GetRankMyself Error On zrank ", err)
		return nil
	}

	reply.Code = proto.GetRankPlayerOk
	reply.Rank = location

	return nil
	//rps := &rpc.RankPlayers{}

	//var p rpc.PlayerInfo

	//exist, err := KVQuery("player", req.Uid, &p)
	//if err != nil {
	//	logger.Error("can't find player", req.Uid)
	//	return nil
	//}

	//if exist {
	//	var p rpc.PlayerInfo
	//	rp := rpc.Player{}
	//	rp.SetType(rpc.Player_Rank)
	//	rp.SetRank(uint32(location))
	//	rp.SetName(p.GetName())
	//	rp.SetUid(p.GetUid())
	//	rp.SetTrophy(p.GetTrophy())
	//	rp.SetLevel(p.GetLevel())
	//	rp.SetClanName(p.GetClan())
	//	rp.SetClanSymbol(p.GetClanSymbol())

	//	rps.RpsTop = append(rps.RpsMe, &rp)
	//}

	//buff, err := gp.Marshal(rps)
	//if err != nil {
	//	logger.Error("SearchClan Error On Marshal (%s, %v)", err.Error(), buff)
	//	return nil
	//}

	//return nil

}

func (self *Center) getRankMyselfGlobal(req *proto.GetMyself, reply *proto.GetMyselfResult) error {
	logger.Info("进入center上的GetRankMyself")
	global, err := self.zrank("rank", "player", req.Uid)

	if err != nil {
		logger.Error("GetRankMyself Error On zrank ", err)
		return nil
	}

	reply.Code = proto.GetRankPlayerOk
	reply.Rank = global

	return nil

	//rps := &rpc.RankPlayers{}

	//var p rpc.PlayerInfo

	//exist, err := KVQuery("player", req.Uid, &p)
	//if err != nil {
	//	logger.Error("can't find player", req.Uid)
	//	return nil
	//}

	//if exist {
	//	var p rpc.PlayerInfo
	//	rp := rpc.Player{}
	//	rp.SetType(rpc.Player_Rank)
	//	rp.SetRank(uint32(global))
	//	rp.SetName(p.GetName())
	//	rp.SetUid(p.GetUid())
	//	rp.SetTrophy(p.GetTrophy())
	//	rp.SetLevel(p.GetLevel())
	//	rp.SetClanName(p.GetClan())
	//	rp.SetClanSymbol(p.GetClanSymbol())

	//	rps.RpsTop = append(rps.RpsMe, &rp)
	//}

	//buff, err := gp.Marshal(rps)
	//if err != nil {
	//	logger.Error("SearchClan Error On Marshal (%s, %v)", err.Error(), buff)
	//	return nil
	//}

	//return nil

}
