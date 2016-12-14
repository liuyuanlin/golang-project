package connector

import (
	"golang-project/dpsg/common"
	"golang-project/dpsg/proto"
	"golang-project/dpsg/rpc"
	"strconv"
	"strings"
)

func (self *CNServer) ChatP2P(conn rpc.RpcConn, msg rpc.C2SChatP2P) error {
	ts("CNServer:ChatP2P", conn.GetId())
	defer te("CNServer:ChatP2P", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()
	if !exist {
		return nil
	}

	cmd := proto.PlayerChatToPlayer{
		FromPlayerId: p.GetUid(),
		ToPlayerId:   msg.GetToPlayerId(),
		Content:      msg.GetChatContent(),
	}

	var ret proto.PlayerChatToPlayerResult
	self.chatRpcConn.Go("ChatServices.PlayerChatToPlayer", cmd, &ret, nil)

	return nil
}

func (self *CNServer) ChatAlliance(conn rpc.RpcConn, msg rpc.C2SChatAlliance) error {
	ts("CNServer:ChatAlliance", conn.GetId())
	defer te("CNServer:ChatAlliance", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()
	if !exist {
		return nil
	}

	if p.GetClan() == "" {
		return nil
	}

	msgCast := rpc.ClanChatMessage{}
	msgCast.SetType(rpc.ClanChatMessage_Chat)
	msgCast.SetUid(p.GetUid())
	msgCast.SetName(p.GetName())
	msgCast.SetLevel(p.GetLevel())
	msgCast.SetPower(p.GetClanPlayerPower())
	msgCast.Args = append(msgCast.Args, msg.GetChatContent())

	CastClanChatMsg(p.GetClan(), msgCast)

	return nil
}

//世界聊天
func (self *CNServer) ChatWorld(conn rpc.RpcConn, msg rpc.C2SChatWorld) error {
	ts("CNServer:ChatWorld", conn.GetId())
	defer te("CNServer:ChatWorld", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()
	if !exist {
		return nil
	}

	//gm指令判断
	if self.checkGmCommand(conn, p, msg.GetChatContent()) {
		return nil
	}

	// 扣元宝
	chatCost := GetGlocalChatCost()

	if p.GetPlayerTotalGem() < chatCost {
		return nil
	}

	//先确定扣除成功再做后面的操作
	if !p.CostResource(chatCost, proto.ResType_Gem, proto.Lose_WorldChat) {
		return nil
	}

	cmd := proto.PlayerWorldChat{
		FromPlayerId: p.GetUid(),
		Content:      msg.GetChatContent(),
		CName:        p.GetClan(),
		CSymbol:      p.GetClanSymbol(),
	}

	var ret proto.PlayerWorldChatResult
	self.chatRpcConn.Go("ChatServices.PlayerWorldChat", cmd, &ret, nil)

	return nil
}

func (self *CNServer) BeginChat(conn rpc.RpcConn, msg rpc.BeginChat) error {
	ts("CNServer:EnterChatServer", conn.GetId())
	defer te("CNServer:EnterChatServer", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()
	if !exist {
		return nil
	}

	cmd := proto.AddPlayer{}
	cmd.PlayerId = p.GetUid()
	cmd.AuthKey = *msg.AuthKey
	cmd.PlayerName = p.GetName()
	cmd.PlayerLevel = uint8(p.GetLevel())
	cmd.AllianceName = p.GetClan()
	cmd.AllianceSymbol = p.GetClanSymbol()
	cmd.AlliancePower = uint32(p.GetClanPlayerPower())
	cmd.ChannelId = p.GetGamelocation()
	var ret proto.AddPlayerResult

	self.chatRpcConn.Go("ChatServices.AddPlayer", cmd, &ret, nil)

	return nil
}

func (self *CNServer) EndChat(p *player) error {

	cmd := proto.DelPlayer{}
	cmd.PlayerId = p.GetUid()

	var ret proto.DelPlayerResult
	self.chatRpcConn.Go("ChatServices.DelPlayer", cmd, &ret, nil)

	return nil
}

/*
func (self *CNServer) SendJuanBinMsg(playerId string, content string) error {
	cmd := proto.JuanBinMsg{}
	cmd.FromPlayerId = playerId
	cmd.Content = content

	var ret proto.JuanBinMsgResult
	self.chatRpcConn.Go("ChatServices.JuanBin", cmd, &ret, nil)

	return nil
}
*/

//gm指定
func (self *CNServer) checkGmCommand(conn rpc.RpcConn, p *player, content string) bool {
	//检查开关
	if !common.IsOpenGm() {
		return false
	}

	if len(content) < 2 || content[:2] != "$$" {
		return false
	}

	content = strings.Trim(content[2:], " ")
	pos := strings.Index(content, " ")
	if pos == -1 {
		return false
	}

	cmd, args := strings.ToLower(content[:pos]), content[pos+1:]
	intarg, err := strconv.Atoi(args)
	if err != nil && cmd != "pvp" {
		return false
	}
	switch cmd {
	//加钱
	case "am":
		{
			if intarg > 0 {
				p.GainResource(uint32(intarg), proto.ResType_Gold, proto.Gain_GM)
			} else {
				p.CostResource(uint32(-intarg), proto.ResType_Gold, proto.Lose_GM)
			}
		}
	//加粮草
	case "af":
		{
			if intarg > 0 {
				p.GainResource(uint32(intarg), proto.ResType_Food, proto.Gain_GM)
			} else {
				p.CostResource(uint32(-intarg), proto.ResType_Food, proto.Lose_GM)
			}
		}
	//加武魂
	case "aw":
		{
			if intarg > 0 {
				p.GainResource(uint32(intarg), proto.ResType_Wuhun, proto.Gain_GM)
			} else {
				p.CostResource(uint32(-intarg), proto.ResType_Wuhun, proto.Lose_GM)
			}
		}
	//加体力
	case "at":
		{
			if intarg > 0 {
				p.GainResource(uint32(intarg), proto.ResType_TiLi, proto.Gain_GM)
			} else {
				p.CostResource(uint32(-intarg), proto.ResType_TiLi, proto.Lose_GM)
			}
		}
	//加令牌
	case "al":
		{
			if intarg > 0 {
				p.GainResource(uint32(intarg), proto.ResType_Trophy, proto.Gain_GM)
			} else {
				p.CostResource(uint32(-intarg), proto.ResType_Trophy, proto.Lose_GM)
			}
		}
	//加宝石
	case "ag":
		{
			if intarg > 0 {
				p.GainResource(uint32(intarg), proto.ResType_Gem, proto.Gain_GM)
			} else {
				p.CostResource(uint32(-intarg), proto.ResType_Gem, proto.Lose_GM)
			}
		}
	//加演习次数
	case "ay":
		p.SetDrillTimes(uint32(int(p.GetDrillTimes()) + intarg))
	//加免战时间（时）
	case "as":
		{
			if intarg > 0 {
				p.AddShield(uint32(intarg))
			} else {
				p.RemoveShield()
			}
		}
	//完成任务
	case "dt":
		{
			cfg := GetTaskCfg(args)
			if cfg == nil {
				return false
			}
			task := &rpc.UpdateTaskInfo{}
			task.SetName(args)
			task.SetProgress(cfg.Progress)
			p.UpdateTaskInfo(task)
		}
	//攻打玩家
	case "pvp":
		{
			req := &proto.QueryName{
				Name:   args,
				BQuery: true,
			}

			rst := &proto.QueryNameResult{Success: false}
			if err := cns.center.Call("Center.CheckPlayerName", req, rst); err != nil || !rst.Success {
				return false
			}

			p.ClearFight()
			cns.beginFightWith(conn, p, rst.Id)
		}
	//解锁关卡
	case "ul":
		{
			stage := rpc.Stage{}
			stage.SetStageId(uint32(intarg))
			stage.SetStars(3)
			stage.SetCurrentFood(0)
			stage.SetCurrentGold(0)
			msg := &rpc.PVEAttackEnd{}
			msg.SetPlayerid(p.lid)
			msg.SetStage(&stage)
			msg.SetGoldstolen(0)
			msg.SetFoodstolen(0)
			msg.SetExp(0)
			p.OnPlayerPveFightResult(msg)
		}
	default:
		return false
	}

	return true
}
