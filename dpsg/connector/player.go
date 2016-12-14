package connector

import (
	"fmt"
	"golang-project/dpsg/common"
	"golang-project/dpsg/lockclient"
	"golang-project/dpsg/logger"
	"golang-project/dpsg/proto"
	"golang-project/dpsg/rpc"
	"golang-project/dpsg/timer"
	"math/rand"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"
)

const MaxLogNumber = 30
const MaxRepNumber = 4

type playerfightinfo struct {
	pvpmsg         rpc.AttackBegin
	pvemsg         rpc.PVEAttackBegin
	def_log        rpc.BattleLog
	replay         rpc.BattleReplay
	uVillageId     uint64
	bPve           bool
	bInFight       bool
	bFriendExecise bool
	bTTTFight      bool
	//add for challenge
	bIsChallenge bool
}

func (self *playerfightinfo) HasAttacker() bool {
	if self.bPve {
		return len(self.pvemsg.Attackunits) > 0 || (self.pvemsg.ClanForceInfo != nil)
	}
	return len(self.pvpmsg.Attackunits) > 0 || (self.pvpmsg.ClanForceInfo != nil)
}
func (self *playerfightinfo) HasSpell() bool {
	if self.bPve {
		return len(self.pvemsg.Spells) > 0
	}
	return len(self.pvpmsg.Spells) > 0
}
func (self *playerfightinfo) AddAttacker(msg *rpc.AttackerInfo) {
	if self.bPve {
		self.pvemsg.Attackunits = append(self.pvemsg.Attackunits, msg)
	} else {
		char_exist := false
		for _, char := range self.def_log.Chars {
			if char.GetType() == *msg.Type {
				char.SetCount(char.GetCount() + 1)
				char_exist = true
				break
			}
		}

		if !char_exist {
			char := &rpc.Character{}
			char.SetType(*msg.Type)
			char.SetLevel(*msg.Level)
			char.SetCount(1)
			self.def_log.Chars = append(self.def_log.Chars, char)
		}

		self.replay.Attackunits = append(self.replay.Attackunits, msg)
		self.pvpmsg.Attackunits = append(self.pvpmsg.Attackunits, msg)
	}
}
func (self *playerfightinfo) AddSpell(msg *rpc.SpellInfo) {
	if self.bPve {
		self.pvemsg.Spells = append(self.pvemsg.Spells, msg)
	} else {
		spell_exist := false
		for _, spell := range self.def_log.Spells {
			if spell.GetType() == *msg.Type {
				spell.SetCount(spell.GetCount() + 1)
				spell_exist = true
				break
			}
		}

		if !spell_exist {
			spell := &rpc.Spell{}
			spell.SetType(*msg.Type)
			spell.SetLevel(*msg.Level)
			spell.SetCount(1)
			self.def_log.Spells = append(self.def_log.Spells, spell)
		}

		self.replay.Spells = append(self.replay.Spells, msg)
		self.pvpmsg.Spells = append(self.pvpmsg.Spells, msg)
	}
}

func (self *playerfightinfo) AddClanForce(msg *rpc.ClanForceInfo) {
	if self.bPve {
		self.pvemsg.ClanForceInfo = msg
	} else {
		self.def_log.ClanForce = msg.ClanForce
		self.replay.ClanForceInfo = msg
		self.pvpmsg.ClanForceInfo = msg
	}
}

//移动QQ保存的数据
type MobileQQInfo struct {
	Openid    string
	Openkey   string
	Pay_token string
	Pf        string
	Pfkey     string
	Balance   uint32
}

type player struct {
	*rpc.PlayerBaseInfo
	*rpc.PlayerExtraInfo
	lid       uint64
	v         *village
	t         *timer.Timer
	conn      rpc.RpcConn
	fightinfo *playerfightinfo
	pve       *rpc.PveStages
	d_logs    *rpc.BattleLogs
	a_logs    *rpc.BattleLogs
	f_logs    *rpc.BattleLogs
	*rpc.FriendsExeciseInfo
	//好友缓存
	friendscache    *rpc.FriendsList
	googlePayNonces map[string]bool //google支付随机值
	lastmatch       string
	myselfGlobal    *rpc.RankPlayers //add for save myself
	myselfLocation  *rpc.RankPlayers //add for save myself
	mobileqqinfo    *MobileQQInfo
	refreshTiLiTick *timer.Timer
	paylock         sync.Mutex
}

func decodeFriends(f *rpc.FriendsIdList) map[string]bool {
	mapfriends := make(map[string]bool)

	for _, fid := range f.Friends {
		mapfriends[fid] = true
	}

	return mapfriends
}

func LoadPlayer(uid string, lid uint64, gl rpc.GameLocation) *player {
	ts("LoadPlayer", uid, lid)
	defer te("LoadPlayer", uid, lid)

	var base rpc.PlayerBaseInfo
	var extra rpc.PlayerExtraInfo

	//fmt.Println("LoadPlayer FMT Print:", "player", uid, p)
	exist, err := KVQueryBase(common.PlayerBase, uid, &base)
	if err != nil {
		logger.Error("query PlayerBase failed!", err)
		return nil
	}

	//有基础信息才查询额外信息
	if exist {
		if exist, err = KVQueryExt(common.PlayerExtra, uid, &extra); err != nil {
			logger.Error("query PlayerExtra failed!", err)
			return nil
		}
	}

	if exist {
		//玩家上线从离线表离拿走，不能再被匹配出来战斗
		if err = NotifyOnline(uid); err != nil {
			return nil
		}

		// 完成基本成员变量组装
		ret := &player{lid: lid, PlayerBaseInfo: &base, PlayerExtraInfo: &extra, fightinfo: nil, pve: nil}

		// 这里去加载pve数据, 因为有可能有角色数据，但是没有任何pve数据
		var pve rpc.PveStages
		exist, err := KVQueryExt("pve", uid, &pve)

		if err != nil {
			return nil
		}

		if exist {
			ret.pve = &pve
		}

		/*var f rpc.FriendsIdList
		exist, err = KVQueryExt("friends", uid, &f)

		if exist {
			ts("Loadfriend", uid, f)
			ret.friends = decodeFriends(&f)

		} else {
			ts("UnLoadfriend", uid, f)
			ret.friends = make(map[string]bool)
		}

		ret.savefriends = true*/

		ret.googlePayNonces = make(map[string]bool)
		ret.SetPower(ret.GetClanPlayerPower())

		return ret
	}

	logger.Error("query Player nothing!")

	//新建
	return NewPlayer(uid, lid, gl)
}

func LoadOtherPlayer(uid string, lid uint64) *player {
	ts("LoadOtherPlayer", uid, lid)
	defer te("LoadOtherPlayer", uid, lid)

	var base rpc.PlayerBaseInfo
	var extra rpc.PlayerExtraInfo

	if exist, err := KVQueryBase(common.PlayerBase, uid, &base); err != nil || !exist {
		return nil
	}

	if exist, err := KVQueryExt(common.PlayerExtra, uid, &extra); err != nil || !exist {
		return nil
	}

	if err := NotifyOnline(uid); err != nil {
		return nil
	}

	ret := &player{lid: lid, PlayerBaseInfo: &base, PlayerExtraInfo: &extra}

	/*
		var f rpc.FriendsIdList
		exist, err = KVQuery("friends", uid, &f)

		if exist {
			ts("Loadfriend OnVisit", uid, f)
			ret.friends = decodeFriends(&f)

		} else {
			ts("UnLoadfriend OnVisit", uid, f)
			ret.friends = make(map[string]bool)
		}
	*/

	return ret
}

func LoadPlayerToVisit(uid string) *player {
	ts("LoadPlayerToVisit", uid)
	defer te("LoadPlayerToVisit", uid)

	var base rpc.PlayerBaseInfo
	var extra rpc.PlayerExtraInfo

	if exist, err := KVQueryBase(common.PlayerBase, uid, &base); err != nil || !exist {
		return nil
	}

	if exist, err := KVQueryExt(common.PlayerExtra, uid, &extra); err != nil || !exist {
		return nil
	}

	ret := &player{PlayerBaseInfo: &base, PlayerExtraInfo: &extra}

	//logger.Info("Player VillageId:<%d>", p.GetVillageId())
	/*
		var f rpc.FriendsIdList
		exist, err = KVQuery("friends", uid, &f)

		if exist {
			ts("Loadfriend OnVisit", uid, f)
			ret.friends = decodeFriends(&f)

		} else {
			ts("UnLoadfriend OnVisit", uid, f)
			ret.friends = make(map[string]bool)
		}
	*/

	return ret
}

func NewPlayer(uid string, lid uint64, gl rpc.GameLocation) *player {
	ts("NewPlayer", uid, lid)
	defer te("NewPlayer", uid, lid)

	var base rpc.PlayerBaseInfo
	var extra rpc.PlayerExtraInfo

	base.SetUid(uid)
	base.SetVillageId(0)
	base.SetName("")
	base.SetTrophy(0)
	base.SetClan("")
	base.SetClanSymbol(0)
	base.SetLevel(1)
	base.SetExp(0)
	base.SetGamelocation(gl)

	extra.SetWuhun(0)
	//读取配置表
	defaultDiamonds := GetGlobalCfg("INIT_GIVE_GEM")
	extra.SetDiamonds(defaultDiamonds)
	extra.SetDrillTimes(10)
	extra.SetFriendDrillTimes(10)
	//.need fix
	tiLi_Ini := GetGlobalCfg("TTT_TILI_INI")
	extra.SetTili(tiLi_Ini)
	extra.SetBattleAccelerateTimes(0)
	extra.SetOnQuitTime(0)

	ret := &player{lid: lid, PlayerBaseInfo: &base, PlayerExtraInfo: &extra, fightinfo: nil, pve: nil, googlePayNonces: make(map[string]bool)}
	//ret.AddShield(3 * 24) //新手3天护盾，策划要求取消

	if result, err := KVWriteBase(common.PlayerBase, uid, &base); err != nil || result == false {
		return nil
	}

	if result, err := KVWriteExt(common.PlayerExtra, uid, &extra); err != nil || result == false {
		return nil
	}

	return ret
}

// 检查是否能够PVP战斗
func (self *player) CanPvPFight() bool {

	if self.fightinfo == nil {
		return self.v.collect_HasEnoughGold(GetAttackCost(self.v.getCenterLevel()))
	}

	if self.fightinfo.bInFight {
		return false
	}

	return self.v.collect_HasEnoughGold(GetAttackCost(self.v.getCenterLevel()))
}

// 检查是否能够PVP战斗
func (self *player) CanPvEFight() bool {
	if self.fightinfo == nil {
		return true
	}

	if self.fightinfo.bInFight {
		return false
	}

	return true
}

//能不能复仇，不要银两
func (self *player) CanRevengeFight() bool {
	if self.fightinfo != nil && self.fightinfo.bInFight {
		return false
	}

	return true
}

func (self *player) ClearFight() {
	self.fightinfo = nil
}

//	 PVE 准备开始战斗
func (self *player) NotifyPVEBattleStart(uConnId uint64, msg rpc.NotifyPVEBattleStart) {
	self.fightinfo = &playerfightinfo{
		pvemsg:   rpc.PVEAttackBegin{},
		bInFight: false,
		bPve:     true,
	}
	self.fightinfo.pvemsg.SetPlayerlid(uConnId)
	logger.Info("NotifyPVEBattleStart", uConnId)

	//是否新手关卡
	if cfg := GetPVEStageCfg(msg.GetStageId()); cfg == nil || !cfg.GuideStage {
		return
	}

	bFirstStart := true
	if self.pve == nil {
		self.pve = &rpc.PveStages{}
	} else {
		for _, stage := range self.pve.Stages {
			if stage.GetStageId() == msg.GetStageId() {
				bFirstStart = false
				break
			}
		}
	}

	//没有打过
	if bFirstStart {
		NewStage := &rpc.Stage{}
		NewStage.SetStageId(msg.GetStageId())
		//如果为100星的表示有开始没结束的情况
		NewStage.SetStars(100)

		self.pve.Stages = append(self.pve.Stages, NewStage)
	}
}

