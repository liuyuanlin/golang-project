package connector

import (
	gp "github.com/golang/protobuf/proto"
	//"fmt"
	"golang-project/dpsg/common"
	"golang-project/dpsg/logger"
	"golang-project/dpsg/proto"
	"golang-project/dpsg/rpc"
	"strings"
)

func CreateClan(p *player, claninfo *rpc.ClanInfo) (err error) {
	ts("CreateClan", p.GetUid(), claninfo.GetName())
	defer te("CreateClan", p.GetUid(), claninfo.GetName())

	if p.GetClan() != "" {
		logger.Info("CreateClan failed! <%s> already have a clan!", p.GetClan()) //test

		return
	}

	//合法性检查
	if strings.Contains(claninfo.GetName(), " ") || len(claninfo.GetName()) == 0 || len(claninfo.GetName()) > 21 {
		return
	}

	_, totalGold := p.v.collect_GetStorageGoldLimit()

	cost := uint32(40000)
	if totalGold < cost {
		logger.Error("CreateClan Error NotEnough Gold!!!")
		return
	}

	cp := p.CreateClanPlayer(rpc.Player_Leader)

	cps := make([]*rpc.Player, 0)
	cps = append(cps, cp)

	c := &rpc.Clan{Info: claninfo, Players: cps}

	buf, err := gp.Marshal(c)
	if err != nil {
		logger.Error("CreateClan Error On Marshal (%s, %v)", err.Error(), c)
		return
	}

	req := &proto.CreateClan{Value: buf}
	rst := &proto.CreateClanResult{}

	err = cns.center.Call("Center.CreateClan", req, rst)
	//logger.Info("CreateClan:", c, rst) //test
	if err != nil {
		logger.Error("Error On CreateClan : %s", err.Error())

		return
	}

	switch rst.Value {
	case proto.CreateClanOk:
		p.SetClan(claninfo.GetName())
		p.SetClanSymbol(claninfo.GetSymbol())
		p.v.collect_CostGold(cost)

		WriteResult(p.conn, claninfo)

		SendMsg(p.conn, "TID_ALLIANCE_CREATE_SUCCESS")

		UpdatePlayerChatInfo(p.GetUid(), p.GetLevel(), claninfo.GetName(), rpc.Player_Leader)
		//SendText(p.conn, fmt.Sprintf("You are one of <%s> now!", claninfo.GetName()))

		clanMsg := &rpc.ClanMessage{}
		clanMsg.SetType(rpc.ClanChatMessage_Create)
		WriteResult(p.conn, clanMsg)

		return nil
	case proto.CreateClanFailed_Exist:
		SendMsg(p.conn, "TID_ALLIANCE_NAME_REPEAT")
		logger.Info("CreateClan failed! <%s> already exist!", claninfo.GetName()) //test
		//SendText(p.conn, fmt.Sprintf("CreateClan failed! <%s> already exist!", claninfo.GetName()))
		return
	}

	return
}

func SaveClan(p *player, claninfo *rpc.ClanInfo) (err error) {
	ts("SaveClan", p.GetUid(), claninfo.GetName())
	defer te("SaveClan", p.GetUid(), claninfo.GetName())

	if p.GetClan() == "" {
		logger.Info("SaveClan failed! <%s> have none clan!", p.GetClan()) //test
		return
	}

	//todo 合法性检查

	buf, err := gp.Marshal(claninfo)
	if err != nil {
		logger.Error("SaveClan Error On Marshal (%s, %v)", err.Error(), claninfo)
		return
	}

	req := &proto.SaveClan{Value: buf}
	rst := &proto.SaveClanResult{}

	err = cns.center.Call("Center.SaveClan", req, rst)
	//logger.Info("SaveClan:", c, rst) //test
	if err != nil {
		logger.Error("Error On SaveClan : %s", err.Error())

		return
	}

	switch rst.Value {
	case proto.SaveClanOk:
		p.SetClanSymbol(claninfo.GetSymbol())

		WriteResult(p.conn, claninfo)

		//result := rpc.SaveClanResult{}
		//result.SetFail(false)
		//WriteResult(p.conn, &result)

		SendMsg(p.conn, "TID_SAVE_ALLIANCE_SETTINGS")
		//SendText(p.conn, fmt.Sprintf("Save clan <%s> ok!", claninfo.GetName()))

		return nil
	}

	return
}

