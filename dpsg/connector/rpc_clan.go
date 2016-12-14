package connector

import (
	"fmt"
	//"language"
	"golang-project/dpsg/logger"
	"golang-project/dpsg/proto"
	"golang-project/dpsg/rpc"
)

func (self *CNServer) CreateClan(conn rpc.RpcConn, claninfo rpc.ClanInfo) error {
	ts("CNServer:CreateClan", conn.GetId())
	defer te("CNServer:CreateClan", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	return CreateClan(p, &claninfo)
}

func (self *CNServer) SaveClan(conn rpc.RpcConn, claninfo rpc.ClanInfo) error {
	ts("CNServer:SaveClan", conn.GetId())
	defer te("CNServer:SaveClan", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	return SaveClan(p, &claninfo)
}

func (self *CNServer) GetClan(conn rpc.RpcConn, try rpc.TryGetClan) error {
	ts("CNServer:GetClan", conn.GetId())
	defer te("CNServer:GetClan", conn.GetId())

	self.l.RLock()
	_, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	value, clan := GetClan(try.GetName())
	if value != proto.GetClanOk {
		return nil
	}

	clan.SetType(try.GetType())

	WriteResult(conn, clan)

	return nil
}

func (self *CNServer) RandomGetClans(conn rpc.RpcConn, try rpc.TryGetClans) error {
	ts("CNServer:RandomGetClans", conn.GetId())
	defer te("CNServer:RandomGetClans", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	value, clans := RandomGetClans(p.GetTrophy())
	if value != proto.GetClanOk {
		return nil
	}

	clans.SetType(try.GetType())

	WriteResult(conn, clans)

	return nil
}

func (self *CNServer) TryJoinClan(conn rpc.RpcConn, try rpc.TryJoinClan) error {
	ts("CNServer:TryJoinClan", conn.GetId())
	defer te("CNServer:TryJoinClan", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	value := TryJoinClan(p, try.GetName())
	if value != proto.JoinClanOk {
		if value == proto.JoinClanFailed_NotEnoughTrophy {
			SendMsg(p.conn, "TID_TROPHY_NOT_ENOUGH")
		}
		return nil
	}

	p.SetClan(try.GetName())
	p.SetClanSymbol(p.GetClanInfo().GetSymbol())
	p.Save()

	value, clan := GetClan(try.GetName())
	if value != proto.GetClanOk {
		return nil
	}

	clan.SetType(try.GetType())

	WriteResult(conn, clan.GetInfo())
	WriteResult(conn, clan)

	//更新chatserver上该玩家的信息
	UpdatePlayerChatInfo(p.GetUid(), p.GetLevel(), p.GetClan(), rpc.Player_Member)

	clanMsg := &rpc.ClanMessage{}
	clanMsg.SetType(rpc.ClanChatMessage_Join)
	WriteResult(conn, clanMsg)

	//公告：加入了公会.
	msgCast := rpc.ClanChatMessage{}
	msgCast.SetType(rpc.ClanChatMessage_Join)
	msgCast.SetUid(p.GetUid())
	msgCast.SetName(p.GetName())
	msgCast.SetLevel(p.GetLevel())
	msgCast.SetPower(rpc.Player_Member)
	CastClanChatMsg(try.GetName(), msgCast) //广播公告

	return nil
}

func (self *CNServer) TryLeaveClan(conn rpc.RpcConn, try rpc.TryLeaveClan) error {
	ts("CNServer:TryLeaveClan", conn.GetId())
	defer te("CNServer:TryLeaveClan", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		logger.Info("没找到玩家")
		return nil
	}

	power := p.GetClanPlayerPower()
	cname := p.GetClan()

	value := TryLeaveClan(p)
	logger.Info("返回值是", value)
	if value != proto.LeaveClanOk && value != proto.DeleteClanOK {
		logger.Info("离开失败")
		return nil
	}

	p.SetClan("")
	p.SetClanSymbol(0)
	p.Save()
	WriteResult(conn, p.GetClanInfo())

	//更新chatserver上该玩家的信息
	UpdatePlayerChatInfo(p.GetUid(), p.GetLevel(), "", rpc.Player_None)

	clanMsg := &rpc.ClanMessage{}
	clanMsg.SetType(rpc.ClanChatMessage_Leave)
	WriteResult(conn, clanMsg)

	//公告：加入了公会.
	logger.Info("发公告了")
	msgCast := rpc.ClanChatMessage{}
	msgCast.SetType(rpc.ClanChatMessage_Leave)
	msgCast.SetUid(p.GetUid())
	msgCast.SetName(p.GetName())
	msgCast.SetLevel(p.GetLevel())
	msgCast.SetPower(power)
	CastClanChatMsg(cname, msgCast) //广播公告

	return nil
}

func (self *CNServer) TryKickPlayer(conn rpc.RpcConn, try rpc.TryKickPlayer) error {
	ts("CNServer:TryKickPlayer", conn.GetId())
	defer te("CNServer:TryKickPlayer", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	value := TryKickPlayer(p, try.GetTarUid())
	if value != proto.KickPlayerOk {
		return nil
	}

	value, clan := GetClan(p.GetClan())
	if value != proto.GetClanOk {
		return nil
	}

	clan.SetType(try.GetType())
	WriteResult(conn, clan)

	//发送邮件
	req := &proto.SendSystemMail{
		ToPlayerId: try.GetTarUid(),
		Title:      "",
		Content:    fmt.Sprintf("$$L:TID_MAIL_KICKET_FROM_ALLIANCE$$"),
		Attach:     "",
	}
	rst := &proto.SendSystemMailResult{}
	self.chatRpcConn.Go("ChatServices.SendSysMail2Player", req, rst, nil)

	return nil
}

func (self *CNServer) TryAppointPlayer(conn rpc.RpcConn, try rpc.TryAppointPlayer) error {
	ts("CNServer:TryAppointPlayer", conn.GetId())
	defer te("CNServer:TryAppointPlayer", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	value := TryAppointPlayer(p, try.GetTarUid(), try.GetPower())
	if value != proto.AppointPlayerOk {
		return nil
	}

	value, clan := GetClan(p.GetClan())
	if value != proto.GetClanOk {
		return nil
	}

	clan.SetType(try.GetType())
	WriteResult(conn, clan)

	return nil
}

func (self *CNServer) SearchClan(conn rpc.RpcConn, try rpc.TryGetClans) error {
	ts("CNServer:SearchClan", conn.GetId())
	defer te("CNServer:SearchClan", conn.GetId())

	self.l.RLock()
	_, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	value, clans := SearchClan(try.GetKey())
	if value != proto.SearchClanOk {
		return nil
	}

	clans.SetType(try.GetType())

	WriteResult(conn, clans)

	return nil
}

//来自Center的通知
func (s *CenterService) NotifyUpdateClanInfo(req *proto.NotifyUpdateClanInfo, reply *proto.NotifyUpdateClanInfoResult) (err error) {
	logger.Info("CenterService.NotifyUpdateClanInfo:%d, %s, %s", req.Type, req.Uid, req.CName)

	cns.l.Lock()
	defer cns.l.Unlock()

	p, exist := cns.playersbyid[req.Uid]
	if exist {
		switch rpc.ClanChatMessage_MsgType(req.Type) {
		case rpc.ClanChatMessage_Kick:
			p.SetClan("")
			p.SetClanSymbol(0)
			WriteResult(p.conn, p.GetClanInfo())

			//更新chatserver上该玩家的信息
			UpdatePlayerChatInfo(p.GetUid(), p.GetLevel(), "", rpc.Player_None)
		case rpc.ClanChatMessage_PromoteLeader:
			//更新chatserver上该玩家的信息
			UpdatePlayerChatInfo(p.GetUid(), p.GetLevel(), p.GetClan(), rpc.Player_Leader)
		case rpc.ClanChatMessage_PromoteElder:
			//更新chatserver上该玩家的信息
			UpdatePlayerChatInfo(p.GetUid(), p.GetLevel(), p.GetClan(), rpc.Player_Elder)
		case rpc.ClanChatMessage_DemoteMember:
			//更新chatserver上该玩家的信息
			UpdatePlayerChatInfo(p.GetUid(), p.GetLevel(), p.GetClan(), rpc.Player_Member)
		}

		clanMsg := &rpc.ClanMessage{}
		clanMsg.SetType(rpc.ClanChatMessage_MsgType(req.Type))
		WriteResult(p.conn, clanMsg)
	}

	return nil
}

//send mail to everyone
func (self *CenterService) SendMailtoClanplayer(req *proto.SendClanMail, ret *proto.SendClanMailResult) error {

	logger.Info("")
	mailreq := &proto.SendSystemMail{
		ToPlayerId: req.Uid,
		Title:      fmt.Sprintf("$$L:TID_LT_MENGZHU_CHANGE_TITLE"),
		Content: fmt.Sprintf("$$L:TID_LT_MENGZHU_CHANGE	%s", req.NewLeader),
		Attach: "",
	}
	mailreqReselt := &proto.SendSystemMailResult{}

	cns.chatRpcConn.Go("ChatServices.SendSysMail2Player", mailreq, mailreqReselt, nil)

	return nil
}
