package center

import (
	"golang-project/dpsg/logger"
	"golang-project/dpsg/proto"
	"golang-project/dpsg/rpc"
	//"github.com/garyburd/redigo/redis"
	//"rpcplus"
)

type clan struct {
	*rpc.Clan
}

func LoadClanFromBuf(c *rpc.Clan) *clan {
	return &clan{Clan: c}
}

func (c *clan) GetClan() *rpc.Clan {
	c.Info.SetMembers(uint32(len(c.Players)))
	return c.Clan
}

func (c *clan) GetInfo() *rpc.ClanInfo {
	c.Info.SetMembers(uint32(len(c.Players)))
	return c.Info
}

func (c *clan) GetPlayer(uid string) *rpc.Player {
	for _, cp := range c.Players {
		if cp.GetUid() == uid {
			return cp
		}
	}
	return nil
}

func (c *clan) UpdateInfo(info *rpc.ClanInfo) {
	c.Info.SetType(info.GetType())
	c.Info.SetSymbol(info.GetSymbol())
	c.Info.SetRequire(info.GetRequire())
	c.Info.SetDescribe(info.GetDescribe())
}

func (c *clan) UpdateTrophy(self *Center) {
	total := uint32(0)

	for _, cp := range c.Players {
		trophy := self.getPlayerTrophy(cp.GetUid())
		cp.SetTrophy(trophy)
		total += trophy
	}

	c.Info.SetTrophy(total)

	//logger.Info("UpdateTrophy:", c.GetInfo().GetName(), total) //test

	self.zadd("rank", "clan", c.GetInfo().GetName(), total) //更新杯数
}

func (c *clan) TryJoinClan(cp *rpc.Player) int {
	if len(c.Players) >= 50 {
		return proto.JoinClanFailed_ClanFull
	}

	if c.GetPlayer(cp.GetUid()) != nil {
		return proto.JoinClanFailed_AleadyIn
	}

	if cp.GetTrophy() < c.Info.GetRequire() {
		return proto.JoinClanFailed_NotEnoughTrophy
	}

	c.Players = append(c.Players, cp)

	return proto.JoinClanOk
}

func (c *clan) TryLeaveClan(puid string) int {
	logger.Info("尝试离开联盟")

	clanPlayers := 0
	for index, cp := range c.Players {
		clanPlayers += 1
		if puid == cp.GetUid() {
			if cp.GetPower() == rpc.Player_Leader {
				//return proto.LeaveClanFailed_LeaderCannotLeave

				playerCount := uint32(0)
				elderCount := uint32(0)
				Elder := cp
				Player := cp
				bHaveMaster := false
				for _, value := range c.Players { //查找排名第一的长老
					if value.GetPower() == rpc.Player_Elder {
						bHaveMaster = true
						if elderCount <= value.GetTrophy() {
							elderCount = value.GetTrophy()
							Elder = value
						}
					} else if !bHaveMaster { //查找排名第一的玩家
						if playerCount <= value.GetTrophy() && value.GetPower() != rpc.Player_Leader {
							playerCount = value.GetTrophy()
							Player = value
						}
					}
				}
				if bHaveMaster { //如果有长老，则认命杯数最多的长老为盟主
					logger.Info("set the first elder as the leader!")
					Elder.SetPower(rpc.Player_Leader)
				} else { //否则认命杯数最多的玩家
					logger.Info("no elder , set the first player as the leader!")
					Player.SetPower(rpc.Player_Leader)
				}

				c.Players = append(c.Players[:index], c.Players[index+1:]...)

				//send mail
				c.SendMailToClanPlayers(cp.GetName(), Player.GetName())

				//都没有就直接解散
				if clanPlayers == 1 {
					logger.Info("only leader here!")
					return proto.DeleteClanOK
				}

				return proto.LeaveClanOk
			}

			c.Players = append(c.Players[:index], c.Players[index+1:]...)
			return proto.LeaveClanOk
		}
	}

	return proto.LeaveClanFailed_NotFoundPlayer
}