func GetClan(name string) (int, *rpc.Clan) {
	//ts("GetClan", trophy)
	//defer te("GetClan", trophy)

	try := &proto.GetClan{Name: name}
	ret := &proto.GetClanResult{}

	err := cns.center.Call("Center.GetClan", try, ret)
	//logger.Info("GetClan", try, ret, err)
	if err != nil {
		logger.Error("Error On GetClan : %s, %s", name, err.Error())

		return proto.CenterRpcError, nil
	}

	c := &rpc.Clan{}
	err = gp.Unmarshal(ret.Value, c)
	if err != nil {
		logger.Error("GetClan Unmarshal Error: %s (%v)", err.Error(), c)
		return proto.CenterRpcError, nil
	}

	for _, cp := range c.Players {
		var p rpc.PlayerBaseInfo

		exist, err := KVQueryBase(common.PlayerBase, cp.GetUid(), &p)
		if err != nil {
			continue
		}

		if exist {
			cp.SetDonateNum(p.GetDonateNum())
		}

		cp.SetClanName(c.GetInfo().GetName())
		cp.SetClanSymbol(c.GetInfo().GetSymbol())
	}
	//logger.Info("GetClan11111: %d, %v", ret.Code, c)
	return ret.Code, c
}

func GetClanInfo(name string) (int, *rpc.ClanInfo) {
	//ts("GetClanInfo", trophy)
	//defer te("GetClanInfo", trophy)

	try := &proto.GetClan{Name: name}
	ret := &proto.GetClanInfoResult{}

	err := cns.center.Call("Center.GetClanInfo", try, ret)
	//logger.Info("GetClanInfo", try, ret, err)
	if err != nil {
		logger.Error("Error On GetClanInfo : %s, %s", name, err.Error())

		return proto.CenterRpcError, nil
	}

	info := &rpc.ClanInfo{}
	err = gp.Unmarshal(ret.Value, info)
	if err != nil {
		logger.Error("GetClanInfo Unmarshal Error: %s (%v)", err.Error(), info)
		return proto.CenterRpcError, nil
	}

	return ret.Code, info
}

func GetClanPlayer(name string, uid string) (int, *rpc.Player) {
	//ts("GetClanPlayer", name, uid)
	//defer te("GetClanPlayer", name, uid)

	try := &proto.GetClanPlayer{Name: name, Uid: uid}
	ret := &proto.GetClanPlayerResult{}

	err := cns.center.Call("Center.GetClanPlayer", try, ret)
	logger.Info("GetClanPlayer", try, ret, err)
	if err != nil {
		logger.Error("Error On GetClanPlayer : %s, %s", name, err.Error())
		return proto.CenterRpcError, nil
	}
	if ret.Code != proto.GetClanPlayerOk {
		return ret.Code, nil
	}

	player := &rpc.Player{}
	err = gp.Unmarshal(ret.Value, player)
	if err != nil {
		logger.Error("GetClanPlayer Unmarshal Error: %s (%v)", err.Error(), *player)
		return proto.CenterRpcError, nil
	}

	return ret.Code, player
}

func RandomGetClans(trophy uint32) (int, *rpc.ClanInfos) {
	//ts("RandomGetClans", trophy)
	//defer te("RandomGetClans", trophy)

	try := &proto.RandomGetClans{Trophy: trophy, Num: 10}
	ret := &proto.RandomGetClansResult{}

	err := cns.center.Call("Center.RandomGetClans", try, ret)
	//logger.Info("RandomGetClans", try, ret, err)
	if err != nil {
		logger.Error("Error On RandomGetClans : %s", err.Error())

		return proto.CenterRpcError, nil
	}

	infos := &rpc.ClanInfos{}
	err = gp.Unmarshal(ret.Value, infos)
	if err != nil {
		logger.Error("RandomGetClans Unmarshal Error: %s (%v)", err.Error(), infos)
		return proto.CenterRpcError, nil
	}

	return ret.Code, infos
}