func (self *player) NotifyPVEBattleEnd(msg *rpc.NotifyPVEBattleEnd) bool {
	fmt.Printf("Enter OnFightEnd \n")
	if self.fightinfo == nil {
		return false
	}

	if !self.fightinfo.HasAttacker() {
		self.ClearFight()

		return true
	}

	//找是否有目标stage数据
	StageId := *msg.StageId
	self.fightinfo.pvemsg.Stage = &rpc.Stage{}

	self.fightinfo.pvemsg.Stage.SetStageId(StageId)
	self.fightinfo.pvemsg.SetTotaltime(*msg.Totaltime) // 验证数据。发送

	if self.pve != nil {
		for _, stage := range self.pve.Stages {
			if stage.GetStageId() == StageId {

				if stage.Stars != nil {
					Stars := *stage.Stars
					self.fightinfo.pvemsg.Stage.SetStars(Stars)

					logger.Info("NotifyPVEBattleEnd, Stars:%d", *stage.Stars)
				}

				if stage.CurrentGold != nil {
					CurrentGold := *stage.CurrentGold
					self.fightinfo.pvemsg.Stage.SetCurrentGold(CurrentGold)

					logger.Info("NotifyPVEBattleEnd, Gold:%d", *stage.CurrentGold)
				}

				if stage.CurrentFood != nil {
					CurrentFood := *stage.CurrentFood
					self.fightinfo.pvemsg.Stage.SetCurrentFood(CurrentFood)

					logger.Info("NotifyPVEBattleEnd, Food:%d", *stage.CurrentFood)
				}

				break
			}
		}
	}

	logger.Info("NotifyPVEBattleEnd2, %v", self.fightinfo.pvemsg.Stage)
	fmt.Printf("----pve-------SendFightJob------start--------")
	err := cns.FsMgr.SendFightJob(&self.fightinfo.pvemsg)
	if err != nil {
		fmt.Printf("----pve-------SendFightJob-------errror-------")
		fmt.Print(err.Error())
		return false
	}
	fmt.Printf("OnClientPveFightEnd End \n")
	return true
}

// 准备开始战斗
func (self *player) OnFightBegin(uConnId uint64, op *player, revBattleLogId uint64, bTTTF bool, bFriendExe bool) bool {
	self.fightinfo = &playerfightinfo{
		pvpmsg:         rpc.AttackBegin{},
		uVillageId:     op.GetVillageId(),
		def_log:        rpc.BattleLog{},
		replay:         rpc.BattleReplay{},
		bInFight:       false,
		bPve:           false,
		bFriendExecise: false,
		bTTTFight:      false,
		bIsChallenge:   false,
	}
	//ttt
	fmt.Println("*********OnFightBegin******准备开始战斗****")
	self.fightinfo.bTTTFight = bTTTF
	self.fightinfo.bFriendExecise = bFriendExe
	self.fightinfo.pvpmsg.V = op.v.VillageInfo
	self.fightinfo.pvpmsg.SetPlayerlid(uConnId)
	self.fightinfo.pvpmsg.SetSrcTrophy(self.GetTrophy())
	self.fightinfo.pvpmsg.SetTarTrophy(op.GetTrophy())

	claninfo := self.GetClanInfo()

	def_log := &self.fightinfo.def_log
	def_log.SetBid(revBattleLogId)
	def_log.SetPid(self.GetUid())
	def_log.SetName(self.GetName())
	def_log.SetLevel(self.GetLevel())
	def_log.SetClanName(claninfo.GetName())
	def_log.SetClanSymbol(claninfo.GetSymbol())
	def_log.SetTrophy(self.GetTrophy())
	def_log.SetTime(uint32(time.Now().Unix()))
	def_log.SetState(rpc.BattleLog_UnRead)
	def_log.SetRevstate(rpc.BattleLog_UnRevenged)

	replay := &self.fightinfo.replay
	replay.SetV(op.v.VillageInfo)

	// 破盾
	if self.HasShield() && (self.fightinfo.bFriendExecise == false || self.fightinfo.bTTTFight == false) {
		self.RemoveShield()
	}

	return true
}

func (self *player) OnFightEnd(msg *rpc.NotifyBattleEnd) bool {
	if self.fightinfo == nil {
		return false
	}

	// 如果战斗根本就没开始，那么就不用验证了
	if !self.fightinfo.HasAttacker() && !self.fightinfo.HasSpell() {
		cns.delOtherPlayer(self.fightinfo.pvpmsg.GetPlayerlid())

		self.ClearFight()

		return true
	}

	// 验证数据。发送
	self.fightinfo.pvpmsg.Totaltime = msg.Totaltime
	self.fightinfo.replay.SetTotaltime(msg.GetTotaltime())

	//fmt.Println("send to fightserver :", self.fightinfo.msg.V)
	err := cns.FsMgr.SendFightJob(&self.fightinfo.pvpmsg)

	if err != nil {
		fmt.Print("OnFightEnd Error:", err.Error())
		return false
	}

	//fmt.Printf("OnFightEnd End \n")
	return true
}

//战斗时点击屏幕增加一个攻击单位，就真开始战斗了
func (self *player) AddAttacker(msg *rpc.AttackerInfo) bool {
	if self.fightinfo == nil {
		return false
	}

	self.fightinfo.bInFight = true

	cfg := GetCharacterCfgByTypeId(*msg.Type, 1)
	if cfg == nil {
		return false
	}
	if self.fightinfo.bFriendExecise || self.fightinfo.bTTTFight {
		self.fightinfo.AddAttacker(msg)
		return true
	}

	if cfg.IsHero {
		gs := self.v.buildings_GetAllOf(rpc.BuildingId_GeneralHouse)
		bFound := false
		for _, g := range gs {
			gh := g.(*rpc.GeneralHouse)

			if Hero_Has(gh, *msg.Type) && gh.GetSelectedhero() == *msg.Type {
				bFound = true
				break
			}

			if bFound {
				break
			}
		}
	} else {
		c := self.v.barrack_TroopHousingPopCharacter(*msg.Type)
		if c == nil {
			return false
		}
	}

	self.fightinfo.AddAttacker(msg)

	//logger.Info("AddAttacker End <%d, %d>\n", msg.GetP().GetX(), msg.GetP().GetY())

	return true

}

func (self *player) AddSpell(msg *rpc.SpellInfo) bool {
	if self.fightinfo == nil {
		return false
	}
	self.fightinfo.bInFight = true
	cfg := GetSpellCfgByTypeId(*msg.Type, 1)
	if cfg == nil {
		return false
	} else {
		if self.fightinfo.bFriendExecise || self.fightinfo.bTTTFight {
			self.fightinfo.AddSpell(msg)
			return true
		}

		c := self.v.spell_SpellForgePopSpell(*msg.Type)
		if c == nil {
			return false
		}
	}
	self.fightinfo.AddSpell(msg)

	return true
}

//战斗时点击屏幕增加一个攻击单位，就真开始战斗了
func (self *player) AddClanForce(msg *rpc.ClanForceInfo) bool {
	if self.fightinfo == nil {
		return false
	}
	if self.fightinfo.bFriendExecise {
		return false
	}
	self.fightinfo.bInFight = true

	self.v.castle_PopCharacters(msg.ClanForce.Char)

	self.fightinfo.AddClanForce(msg)

	logger.Info("AddClanForce End <%d, %d>\n", msg.GetP().GetX(), msg.GetP().GetY())

	return true
}

