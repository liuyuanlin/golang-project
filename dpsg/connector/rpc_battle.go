package connector

import (
	"fmt"
	"golang-project/dpsg/common"
	"golang-project/dpsg/language"
	"golang-project/dpsg/lockclient"
	"golang-project/dpsg/logger"
	"golang-project/dpsg/proto"
	"golang-project/dpsg/pushmsg"
	"golang-project/dpsg/rpc"
	"strconv"
)

func (self *CNServer) RandomMatch(conn rpc.RpcConn, ping rpc.Ping) error {
	ts("CNServer:RandomMatch", conn.GetId())
	defer te("CNServer:RandomMatch", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()
	if !exist {
		return nil
	}

	// 如果已经处于战斗状态，则不能再次匹配
	if !p.CanPvPFight() {
		logger.Info("RandomMatch Error: In Fight:%s", p.GetUid())
		return nil
	}

	// 删除上一个匹配到的对手
	self.delOtherPlayer(conn.GetId())

	//match
	ouid, err := TryMatch(p.GetTrophy(), p.lastmatch)
	if err != nil || ouid == "" {
		SendMsg(conn, "TID_SEARCH_NOTHING")
		p.lastmatch = ""
		WriteMatchResult(conn, rpc.MatchPlayerResult_MATCHNOTHING)
		return nil
	}
	p.lastmatch = ouid

	return self.beginFightWith(conn, p, ouid)
}

func (self *CNServer) beginFightWith(conn rpc.RpcConn, p *player, ouid string) error {
	//lock
	lid := GenLockMessage(self.GetServerId(), proto.MethodPlayerMatch, 0)
	ok, _, err := lockclient.TryLock("player", ouid, lid)
	if err != nil {
		logger.Error("beginFightWith LockGet err:", err)
		WriteMatchResult(conn, rpc.MatchPlayerResult_SERVERERROR)
		return nil
	}
	if !ok {
		logger.Info("LockGet:Onfire!") //test

		SendMsg(conn, "TID_ENEMY_ONLINE")

		WriteMatchResult(conn, rpc.MatchPlayerResult_ISONFIRE)
		return nil
	}

	//load otherplayer
	op := LoadOtherPlayer(ouid, lid)
	if op == nil {
		lockclient.TryUnlock("player", ouid, lid)

		WriteMatchResult(conn, rpc.MatchPlayerResult_SERVERERROR)
		return nil
	}

	ov := op.GetVillage()
	if ov == nil {
		lockclient.TryUnlock("player", ouid, lid)

		WriteMatchResult(conn, rpc.MatchPlayerResult_SERVERERROR)
		return nil
	}

	//cost yuanbao
	p.CostResource(GetAttackCost(p.v.getCenterLevel()), proto.ResType_Gold, proto.Lose_Search)

	//send result to client
	rep := rpc.MatchPlayer{}
	rep.SetAct(rpc.MatchPlayer_ATTACK)
	rep.SetName(op.GetName())
	rep.SetTrophy(op.GetTrophy())
	rep.SetLevel(op.GetLevel())
	rep.SetV(ov.VillageInfo)

	maxGold, totalGold := p.v.collect_GetStorageGoldLimit()
	maxFood, totalFood := p.v.collect_GetStorageFoodLimit()
	rep.SetOwnGold(totalGold)
	rep.SetOwnFood(totalFood)
	rep.SetMaxGold(maxGold)
	rep.SetMaxFood(maxFood)
	rep.SetOwnDiamond(p.GetPlayerTotalGem())
	rep.SetOwnTrophy(p.GetTrophy())
	rep.SetOwnTownlevel(p.v.getCenterLevel())

	rep.OwnChar = p.v.barrack_GetAllCharacters()
	rep.OwnSpell = p.v.spellForge_GetAllSpells()

	rep.ClanForce = p.v.castle_GetClanForce()

	self.addOtherPlayer(conn.GetId(), op)

	p.OnFightBegin(conn.GetId(), op, 0, false, false)
	ts("this by")
	logger.Info("this is by tung", rep.OwnSpell)

	if WriteMatchResult(conn, rpc.MatchPlayerResult_OK) && WriteResult(conn, &rep) {
		//logger.Info("WriteResult:", ouid) //test
	}

	return nil
}

func (self *CNServer) TryDrill(conn rpc.RpcConn, msg rpc.TryDrill) error {
	ts("CNServer:TryDrill", conn.GetId())
	defer te("CNServer:TryDrill", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	p.TryDrill()

	return nil
}

func (self *CNServer) TryFriendDrill(conn rpc.RpcConn, msg rpc.TryFriendDrill) error {
	ts("CNServer:TryFriendDrill", conn.GetId())
	defer te("CNServer:TryFriendDrill", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	p.TryFriendDrill()

	return nil
}

//客户端战斗完毕
func (self *CNServer) NotifyBattleEnd(conn rpc.RpcConn, msg rpc.NotifyBattleEnd) error {
	ts("CNServer:NotifyBattleEnd", conn.GetId())
	defer te("CNServer:NotifyBattleEnd", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	p.OnFightEnd(&msg)

	return nil
}

//客户端PVE战斗开始
func (self *CNServer) NotifyPVEBattleStart(conn rpc.RpcConn, info rpc.NotifyPVEBattleStart) error {
	ts("CNServer:NotifyPVEBattleStart", conn.GetId())
	defer te("CNServer:NotifyPVEBattleStart", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()
	if !exist {
		return nil
	}

	// 如果已经处于战斗状态，则不能再次战斗，不管是PVE，还是PVP
	if !p.CanPvEFight() {
		return nil
	}

	msg := rpc.PVEAttackerInfo{}

	msg.OwnChar = p.v.barrack_GetAllCharacters()
	msg.OwnSpell = p.v.spellForge_GetAllSpells()
	//test
	p.NotifyPVEBattleStart(conn.GetId(), info)

	WriteResult(conn, &msg)

	return nil
}

// 客户端PVE战斗结束
func (self *CNServer) NotifyPVEBattleEnd(conn rpc.RpcConn, msg rpc.NotifyPVEBattleEnd) error {
	ts("CNServer:NotifyPVEBattleEnd", conn.GetId())
	defer te("CNServer:NotifyPVEBattleEnd", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	p.NotifyPVEBattleEnd(&msg)

	return nil
}

func (self *CNServer) Visit(conn rpc.RpcConn, try rpc.TryVisit) error {
	ts("CNServer:Visit", conn.GetId())
	defer te("CNServer:Visit", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()
	if !exist {
		return nil
	}

	self.l.RLock()
	vp, exist := self.otherplayers[conn.GetId()]
	self.l.RUnlock()
	if !exist {
		//load otherplayer
		vp = LoadPlayerToVisit(try.GetUid())
		if vp == nil {
			return nil
		}
	}

	vv := vp.GetVillage()
	if vv == nil {
		return nil
	}

	//send result to client
	rep := rpc.MatchPlayer{}
	rep.SetAct(rpc.MatchPlayer_VISIT)
	rep.SetName(vp.GetName())
	rep.SetTrophy(vp.GetTrophy())
	rep.SetLevel(vp.GetLevel())
	rep.SetV(vv.VillageInfo)

	maxGold, totalGold := p.v.collect_GetStorageGoldLimit()
	maxFood, totalFood := p.v.collect_GetStorageFoodLimit()
	rep.SetOwnGold(totalGold)
	rep.SetOwnFood(totalFood)
	rep.SetMaxGold(maxGold)
	rep.SetMaxFood(maxFood)
	rep.SetOwnDiamond(p.GetPlayerTotalGem())
	rep.SetOwnTrophy(p.GetTrophy())

	if WriteResult(conn, &rep) {
		logger.Info("WriteResult:", try.GetUid()) //test
	}

	return nil
}

func (self *CNServer) Revenge(conn rpc.RpcConn, try rpc.TryRevenge) error {
	ts("CNServer:Revenge", conn.GetId())
	defer te("CNServer:Revenge", conn.GetId())

	ouid := try.GetUid() //PlayerUid
	bid := try.GetBid()  //BattleLogId

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()
	if !exist || !p.CanRevengeFight() || !p.CanRevenge(bid) {
		logger.Error("Revenge self error")
		return nil
	}

	//lock
	lid := GenLockMessage(self.GetServerId(), proto.MethodPlayerRevenge, 0)
	ok, _, err := lockclient.TryLock("player", ouid, lid)
	if err != nil {
		WriteMatchResult(conn, rpc.MatchPlayerResult_SERVERERROR)

		logger.Error("Revenge LockGet err:", err)
		return nil
	}
	if !ok {
		logger.Info("LockGet:Onfire") //test

		SendMsg(conn, "TID_ENEMY_ONLINE")

		WriteMatchResult(conn, rpc.MatchPlayerResult_ISONFIRE)
		return nil
	}

	//load otherplayer
	op := LoadOtherPlayer(ouid, lid)
	if op == nil {
		logger.Error("Revenge load other failed")

		lockclient.TryUnlock("player", ouid, lid)

		WriteMatchResult(conn, rpc.MatchPlayerResult_SERVERERROR)
		return nil
	}

	ov := op.GetVillage()
	if ov == nil {
		logger.Error("Revenge get other village failed")

		lockclient.TryUnlock("player", ouid, lid)

		WriteMatchResult(conn, rpc.MatchPlayerResult_SERVERERROR)
		return nil
	}

	//send result to client
	rep := rpc.MatchPlayer{}
	rep.SetAct(rpc.MatchPlayer_REVENGE)
	rep.SetName(op.GetName())
	rep.SetTrophy(op.GetTrophy())
	rep.SetLevel(op.GetLevel())
	rep.SetV(ov.VillageInfo)

	maxGold, totalGold := p.v.collect_GetStorageGoldLimit()
	maxFood, totalFood := p.v.collect_GetStorageFoodLimit()
	rep.SetOwnGold(totalGold)
	rep.SetOwnFood(totalFood)
	rep.SetMaxGold(maxGold)
	rep.SetMaxFood(maxFood)
	rep.SetOwnDiamond(p.GetPlayerTotalGem())
	rep.SetOwnTrophy(p.GetTrophy())
	rep.SetOwnTownlevel(p.v.getCenterLevel())

	rep.OwnChar = p.v.barrack_GetAllCharacters()
	rep.OwnSpell = p.v.spellForge_GetAllSpells()
	rep.ClanForce = p.v.castle_GetClanForce()

	self.addOtherPlayer(conn.GetId(), op)

	p.OnFightBegin(conn.GetId(), op, bid, false, false)

	if WriteMatchResult(conn, rpc.MatchPlayerResult_OK) && WriteResult(conn, &rep) {
		//logger.Info("WriteResult:", ouid) //test
	}

	return nil
}

func (self *CNServer) Replay(conn rpc.RpcConn, try rpc.TryReplay) error {
	ts("CNServer:Replay", conn.GetId())
	defer te("CNServer:Replay", conn.GetId())

	replay := rpc.BattleReplay{}

	exist, err := KVQueryExt("replay", strconv.FormatUint(try.GetRid(), 16), &replay)
	if err != nil {
		logger.Info("KVQuery error:%s", err.Error())
		return nil
	}

	if exist {
		WriteResult(conn, &replay)
	} else {
		logger.Info("KVQuery nothing:%d, %d", try.GetRid(), strconv.FormatUint(try.GetRid(), 16))
	}

	return nil
}

func (self *CNServer) ReturnHome(conn rpc.RpcConn, ping rpc.Ping) error {
	ts("CNServer:ReturnHome", conn.GetId())
	defer te("CNServer:ReturnHome", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if exist {
		p.ReturnHome()

		self.delOtherPlayer(conn.GetId())
	}

	p.SyncPlayerGem()

	return nil
}

func (self *CNServer) FriendExecise(conn rpc.RpcConn, try rpc.TryFriendExecise) error {
	ts("CNServer:FriendExecise", conn.GetId())
	defer te("CNServer:FriendExecise", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()
	if !exist || p.GetFriendDrillTimes() == 0 {
		logger.Error("FriendExecise self error")
		return nil
	}

	self.l.RLock()
	vp, exist := self.otherplayers[conn.GetId()]
	self.l.RUnlock()
	if !exist {
		//load otherplayer
		vp = LoadPlayerToVisit(try.GetUid())
		if vp == nil {
			return nil
		}
	}

	vv := vp.GetVillage()
	if vv == nil {
		return nil
	}

	//send result to client
	rep := rpc.MatchPlayer{}
	rep.SetAct(rpc.MatchPlayer_FRIEND_EXECISE)
	rep.SetName(vp.GetName())
	rep.SetTrophy(vp.GetTrophy())
	rep.SetLevel(vp.GetLevel())
	rep.SetV(vv.VillageInfo)

	maxGold, totalGold := p.v.collect_GetStorageGoldLimit()
	maxFood, totalFood := p.v.collect_GetStorageFoodLimit()
	rep.SetOwnGold(totalGold)
	rep.SetOwnFood(totalFood)
	rep.SetMaxGold(maxGold)
	rep.SetMaxFood(maxFood)
	rep.SetOwnDiamond(p.GetPlayerTotalGem())
	rep.SetOwnTrophy(p.GetTrophy())
	rep.SetOwnTownlevel(p.v.getCenterLevel())

	rep.OwnChar = p.v.barrack_GetAllCharacters()
	rep.OwnSpell = p.v.spellForge_GetAllSpells()
	rep.ClanForce = p.v.castle_GetClanForce()

	p.TryFriendDrill()
	self.addOtherPlayer(conn.GetId(), vp)
	p.OnFightBegin(conn.GetId(), vp, 0, false, true)

	if WriteMatchResult(conn, rpc.MatchPlayerResult_OK) && WriteResult(conn, &rep) {
		//logger.Info("WriteResult:", ouid) //test
	}

	return nil
}

func (self *CNServer) GetFriendExeciseData(conn rpc.RpcConn, msg rpc.Ping) error {
	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()
	if !exist {
		return nil
	}

	p.InitFriendExeciseLog(true)
	return nil
}

func (self *CNServer) UpdateFriendExeciseDaily(conn rpc.RpcConn, msg rpc.Ping) error {
	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()
	if !exist {
		return nil
	}

	if p.GetFriendDrillTimes() < 10 {
		p.SetFriendDrillTimes(p.GetFriendDrillTimes() + 1)
	}
	return nil
}

//玩家放兵
func (self *CNServer) AttackerInfo(conn rpc.RpcConn, msg rpc.AttackerInfo) error {
	//ts("CNServer:AttackerInfo", conn.GetId())
	//defer te("CNServer:AttackerInfo", conn.GetId())
	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	p.AddAttacker(&msg)
	return nil
}

// 玩家使用药水
func (self *CNServer) SpellInfo(conn rpc.RpcConn, msg rpc.SpellInfo) error {
	//ts("CNServer:SpellInfo", conn.GetId())
	//defer te("CNServer:SpellInfo", conn.GetId())
	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	p.AddSpell(&msg)
	return nil
}

//玩家放联盟
func (self *CNServer) ClanForceInfo(conn rpc.RpcConn, msg rpc.ClanForceInfo) error {
	//ts("CNServer:ClanForceInfo", conn.GetId())
	//defer te("CNServer:ClanForceInfo", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	p.AddClanForce(&msg)
	return nil
}

//战斗服务器计算返回的结果结算
func (self *CNServer) AttackEnd(conn rpc.RpcConn, msg rpc.AttackEnd) error {
	ts("CNServer:AttackEnd", conn.GetId())
	defer te("CNServer:AttackEnd", conn.GetId())
	uLid := *msg.Playerlid
	self.l.RLock()
	p, exist := self.players[uLid]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	//add for challenge
	bIsChallenge := p.fightinfo.bIsChallenge

	bFriendExecise := p.fightinfo.bFriendExecise
	bTTTFighting := p.fightinfo.bTTTFight
	op, exist := self.otherplayers[uLid]
	if exist {
		if bTTTFighting || bFriendExecise || bIsChallenge {
			if bIsChallenge {
				fmt.Println("现在结算的是擂台赛")
			}

			p.OnPlayerBattleResult(op, &msg)
			delete(self.otherplayers, uLid)
			delete(self.playersbyid, op.GetUid())
			// 刷新擂台列表
			Normallist := &rpc.GetNormalChallengeList{}
			Moneylist := &rpc.GetMoneylChallengeList{}
			self.GetNormalChallengeList(conn, *Normallist)
			self.GetMoneylChallengeList(conn, *Moneylist)
			return nil
		} else {
			p.OnPlayerBattleResult(op, &msg)
		}
	}

	//op, exist := self.otherplayers[uLid]
	//if exist {
	//	fmt.Println("战斗结束 ")
	//	p.OnPlayerBattleResult(op, &msg)
	//}

	self.delOtherPlayer(uLid)

	//推送消息
	if exist && op.PlayerBaseInfo != nil && op.PlayerExtraInfo != nil {
		if p.PlayerBaseInfo != nil && p.PlayerExtraInfo != nil {
			go pushmsg.PushMsg(op.GetUid(), "", language.GetLocationLanguage("TID_NOTIFY_ONATTACK_", op.GetGamelocation(), p.GetName()))
		} else {
			logger.Info("push:self player not online", uLid)
		}
	} else {
		logger.Info("push:other player not online")
	}

	return nil
}

//战斗服务器计算返回PVE结果结算
func (self *CNServer) PVEAttackEnd(conn rpc.RpcConn, msg rpc.PVEAttackEnd) error {
	ts("CNServer:PVEAttackEnd", conn.GetId(), msg)
	defer te("CNServer:PVEAttackEnd", conn.GetId(), msg)

	uLid := *msg.Playerid

	self.l.RLock()
	p, exist := self.players[uLid]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	p.OnPlayerPveFightResult(&msg)

	return nil
}

func (self *CNServer) AddTrophy(conn rpc.RpcConn, msg rpc.UpdatePlayerInfo) error {
	ts("CNServer:AddTrophy", conn.GetId(), msg)
	defer te("CNServer:AddTrophy", conn.GetId(), msg)

	if common.IsOpenTrophyAdd() == false {
		return nil
	}

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	p.GainResource(msg.GetTrophy(), proto.ResType_Trophy, proto.Gain_Plunder)

	return nil
}

//通天塔相关功能，开始
func (self *CNServer) TryRandomTTTCharacter(conn rpc.RpcConn, msg rpc.TryRandomTTTCharacter) error {
	self.l.RLock()
	_, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	//先屏蔽，以后要打开请修改里面的扣元宝流程
	//p.RandomTTTCharacterMultiples(msg.GetFree())

	return nil
}

func (self *CNServer) TryStartTTT(conn rpc.RpcConn, msg rpc.TryStartTTT) error {
	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	if msg.GetIsBreakORReast() {
		pTTTInfo1 := p.GetSelftttinfo()
		if pTTTInfo1 == nil {
			return nil
		}
		fmt.Println("*****TryStartTTT***断线重连接******1****")
		if msg.GetIsRelive() {
			relive_Times := GetGlobalCfg("TTT_RELIVE_TIMES")
			if pTTTInfo1.GetReLiveCount() > relive_Times {
				logger.Error("relive times more than config")
				return nil
			}
			fmt.Println("*****TryStartTTT***断线重连接*****复活**2***")
			//复活花费宝石
			//.need fix
			spendnumber_Dem := GetGlobalCfg("TTT_RELIVE_COST")
			if p.GetPlayerTotalGem() < spendnumber_Dem {
				logger.Error(" Changetimes no Dem ! ")
				return nil
			} else {
				fmt.Println("复活  ：：花费  前   ", p.GetPlayerTotalGem())
				//先确定扣除成功再做后面的操作
				if !p.CostResource(spendnumber_Dem, proto.ResType_Gem, proto.Lose_TTTBattle) {
					logger.Error("relive cost Dimonds Failed ")
					return nil
				}
			}
			//复活的时候初始化数据
			pTTTInfo1.SetIsTTTOver(false)
			pTTTInfo1.SetIsStart(true)
			pTTTInfo1.Characters = nil
			pTTTInfo1.Spells = nil
			pTTTInfo1.Tttbuffs = nil

			if len(pTTTInfo1.Startcharacters) != 0 {
				for _, cfromP := range pTTTInfo1.Startcharacters {
					//count := cfromP.GetCount() / 2
					//if (cfromP.GetCount() % 2) != 0 {
					//	count += 1
					//}
					chr := &rpc.Character{}
					chr.SetType(cfromP.GetType())
					chr.SetCount(cfromP.GetCount())
					chr.SetLevel(cfromP.GetLevel())
					fmt.Println("charater  记录 charater 数量", cfromP.GetCount())
					pTTTInfo1.Characters = append(pTTTInfo1.Characters, chr)
				}
			}
			if len(pTTTInfo1.Startspells) != 0 {
				for _, sfromP := range pTTTInfo1.Startspells {
					//count := sfromP.GetCount() / 2
					//if (sfromP.GetCount() % 2) != 0 {
					//	count += 1
					//}
					spell := &rpc.Spell{}
					spell.SetType(sfromP.GetType())
					spell.SetCount(sfromP.GetCount())
					spell.SetLevel(sfromP.GetLevel())
					pTTTInfo1.Spells = append(pTTTInfo1.Spells, spell)
				}
			}
			if len(pTTTInfo1.Starttttbuffs) != 0 {
				for _, tbuff := range pTTTInfo1.Starttttbuffs {
					tttbuff := &rpc.TTTBuff{}
					tttbuff.SetType(tbuff.GetType())
					tttbuff.SetCount(tbuff.GetCount())
					fmt.Println("buff  记录 buff类型", tttbuff.GetType())
					pTTTInfo1.Tttbuffs = append(pTTTInfo1.Tttbuffs, tttbuff)
				}
			}
			pTTTInfo1.SetReLiveCount(pTTTInfo1.GetReLiveCount() + 1)
			fmt.Println("打印 复活次数   ", pTTTInfo1.GetReLiveCount())
			fmt.Println("导入 Startcharacters: Startspells: Starttttbuffs:", len(pTTTInfo1.Startcharacters), len(pTTTInfo1.Startspells), len(pTTTInfo1.Starttttbuffs))
		}
	}
	//下发开始战斗的关卡
	if p.conn == nil {
		return nil
	}
	fmt.Println("*****TryStartTTT****上传战斗消息******1****")
	//去组装数据
	if !p.BeginTTTFight(msg) {
		//首先应该还原和反馈数据

		logger.Error(" Made data error ")
		fmt.Println("*****去组装数据 错误*****")
		//ping := &rpc.Ping{}
		//self.ReturnHome(conn, *ping)
		//fmt.Println("返回城镇 ")
		return nil
	}

	//保证一定有，且次数等判断已经在外面处理过了
	pTTTInfo := p.GetSelftttinfo()
	if pTTTInfo == nil {
		return nil
	}
	fmt.Println("*****TryStartTTT****上传战斗消息*****2***对手ID**", pTTTInfo.GetMatchedplayerid())
	//todo 如果有匹配的直接使用，否则从center根据关卡重新匹配
	if pTTTInfo.GetMatchedplayerid() == "" {
		fmt.Println("*TryStartTTT*战斗开始 对手id 被 重置 为空 需要重新匹配一个 id*")
		curc := pTTTInfo.GetCurcheckpoint()
		fmt.Println("the curcheckpoint = ", curc)
		cfg := GetTTTCfg(strconv.FormatInt(int64(curc), 10))
		if cfg == nil || cfg.TownhallLevel == 0 {
			fmt.Println("no config or config error 当前关卡  配置的等级 ", cfg.TownhallLevel)
			logger.Info(" no config or config error", cfg.TownhallLevel)
			return nil
		}
		req := &proto.RandomGetPlayerIdByLevel{Level: cfg.TownhallLevel}
		rst := &proto.RandomGetPlayerIdByLevelResult{}
		fmt.Println("每次正常战斗开始匹配 对手 当前的关卡   匹配的对手的等级数  ", curc, cfg.TownhallLevel)
		err := cns.center.Call("Center.RandomGetPlayerIdByLevel", req, rst)
		if err != nil {
			logger.Info(" Get match player reeor")
			fmt.Println("*****取得对手错误******3****", err)
			ping := &rpc.Ping{}
			self.ReturnHome(conn, *ping)
			fmt.Println("返回城镇 ")
			return nil
		}
		fmt.Println("取得 对手的id", rst.Id)
		if len(rst.Id) == 0 {
			fmt.Println("取得对手的id长度== 0 表示没有取得对手的uid")
			return nil
		}
		pTTTInfo.SetMatchedplayerid(rst.Id)
	} else {

	}
	self.l.RLock()
	vp, exist := self.otherplayers[conn.GetId()]
	self.l.RUnlock()
	if !exist {
		//load otherplayer
		vp = LoadPlayerToVisit(pTTTInfo.GetMatchedplayerid())
		if vp == nil {
			logger.Info(" LoadPlayerToVisit match player error")
			return nil
		}
	}
	vv := vp.GetVillage()
	if vv == nil {
		logger.Info(" GetVillage  data error")
		return nil
	}
	//todo 这里保存一次
	pTTTCheckpoint := &rpc.SynTTTCheckpoint{}
	pTTTCheckpoint.SetCurcheckpoint(pTTTInfo.GetCurcheckpoint())
	pTTTCheckpoint.SetChangetimes(pTTTInfo.GetChangetimes())
	pTTTCheckpoint.SetReliveCount(pTTTInfo.GetReLiveCount())
	if len(pTTTInfo.Characters) != 0 {
		for _, cfromP := range pTTTInfo.Characters {
			chr := &rpc.Character{}
			chr.SetType(cfromP.GetType())
			chr.SetCount(cfromP.GetCount())
			chr.SetLevel(cfromP.GetLevel())
			fmt.Println("2 兵 保存 类型 等级 数量", cfromP.GetType(), cfromP.GetCount(), cfromP.GetLevel())
			pTTTCheckpoint.Characters = append(pTTTCheckpoint.Characters, chr)
		}
	}
	if len(pTTTInfo.Spells) != 0 {
		for _, sfromP := range pTTTInfo.Spells {
			spell := &rpc.Spell{}
			spell.SetType(sfromP.GetType())
			spell.SetCount(sfromP.GetCount())
			spell.SetLevel(sfromP.GetLevel())
			fmt.Println("2 丹 保存 类型 等级 数量", sfromP.GetType(), sfromP.GetCount(), sfromP.GetLevel())
			pTTTCheckpoint.Spells = append(pTTTCheckpoint.Spells, spell)
		}
	}
	//为了测试这里的村庄是指定的自己的村庄
	pTTTCheckpoint.SetV(vv.VillageInfo)
	if msg.GetIsBreakORReast() {
		pTTTCheckpoint.SetIsBreakORReast(true)
	}
	if msg.GetIsContinueBattle() {
		pTTTCheckpoint.SetIsContinueBattle(true)
	}
	if msg.GetIsRelive() {
		pTTTCheckpoint.SetIsRelive(true)
	}
	self.addOtherPlayer(conn.GetId(), vp)
	p.OnFightBegin(conn.GetId(), vp, 0, true, false)
	fmt.Println("每次闯关 开始战斗关卡 切换对手的次数 ", pTTTCheckpoint.GetCurcheckpoint(), pTTTCheckpoint.GetChangetimes)
	//todo 还要添加配置到的玩家的数据
	//没次进入消耗一点体力
	spendNum_Tili := GetGlobalCfg("TTT_ENTRE_COST")
	if !p.CostResource(spendNum_Tili, proto.ResType_TiLi, proto.Lose_TTTBattle) {
		logger.Error("BeginTTTFight Failed : wrong spend tili count(%d/%d)", spendNum_Tili, p.GetTili())
		return nil
	}
	fmt.Println("当前体力 ", p.GetTili())
	WriteResult(p.conn, pTTTCheckpoint)

	return nil
}
func (self *CNServer) TryMatchNextCheckpointPlayer(conn rpc.RpcConn, msg rpc.TryMatchNextCheckpointPlayer) error {
	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}
	// 删除上一个匹配到的对手
	self.delOtherPlayer(conn.GetId())

	p.MatchNextCheckpointPlayer(msg)
	//下发开始战斗的关卡
	if p.conn == nil {
		return nil
	}
	//保证一定有，且次数等判断已经在外面处理过了
	pTTTInfo := p.GetSelftttinfo()
	if pTTTInfo == nil {
		return nil
	}
	pTTTInfo.SetIsTTTEndEveryCheckpoint(true)
	//todo 如果有匹配的直接使用，否则从center根据关卡重新匹配
	fmt.Println("*****TryStartTTT****上传战斗消息***match***对手ID**", pTTTInfo.GetMatchedplayerid())
	if pTTTInfo.GetMatchedplayerid() == "" {
		fmt.Println("* TryMatchNext*战斗开始 对手id 被 重置 为空 需要重新匹配一个 id*")
		curc := pTTTInfo.GetCurcheckpoint()
		cfg := GetTTTCfg(strconv.FormatInt(int64(curc), 10))
		if cfg == nil || cfg.TownhallLevel == 0 {
			fmt.Println("no config or config error 当前关卡  配置的等级 ", cfg.TownhallLevel)
			logger.Info(" no config or config error", cfg.TownhallLevel)
			return nil
		}
		req := &proto.RandomGetPlayerIdByLevel{Level: cfg.TownhallLevel}
		rst := &proto.RandomGetPlayerIdByLevelResult{}
		fmt.Println("切换对手战斗开始匹配 对手 当前的关卡   匹配的对手的等级数  ", curc, cfg.TownhallLevel)
		err := cns.center.Call("Center.RandomGetPlayerIdByLevel", req, rst)
		if err != nil {
			fmt.Println("*****TryStartTTT****上传战斗消息******3****", err)
			fmt.Println("取得对手错误 ")
			ping := &rpc.Ping{}
			self.ReturnHome(conn, *ping)
			fmt.Println("返回城镇 ")
			return nil
		}
		if len(rst.Id) == 0 {
			fmt.Println("取得对手的id长度== 0")
			return nil
		}
		fmt.Println("根据本数匹配到的  对手 id ", rst.Id)
		pTTTInfo.SetMatchedplayerid(rst.Id)
		//todo 判断关卡上限是否要加1
	}
	self.l.RLock()
	vp, exist := self.otherplayers[conn.GetId()]
	self.l.RUnlock()
	if !exist {
		//load otherplayer
		vp = LoadPlayerToVisit(pTTTInfo.GetMatchedplayerid())
		if vp == nil {
			return nil
		}
	}

	vv := vp.GetVillage()
	if vv == nil {
		return nil
	}
	//todo 这里向db取玩家数据，这里的玩家数据则为村庄id
	//这里花元宝 对配置表的问题
	//fmt.Println("切换对手  ：：  次数  ", pTTTInfo.GetChangetimes())
	var spendnumber uint32 = 0
	if pTTTInfo.GetChangetimes() <= 8 {
		spendnumber = GetGlobalCfg("TTT_CHANGE_COST_" + strconv.FormatInt(int64(pTTTInfo.GetChangetimes()), 10))
		fmt.Println("切换对手  ：：< 8  次数  花费  ", pTTTInfo.GetChangetimes(), spendnumber)
	} else {
		spendnumber = GetGlobalCfg("TTT_CHANGE_COST_MAX")
		fmt.Println("切换对手  ：： >8  次数  花费  ", pTTTInfo.GetChangetimes(), spendnumber)
	}

	if p.GetPlayerTotalGem() < spendnumber {
		pTTTInfo.SetChangetimes(pTTTInfo.GetChangetimes() - 1)
		logger.Error(" Changetimes no Dem ! ")
		return nil
	} else {
		fmt.Println("切换对手  ：：花费  前   ", p.GetPlayerTotalGem())

		if !p.CostResource(spendnumber, proto.ResType_Gem, proto.Lose_TTTBattle) {
			logger.Error("relive  cost  failed")
			return nil
		}
		fmt.Println("切换对手  ：：花费  后   ", p.GetPlayerTotalGem())
	}
	//todo 这里保存一次
	pTTTCheckpoint := &rpc.SynTTTCheckpoint{}
	pTTTCheckpoint.SetCurcheckpoint(pTTTInfo.GetCurcheckpoint())
	pTTTCheckpoint.SetChangetimes(pTTTInfo.GetChangetimes())
	pTTTCheckpoint.SetReliveCount(pTTTInfo.GetReLiveCount())
	if len(pTTTInfo.Characters) != 0 {
		for _, cfromP := range pTTTInfo.Characters {
			chr := &rpc.Character{}
			chr.SetType(cfromP.GetType())
			chr.SetCount(cfromP.GetCount())
			chr.SetLevel(cfromP.GetLevel())
			fmt.Println("2a 兵 保存 类型 等级 数量", cfromP.GetType(), cfromP.GetCount(), cfromP.GetLevel())
			pTTTCheckpoint.Characters = append(pTTTCheckpoint.Characters, chr)
		}
	}
	if len(pTTTInfo.Spells) != 0 {
		for _, sfromP := range pTTTInfo.Spells {
			spell := &rpc.Spell{}
			spell.SetType(sfromP.GetType())
			spell.SetCount(sfromP.GetCount())
			spell.SetLevel(sfromP.GetLevel())
			fmt.Println("2a 丹 保存 类型 等级 数量", sfromP.GetType(), sfromP.GetCount(), sfromP.GetLevel())
			pTTTCheckpoint.Spells = append(pTTTCheckpoint.Spells, spell)
		}
	}
	pTTTCheckpoint.SetIsBreakORReast(false)
	pTTTCheckpoint.SetIsContinueBattle(true)
	//为了测试这里的村庄是指定的自己的村庄
	pTTTCheckpoint.SetV(vv.VillageInfo)
	fmt.Println("切换关卡 开始战斗关卡 切换对手的次数 ", pTTTCheckpoint.GetCurcheckpoint(), pTTTCheckpoint.GetChangetimes)
	self.addOtherPlayer(conn.GetId(), vp)
	p.OnFightBegin(conn.GetId(), vp, 0, true, false)

	//todo 还要添加配置到的玩家的数据
	WriteResult(p.conn, pTTTCheckpoint)

	return nil
}

/*
//修改查询接口，该函数被弃用
func (self *CNServer) GetTTTRankPlayers(conn rpc.RpcConn, msg rpc.TryGetTTTRankPlayers) error {

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}
	if p.conn == nil {
		return nil
	}

	try := &proto.GetRankPlayerTTTScore{Start: 0, Stop: 99}
	ret := &proto.GetRankPlayerTTTScoreResult{}

	err := cns.center.Call("Center.GetRankPlayerTTTScore", try, ret)
	if err != nil {
		return nil
	}

	rps := &rpc.TTTRankPlayers{}
	for _, PlayerTTTScoreStructinfo := range ret.Value {
		var TTTp rpc.PlayerInfo
		fmt.Println("*****GetTTTRankPlayers****取出的数据里面的uid******score****", PlayerTTTScoreStructinfo.Id, PlayerTTTScoreStructinfo.Score)
		exist, err := KVQuery("player", PlayerTTTScoreStructinfo.Id, &TTTp)
		if err != nil {
			continue
		}
		fmt.Println("*****GetTTTRankPlayers****有对象存在**********")
		if exist {
			rp := rpc.TTTPlayer{}
			rp.SetType(rpc.TTTPlayer_Rank)
			rp.SetName(TTTp.GetName())
			rp.SetUid(TTTp.GetUid())
			rp.SetLevel(TTTp.GetLevel())
			rp.SetTrophy(TTTp.GetTrophy())
			rp.SetClanName(TTTp.GetClan())
			rp.SetClanSymbol(TTTp.GetClanSymbol())
			rp.SetTttSCoreQuery(PlayerTTTScoreStructinfo.Score)
			rp.SetRanknumberQuery(0)

			rps.RpsTop = append(rps.RpsTop, &rp)
		}

	}
	fmt.Println("*****GetTTTRankPlayers****下发TTT排行榜x**********")
	WriteResult(p.conn, rps)

	return nil
}
*/
func (self *CNServer) TryGetScoreAndRank(conn rpc.RpcConn, msg rpc.TryGetScoreAndRank) error {
	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}
	fmt.Println("请求 最高积分  通天塔排行名次  ")
	req := &proto.QueryPlayerTTTScore{Id: p.GetUid()}
	rst := &proto.QueryPlayerTTTScoreResult{}

	err := cns.center.Call("Center.GetPlayerTTTScore", req, rst)
	if err != nil {
		return nil
	}

	tttScoreAndRank := &rpc.TTTScoreAndRank{}
	tttScoreAndRank.SetMostScore(rst.Score)
	tttScoreAndRank.SetTttScoreRank(rst.Rank + 1)
	fmt.Println("服务器下发的 本玩家的 最高积分  通天塔排行名次  ", tttScoreAndRank.GetMostScore(), tttScoreAndRank.GetTttScoreRank())

	WriteResult(p.conn, tttScoreAndRank)

	return nil
}

//add for challenge
func (self *CNServer) Challenge(conn rpc.RpcConn, hostid string, challenge *player, types rpc.MatchPlayer_Act) error {
	ts("CNServer:Challenge", conn.GetId())
	defer te("CNServer:Challenge", conn.GetId())

	self.l.RLock()
	host, exist := self.playersbyid[hostid]
	self.l.RUnlock()

	if !exist {
		//load otherplayer
		host = LoadPlayerToVisit(hostid)

		if host == nil {
			logger.Error("No player")
		}
	}

	if challenge.getCenterLevel() < 5 || challenge.getCenterLevel() > 9 {
		logger.Error("挑战者不存在或者主营等级不在范围之内")
		return nil
	}

	vv := host.GetVillage()
	if vv == nil {
		return nil
	}

	//send result to client
	rep := rpc.MatchPlayer{}
	rep.SetAct(types)
	rep.SetName(host.GetName())
	rep.SetTrophy(host.GetTrophy())
	rep.SetLevel(host.GetLevel())
	rep.SetV(vv.VillageInfo)

	maxGold, totalGold := challenge.v.collect_GetStorageGoldLimit()
	maxFood, totalFood := challenge.v.collect_GetStorageFoodLimit()
	rep.SetOwnGold(totalGold)
	rep.SetOwnFood(totalFood)
	rep.SetMaxGold(maxGold)
	rep.SetMaxFood(maxFood)
	rep.SetOwnDiamond(challenge.GetPlayerTotalGem())
	rep.SetOwnTrophy(challenge.GetTrophy())
	rep.SetOwnTownlevel(challenge.v.getCenterLevel())

	rep.OwnChar = challenge.v.barrack_GetAllCharacters()
	rep.OwnSpell = challenge.v.spellForge_GetAllSpells()
	rep.ClanForce = challenge.v.castle_GetClanForce()

	self.addOtherPlayer(conn.GetId(), host)
	challenge.ChallengeBegin(conn.GetId(), host, 0, true)

	if WriteMatchResult(conn, rpc.MatchPlayerResult_OK) && WriteResult(conn, &rep) {

	}

	return nil
}