func TryJoinClan(p *player, name string) int {
	ts("TryJoinClan", p.GetUid(), name)
	defer te("TryJoinClan", p.GetUid(), name)

	obj := p.v.buildings_Get(rpc.BuildingId_AllianceCastle, 0)
	if obj == nil {
		logger.Error("TryJoinClan Error:AllianceCastle Not Found!")
		return proto.JoinClanFailed_NoCastle
	}

	cp := p.CreateClanPlayer(rpc.Player_Member)

	buf, err := gp.Marshal(cp)
	if err != nil {
		logger.Error("TryJoinClan Error On Marshal (%s, %v)", err.Error(), cp)
		return proto.CenterRpcError
	}

	try := &proto.JoinClan{Value: buf, Name: name}
	ret := &proto.JoinClanResult{}

	err = cns.center.Call("Center.TryJoinClan", try, ret)
	//logger.Info("TryJoinClan", try, ret, err)
	if err != nil {
		logger.Error("Error On TryJoinClan : %s, %s", name, err.Error())

		return proto.CenterRpcError
	}

	return ret.Value
}

func TryLeaveClan(p *player) int {
	ts("TryLeaveClan", p.GetUid())
	defer te("TryLeaveClan", p.GetUid())

	cname := p.GetClan()

	try := &proto.LeaveClan{PUid: p.GetUid(), Name: cname}
	ret := &proto.LeaveClanResult{}

	err := cns.center.Call("Center.TryLeaveClan", try, ret)
	//logger.Info("TryLeaveClan", try, ret, err)
	if err != nil {
		logger.Error("Error On TryLeaveClan : %s, %s", cname, err.Error())

		return proto.CenterRpcError
	}

	return ret.Value
}

func TryKickPlayer(p *player, taruid string) int {
	ts("TryKickPlayer", p.GetUid())
	defer te("TryKickPlayer", p.GetUid())

	cname := p.GetClan()

	try := &proto.KickPlayer{Uid: p.GetUid(), CName: cname, TarUid: taruid}
	ret := &proto.KickPlayerResult{}

	err := cns.center.Call("Center.TryKickPlayer", try, ret)
	//logger.Info("TryKickPlayer", try, ret, err)
	if err != nil {
		logger.Error("Error On TryKickPlayer : %s, %s, %s(Error:%s)", p.GetUid(), cname, taruid, err.Error())

		return proto.CenterRpcError
	}

	if ret.Value == proto.KickPlayerOk {
		player := LoadPlayerToVisit(taruid)
		if player == nil {
			logger.Error("TryKickPlayer:player(%s) == nil", taruid)
		}
		//公告：被xx移出了公会
		msgCast := rpc.ClanChatMessage{}
		msgCast.SetType(rpc.ClanChatMessage_Kick)
		msgCast.SetUid(taruid)
		msgCast.SetName(player.GetName())
		msgCast.SetLevel(player.GetLevel())
		msgCast.SetPower(rpc.Player_ClanPower(ret.Power))
		msgCast.Args = append(msgCast.Args, p.GetName())
		CastClanChatMsg(cname, msgCast) //广播公告

		//请求center广播以通知对方更新数据。若不在线则不管
		req := &proto.NotifyUpdateClanInfo{Uid: taruid, Type: int32(rpc.ClanChatMessage_Kick), CName: cname}
		rst := &proto.NotifyUpdateClanInfoResult{}
		cns.center.Go("Center.NotifyUpdateClanInfo", req, rst, nil)

		//不管对方在不在线先更新保存公会相关信息
		player.SetClan("")
		player.SetClanSymbol(0)
		player.Save()
	}

	return ret.Value
}