//TTT结算
func (self *player) onPlayerBattleResult_TTT(msg *rpc.AttackEnd) {

	PlayerTTTInfo := self.GetSelftttinfo()
	cfg := GetTTTCfg(strconv.FormatInt(int64(PlayerTTTInfo.GetCurcheckpoint()), 10))
	countGold := cfg.Award1 * msg.GetDamagepercent() / 100
	countFood := cfg.Award2 * msg.GetDamagepercent() / 100
	countWuHun := cfg.Award3 * msg.GetDamagepercent() / 100

	TTTRankPlayersInfo := self.GetTttrankplayerinfo()
	curtttScore := cfg.Mark * msg.GetDamagepercent() / 100
	curtttScore += TTTRankPlayersInfo.GetCurTTTScore()
	TTTRankPlayersInfo.SetCurTTTScore(curtttScore)
	if curtttScore > TTTRankPlayersInfo.GetCurMostTTTScore() {

		TTTRankPlayersInfo.SetCurMostTTTScore(curtttScore)
		req := &proto.UpdatePlayerTTTScore{Id: self.GetUid(), Score: TTTRankPlayersInfo.GetCurMostTTTScore()}
		rst := &proto.UpdatePlayerTTTScoreResult{}
		err := cns.center.Call("Center.UpdatePlayerTTTScore", req, rst)
		if err != nil {
			logger.Error(" TTTScore save error ")
			return
		}
	}
	self.GainResource(countGold, proto.ResType_Gold, proto.Gain_Plunder)
	self.GainResource(countFood, proto.ResType_Food, proto.Gain_Plunder)
	self.GainResource(countWuHun, proto.ResType_Wuhun, proto.Gain_Plunder)
	v := self.GetVillage()
	obj := v.buildings_Get(rpc.BuildingId_GeneralHouse, 0)
	gh := obj.(*rpc.GeneralHouse)
	ctype := gh.GetSelectedhero()

	//也就是说在fight的信息清空之前我们都需要存放一次结果
	var alifecount uint32 = 0
	if len(PlayerTTTInfo.Characters) != 0 {
		for _, cfromP := range PlayerTTTInfo.Characters {
			for _, cfromD := range self.fightinfo.def_log.Chars {
				if cfromP.GetType() == cfromD.GetType() {
					if cfromP.GetType() == ctype {
						cfromP.SetCount(1)
					} else {
						cfromP.SetCount(cfromP.GetCount() - cfromD.GetCount())
					}
				}
			}
			if cfromP.GetType() != ctype {
				alifecount += cfromP.GetCount()
			}
		}
	}
	if len(PlayerTTTInfo.Spells) != 0 {
		for _, sfromP := range PlayerTTTInfo.Spells {
			for _, sfromD := range self.fightinfo.def_log.Spells {
				if sfromP.GetType() == sfromD.GetType() {
					sfromP.SetCount(sfromP.GetCount() - sfromD.GetCount())
				}

			}
			alifecount += sfromP.GetCount()
		}
	}

	//对手信息放空
	PlayerTTTInfo.SetMatchedplayerid("")
	PlayerTTTInfo.SetIsTTTEndEveryCheckpoint(true)
	update := &rpc.UpdatePlayerInfo{}
	//buff
	if msg.GetDamagepercent() >= 100 && alifecount > 0 {
		fmt.Println("TTT 战斗结果成功 ")
		PlayerTTTInfo.SetIsTTTOver(false)
		PlayerTTTInfo.SetIsStart(true)
		update.SetIsTTTStart(PlayerTTTInfo.GetIsStart())
		update.SetIsTTTOver(PlayerTTTInfo.GetIsTTTOver())
		update.SetIsTTTwined(true)
		//.buff
		//每次处理之前要把上一次的清空
		if len(PlayerTTTInfo.Tttbuffs) != 0 {
			for _, tbuff := range PlayerTTTInfo.Tttbuffs {
				tbuff.Characters = nil
				tbuff.Spells = nil
				switch tbuff.GetType() {
				case rpc.TTTBuff_TTTBuffAddArmy:
					{
						//添加兵力buff的处理,随机倍数
						arrTemp := []rpc.CharacterType{}
						for a := int(rpc.CharacterType_Barbarian); a <= int(rpc.CharacterType_PEKKA); a++ {
							charaterCfg := GetCharacterCfgByTypeId(rpc.CharacterType(a), v.laboratory_GetInfoLevel(rpc.CharacterType(a)))
							if charaterCfg == nil {
								logger.Error(" get charatercfg error ")
								break
							}
							var curMostHighLevel uint32 = 1
							for _, ba := range v.Barrack {
								if ba.GetLevel() > curMostHighLevel {
									curMostHighLevel = ba.GetLevel()
								}
							}
							if charaterCfg.BarrackLevel <= curMostHighLevel {
								arrTemp = append(arrTemp, rpc.CharacterType(a))
							}
						}
						raandtypeNumBegin := arrTemp[0]
						raandtypeNumEnd := arrTemp[len(arrTemp)-1]
						raandtypeNum := int(raandtypeNumBegin) + rand.Intn(int(raandtypeNumEnd)-int(raandtypeNumBegin)+1)
						randtype := rpc.CharacterType(raandtypeNum)
						//先挑选解锁的兵种
						var randnumber uint32 = uint32(1 + rand.Intn(10-1))
						for i := 1; i <= int(randnumber); i++ {
							//将buff获得的兵加入到下发的消息里面
							char_exist := false
							for _, char := range tbuff.Characters {
								if char.GetType() == randtype {
									char.SetCount(char.GetCount() + 1)
									char_exist = true
									break
								}
							}
							if !char_exist {
								char := &rpc.Character{}
								char.SetType(randtype)
								char.SetLevel(v.laboratory_GetInfoLevel(randtype))
								char.SetCount(1)
								tbuff.Characters = append(tbuff.Characters, char)
							}
							//将buff获得的兵加入到储存的信息里面
							char_exist1 := false
							for _, char1 := range PlayerTTTInfo.Characters {
								if char1.GetType() == randtype {
									char1.SetCount(char1.GetCount() + 1)
									char_exist1 = true
									break
								}
							}
							if !char_exist1 {
								char1 := &rpc.Character{}
								char1.SetType(randtype)
								char1.SetLevel(v.laboratory_GetInfoLevel(randtype))
								char1.SetCount(1)
								PlayerTTTInfo.Characters = append(PlayerTTTInfo.Characters, char1)
							}
						}
					}
				case rpc.TTTBuff_TTTBuffAddSpell:
					{
						//	//添加丹药buff的处理
						//	raandtypeNumBegin := rpc.SpellType_LighningStorm
						//	raandtypeNumEnd := rpc.SpellType_Haste
						//	raandtypeNum := int(raandtypeNumBegin) + rand.Intn(int(raandtypeNumEnd)-int(raandtypeNumBegin)+1)
						//	randtype := rpc.SpellType(raandtypeNum)

						//	//将buff获得的兵加入到下发的消息里面
						//	spell_exist := false
						//	for _, sp := range tbuff.Spells {
						//		if sp.GetType() == randtype {
						//			sp.SetCount(sp.GetCount() + 1)
						//			spell_exist = true
						//			break
						//		}
						//	}
						//	if !spell_exist {
						//		sp := &rpc.Spell{}
						//		sp.SetType(randtype)
						//		sp.SetLevel(v.laboratory_GetSpellInfoLevel(randtype))
						//		sp.SetCount(1)
						//		tbuff.Spells = append(tbuff.Spells, sp)
						//	}
						//	//将buff获得的兵加入到储存的信息里面
						//	spell_exist1 := false
						//	for _, sp1 := range PlayerTTTInfo.Spells {
						//		if sp1.GetType() == randtype {
						//			sp1.SetCount(sp1.GetCount() + 1)
						//			spell_exist1 = true
						//			break
						//		}
						//	}
						//	if !spell_exist1 {
						//		sp1 := &rpc.Spell{}
						//		sp1.SetType(randtype)
						//		sp1.SetLevel(v.laboratory_GetSpellInfoLevel(randtype))
						//		sp1.SetCount(1)
						//		PlayerTTTInfo.Spells = append(PlayerTTTInfo.Spells, sp1)
						//	}

					}
				}
			}
			update.Tttbuffs = PlayerTTTInfo.Tttbuffs
		}

	} else {
		fmt.Println("TTT 战斗结果失败 ")
		//PlayerTTTInfo.Tttbuffs = nil
		PlayerTTTInfo.SetIsTTTOver(true)
		PlayerTTTInfo.SetIsStart(false)
		update.SetIsTTTStart(PlayerTTTInfo.GetIsStart())
		update.SetIsTTTOver(PlayerTTTInfo.GetIsTTTOver())
		update.SetIsTTTwined(false)
		//.buff
		//判断复活次数使用完才刷新buff
	}
	update.SetWuhun(self.GetWuhun())
	if WriteResult(self.conn, update) {
	}
	self.ClearFight()
	return
}
func (self *player) onPlayerBattleResult_TTT_BUFF() {

}

// 进攻结果，存储战利品，金币和食物
func (self *player) OnPlayerBattleResult(op *player, msg *rpc.AttackEnd) {
	if self.fightinfo == nil {
		return
	}
	if self.fightinfo.bFriendExecise {
		ts("FriendExecise BattleResult")
		op.InitFriendExeciseLog(false)

		friend_log := &self.fightinfo.def_log
		friend_log.SetGoldstolen(msg.GetGoldstolen())
		friend_log.SetFoodstolen(msg.GetFoodstolen())
		friend_log.SetGaintrophy(-msg.GetTrophy())
		friend_log.SetDmgpercent(msg.GetDamagepercent())
		friend_log.SetStarts(msg.GetStarts())
		friend_log.SetWuhun(msg.GetWuhun())
		friend_log.SetRevstate(rpc.BattleLog_Revenged)

		op.InsertFriendExeciseLog(friend_log, &self.fightinfo.replay)
		self.ClearFight()
		return
	}
	//add for challenge

	if self.fightinfo.bIsChallenge {
		self.ClearFight()
		bSuccess := (msg.GetStarts() == 3)
		req := &proto.SendtoCenter{Challengeid: self.GetUid(), Hostid: op.GetUid(), Name: self.GetName(), IsSuccess: bSuccess, IsFinished: true}
		ret := &proto.SendtoCenterResult{}
		cns.center.Go("Center.ChllengeEnd", req, ret, nil)
		return
	}

	if self.fightinfo.bTTTFight {
		self.onPlayerBattleResult_TTT(msg)
		return
	}
	logger.Info("战利品：", msg.GetGoldstolen(), msg.GetFoodstolen(), msg.GetTrophy(), msg.GetWuhun())

	self.GainResource(msg.GetGoldstolen(), proto.ResType_Gold, proto.Gain_Plunder)
	self.GainResource(msg.GetFoodstolen(), proto.ResType_Food, proto.Gain_Plunder)
	self.GainResource(msg.GetWuhun(), proto.ResType_Wuhun, proto.Gain_Plunder)

	if msg.GetTrophy() > 0 {
		self.GainResource(uint32(msg.GetTrophy()), proto.ResType_Trophy, proto.Gain_Plunder)
	} else {
		self.CostResource(uint32(-msg.GetTrophy()), proto.ResType_Trophy, proto.Lose_Plunder)
	}

	if msg.GetWuhun() > 0 {
		update := &rpc.UpdatePlayerInfo{}
		update.SetWuhun(self.GetWuhun())
		WriteResult(self.conn, update)
	}

	LOG_Resources(op.GetGamelocation(), op.GetUid(), false, proto.ResType_Gold, msg.GetGoldstolen(), proto.Lose_Robed)
	LOG_Resources(op.GetGamelocation(), op.GetUid(), false, proto.ResType_Food, msg.GetFoodstolen(), proto.Lose_Robed)
	if msg.GetTrophy() > 0 {
		op.CostResource(uint32(msg.GetTrophy()), proto.ResType_Trophy, proto.Lose_Robed)
	} else {
		op.GainResource(uint32(-msg.GetTrophy()), proto.ResType_Trophy, proto.Gain_Robed)
	}

	//根据百分比加护盾
	if msg.GetDamagepercent() >= 90 {
		op.AddShield(16)
	} else if msg.GetDamagepercent() >= 40 {
		op.AddShield(12)
	}
	//自我演习次数加一
	op.SetDrillTimes(op.GetDrillTimes() + 1)

	def_log := &self.fightinfo.def_log
	def_log.SetGoldstolen(msg.GetGoldstolen())
	def_log.SetFoodstolen(msg.GetFoodstolen())
	def_log.SetGaintrophy(-msg.GetTrophy())
	def_log.SetDmgpercent(msg.GetDamagepercent())
	def_log.SetStarts(msg.GetStarts())
	def_log.SetWuhun(msg.GetWuhun())

	//如果已经有bid了，那么说明是复仇
	if def_log.GetBid() > 0 {
		def_log.SetRevstate(rpc.BattleLog_Revenged)

		self.SetBattleLogRevenged(def_log.GetBid())
	}

	//log.SetRid()
	op.InsertBattleLog(def_log, &self.fightinfo.replay)

	claninfo := op.GetClanInfo()
	atc_log := self.fightinfo.def_log
	atc_log.SetPid(op.GetUid())
	atc_log.SetName(op.GetName())
	atc_log.SetLevel(op.GetLevel())
	atc_log.SetClanName(claninfo.GetName())
	atc_log.SetClanSymbol(claninfo.GetSymbol())
	atc_log.SetTrophy(op.GetTrophy())
	atc_log.SetGaintrophy(msg.GetTrophy())
	self.InsertAttackLog(&atc_log, &self.fightinfo.replay)

	//数量为0的联盟兵
	if msg.V.Alliancecastle != nil {
		for _, castle := range msg.V.Alliancecastle {
			if castle.Characters != nil {
				bCycle := true
				for bCycle {
					bCycle = false
					for index, character := range castle.Characters {
						if character.GetCount() <= 0 {
							castle.Characters = append(castle.Characters[:index], castle.Characters[index+1:]...)
							bCycle = true
							break
						}
					}
				}
			}
		}
	}
	op.v.VillageInfo = msg.V

	self.ClearFight()

	return
}

//PVE战斗结果处理
func (self *player) OnPlayerPveFightResult(msg *rpc.PVEAttackEnd) {
	if self.pve == nil {
		self.pve = &rpc.PveStages{}
	}

	StageId := *msg.Stage.StageId

	//新手全得
	for _, stage := range self.pve.Stages {
		if stage.GetStageId() == StageId {
			if stage.GetStars() == 100 {
				cfg := GetPVEStageCfg(stage.GetStageId())
				msg.SetGoldstolen(cfg.GoldStorage)
				msg.SetFoodstolen(cfg.FoodStorage)
				msg.Stage.SetStars(3)
				msg.Stage.SetCurrentFood(0)
				msg.Stage.SetCurrentGold(0)
			}
			break
		}
	}

	//处理完新手后再赋值
	StageStar := *msg.Stage.Stars
	food := *msg.Stage.CurrentFood
	gold := *msg.Stage.CurrentGold
	logger.Info("PVE(%d)战利品：%d, %d.关卡剩余：%d, %d,星级：%d", StageId, msg.GetGoldstolen(), msg.GetFoodstolen(), food, gold, StageStar)

	//如果此关卡已经攻打过，就更新关卡数据
	NewPveStage := true
	for index, stage := range self.pve.Stages {
		if stage.GetStageId() == StageId {

			stage := self.pve.Stages[index]
			//100表示新手
			if stage.GetStars() == 100 || stage.GetStars() < StageStar {
				stage.SetStars(StageStar)
			}
			stage.SetCurrentFood(food)
			stage.SetCurrentGold(gold)
			NewPveStage = false
			break
		}
	}

	//如果是第一次攻打的关卡，就创建之
	if NewPveStage {
		NewStage := &rpc.Stage{}
		NewStage.SetStageId(StageId)
		NewStage.SetStars(StageStar)
		NewStage.SetCurrentGold(gold)
		NewStage.SetCurrentFood(food)

		self.pve.Stages = append(self.pve.Stages, NewStage)
	}

	// 加粮食，加金
	self.GainResource(msg.GetGoldstolen(), proto.ResType_Gold, proto.Gain_Pve)
	self.GainResource(msg.GetFoodstolen(), proto.ResType_Food, proto.Gain_Pve)

	//logger.Info("OnPlayerPveFightResult11, %v", self.pve)
	//logger.Info("OnPlayerPveFightResult Gold:%d, Food:%d", *msg.Goldstolen, *msg.Foodstolen)

	//limitf, totalf := self.v.collect_GetStorageFoodLimit()
	//logger.Info("collect_GetStorageFoodLimit, %d/%d", limitf, totalf)
	//limitg, totalg := self.v.collect_GetStorageGoldLimit()
	//logger.Info("collect_GetStorageFoodLimit, %d/%d", limitg, totalg)

	self.ClearFight()
}

