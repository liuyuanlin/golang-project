package connector

import (
	"fmt"
	"golang-project/dpsg/logger"
	//"language"
	"golang-project/dpsg/accountclient"
	"golang-project/dpsg/common"
	"golang-project/dpsg/proto"
	"golang-project/dpsg/rpc"
)

func (self *CNServer) GetFriendsList(conn rpc.RpcConn, unuse rpc.C2SGetFriendsList) error {
	ts("CNServer:C2SGetFriendsList %d", conn.GetId())
	defer te("CNServer:C2SGetFriendsList", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	rps := &rpc.FriendsList{}
	if p.friendscache == nil {
		success, errmsg, list := MobileQQFriends(p)
		if !success {
			logger.Error("GetFriendsList MobileQQFriends Failed!, %s", errmsg)
			return nil
		}

		for _, v := range list {
			//去掉QQ好友上的自己，后面统一加上自己
			if p.mobileqqinfo != nil && v.OpenId == p.mobileqqinfo.Openid {
				continue
			}

			uid, err := accountclient.QueryPlayerIdByPartnerId(common.TableName_TencentAccount, v.OpenId)
			if err != nil || len(uid) == 0 {
				continue
			}

			var friend rpc.PlayerBaseInfo
			exist, err := KVQueryBase(common.PlayerBase, uid, &friend)
			if err != nil || !exist {
				continue
			}

			rp := rpc.Player{}
			rp.SetType(rpc.Player_Rank)
			rp.SetName(v.NickName)
			rp.SetUid(uid)
			rp.SetTrophy(friend.GetTrophy())
			rp.SetLevel(friend.GetLevel())
			rp.SetClanName(friend.GetClan())
			rp.SetClanSymbol(friend.GetClanSymbol())
			rp.SetGender(v.Gender)
			rp.SetHeadurl(v.Picture)

			rps.Friends = append(rps.Friends, &rp)
		}

		//加上自己
		if p.mobileqqinfo != nil {
			rp := rpc.Player{}
			rp.SetType(rpc.Player_Rank)
			rp.SetName(p.GetName())
			rp.SetUid(p.GetUid())
			rp.SetTrophy(p.GetTrophy())
			rp.SetLevel(p.GetLevel())
			rp.SetClanName(p.GetClan())
			rp.SetClanSymbol(p.GetClanSymbol())
			rp.SetGender(p.GetGender())
			rp.SetHeadurl(p.GetHeadurl())

			rps.Friends = append(rps.Friends, &rp)
		}

		p.friendscache = rps
	} else {
		rps = p.friendscache
	}

	WriteResult(conn, rps)

	return nil

}

/*func (self *CNServer) AddFriend(conn rpc.RpcConn, key rpc.C2SAddFriend) error {
	ts("CNServer:AddFriend %d", conn.GetId())
	defer te("CNServer:AddFriend", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	rps := &rpc.S2CAddFriendResult{}

	if !p.CanAddFriend() {
		//如果已经添加满好友
		rps.SetRst(rpc.S2CAddFriendResult_HasFull)
		WriteResult(conn, rps)
		return nil
	}

	//参数可能是传入的id，也可能是名字。
	//先看一下是否传入了id
	id := key.GetId()

	//如果传入的id为空，则需要向center通过名字请求id
	if id == "" {
		//通过名字查询id
		req := &proto.QueryName{
			Name:   key.GetName(),
			BQuery: true,
		}

		ts("AddFriend by Name", key.GetName())

		rst := &proto.QueryNameResult{Success: false}
		if err := self.center.Call("Center.CheckPlayerName", req, rst); err != nil || !rst.Success {
			//查询失败，未能跟据名字找到对应玩家id
			ts("AddFriend by Name", key.GetName())
			rps.SetRst(rpc.S2CAddFriendResult_FindntPlayer)
			WriteResult(conn, rps)
			return nil
		} else {
			id = rst.Id
		}
	}

	if id == "" {
		//如果id还是为空表示没找到该玩家
		rps.SetRst(rpc.S2CAddFriendResult_FindntPlayer)
		WriteResult(conn, rps)
		return nil
	}

	//先查一下该玩家是否已经是好友
	_, ok := p.friends[id]
	if ok {
		rps.SetRst(rpc.S2CAddFriendResult_RepeatAdd)
		WriteResult(conn, rps)
		return nil
	}

	//反之跟据id开始添加好友
	var friend rpc.PlayerBaseInfo

	_, err := KVQueryBase(common.PlayerBase, id, &friend)

	if err != nil {
		rps.SetRst(rpc.S2CAddFriendResult_FindntPlayer)
		WriteResult(conn, rps)
		return nil
	}

	//不存在则添加好友
	p.friends[id] = true

	rp := rpc.Player{}
	rp.SetType(rpc.Player_Rank)
	rp.SetName(friend.GetName())
	rp.SetUid(friend.GetUid())
	rp.SetTrophy(friend.GetTrophy())
	rp.SetLevel(friend.GetLevel())
	rp.SetClanName(friend.GetClan())
	rp.SetClanSymbol(friend.GetClanSymbol())

	rps.SetFriend(&rp)
	rps.SetRst(rpc.S2CAddFriendResult_OK)

	WriteResult(conn, rps)

	//发送通知邮件
	req := &proto.SendSystemMail{
		ToPlayerId: friend.GetUid(),
		Title:      "",
		Content:    fmt.Sprintf("$$L:TID_FRIEND_11,%s$$", p.GetName()), //language.GetLanguage("TID_FRIEND_11", p.GetName()),
		Attach:     "",
	}
	rst := &proto.SendSystemMailResult{}
	self.chatRpcConn.Go("ChatServices.SendSysMail2Player", req, rst, nil)

	return nil

}

func (self *CNServer) DelFriend(conn rpc.RpcConn, id rpc.C2SDelFriend) error {
	ts("CNServer:DelFriend %d", conn.GetId())
	defer te("CNServer:DelFriend", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	rps := &rpc.S2CDelFriendResult{}
	rps.SetId(id.GetId())

	_, ok := p.friends[id.GetId()]

	if ok {
		delete(p.friends, id.GetId())
		rps.SetRst(rpc.S2CDelFriendResult_OK)
	} else {
		rps.SetRst(rpc.S2CDelFriendResult_FindntFried)
	}

	WriteResult(conn, rps)

	return nil
}*/

func (self *CNServer) PlayerGiveGift(conn rpc.RpcConn, info rpc.PlayerGiveGift) error {
	common.LoadPresentCfg()

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()
	if !exist {
		return nil
	}

	req := &proto.PlayerGiveGift{Uid: *info.Uid}
	ret := &proto.PlayerGiveGiftResult{}

	if err := self.center.Call("Center.FindToPlayerFromDB", req, ret); err != nil {
		logger.Error("Center.FindToPlayerFromDB err", err)
		return err
	}

	if ret.Code == proto.PlayNotExist {
		logger.Info("--No player--")
		return nil
	}

	if ret.Code == proto.FindPlayerOK {
		mailreq := &proto.GiveGiftToPlayer{ToUid: *info.Uid, FromUid: p.GetUid()}
		mailret := &proto.GiveGiftToPlayerResult{}
		self.SendMailtoplayer(mailreq, mailret)
	}
	return nil
}

func (self *CNServer) SendMailtoplayer(req *proto.GiveGiftToPlayer, ret *proto.GiveGiftToPlayerResult) {
	//这里读表，发体力，粮草，银子给玩家

	sendInfo1 := common.GetPresentCfg("1")
	sendInfo2 := common.GetPresentCfg("2")
	logger.Info("sendInfo1", sendInfo1)
	logger.Info("sendInfo2", sendInfo2)
	mailreq := &proto.SendSystemMail{
		ToPlayerId: req.ToUid,
		Title:      fmt.Sprintf("$$L:%s TID_LT_MAIL_TITLE$$", req.FromUid),
		Content:    "",
		Attach:     fmt.Sprintf("%d:%d,%d:%d", sendInfo1.Type, sendInfo1.Number, sendInfo2.Type, sendInfo2.Number),
	}
	mailreqReselt := &proto.SendSystemMailResult{}

	cns.chatRpcConn.Go("ChatServices.SendSysMail2Player", mailreq, mailreqReselt, nil)
}