func (c *clan) TryKickPlayer(uid string, taruid string) (Value int, Power int32) {
	if uid == taruid {
		Value = proto.KickPlayerFailed_CannotKickSelf
		return
	}

	var index int = 0
	var p1 *rpc.Player = nil
	var p2 *rpc.Player = nil
	for i, cp := range c.Players {
		if uid == cp.GetUid() {
			p1 = cp
		}
		if taruid == cp.GetUid() {
			p2 = cp
			Power = int32(cp.GetPower())
			index = i
		}
	}

	if p1 == nil || p2 == nil {
		Value = proto.KickPlayerFailed_NotFoundPlayer
		return
	}

	if p1.GetPower() == rpc.Player_Member {
		Value = proto.KickPlayerFailed_NotEnoughPower
		return
	}

	if p1.GetPower() == rpc.Player_Elder {
		if p2.GetPower() == rpc.Player_Leader {
			Value = proto.KickPlayerFailed_NotEnoughPower
			return
		}
	}

	c.Players = append(c.Players[:index], c.Players[index+1:]...)

	Value = proto.KickPlayerOk
	return
}

func (c *clan) TryAppointPlayer(uid string, taruid string, power rpc.Player_ClanPower) (Value int, Power int32) {
	if uid == taruid {
		Value = proto.AppointPlayerFailed_CannotAppointSelf
		return
	}

	var p1 *rpc.Player = nil
	var p2 *rpc.Player = nil
	for _, cp := range c.Players {
		if uid == cp.GetUid() {
			p1 = cp
		}
		if taruid == cp.GetUid() {
			p2 = cp
			Power = int32(cp.GetPower())
		}
	}

	if p1 == nil || p2 == nil {
		Value = proto.AppointPlayerFailed_NotFoundPlayer
		return
	}

	if p2.GetPower() == power {
		Value = proto.AppointPlayerFailed_PowerNotChange
		return
	}

	if p2.GetPower() == rpc.Player_Leader { //不能对会长进行任何任命
		Value = proto.AppointPlayerFailed_NotEnoughPower
		return
	} else if p2.GetPower() == rpc.Player_Elder { //如果目标是长老，则只有会长可以对其任命
		if p1.GetPower() == rpc.Player_Leader {
			if power == rpc.Player_Leader { //升任会长
				//合法
				p1.SetPower(rpc.Player_Elder) //会长自动降职为长老
			} else if power == rpc.Player_Member { //降职为成员
				//合法
			} else {
				logger.Error("Center:TryAppointPlayer:Unexpect Error1!")
				Value = proto.AppointPlayerFailed_UnKnown
				return
			}
		} else {
			Value = proto.AppointPlayerFailed_NotEnoughPower
			return
		}
	} else if p2.GetPower() == rpc.Player_Member {
		if power == rpc.Player_Leader { //升任会长
			if p1.GetPower() == rpc.Player_Leader {
				//合法
				p1.SetPower(rpc.Player_Elder) //会长自动降职为长老
			} else {
				Value = proto.AppointPlayerFailed_NotEnoughPower
				return
			}
		} else if power == rpc.Player_Elder { //升任长老
			if p1.GetPower() == rpc.Player_Leader || p1.GetPower() == rpc.Player_Elder {
				//合法
			} else {
				Value = proto.AppointPlayerFailed_NotEnoughPower
				return
			}
		} else {
			logger.Error("Center:TryAppointPlayer:Unexpect Error2!")
			Value = proto.AppointPlayerFailed_UnKnown
			return
		}
	}

	p2.SetPower(power)

	Value = proto.AppointPlayerOk

	return
}

//send mail to clan players

func (c *clan) SendMailToClanPlayers(Old, New string) error {
	ts("ComeInto SendMailToClanPlayers")
	defer te("Out SendMailToClanPlayers")

	for _, cp := range c.Players {
		logger.Info("find player", cp)

		req := &proto.SendClanMail{Uid: cp.GetUid(), OldLeader: Old, NewLeader: New}
		ret := &proto.SendClanMailResult{}
		conn := centerServer.cnss[0]
		if err := conn.Call("CenterService.SendMailtoClanplayer", req, ret); err != nil {
			logger.Error("Send mail error", err.Error())
			return nil
		} else {
			logger.Info("I just send clan mail to player")
		}
	}

	return nil
}