func (p *player) TryDrill() {
	ret := &rpc.TryDrillResult{}
	defer WriteResult(p.conn, ret)

	if p.GetDrillTimes() == 0 {
		ret.SetOk(false)

		return
	}

	p.SetDrillTimes(p.GetDrillTimes() - 1)

	ret.SetOk(true)

	update := &rpc.UpdatePlayerInfo{}
	update.SetDrillTimes(p.GetDrillTimes())
	WriteResult(p.conn, update)
}

func (p *player) TryFriendDrill() {
	if p.GetFriendDrillTimes() == 0 {
		return
	}

	p.SetFriendDrillTimes(p.GetFriendDrillTimes() - 1)

	update := &rpc.UpdatePlayerInfo{}
	update.SetFriendDrillTimes(p.GetFriendDrillTimes())
	WriteResult(p.conn, update)
}
func (p *player) onDayTick() {

	p.refreshTiLiTick.Stop()
	p.refreshTiLiTick = nil
	//.need fix
	tiLi_Max := GetGlobalCfg("TTT_TILI_MAX")
	refresh_Time := GetGlobalCfg("TTT_TILI_REFRESH")
	if p.GetTili() < tiLi_Max {
		preTiLiCount := p.GetTili()
		if 1+p.GetTili() >= tiLi_Max {
			p.SetTili(tiLi_Max)
		} else {
			p.SetTili(p.GetTili() + 1)
		}
		if preTiLiCount != p.GetTili() {
			update := &rpc.UpdatePlayerInfo{}
			update.SetTili(p.GetTili())
			WriteResult(p.conn, update)
		}
	}
	p.refreshTiLiTick = timer.NewTimer(time.Duration(refresh_Time) * time.Second)
	p.refreshTiLiTick.Start(
		func() {
			p.onDayTick()
		},
	)
}

func (p *player) OnInit(conn rpc.RpcConn) {
	p.conn = conn
	p.OnTick()

	p.t = timer.NewTimer(time.Second)
	p.t.Start(
		func() {

			defer func() {
				p.conn.Unlock()
				if r := recover(); r != nil {
					fmt.Println("player tick runtime error begin: ", r)
					debug.PrintStack()

					cns.onDisConn(conn)
					conn.Close()
					fmt.Println("player tick runtime error end: ", r)
				}
			}()

			p.conn.Lock()
			p.OnTick()
		},
	)

	fmt.Println("登陆进入 ")
	t := time.Now()
	allsec := t.Unix()
	var daysec int64
	daysec = 24 * 3600
	curDay := allsec / daysec

	//计算添加的体力数
	//.need fix
	tiLi_Max := GetGlobalCfg("TTT_TILI_MAX")
	refresh_Time := GetGlobalCfg("TTT_TILI_REFRESH")
	if p.GetOnQuitTime() != 0 {
		seprateTime := allsec - p.GetOnQuitTime()
		addTiLiCount := (uint32(seprateTime) / refresh_Time)
		if p.GetTili() < tiLi_Max {
			if addTiLiCount+p.GetTili() >= tiLi_Max {
				p.SetTili(tiLi_Max)
			} else {
				p.SetTili(p.GetTili() + addTiLiCount)
			}
		}
	}
	//开始时间添加体力刷新tick
	p.refreshTiLiTick = timer.NewTimer(time.Duration(refresh_Time) * time.Second)
	p.refreshTiLiTick.Start(
		func() {
			p.onDayTick()
		},
	)
	//每日分享奖励
	shareinfo := p.GetShareinfo()
	if shareinfo == nil {
		shareinfo = &rpc.ShareInfo{}
		shareinfo.SetStep(0)
		shareinfo.SetSharetime(allsec)
	}
	curStepShare := shareinfo.GetStep()
	if !common.IsTheSameDay(uint32(time.Now().Unix()), uint32(shareinfo.GetSharetime())) {
		curStepShare = 0
	}
	shareinfo.SetStep(curStepShare)
	shareinfo.SetSharetime(allsec)
	p.SetShareinfo(shareinfo)

	//连续登陆奖励
	landinfo := p.GetLandedrewardinfo()
	if landinfo == nil {
		landinfo = &rpc.LandedRewardInfo{}
		landinfo.SetLandCount(1)
		landinfo.SetLastLandedTime(allsec)
		landinfo.SetNeedGetReward(true)
	}
	lastsec := landinfo.GetLastLandedTime()
	lastDay := lastsec / daysec
	curCount := landinfo.GetLandCount()
	if curDay == lastDay+1 {
		if curCount < 7 {
			curCount = curCount + 1
		}
	} else {
		if curDay-lastDay > 1 {
			curCount = 1
		}
	}
	needGetReward := landinfo.GetNeedGetReward()
	if curDay != lastDay {
		needGetReward = true
	}

	pMultipleInfo := p.GetSelftttmultiples()
	//检测是否刷新每日的ttt次数
	uTimeCur := uint32(time.Now().Unix())
	if pMultipleInfo == nil || !common.IsTheSameDay(uTimeCur, pMultipleInfo.GetRandomtime()) {
		pMultipleInfo = &rpc.PlayerTTTMutliples{}
		pMultipleInfo.SetMultiple(1)
		pMultipleInfo.SetRandomtime(uTimeCur)
		pMultipleInfo.SetFreetimes(0)
		pMultipleInfo.SetCurRandtimes(0)
	}
	p.SetSelftttmultiples(pMultipleInfo)
	//这里要存储数据
	landinfo.SetLandCount(curCount)
	landinfo.SetLastLandedTime(allsec)
	landinfo.SetNeedGetReward(needGetReward)
	p.SetLandedrewardinfo(landinfo)

	//刷新好友演习次数
	if p.GetFriendDrillTimes() < 10 {
		if lastDay-curDay > 0 {
			p.SetFriendDrillTimes(p.GetFriendDrillTimes() + uint32(lastDay-curDay))
			if p.GetFriendDrillTimes() > 10 {
				p.SetFriendDrillTimes(10)
			}
		}
	}

	pTTTRankPlayersInfo := p.GetTttrankplayerinfo()
	if pTTTRankPlayersInfo == nil {
		pTTTRankPlayersInfo = &rpc.TTTRankPlayersInfo{}
		pTTTRankPlayersInfo.SetCurTTTScore(0)
		pTTTRankPlayersInfo.SetCurMostTTTScore(0)
		pTTTRankPlayersInfo.SetLastFreshTime(uTimeCur)
	}

	if pTTTRankPlayersInfo != nil && !common.IsTheSameWeek(uTimeCur, pTTTRankPlayersInfo.GetLastFreshTime()) {
		pTTTRankPlayersInfo.SetCurMostTTTScore(0)
		pTTTRankPlayersInfo.SetLastFreshTime(uTimeCur)
	} else {
		p.SetTttrankplayerinfo(pTTTRankPlayersInfo)
	}

	//初始化通天塔信息
	pNewTTTInfo := p.GetSelftttinfo()
	if pNewTTTInfo == nil {
		pNewTTTInfo := &rpc.PlayerTTTInfo{}
		pNewTTTInfo.SetFightbegintime(uint32(allsec))
		pNewTTTInfo.SetFightdaytimes(1)
		pNewTTTInfo.SetCurcheckpoint(0)
		pNewTTTInfo.SetChangetimes(0)
		pNewTTTInfo.SetReLiveCount(0)
		pNewTTTInfo.SetIsTTTOver(true)
		pNewTTTInfo.SetIsStart(false)
		pNewTTTInfo.SetIsTTTEndEveryCheckpoint(true)

		pNewTTTInfo.Startcharacters = nil
		pNewTTTInfo.Startspells = nil
		pNewTTTInfo.Starttttbuffs = nil

		//save info
		p.SetSelftttinfo(pNewTTTInfo)
	}
	//.need  每日刷新TTT数据，改为每周刷新数据
	if pNewTTTInfo != nil {
		if !common.IsTheSameDay(uint32(allsec), pNewTTTInfo.GetFightbegintime()) {
			//不是同一天就从头开始
			pNewTTTInfo.SetFightbegintime(uint32(allsec))
			pNewTTTInfo.SetFightdaytimes(1)
			pNewTTTInfo.SetCurcheckpoint(0)
			pNewTTTInfo.SetChangetimes(0)
			pNewTTTInfo.SetReLiveCount(0)
			//这两个是判断ttt流程是否完成
			pNewTTTInfo.SetIsTTTOver(true)
			pNewTTTInfo.SetIsStart(false)

			//这一个是判断当前的关卡是否顺利结束
			pNewTTTInfo.SetIsTTTEndEveryCheckpoint(true)

			//初始化首次进入的数据
			pNewTTTInfo.Startcharacters = nil
			pNewTTTInfo.Startspells = nil
			pNewTTTInfo.Starttttbuffs = nil
		}
	}
}

func (p *player) Unlock() (err error) {
	ts("player Unlock", p.GetUid())
	defer te("player Unlock", p.GetUid())

	_, err = lockclient.TryUnlock("player", p.GetUid(), p.lid)

	return
}

//好友全从腾讯取
/*func (p *player) getFriends() (f *rpc.FriendsIdList) {
	f = &rpc.FriendsIdList{}
	for fid, _ := range p.friends {
		f.Friends = append(f.Friends, fid)
	}

	return f
}

func (p *player) CanAddFriend() bool {
	fmax := GetGlobalCfg("MAX_FRIEND_NUM")

	//如果没有配置或配成0，则默认为100
	if fmax == 0 {
		fmax = 100
	}

	//当前好友数
	fcur := len(p.friends)

	//已超过上限
	if uint32(fcur) >= uint32(fmax) {
		return false
	}

	return true
}*/

func (p *player) getGooglePayNonce() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	nonce := strconv.FormatUint(uint64(r.Uint32()), 10)
	_, ok := p.googlePayNonces[nonce]
	if ok {
		nonce = p.AddGooglePayNonce()
	}
	return nonce
}

func (p *player) AddGooglePayNonce() string {
	nonce := p.getGooglePayNonce()
	p.googlePayNonces[nonce] = true
	ts("增加googlePay随机数", nonce)
	return nonce
}

func (p *player) RemoveGooglePayNonce(nonce string) {
	delete(p.googlePayNonces, nonce)
	ts("删除googlePay随机数", nonce)
}

func (p *player) FindGooglePayNonce(nonce string) bool {
	ts("检查googlePay随机数", nonce)
	_, ok := p.googlePayNonces[nonce]
	ts("检查googlePay随机数 结果：", ok)
	return ok
}