func TryAppointPlayer(p *player, taruid string, power rpc.Player_ClanPower) int {
	ts("TryAppointPlayer:%s, %s, %d", p.GetUid(), taruid, power)
	defer te("TryAppointPlayer:%s, %s, %d", p.GetUid(), taruid, power)

	cname := p.GetClan()

	try := &proto.AppointPlayer{Uid: p.GetUid(), CName: cname, TarUid: taruid, Power: int32(power)}
	ret := &proto.AppointPlayerResult{}

	err := cns.center.Call("Center.TryAppointPlayer", try, ret)
	if err != nil {
		logger.Error("Error On TryAppointPlayer : %s, %s, %s, %d(Error:%s)", p.GetUid(), cname, taruid, power, err.Error())

		return proto.CenterRpcError
	}

	if ret.Value == proto.AppointPlayerOk {
		player := LoadPlayerToVisit(taruid)
		if player == nil {
			logger.Error("CN:TryAppointPlayer:player(%s) == nil", taruid)
		}
		//公告：被xx提升/降职为yy
		msgCast := rpc.ClanChatMessage{}
		if ret.OldPower == int32(rpc.Player_Elder) {
			if power == rpc.Player_Leader {
				msgCast.SetType(rpc.ClanChatMessage_PromoteLeader)

				//自动降级为长老
				UpdatePlayerChatInfo(p.GetUid(), p.GetLevel(), p.GetClan(), rpc.Player_Elder)

				clanMsg := &rpc.ClanMessage{}
				clanMsg.SetType(rpc.ClanChatMessage_DemoteElder)
				WriteResult(p.conn, clanMsg)
			} else if power == rpc.Player_Member {
				msgCast.SetType(rpc.ClanChatMessage_DemoteMember)
			} else {
				logger.Error("CN:TryAppointPlayer:Unexpect Error1!")
				return ret.Value
			}
		} else if ret.OldPower == int32(rpc.Player_Member) {
			if power == rpc.Player_Leader {
				msgCast.SetType(rpc.ClanChatMessage_PromoteLeader)

				//自动降级为长老
				UpdatePlayerChatInfo(p.GetUid(), p.GetLevel(), p.GetClan(), rpc.Player_Elder)

				clanMsg := &rpc.ClanMessage{}
				clanMsg.SetType(rpc.ClanChatMessage_DemoteElder)
				WriteResult(p.conn, clanMsg)
			} else if power == rpc.Player_Elder {
				msgCast.SetType(rpc.ClanChatMessage_PromoteElder)
			} else {
				logger.Error("CN:TryAppointPlayer:Unexpect Error2!")
				return ret.Value
			}
		} else {
			logger.Error("CN:TryAppointPlayer:Unexpect Error3!")
			return ret.Value
		}
		msgCast.SetUid(taruid)
		msgCast.SetName(player.GetName())
		msgCast.SetLevel(player.GetLevel())
		msgCast.SetPower(power)
		msgCast.Args = append(msgCast.Args, p.GetName())
		CastClanChatMsg(cname, msgCast) //广播公告

		//请求center广播以通知对方更新数据。若不在线则不管
		req := &proto.NotifyUpdateClanInfo{Uid: taruid, Type: int32(msgCast.GetType()), CName: cname}
		rst := &proto.NotifyUpdateClanInfoResult{}
		cns.center.Go("Center.NotifyUpdateClanInfo", req, rst, nil)
	}

	return ret.Value
}

func SearchClan(key string) (int, *rpc.ClanInfos) {
	//ts("SearchClan", key)
	//defer te("SearchClan", key)

	try := &proto.SearchClan{Key: key}
	ret := &proto.SearchClanResult{}

	err := cns.center.Call("Center.SearchClan", try, ret)
	//logger.Info("SearchClan", try, ret, err)
	if err != nil {
		logger.Error("Error On SearchClan : %s", err.Error())

		return proto.CenterRpcError, nil
	}

	infos := &rpc.ClanInfos{}
	err = gp.Unmarshal(ret.Value, infos)
	if err != nil {
		logger.Error("SearchClan Unmarshal Error: %s (%v)", err.Error(), infos)
		return proto.CenterRpcError, nil
	}

	return ret.Code, infos
}