func (p *player) Save() (err error) {
	ts("player Save", p.GetUid())
	defer te("player Save", p.GetUid())

	//保存玩家PVE数据，如果有的话，pve放在基础信息之前保存
	if p.pve != nil {
		//新手的pve的关卡全得功能
		for index, stage := range p.pve.Stages {
			if stage.GetStars() == 100 {
				cfg := GetPVEStageCfg(stage.GetStageId())
				p.GainResource(cfg.GoldStorage, proto.ResType_Gold, proto.Gain_Pve)
				p.GainResource(cfg.FoodStorage, proto.ResType_Food, proto.Gain_Pve)

				stage := p.pve.Stages[index]
				stage.SetStars(3)
				stage.SetCurrentGold(0)
				stage.SetCurrentFood(0)
				stage.SetCurrentDiamond(0)
			}
		}
		_, err = KVWriteExt("pve", p.GetUid(), p.pve)
	}

	//保存玩家记本数据
	_, err = KVWriteBase(common.PlayerBase, p.GetUid(), p.PlayerBaseInfo)
	_, err = KVWriteExt(common.PlayerExtra, p.GetUid(), p.PlayerExtraInfo)

	//只有是自己读取上线下线保存才会存取玩家数据，其它情况不管
	/*if p.savefriends {
		f := p.getFriends()
		_, err = KVWriteExt("friends", p.GetUid(), f)
	}*/

	return
}

func (p *player) OnTick() {

	if p.v != nil {
		p.v.OnTick()
	}
}

func (p *player) OnQuit() {
	ts("player:OnQuit", p.GetUid())
	defer te("player:OnQuit", p.GetUid())

	cns.EndChat(p) //聊天放在这里是因为不用关心客户端连接是否健康

	fmt.Println("退出 ")
	if p.refreshTiLiTick != nil {
		p.refreshTiLiTick.Stop()
		p.refreshTiLiTick = nil
		p.SetOnQuitTime(time.Now().Unix())
	}

	logger.Info("OnQuit end chat end")

	if p.conn != nil {
		p.conn.Lock()
		defer p.conn.Unlock()
	}

	logger.Info("OnQuit p.conn.Lock() end")

	if p.t != nil {
		p.t.Stop()
	}

	p.Save()

	if p.v != nil {
		p.v.OnQuit()
	}

	p.Unlock()

	SetPlayerTrophy(p.GetTrophy(), int64(p.GetGamelocation()), p.GetUid())

	NotifyOffline(p.GetUid(), p.GetIsUserguideFinish())

	//纪录登出log
	msg := proto.LogPlayerLoginLogout{
		ChannelId: uint8(p.GetGamelocation()),
		Playerid:  p.GetUid(),
		Time:      time.Now().Unix(),
		Logout:    true,
		Ip:        "",
	}

	var ret proto.LogPlayerLoginLogoutResult

	cns.logRpcConn.Go("LogServices.LogPlayerLoginLogoutGame", msg, &ret, nil)
}

func (p *player) ReturnHome() {
	ts("player ReturnHome", p.GetUid())
	defer te("player ReturnHome", p.GetUid())

	//add for challenge
	if p.fightinfo != nil && p.fightinfo.bIsChallenge {
		logger.Info("进入了擂台的返回删除")
		req := &proto.PlayerReturnHome{Uid: p.GetUid()}
		ret := &proto.PlayerReturnHomeResult{}
		cns.center.Go("Center.PlayerReturnHome", req, ret, nil)
	}

	v := p.GetVillage()
	if v == nil {
		WriteLoginResult(p.conn, rpc.LoginResult_SERVERERROR)
		logger.Fatal("ReturnHome Error:player's village not found!")

		return
	}

	v.buildings_ProcessUpgrade()
	v.ResetBuildingHp()
	WriteResult(p.conn, v.VillageInfo)

	return
}

func (p *player) getCenterLevel() uint32 {
	if p.v != nil {
		return p.v.getCenterLevel()
	}

	return 0
}

func (p *player) GetVillage() *village {
	ts("player GetVillage", p.GetUid())
	defer te("player GetVillage", p.GetUid())

	if p.v == nil {
		p.v = LoadVillage(p.GetVillageId(), !p.GetIsUserguideFinish())
		if p.v != nil {
			p.VillageId = &p.v.vid
			p.v.p = p
			p.v.castle_GetDonate(nil)
			p.v.buildings_ProcessUpgrade()
			p.v.obstacle_ProcessGenerate()
			p.v.laboratory_ProcessUpgrade()
			p.v.barrack_Process()
			p.v.spellForge_Process()
			p.v.hero_ProcessUpgrade()
			//更新建筑新数据
			//p.v.UpdetaBuildingNewData()
		}
	}
	//数据库中都加载失败了
	if p.v == nil {
		logger.Error("load village failed!")
		return nil
	}

	p.v.castle_UpdateClanInfo(p.GetClan(), p.GetClanSymbol())

	//向center发送消息
	req := &proto.UpdaePlayerLevel2Id{Id: p.GetUid(), Level: p.v.getCenterLevel()}
	rst := &proto.UpdaePlayerLevel2IdResult{}
	cns.center.Go("Center.UpdatePlayerLevel2Id", req, rst, nil)

	return p.v
}

func (p *player) IsShopItemInCD(sit rpc.ShopItemType) bool {
	for _, info := range p.Shop {
		if info.GetType() == sit {
			start_time := info.GetCdTime()
			if start_time == 0 {
				return false
			}

			total_Time := uint32(0)
			if sit == rpc.ShopItemType_ShopItem_OneDayShield {
				total_Time = GetGlobalCfg("SHIELD_1DAY_BUY_CD") * 24 * 3600
			} else if sit == rpc.ShopItemType_ShopItem_TwoDayShield {
				total_Time = GetGlobalCfg("SHIELD_2DAY_BUY_CD") * 24 * 3600
			} else if sit == rpc.ShopItemType_ShopItem_OneWeekShield {
				total_Time = GetGlobalCfg("SHIELD_7DAY_BUY_CD") * 24 * 3600
			}

			if total_Time == 0 {
				return false
			}

			return (start_time + total_Time) > uint32(time.Now().Unix())
		}
	}

	return false
}

func (p *player) SetShopItemCD(itemType rpc.ShopItemType) *rpc.ShopInfo {
	for _, info := range p.Shop {
		if info.GetType() == itemType {
			info.SetCdTime(uint32(time.Now().Unix()))

			return nil
		}
	}

	info := &rpc.ShopInfo{}
	info.SetType(itemType)
	info.SetCdTime(uint32(time.Now().Unix()))

	p.Shop = append(p.Shop, info)

	return info
}

func (p *player) BuyResource(buy rpc.TryBuy, ResFrom uint32) (update *rpc.UpdatePlayerInfo) {
	ts("player BuyResource", p.GetUid(), buy.GetType(), buy.GetNum())
	defer te("player BuyResource", p.GetUid())

	//检查是否在cd中
	if p.IsShopItemInCD(buy.GetType()) {
		return nil
	}

	update = &rpc.UpdatePlayerInfo{}
	logger.Info("this is done 1")
	if buy.GetType() == rpc.ShopItemType_ShopItem_Gem {
		logger.Info("this is done 2")
		p.GainResource(buy.GetNum(), proto.ResType_Gem, ResFrom)

		update.SetDiamonds(p.GetPlayerTotalGem())
	} else if buy.GetType() == rpc.ShopItemType_ShopItem_Food {
		cost := GetCostGem(buy.GetNum())
		if cost > p.GetPlayerTotalGem() {
			logger.Info("Not enough gem, need:", cost)
			return nil
		}

		//先确定扣除成功再做后面的操作
		if !p.CostResource(cost, proto.ResType_Gem, proto.Lose_BuyRes) {
			return nil
		}

		left := p.v.collect_StorageFood(buy.GetNum())
		update.SetFood(buy.GetNum() - left)

		update.SetDiamonds(p.GetPlayerTotalGem())
	} else if buy.GetType() == rpc.ShopItemType_ShopItem_Gold {
		cost := GetCostGem(buy.GetNum())
		if cost > p.GetPlayerTotalGem() {
			logger.Info("Not enough gem, need:", cost)
			return nil
		}

		//先确定扣除成功再做后面的操作
		if !p.CostResource(cost, proto.ResType_Gem, proto.Lose_BuyRes) {
			return nil
		}

		left := p.v.collect_StorageGold(buy.GetNum())
		update.SetGold(buy.GetNum() - left)

		update.SetDiamonds(p.GetPlayerTotalGem())
	} else if buy.GetType() == rpc.ShopItemType_ShopItem_Wuhun {
		cost := GetWuhunCostGem(buy.GetNum())
		fmt.Println("--buy resources  wu hun buy.GetNum() and cost--", buy.GetNum(), cost)
		if cost > p.GetPlayerTotalGem() {
			logger.Info("Not enough gem, need:", cost)
			return nil
		}

		//先确定扣除成功再做后面的操作
		if !p.CostResource(cost, proto.ResType_Gem, proto.Lose_BuyRes) {
			return nil
		}

		p.SetWuhun(p.GetWuhun() + buy.GetNum())

		update.SetDiamonds(p.GetPlayerTotalGem())
	} else if buy.GetType() == rpc.ShopItemType_ShopItem_OneDayShield {
		cost := GetGlobalCfg("SHIELD_PRICE_1DAY")
		if cost > p.GetPlayerTotalGem() {
			logger.Info("Not enough gem, need:", cost)
			return nil
		}

		//先确定扣除成功再做后面的操作
		if !p.CostResource(cost, proto.ResType_Gem, proto.Lose_BuyShield) {
			return nil
		}

		if !p.AddShield(1 * 24) {
			return nil
		}

		shopinfo := p.SetShopItemCD(rpc.ShopItemType_ShopItem_OneDayShield)

		update.SetShop(shopinfo)
		update.SetDiamonds(p.GetPlayerTotalGem())
	} else if buy.GetType() == rpc.ShopItemType_ShopItem_TwoDayShield {
		cost := GetGlobalCfg("SHIELD_PRICE_2DAY")
		if cost > p.GetPlayerTotalGem() {
			logger.Info("Not enough gem, need:", cost)
			return nil
		}

		//先确定扣除成功再做后面的操作
		if !p.CostResource(cost, proto.ResType_Gem, proto.Lose_BuyShield) {
			return nil
		}

		if !p.AddShield(2 * 24) {
			return nil
		}

		shopinfo := p.SetShopItemCD(rpc.ShopItemType_ShopItem_TwoDayShield)

		update.SetShop(shopinfo)
		update.SetDiamonds(p.GetPlayerTotalGem())
	} else if buy.GetType() == rpc.ShopItemType_ShopItem_OneWeekShield {
		cost := GetGlobalCfg("SHIELD_PRICE_7DAY")
		if cost > p.GetPlayerTotalGem() {
			logger.Info("Not enough gem, need:", cost)
			return nil
		}

		//先确定扣除成功再做后面的操作
		if !p.CostResource(cost, proto.ResType_Gem, proto.Lose_BuyShield) {
			return nil
		}

		if !p.AddShield(7 * 24) {
			return nil
		}

		shopinfo := p.SetShopItemCD(rpc.ShopItemType_ShopItem_OneWeekShield)

		update.SetShop(shopinfo)
		update.SetDiamonds(p.GetPlayerTotalGem())
	} else if buy.GetType() == rpc.ShopItemType_ShopItem_Drill1 {
		cost := GetGlobalCfg("YANXI_BUY_1_COST")
		if cost > p.GetPlayerTotalGem() {
			logger.Info("Not enough gem, need:", cost)
			return nil
		}

		//先确定扣除成功再做后面的操作
		if !p.CostResource(cost, proto.ResType_Gem, proto.Lose_BuyDrill) {
			return nil
		}

		p.SetDrillTimes(p.GetDrillTimes() + 1)

		update.SetDrillTimes(p.GetDrillTimes())
	} else if buy.GetType() == rpc.ShopItemType_ShopItem_Drill12 {
		cost := GetGlobalCfg("YANXI_BUY_12_COST")
		if cost > p.GetPlayerTotalGem() {
			logger.Info("Not enough gem, need:", cost)
			return nil
		}

		//先确定扣除成功再做后面的操作
		if !p.CostResource(cost, proto.ResType_Gem, proto.Lose_BuyDrill) {
			return nil
		}

		p.SetDrillTimes(p.GetDrillTimes() + 12)

		update.SetDrillTimes(p.GetDrillTimes())
	} else if buy.GetType() == rpc.ShopItemType_ShopItem_Drill150 {
		cost := GetGlobalCfg("YANXI_BUY_150_COST")
		if cost > p.GetPlayerTotalGem() {
			logger.Info("Not enough gem, need:", cost)
			return nil
		}

		//先确定扣除成功再做后面的操作
		if !p.CostResource(cost, proto.ResType_Gem, proto.Lose_BuyDrill) {
			return nil
		}

		p.SetDrillTimes(p.GetDrillTimes() + 150)

		update.SetDrillTimes(p.GetDrillTimes())
	} else if buy.GetType() == rpc.ShopItemType_ShopItem_FriendDrill1 {
		cost := GetGlobalCfg("YANXI_BUY_1_COST")
		if cost > p.GetPlayerTotalGem() {
			logger.Info("Not enough gem, need:", cost)
			return nil
		}

		//先确定扣除成功再做后面的操作
		if !p.CostResource(cost, proto.ResType_Gem, proto.Lose_BuyFriendDrill) {
			return nil
		}

		p.SetFriendDrillTimes(p.GetFriendDrillTimes() + 1)

		update.SetFriendDrillTimes(p.GetFriendDrillTimes())
	} else if buy.GetType() == rpc.ShopItemType_ShopItem_FriendDrill12 {
		cost := GetGlobalCfg("YANXI_BUY_12_COST")
		if cost > p.GetPlayerTotalGem() {
			logger.Info("Not enough gem, need:", cost)
			return nil
		}

		//先确定扣除成功再做后面的操作
		if !p.CostResource(cost, proto.ResType_Gem, proto.Lose_BuyFriendDrill) {
			return nil
		}

		p.SetFriendDrillTimes(p.GetFriendDrillTimes() + 12)

		update.SetFriendDrillTimes(p.GetFriendDrillTimes())
	} else if buy.GetType() == rpc.ShopItemType_ShopItem_FriendDrill150 {
		cost := GetGlobalCfg("YANXI_BUY_150_COST")
		if cost > p.GetPlayerTotalGem() {
			logger.Info("Not enough gem, need:", cost)
			return nil
		}

		//先确定扣除成功再做后面的操作
		if !p.CostResource(cost, proto.ResType_Gem, proto.Lose_BuyFriendDrill) {
			return nil
		}

		p.SetFriendDrillTimes(p.GetFriendDrillTimes() + 150)

		update.SetFriendDrillTimes(p.GetFriendDrillTimes())
	} else if buy.GetType() == rpc.ShopItemType_ShopItem_BattleAccelerate {
		cost := GetGlobalCfg("BATLLE_ACCELERATE_BUY_10_COST")
		if cost > p.GetPlayerTotalGem() {
			logger.Info("Not enough gem, need:", cost)
			return nil
		}

		//先确定扣除成功再做后面的操作
		if !p.CostResource(cost, proto.ResType_Gem, proto.Lose_BuyBattleAcc) {
			return nil
		}

		p.SetBattleAccelerateTimes(p.GetBattleAccelerateTimes() + 10)

		update.SetBattleAccelerateTimes(p.GetBattleAccelerateTimes())

	} else if buy.GetType() == rpc.ShopItemType_ShopItem_TiLi {
		//.need fix
		cost := GetGlobalCfg("BATLLE_ACCELERATE_BUY_10_COST")
		fmt.Println("--buy resources  wu hun buy.GetNum() and cost--", buy.GetNum(), cost)
		if cost > p.GetPlayerTotalGem() {
			logger.Info("Not enough gem, need:", cost)
			return nil
		}

		//先确定扣除成功再做后面的操作
		if !p.CostResource(cost, proto.ResType_Gem, proto.Lose_BuyTiLi) {
			return nil
		}

		p.GainResource(buy.GetNum(), proto.ResType_TiLi, proto.Gain_AddTiLI)
		update.SetTili(p.GetTili())

	}

	return nil
}

func (p *player) GetClanInfo() *rpc.ClanInfo {
	ts("player GetClanInfo", p.GetUid(), p.GetClan())
	defer te("player GetClanInfo", p.GetUid(), p.GetClan())

	c := p.GetClan()
	if c != "" {
		value, claninfo := GetClanInfo(c)
		if value != proto.GetClanOk {
			return nil
		}
		return claninfo
	}

	info := &rpc.ClanInfo{}
	info.SetType(rpc.ClanInfo_Any)
	info.SetName("")
	info.SetSymbol(0)
	info.SetRequire(0)
	info.SetDescribe("")

	return info
}

func (p *player) GetClanSymbol() uint32 {
	c := p.GetClan()

	if c == "" {
		return 0
	}

	value, claninfo := GetClanInfo(c)
	if value != proto.GetClanOk || claninfo == nil {
		return 0
	}
	return claninfo.GetSymbol()
}

func (p *player) GetClanPlayer() (int, *rpc.Player) {
	return GetClanPlayer(p.GetClan(), p.GetUid())
}

func (p *player) GetClanPlayerPower() rpc.Player_ClanPower {
	code, cp := GetClanPlayer(p.GetClan(), p.GetUid())
	if cp != nil {
		logger.Info("p.GetClanPlayerPower():%d, %v", code, *cp)
		return cp.GetPower()
	}

	return rpc.Player_None
}

func (p *player) CreateClanPlayer(power rpc.Player_ClanPower) *rpc.Player {
	ts("player CreateClanPlayer", p.GetUid(), p.GetClan())
	defer te("player CreateClanPlayer", p.GetUid(), p.GetClan())

	cp := &rpc.Player{}
	cp.SetType(rpc.Player_Clan)
	cp.SetName(p.GetName())
	cp.SetUid(p.GetUid())
	cp.SetLevel(p.GetLevel())
	cp.SetExp(p.GetExp())
	cp.SetTrophy(p.GetTrophy())
	cp.SetPower(power)

	return cp
}

func (p *player) SetPlayerLevel(level uint32) bool {
	if level == 0 || level > PLAYERLEVEL_MAX || level == p.GetLevel() {
		return false
	}

	p.SetLevel(level)

	//同步到客户端
	if p.conn != nil {
		update := &rpc.UpdatePlayerInfo{}
		update.SetLevel(p.GetLevel())
		WriteResult(p.conn, update)
	}
	return false
}

func (p *player) AddExp(exp uint32) {
	l := p.GetLevel() //当前等级

	if exp == 0 || l > PLAYERLEVEL_MAX {
		return
	}

	e := p.GetExp() //当前经验值

	if l == PLAYERLEVEL_MAX && e+1 >= GetExpPoints(l) {
		return
	}

	for {
		max := GetExpPoints(l) //当前经验上限值

		if e+exp < max { //如果经验不能超过上限则直接break
			p.SetExp(e + exp)
			break
		}

		if l == PLAYERLEVEL_MAX { //如果等级已为最大值则
			p.SetExp(max - 1)
			break
		}

		//升级
		exp -= (max - e) //计算升级后剩余经验值
		e = 0
		l += 1 //等级＋1
	}

	if l != p.GetLevel() {
		p.SetPlayerLevel(l)
	}

	//同步到客户端
	if p.conn != nil {
		update := &rpc.UpdatePlayerInfo{}
		update.SetExp(p.GetExp())
		WriteResult(p.conn, update)
	}
}

func (p *player) AddTrophy(trophy int32) {
	finalTrophy := int32(p.GetTrophy()) + trophy
	if finalTrophy < 0 {
		finalTrophy = 0
	}

	logger.Info("AddTrophy:<%d> Before:<%d>, Now:<%d>", trophy, p.GetTrophy(), uint32(finalTrophy))

	p.SetTrophy(uint32(finalTrophy))

	if p.GetTrophy() > 20000 {
		logger.Error("FFFFFFFFFFFFFuck!!!!!!!!!!!!!!<%d, %d>", finalTrophy, p.GetTrophy())
	}

	//同步到客户端
	if p.conn != nil {
		update := &rpc.UpdatePlayerInfo{}
		update.SetTrophy(p.GetTrophy())
		WriteResult(p.conn, update)
	}
}

func (p *player) GainResource(num uint32, ResType string, ResFrom uint32) (Res uint32, pass bool) {
	if num == 0 {
		return 0, false
	}
	ts("GainResource Begain%d", p.GetPlayerTotalGem())
	var ResNumber uint32 = 0
	switch ResType {
	case proto.ResType_Gold:
		ResNumber = p.v.collect_StorageGold(num)
	case proto.ResType_Food:
		ResNumber = p.v.collect_StorageFood(num)
	case proto.ResType_Wuhun:
		p.SetWuhun(p.GetWuhun() + num)
	case proto.ResType_Gem:
		p.SetDiamonds(p.GetDiamonds() + num)
	case proto.ResType_Trophy:
		p.AddTrophy(int32(num))
	case proto.ResType_TiLi:
		theExtraTiLi := GetGlobalCfg("TTT_TILI_PLUS_MAX")
		if p.GetTili()+num >= theExtraTiLi {
			p.SetTili(theExtraTiLi)
		} else {
			p.SetTili(p.GetTili() + num)
		}
	default:
		return 0, false
	}
	ts("GainResource END%d", p.GetPlayerTotalGem())
	return ResNumber, LOG_Resources(p.GetGamelocation(), p.GetUid(), true, ResType, num, ResFrom)
}

//调用此接口扣除宝石必须判断返回值，这里应该加锁
func (p *player) CostResource(num uint32, ResType string, ResFrom uint32) bool {
	p.paylock.Lock()
	defer p.paylock.Unlock()

	if num == 0 {
		return true
	}

	switch ResType {
	case proto.ResType_Gold:
		p.v.collect_CostGold(num)
	case proto.ResType_Food:
		p.v.collect_CostFood(num)
	case proto.ResType_Wuhun:
		if p.GetWuhun() < num {
			p.SetWuhun(0)
		} else {
			p.SetWuhun(p.GetWuhun() - num)
		}
	case proto.ResType_Gem:
		//先扣除腾讯的，再扣除自己的游戏币
		tencentNum := uint32(0)
		selfNum := num
		if p.GetDiamonds() < num {
			tencentNum = num - p.GetDiamonds()
			selfNum = num - tencentNum
		}

		//腾讯支付流程
		if p.mobileqqinfo != nil && tencentNum > 0 {
			success, errmsg, _, balance := MobileQQPay(p, int(tencentNum))
			if !success {
				WriteLoginResultWithErrorMsg(p.conn, rpc.LoginResult_TX_AUTH_FAILED, errmsg)

				return false
			}

			p.mobileqqinfo.Balance = uint32(balance)
		}

		p.SetDiamonds(p.GetDiamonds() - selfNum)

		//同步宝石
		p.SyncPlayerGem()
	case proto.ResType_Trophy:
		p.AddTrophy(-int32(num))
	case proto.ResType_TiLi:
		if p.GetTili() < num {
			p.SetTili(0)
		} else {
			p.SetTili(p.GetTili() - num)
		}
	default:
		return false
	}

	//log记录失败不管
	LOG_Resources(p.GetGamelocation(), p.GetUid(), false, ResType, num, ResFrom)

	return true
}

func (p *player) SellBuilding(sell rpc.SellBuilding) {

	//fmt.Println("SellBuilding 1")
	Type := *sell.Id.Type
	Index := *sell.Id.Index

	bCanSell := false
	if Type == rpc.BuildingId_Bomb || Type == rpc.BuildingId_GiantBomb || Type == rpc.BuildingId_Eject {
		bCanSell = true
	}

	if Type >= rpc.BuildingId_Deco1 && Type <= rpc.BuildingId_Deco24 {
		bCanSell = true
	}

	if !bCanSell {
		return
	}

	obj := p.v.buildings_Get(Type, Index)

	if obj == nil {
		return
	}

	cfg := GetBuildingCfgByTypeId(Type, 1)

	if cfg != nil {
		p.GainResource(cfg.SellPrice, strings.ToLower(cfg.BuildResource), proto.Gain_SellBuilding)

		if m, ok := obj.(Movable); ok {
			p.v.mapRemoveFrom(m.GetP().GetX(), m.GetP().GetY(), cfg.BuildSize)
		}

		p.v.buildings_Remove(Type, Index)
	}
}

func (p *player) HasShield() bool {
	s := p.GetShield()
	if s == nil {
		return false
	}

	if s.GetStartTime()+s.GetTotalTime() <= uint32(time.Now().Unix()) {
		p.RemoveShield()

		return false
	}

	return true
}

// 传入参数小时为单位
func (p *player) AddShield(t uint32) bool {
	s := &rpc.Shield{}
	s.SetStartTime(uint32(time.Now().Unix()))
	sh := p.GetShield()
	TotalTime := sh.GetTotalTime()
	s.SetTotalTime(t*60*60 + TotalTime)
	logger.Info("==TotalTime==%d\n", t*60*60+TotalTime)
	//s.SetTotalTime(120) //for test

	//同步到center
	err, ok := AddPlayerShield(s.GetStartTime(), s.GetTotalTime(), p.GetUid())
	if err != nil || !ok {
		return false
	}

	p.SetShield(s)

	logger.Info("AddShield:<%d hour> %s", t, p.GetUid())
	//同步到客户端
	if p.conn != nil {
		update := &rpc.UpdatePlayerInfo{Shield: s}
		WriteResult(p.conn, update)
	}

	return true
}

func (p *player) RemoveShield() {
	s := &rpc.Shield{}
	s.SetStartTime(0)
	s.SetTotalTime(0)

	//同步到center
	err := RemovePlayerShield(p.GetUid())
	if err != nil {
		return
	}

	p.SetShield(nil)

	logger.Info("RemoveShield %s", p.GetUid())
	//同步到客户端
	if p.conn != nil {
		update := &rpc.UpdatePlayerInfo{Shield: s}
		WriteResult(p.conn, update)
	}
}

//通天塔
/*func (p *player) RandomTTTCharacterMultiples(bFree bool) {
	pMultipleInfo := p.GetSelftttmultiples()
	//todo 实现免费随机次数配置表化 do
	freeTimeCfg := GetGlobalCfg("TID_TTT_ROLL_COUNT")
	uFreeTimes := pMultipleInfo.GetFreetimes()
	//展示屏蔽
	if (uFreeTimes >= freeTimeCfg && bFree) || (uFreeTimes < freeTimeCfg && !bFree) {
		//fmt.Println("*********当前的免费次数 和 是否是免费 ******", uFreeTimes, bFree)
		return
	}
	//todo 随机倍数，根据配置表免费非免费的概率 do
	uRandom := uint32(1)
	if bFree {
		uRandom = 2
	} else {
		var randnumber uint32 = uint32(1 + rand.Intn(100))
		var nTotal uint32 = 0
		for i := 2; i <= 10; i++ {
			nTotal += GetGlobalCfg("TID_TTT_ROLL_RATE_" + strconv.FormatInt(int64(i), 10))
			if nTotal > randnumber {
				uRandom = uint32(i)
				break
			}
		}
	}
	if uFreeTimes >= freeTimeCfg && !bFree {
		spendDemCfg := GetGlobalCfg("TID_TTT_ROLL_COST")
		if p.GetPlayerTotalGem() < spendDemCfg {
			logger.Error(" no Dem ! ")
			return
		} else {
			p.SetDiamonds(p.GetDiamonds() - spendDemCfg)
		}
	}
	if bFree {
		pMultipleInfo.SetFreetimes(uFreeTimes + 1)
	} else {
		pMultipleInfo.SetFreetimes(uFreeTimes)
	}
	pMultipleInfo.SetCurRandtimes(pMultipleInfo.GetCurRandtimes() + 1)
	//fmt.Println("*****2****当前的随机次数 ******", pMultipleInfo.GetCurRandtimes())
	uTimeCur := uint32(time.Now().Unix())
	pMultipleInfo.SetMultiple(uRandom)
	pMultipleInfo.SetRandomtime(uTimeCur)
	p.SetSelftttmultiples(pMultipleInfo)
	fmt.Println("*******服务器下发的次数信息  倍数 免费次数 时间*****", p.GetSelftttmultiples().GetMultiple(), p.GetSelftttmultiples().GetFreetimes(), p.GetSelftttmultiples().GetRandomtime())
	//下发
	if p.conn != nil {
		WriteResult(p.conn, pMultipleInfo)
	}
}*/

func (p *player) MatchNextCheckpointPlayer(msg rpc.TryMatchNextCheckpointPlayer) {

	//这里做一些数据处理，去整理一下包括把上一次的对手id置空，获取新的id并且获取新的
	pTTTInfo := p.GetSelftttinfo()
	if pTTTInfo == nil {
		return
	} else {
		pTTTInfo.SetMatchedplayerid("")
		fmt.Println("切换同一关的不同难度的对手 将对手放空")
	}
	pTTTInfo.SetChangetimes(pTTTInfo.GetChangetimes() + 1)
}
func (p *player) BeginTTTFight(msg rpc.TryStartTTT) bool {
	//todo 判断msg中兵力正确性，然后在数量里把随机的倍数设置 do
	v := p.GetVillage()
	var totalcount uint32 = 0
	var totalcountSpell uint32 = 0
	if msg.Characters == nil && msg.Spells == nil && !msg.GetIsBreakORReast() && !msg.GetIsContinueBattle() {
		logger.Error("first fight must have send army or spell")
		return false
	}
	fmt.Println("--BeginTTTFight--GetIsBreakORReast:   GetIsContinueBattle: ", msg.GetIsBreakORReast(), msg.GetIsContinueBattle())
	pNewTTTInfo := p.GetSelftttinfo()
	pMultipleInfo := p.GetSelftttmultiples()
	fmt.Println("--BeginTTTFight--1")
	if !msg.GetIsBreakORReast() {

		////todo 同一天判断最大次数，实现配置表化 do
		//everyDayTimesCfg := GetGlobalCfg("TTT_COUNT_PER_DAY")
		////fmt.Println("每天计算的当前的次数   和 每天可以打的总的天数 ", pNewTTTInfo.GetFightdaytimes(), everyDayTimesCfg)
		//if pNewTTTInfo.GetFightdaytimes() >= everyDayTimesCfg {
		//	logger.Error(" beyong the most times for everyday ")
		//	return
		//}
		//pNewTTTInfo.SetFightdaytimes(pNewTTTInfo.GetFightdaytimes() + 1)

	} else {
		//if len(pNewTTTInfo.Tttbuffs) != 0 {
		//	update := &rpc.UpdatePlayerInfo{}
		//	update.Tttbuffs = pNewTTTInfo.Tttbuffs
		//	if !WriteResult(p.conn, update) {
		//		logger.Error("Send buff data to clicent failed")
		//		return false
		//	}
		//}
	}
	fmt.Println("--BeginTTTFight--2")
	//这里判断是不是第一次进入战斗
	if !msg.GetIsContinueBattle() {
		fmt.Println("-第一次进入战斗--")
		fmt.Println("重置 兵数组数量 丹药数组数量", len(pNewTTTInfo.Characters), len(pNewTTTInfo.Spells))
		pNewTTTInfo.Characters = nil
		pNewTTTInfo.Spells = nil
		pNewTTTInfo.Tttbuffs = nil
		//pNewTTTInfo.SetReLiveCount(0)

		if len(v.Spellforge) != 0 {
			var spellforgelevel uint32 = 1
			for _, sp := range v.Spellforge {
				if sp.GetLevel() >= spellforgelevel {
					spellforgelevel = sp.GetLevel()
				}
			}
			spellspace := v.spellForge_GetTroopHousingTotalSpaces(spellforgelevel)
			fmt.Println("** 兵营空间  丹药容量 **", v.barrack_GetTroopHousingTotalSpaces(), spellspace)
			if totalcount > v.barrack_GetTroopHousingTotalSpaces() || totalcountSpell > spellspace {
				logger.Error("BeginTTTFight Failed : wrong count(%d/%d) and wrong spell", totalcount, v.barrack_GetTroopHousingTotalSpaces())
				return false
			}

		} else {
			if totalcount > v.barrack_GetTroopHousingTotalSpaces() {
				logger.Error("BeginTTTFight Failed : wrong count(%d/%d)", totalcount, v.barrack_GetTroopHousingTotalSpaces())
				return false
			}
		}
		mutiptimes := GetGlobalCfg("TTT_TROOP_MULTIPLE")
		if len(msg.Characters) != 0 {
			var totalspendcount uint32 = 0
			for _, c := range msg.Characters {
				fmt.Println("** 兵的数量 **翻倍的的倍数***=", c.GetCount(), mutiptimes)
				c.SetCount(c.GetCount() * mutiptimes)
				//c.SetCount(c.GetCount() * pMultipleInfo.GetMultiple())
				//扣招兵的钱
				cfg := GetCharacterCfgByTypeId(c.GetType(), c.GetLevel())
				totalspendcount += cfg.TrainingCost * c.GetCount()
			}

			if !p.CostResource(totalspendcount, proto.ResType_Food, proto.Lose_CreateCharacter) {
				logger.Error("BeginTTTFight Failed : wrong spend count(%d/%d)", totalspendcount, p.GetPlayerTotalGem())
				return false
			}
		}
		if len(msg.Spells) != 0 {
			var totalspendcountspell uint32 = 0
			for _, s := range msg.Spells {
				fmt.Println("** 丹药的数量 **翻倍的的倍数***=", s.GetCount(), mutiptimes)
				s.SetCount(s.GetCount() * mutiptimes)
				//c.SetCount(c.GetCount() * pMultipleInfo.GetMultiple())
				//扣建造丹药的钱
				spellCfg := GetSpellCfgByTypeId(s.GetType(), s.GetLevel())
				totalspendcountspell += spellCfg.TrainingCost * s.GetCount()
			}

			if !p.CostResource(totalspendcountspell, proto.ResType_Gold, proto.Lose_CreateSpell) {
				logger.Error("BeginTTTFight Failed : wrong spell spend count(%d/%d)", totalspendcountspell, p.GetPlayerTotalGem())
				return false
			}
		}

		fmt.Println("-重置关卡数，倍数，改变次数（用来）进行下一个对手的匹配的钱的计算")
		//重置关卡数，倍数，改变次数（用来）进行下一个对手的匹配的钱的计算
		pMultipleInfo.SetMultiple(mutiptimes)
		pNewTTTInfo.SetChangetimes(0)
		pNewTTTInfo.SetCurcheckpoint(0)
		fmt.Println("第一次 关卡  GetCurcheckpoint =", pNewTTTInfo.GetCurcheckpoint())
		//下发
		if p.conn != nil {
			WriteResult(p.conn, pMultipleInfo)
		}
		//刷新当前的积分
		tttRankPlayersInfo := p.GetTttrankplayerinfo()
		tttRankPlayersInfo.SetCurTTTScore(0)
		tttRankPlayersInfo.SetCurMostTTTScore(tttRankPlayersInfo.GetCurMostTTTScore())
		tttRankPlayersInfo.SetLastFreshTime(tttRankPlayersInfo.GetLastFreshTime())
		p.SetTttrankplayerinfo(tttRankPlayersInfo)
		fmt.Println("每次战斗开始的时候积分置0  tttscore=", p.GetTttrankplayerinfo().GetCurTTTScore())

		if len(msg.Characters) != 0 {
			for _, cfromP := range msg.Characters {
				chr := &rpc.Character{}
				chr.SetType(cfromP.GetType())
				chr.SetCount(cfromP.GetCount())
				chr.SetLevel(cfromP.GetLevel())
				fmt.Println("1 兵 保存 类型 等级 数量", cfromP.GetType(), cfromP.GetCount(), cfromP.GetLevel())
				pNewTTTInfo.Characters = append(pNewTTTInfo.Characters, chr)
			}
		}
		if len(msg.Spells) != 0 {
			for _, sfromP := range msg.Spells {
				spell := &rpc.Spell{}
				spell.SetType(sfromP.GetType())
				spell.SetCount(sfromP.GetCount())
				spell.SetLevel(sfromP.GetLevel())
				fmt.Println("1 丹 保存 类型 等级 数量", sfromP.GetType(), sfromP.GetCount(), sfromP.GetLevel())
				pNewTTTInfo.Spells = append(pNewTTTInfo.Spells, spell)
			}
		}
		//如果不是断线或是休息就添加英雄
		obj := v.buildings_Get(rpc.BuildingId_GeneralHouse, 0)
		if obj == nil {
			SyncError(v.p.conn, "hero_choose:obj(%d, %d) == nil", 0, 0)
			return false
		}
		gh := obj.(*rpc.GeneralHouse)
		ctype := gh.GetSelectedhero()
		if Hero_Has(gh, ctype) {
			fmt.Println("当前  英雄类型", ctype)
			if ctype != 0 {
				hero := Hero_Get(gh, ctype)
				hc := hero.GetCharacter()
				chr := &rpc.Character{}
				chr.SetType(hc.GetType())
				chr.SetCount(hc.GetCount())
				chr.SetLevel(hc.GetLevel())
				fmt.Println("添加 英雄类型 前 长度", len(pNewTTTInfo.Characters))
				pNewTTTInfo.Characters = append(pNewTTTInfo.Characters, chr)
				fmt.Println("添加 英雄类型 后 长度", len(pNewTTTInfo.Characters))
			}
		}

		if len(msg.Tttbuffs) != 0 {
			for _, tbuff := range msg.Tttbuffs {
				fmt.Println("buff花钱", ctype)
				//buff花钱
				tttBuffCfg := GetBuffCfgByTypeId(tbuff.GetType())
				if tttBuffCfg == nil {
					logger.Error("no buff config ")
					return false
				} else {
					switch tttBuffCfg.CostType {
					case "Gold":
						{
							fmt.Println("buff 花资源 类型 数量 ", tttBuffCfg.CostType, tttBuffCfg.Cost)
							if !p.CostResource(tttBuffCfg.Cost, proto.ResType_Gold, proto.Lose_UseTTTBuff) {
								logger.Error("BeginTTTFight buff cost Failed Gold")
								return false
							}
						}
					case "Food":
						{
							fmt.Println("buff 花资源 类型 数量 ", tttBuffCfg.CostType, tttBuffCfg.Cost)
							if !p.CostResource(tttBuffCfg.Cost, proto.ResType_Food, proto.Lose_UseTTTBuff) {
								logger.Error("BeginTTTFight buff cost Failed Food")
								return false
							}
						}
					}
				}

				if tbuff.GetType() == rpc.TTTBuff_TTTBuffJumpCheckPoint {
					pNewTTTInfo.SetCurcheckpoint(tttBuffCfg.Arg1)
				}
				if tbuff.GetType() == rpc.TTTBuff_TTTBuffAddBattleTime {
					tbuff.SetAddtimes(tttBuffCfg.Arg1)
				}
				tttbuff := &rpc.TTTBuff{}
				tttbuff.SetType(tbuff.GetType())
				tttbuff.SetCount(tbuff.GetCount())
				tttbuff.SetAddtimes(tbuff.GetAddtimes())
				fmt.Println("buff 兵力buff的处理 记录 buff类型", tttbuff.GetType())
				pNewTTTInfo.Tttbuffs = append(pNewTTTInfo.Tttbuffs, tttbuff)
			}
			update := &rpc.UpdatePlayerInfo{}
			update.Tttbuffs = msg.Tttbuffs
			if !WriteResult(p.conn, update) {
				logger.Error("Send buff data to clicent failed")
				return false
			}
		}
		fmt.Println("备份战斗数据 用来复活的时候加载")
		//备份战斗数据
		pNewTTTInfo.Startcharacters = nil
		pNewTTTInfo.Startspells = nil
		pNewTTTInfo.Starttttbuffs = nil

		if len(pNewTTTInfo.Characters) != 0 {
			for _, cfromP := range pNewTTTInfo.Characters {
				chr := &rpc.Character{}
				chr.SetType(cfromP.GetType())
				chr.SetCount(cfromP.GetCount())
				chr.SetLevel(cfromP.GetLevel())
				fmt.Println("备份 兵 保存 类型 等级 数量", cfromP.GetType(), cfromP.GetCount(), cfromP.GetLevel())
				pNewTTTInfo.Startcharacters = append(pNewTTTInfo.Startcharacters, chr)
			}
		}
		if len(pNewTTTInfo.Spells) != 0 {
			for _, sfromP := range pNewTTTInfo.Spells {
				spell := &rpc.Spell{}
				spell.SetType(sfromP.GetType())
				spell.SetCount(sfromP.GetCount())
				spell.SetLevel(sfromP.GetLevel())
				fmt.Println("备份 丹药 保存 类型 等级 数量", sfromP.GetType(), sfromP.GetCount(), sfromP.GetLevel())
				pNewTTTInfo.Startspells = append(pNewTTTInfo.Startspells, spell)
			}
		}
		if len(pNewTTTInfo.Tttbuffs) != 0 {
			for _, tbuff := range pNewTTTInfo.Tttbuffs {
				tttbuff := &rpc.TTTBuff{}
				tttbuff.SetType(tbuff.GetType())
				tttbuff.SetCount(tbuff.GetCount())
				fmt.Println("备份 buff 兵力buff的处理 记录 buff类型", tttbuff.GetType())
				pNewTTTInfo.Starttttbuffs = append(pNewTTTInfo.Starttttbuffs, tttbuff)
			}
		}
		fmt.Println("备份种类 Startcharacters: Startspells: Starttttbuffs:", len(pNewTTTInfo.Startcharacters), len(pNewTTTInfo.Startspells), len(pNewTTTInfo.Starttttbuffs))

	}
	fmt.Println("--BeginTTTFight--3  ，进行关卡数的设置")
	fmt.Println("上一个 关卡  GetCurcheckpoint =", pNewTTTInfo.GetCurcheckpoint())
	fmt.Println("检测GetIsBreakORReast GetIsTTTEndEveryCheckpoint()", msg.GetIsBreakORReast(), pNewTTTInfo.GetIsTTTEndEveryCheckpoint())
	if msg.GetIsBreakORReast() && !pNewTTTInfo.GetIsTTTEndEveryCheckpoint() {
		//pNewTTTInfo.SetCurcheckpoint(pNewTTTInfo.GetCurcheckpoint())
	} else {
		isHaveJumpCheckPoint := false
		if len(msg.Tttbuffs) != 0 {
			for _, tbuff := range msg.Tttbuffs {
				if tbuff.GetType() == rpc.TTTBuff_TTTBuffJumpCheckPoint {
					isHaveJumpCheckPoint = true
				}
			}
		}
		if !msg.GetIsContinueBattle() && isHaveJumpCheckPoint {
			//pNewTTTInfo.SetCurcheckpoint(pNewTTTInfo.GetCurcheckpoint())
		} else {
			pNewTTTInfo.SetCurcheckpoint(pNewTTTInfo.GetCurcheckpoint() + 1)
		}
	}
	fmt.Println("当前的 关卡  GetCurcheckpoint =", pNewTTTInfo.GetCurcheckpoint())
	pNewTTTInfo.SetIsTTTEndEveryCheckpoint(false)
	fmt.Println("检测 前 战斗结束设置  GetIsStart  GetIsTTTOver", pNewTTTInfo.GetIsStart(), pNewTTTInfo.GetIsTTTOver())
	pNewTTTInfo.SetIsStart(true)
	pNewTTTInfo.SetIsTTTOver(false)

	return true
}

//擂台赛开始
func (self *player) ChallengeBegin(uConnId uint64, op *player, revBattleLogId uint64, bIsChallenge bool) bool {
	self.fightinfo = &playerfightinfo{
		pvpmsg:         rpc.AttackBegin{},
		uVillageId:     op.GetVillageId(),
		def_log:        rpc.BattleLog{},
		replay:         rpc.BattleReplay{},
		bInFight:       false,
		bPve:           false,
		bFriendExecise: false,
		bTTTFight:      false,
		bIsChallenge:   false,
	}

	self.fightinfo.bIsChallenge = bIsChallenge
	self.fightinfo.pvpmsg.V = op.v.VillageInfo
	self.fightinfo.pvpmsg.SetPlayerlid(uConnId)
	self.fightinfo.pvpmsg.SetSrcTrophy(self.GetTrophy())
	self.fightinfo.pvpmsg.SetTarTrophy(op.GetTrophy())

	claninfo := self.GetClanInfo()

	def_log := &self.fightinfo.def_log
	def_log.SetBid(revBattleLogId)
	def_log.SetPid(self.GetUid())
	def_log.SetName(self.GetName())
	def_log.SetLevel(self.GetLevel())
	def_log.SetClanName(claninfo.GetName())
	def_log.SetClanSymbol(claninfo.GetSymbol())
	def_log.SetTrophy(self.GetTrophy())
	def_log.SetTime(uint32(time.Now().Unix()))
	def_log.SetState(rpc.BattleLog_UnRead)
	def_log.SetRevstate(rpc.BattleLog_UnRevenged)

	replay := &self.fightinfo.replay
	replay.SetV(op.v.VillageInfo)

	//不破盾

	return true
}

//同步宝石数量
func (self *player) SyncPlayerGem() {
	update := &rpc.UpdatePlayerInfo{}
	update.SetDiamonds(self.GetPlayerTotalGem())

	WriteResult(self.conn, update)
}

func (self *player) GetPlayerTotalGem() uint32 {
	total := self.GetDiamonds()
	if self.mobileqqinfo != nil {
		total += self.mobileqqinfo.Balance
	}

	return total
}

func (self *player) UpdateThirdGem() {
	if self.mobileqqinfo == nil {
		return
	}

	//非qq渠道就不用走这个流程了，否则会被踢下线
	if len(self.mobileqqinfo.Openid) == 0 {
		return
	}

	if success, errmsg, balance := MobileQQBalance(self); success {
		logger.Info("MobileQQBalance Result success:%d", balance)
		//更新
		self.mobileqqinfo.Balance = uint32(balance)
	} else {
		logger.Info("MobileQQBalance Result failed:%s", errmsg)

		WriteLoginResultWithErrorMsg(self.conn, rpc.LoginResult_TX_AUTH_FAILED, errmsg)
	}
}
