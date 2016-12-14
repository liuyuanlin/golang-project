package center

import (
	"golang-project/dpsg/common"
	"golang-project/dpsg/dbclient"
	"golang-project/dpsg/logger"
	"golang-project/dpsg/proto"
	"golang-project/dpsg/rpc"
	"golang-project/dpsg/timer"
	"math/rand"
	"strconv"
	"sync"
	"time"

	gp "github.com/golang/protobuf/proto"
)

//擂台的索引
type ChallengeKey struct {
	level int
	index int
}

//玩家在守的擂台
type PlayerDefenseInfo struct {
	etype rpc.ChallengeType
	key   ChallengeKey
}

//普通擂台
type CenterNormalChallengeInfo struct {
	*rpc.NormalChallengeInfo
	isHavetick bool
	tickend    *timer.Timer //结束tick
	tickdelay  *timer.Timer //战斗超时tick
}

//赏金擂台
type CenterMoneyChallengeInfo struct {
	*rpc.MoneyChallengeInfo
	isHavetick bool
	tickend    *timer.Timer
	tickdelay  *timer.Timer
}

type CenterChallengeService struct {
	//普通擂台
	mapAllNormalChallenge map[ChallengeKey]*CenterNormalChallengeInfo
	//赏金擂台
	mapAllMoneyChallenge map[ChallengeKey]*CenterMoneyChallengeInfo
	//玩家在守的擂台
	mapPlayerId2Challenge map[string]*PlayerDefenseInfo
	//正在挑战的玩家列表
	mapPlayerId2Challenging map[string]*PlayerDefenseInfo
	//锁
	lock                 sync.RWMutex
	bIsFinishedChallenge bool
	bIsTimeEnd           bool
}

//擂台条件返回
type PlayerInfo struct {
	Code int
	Uid  string
}

var pCenterChallengeService *CenterChallengeService

func startChallengeService() {
	pCenterChallengeService = &CenterChallengeService{
		mapAllNormalChallenge:   make(map[ChallengeKey]*CenterNormalChallengeInfo),
		mapAllMoneyChallenge:    make(map[ChallengeKey]*CenterMoneyChallengeInfo),
		mapPlayerId2Challenge:   make(map[string]*PlayerDefenseInfo),
		mapPlayerId2Challenging: make(map[string]*PlayerDefenseInfo),
		bIsFinishedChallenge:    false,
		bIsTimeEnd:              false,
	}

	common.LoadChallengeConfigFiles()

	//锁定
	pCenterChallengeService.lock.Lock()
	defer pCenterChallengeService.lock.Unlock()

	timeCur := uint32(time.Now().Unix())

	//从数据库查找已存在的擂台信息
	pCenterChallengeService.getPlayerFromDB()

	//判断已过期的擂台，并且删除。如果没有过期，则注册剩余时间
	pCenterChallengeService.isOutTime(timeCur)

	//注册结束tick

	//注册所有的普通擂台
	for key, value := range pCenterChallengeService.mapAllNormalChallenge {

		if value.isHavetick {
			continue
		}

		value.tickend = timer.NewTimer(time.Duration(common.GetChallengeInfoCfg("LT_DEFEND_LAST_TIME")) * time.Second)
		value.tickend.Start(
			func() {
				pCenterChallengeService.OnTimeEnd(rpc.ChallengeType_Normal, key.level, key.index, false)
			},
		)
	}

	//注册所有的赏金擂台
	for key, value := range pCenterChallengeService.mapAllMoneyChallenge {

		if value.isHavetick {
			continue
		}

		value.tickend = timer.NewTimer(time.Duration(common.GetChallengeInfoCfg("LT_DEFEND_LAST_TIME")) * time.Second)
		value.tickend.Start(
			func() {
				pCenterChallengeService.OnTimeEnd(rpc.ChallengeType_Money, key.level, key.index, false)
			},
		)
	}

	//todo 关联玩家在守的
	pCenterChallengeService.alreadyHavePlayer()

	//没有的要重建
	pCenterChallengeService.createNewChallenge()
}

//从数据库加载擂台信息
func (self *CenterChallengeService) getPlayerFromDB() {
	//todo 先去数据库里取保存的数据

	//普通的擂台

	for level := common.GetChallengeInfoCfg("LT_LEVEL_MIN"); level <= common.GetChallengeInfoCfg("LT_LEVEL_MAX"); level++ {
		//普通的
		for index := int32(0); index < common.GetChallengeInfoCfg("LT_COUNT_NORMAL"); index++ {

			p := rpc.NormalChallengeInfo{}

			exist, err := dbclient.KVQueryExt(common.ChallengeKeyName_NormalChallenge, strconv.Itoa(int(level))+"_"+strconv.Itoa(int(index)), &p)
			if err != nil {
				continue
			}
			if exist {
				challenge := &CenterNormalChallengeInfo{}

				mykey := ChallengeKey{
					index: int(p.GetIndex()),
					level: int(p.GetLevel()),
				}

				challenge.NormalChallengeInfo = &p

				pCenterChallengeService.mapAllNormalChallenge[mykey] = challenge
			}
		}
	}

	//赏金擂台

	for level := common.GetChallengeInfoCfg("LT_LEVEL_MIN"); level <= common.GetChallengeInfoCfg("LT_LEVEL_MAX"); level++ {
		//赏金擂台
		for index := int32(0); index < common.GetChallengeInfoCfg("LT_RANDOM_PER_COUNT"); index++ {

			p := rpc.MoneyChallengeInfo{}

			exist, err := dbclient.KVQueryExt(common.ChallengeKeyName_MoneyChallenge, strconv.Itoa(int(level))+"_"+strconv.Itoa(int(index)), &p)
			if err != nil {
				continue
			}
			if exist {
				challenge := &CenterMoneyChallengeInfo{}

				mykey := ChallengeKey{
					index: int(p.Base.GetIndex()),
					level: int(p.Base.GetLevel()),
				}

				challenge.MoneyChallengeInfo = &p

				pCenterChallengeService.mapAllMoneyChallenge[mykey] = challenge
			}
		}
	}
}

//判断已过期的
func (self *CenterChallengeService) isOutTime(timeCur uint32) {
	//普通擂台

	for key, value := range pCenterChallengeService.mapAllNormalChallenge {
		if value.GetStarttime()+uint32(common.GetChallengeInfoCfg("LT_DEFEND_LAST_TIME")) <= timeCur {
			delete(pCenterChallengeService.mapAllNormalChallenge, key)
			//todo 从数据库删除, 并且发放成功奖励
			//发放奖励
			pCenterChallengeService.NormalChallenge(rpc.ChallengeType_Normal, key.level, key.index, false, value.NormalChallengeInfo.GetHostid(), "", "")

			_, err := dbclient.KVDeleteExt(common.ChallengeKeyName_NormalChallenge, strconv.Itoa(key.level)+"_"+strconv.Itoa(key.index))

			if err != nil {
				logger.Error("删除错误", key)
			}
		} else {
			//否则就注册剩余时间的擂台
			value.tickend = timer.NewTimer(time.Duration(uint32(common.GetChallengeInfoCfg("LT_DEFEND_LAST_TIME"))+value.GetStarttime()-timeCur) * time.Second)
			value.tickend.Start(
				func() {
					pCenterChallengeService.OnTimeEnd(rpc.ChallengeType_Normal, key.level, key.index, false)
				},
			)
			value.isHavetick = true
		}
	}

	//赏金擂台
	for key, value := range pCenterChallengeService.mapAllMoneyChallenge {
		if value.Base.GetStarttime()+uint32(common.GetChallengeInfoCfg("LT_DEFEND_LAST_TIME")) <= timeCur {
			delete(pCenterChallengeService.mapAllMoneyChallenge, key)
			//todo 从数据库删除, 并且发放成功奖励

			pCenterChallengeService.MoneyChallenge(rpc.ChallengeType_Money, key.level, key.index, false, value.Base.GetHostid(), "", "")

			_, err := dbclient.KVDeleteExt(common.ChallengeKeyName_MoneyChallenge, strconv.Itoa(key.level)+"_"+strconv.Itoa(key.index))
			if err != nil {
				logger.Error("删除错误", key)
			}
		} else {
			value.tickend = timer.NewTimer(time.Duration(uint32(common.GetChallengeInfoCfg("LT_DEFEND_LAST_TIME"))+value.Base.GetStarttime()-timeCur) * time.Second)
			value.tickend.Start(
				func() {
					pCenterChallengeService.OnTimeEnd(rpc.ChallengeType_Money, key.level, key.index, false)
				},
			)
			value.isHavetick = true
		}
	}
}

//关联已经有玩家的擂台
func (self *CenterChallengeService) alreadyHavePlayer() {
	//这里关联普通擂台的玩家
	for _, value := range pCenterChallengeService.mapAllNormalChallenge {
		if value.NormalChallengeInfo.GetHostid() != "" {
			p := &PlayerDefenseInfo{}

			p.etype = rpc.ChallengeType_Normal
			p.key.level = int(value.NormalChallengeInfo.GetLevel())
			p.key.index = int(value.NormalChallengeInfo.GetIndex())
			value.NormalChallengeInfo.SetChallengeid("")

			pCenterChallengeService.mapPlayerId2Challenge[value.NormalChallengeInfo.GetHostid()] = p
		}
	}

	//这里关联赏金擂台的玩家
	for _, value := range pCenterChallengeService.mapAllMoneyChallenge {
		if value.Base.GetHostid() != "" {
			p := &PlayerDefenseInfo{}

			p.etype = rpc.ChallengeType_Money
			p.key.level = int(value.Base.GetLevel())
			p.key.index = int(value.Base.GetIndex())
			value.Base.SetChallengeid("")

			pCenterChallengeService.mapPlayerId2Challenge[value.Base.GetHostid()] = p
		}
	}
}

//没有擂台的要新建

func (self *CenterChallengeService) createNewChallenge() {
	//没有的要重建
	for level := int(common.GetChallengeInfoCfg("LT_LEVEL_MIN")); level <= int(common.GetChallengeInfoCfg("LT_LEVEL_MAX")); level++ {
		//普通的
		for index := 0; index < int(common.GetChallengeInfoCfg("LT_COUNT_NORMAL")); index++ {
			key := ChallengeKey{level: level, index: index}
			if _, ok := pCenterChallengeService.mapAllNormalChallenge[key]; !ok { //数据库没有记录就新建一个
				challenge := &rpc.NormalChallengeInfo{}
				challenge.SetIndex(uint32(index))
				challenge.SetLevel(uint32(level))
				challenge.SetState(rpc.NormalChallengeInfo_None)
				challenge.SetChallengetimes(uint32(0))
				challenge.SetStarttime(uint32(0))
				challenge.SetScore(uint32(0))

				pCenterChallengeService.mapAllNormalChallenge[key] = &CenterNormalChallengeInfo{
					NormalChallengeInfo: challenge,
				}
			}

		}

		//赏金的
		for index := 0; index < int(common.GetChallengeInfoCfg("LT_RANDOM_PER_COUNT")); index++ {
			key := ChallengeKey{level: level, index: index}
			if _, ok := pCenterChallengeService.mapAllMoneyChallenge[key]; !ok { //数据库没有记录就新建一个
				base := &rpc.NormalChallengeInfo{}
				base.SetIndex(uint32(index))
				base.SetLevel(uint32(level))
				base.SetState(rpc.NormalChallengeInfo_None)
				base.SetChallengetimes(uint32(0))
				base.SetStarttime(uint32(0))
				base.SetScore(uint32(0))

				challenge := &rpc.MoneyChallengeInfo{}
				challenge.SetBase(base)
				challenge.SetMoney(uint32(0))

				pCenterChallengeService.mapAllMoneyChallenge[key] = &CenterMoneyChallengeInfo{
					MoneyChallengeInfo: challenge,
				}
			}
		}
	}
}

//保存普通擂台结果
func (self *CenterChallengeService) saveNormalChallengeResult(level, index int) {
	if value, ok := self.mapAllNormalChallenge[ChallengeKey{level: level, index: index}]; ok {
		_, err := dbclient.KVWriteExt(common.ChallengeKeyName_NormalChallenge, strconv.Itoa(int(level))+"_"+strconv.Itoa(int(index)), value.NormalChallengeInfo)

		if err != nil {
			logger.Error("写入普通擂台赛到数据库是发生错误: ", err.Error())
		}
	}
}

//保存赏金擂台结果
func (self *CenterChallengeService) saveMoneyChallengeResult(level, index int) {
	if value, ok := self.mapAllMoneyChallenge[ChallengeKey{level: level, index: index}]; ok {
		_, err := dbclient.KVWriteExt(common.ChallengeKeyName_MoneyChallenge, strconv.Itoa(int(level))+"_"+strconv.Itoa(int(index)), value.MoneyChallengeInfo)

		if err != nil {
			logger.Error("写入赏金擂台赛到数据库是发生错误: ", err.Error())
		}
	}
}

//擂台持续时间到达了，要注意正在挑战的要等战斗结束再结算(下面都有etype的判断，如果函数较大，请拆分)
func (self *CenterChallengeService) OnTimeEnd(etype rpc.ChallengeType, level, index int, isQuitChallenge bool) {
	//锁定
	self.lock.Lock()
	defer self.lock.Unlock()

	key := ChallengeKey{level: level, index: index}
	if etype == rpc.ChallengeType_Normal { //普通擂台
		if challenge, ok := self.mapAllNormalChallenge[key]; ok {

			if challenge.tickend == nil {
				return
			}

			challenge.tickend.Stop()
			challenge.tickend = nil

			//todo 这里给奖励（如果正在战斗的话先不处理，等战斗结束或者超时）

			if challenge.NormalChallengeInfo.GetState() == rpc.NormalChallengeInfo_Challenging && !self.bIsFinishedChallenge {
				self.bIsTimeEnd = true
				return
			}

			var p rpc.PlayerChallengeInfo
			exist, err := dbclient.KVQueryExt(common.ChallengeKeyName_PlayerChallengeInfo, challenge.NormalChallengeInfo.GetHostid(), &p)
			if err != nil {
				return
			}

			//更新玩家身上的数据
			if exist {
				p.SetScore(p.GetScore() + challenge.NormalChallengeInfo.GetScore())
				p.SetSalarytime(p.GetSalarytime())
			} else {
				p.SetScore(challenge.NormalChallengeInfo.GetScore())
				p.SetSalarytime(p.GetSalarytime())
			}

			_, err = dbclient.KVWriteExt(common.ChallengeKeyName_PlayerChallengeInfo, challenge.NormalChallengeInfo.GetHostid(), &p)

			if err != nil {
				logger.Error("写入普通擂台赛到数据库是发生错误: ", err.Error())
			}

			//删除擂主
			if _, exist := self.mapPlayerId2Challenge[challenge.NormalChallengeInfo.GetHostid()]; exist {
				delete(self.mapPlayerId2Challenge, challenge.NormalChallengeInfo.GetHostid())
			}

			//如果有挑战者，就删除挑战者
			if _, exist := self.mapPlayerId2Challenging[challenge.NormalChallengeInfo.GetChallengeid()]; exist {
				delete(self.mapPlayerId2Challenging, challenge.NormalChallengeInfo.GetChallengeid())
			}

			challenge.NormalChallengeInfo.SetState(rpc.NormalChallengeInfo_None)
			challenge.NormalChallengeInfo.SetChallengetimes(0)
			challenge.NormalChallengeInfo.SetStarttime(0)
			challenge.NormalChallengeInfo.SetScore(0)
			challenge.NormalChallengeInfo.SetHostid("")
			challenge.NormalChallengeInfo.SetHostname("")
			challenge.NormalChallengeInfo.SetHostclanName("")
			challenge.NormalChallengeInfo.SetHostclanSymbol(0)
			challenge.NormalChallengeInfo.SetChallengeid("")

			//从数据库里删除该擂台
			_, err = dbclient.KVDeleteExt(common.ChallengeKeyName_NormalChallenge, strconv.Itoa(level)+"_"+strconv.Itoa(index))
			if err != nil {
				logger.Error("删除错误", key)
			}
		}

	} else if etype == rpc.ChallengeType_Money { //赏金擂台
		if challenge, ok := self.mapAllMoneyChallenge[key]; ok {
			if challenge.tickend == nil {
				return
			}

			challenge.tickend.Stop()
			challenge.tickend = nil

			//todo 这里给奖励（如果正在战斗的话先不处理，，等战斗结束或者超时）

			if challenge.MoneyChallengeInfo.Base.GetState() == rpc.NormalChallengeInfo_Challenging && !self.bIsFinishedChallenge {
				self.bIsTimeEnd = true
				return
			}

			var p rpc.PlayerChallengeInfo
			exist, err := dbclient.KVQueryExt(common.ChallengeKeyName_PlayerChallengeInfo, challenge.MoneyChallengeInfo.Base.GetHostid(), &p)
			if err != nil {
				return
			}

			//更新玩家身上的数据
			if exist {
				p.SetScore(p.GetScore() + challenge.MoneyChallengeInfo.Base.GetScore())
				p.SetSalarytime(p.GetSalarytime())

			} else {
				p.SetScore(p.GetScore() + challenge.MoneyChallengeInfo.Base.GetScore())
				p.SetSalarytime(p.GetSalarytime())
			}

			_, err = dbclient.KVWriteExt(common.ChallengeKeyName_PlayerChallengeInfo, challenge.MoneyChallengeInfo.Base.GetHostid(), &p)

			if err != nil {
				logger.Error("写入赏金擂台擂台赛到数据库是发生错误: ", err.Error())
			}

			//这里要告诉cn发邮件给玩家
			req := &proto.SendMail{}
			ret := &proto.SendMailResult{}
			req.Uid = challenge.MoneyChallengeInfo.Base.GetHostid()

			if isQuitChallenge {
				req.Money = uint32(float64(challenge.MoneyChallengeInfo.GetMoney()) * float64(common.GetChallengeInfoCfg("LT_BREAK_SCALE")) / 100.0)
			} else {
				req.Money = challenge.MoneyChallengeInfo.GetMoney()
			}

			conn := centerServer.cnss[0]
			if err := conn.Call("CenterService.SendMailtoplayer", req, ret); err != nil {
				logger.Error("调用发送邮件错误")
			}

			//把自己从守擂列表中删除
			if _, exist := self.mapPlayerId2Challenge[challenge.Base.GetHostid()]; exist {
				delete(self.mapPlayerId2Challenge, challenge.Base.GetHostid())
			}

			//如果有挑战者，就删除挑战者
			if _, exist := self.mapPlayerId2Challenging[challenge.Base.GetChallengeid()]; exist {
				delete(self.mapPlayerId2Challenging, challenge.Base.GetChallengeid())
			}

			//设置擂台为空的属性
			challenge.Base.SetState(rpc.NormalChallengeInfo_None)
			challenge.Base.SetChallengetimes(0)
			challenge.Base.SetStarttime(0)
			challenge.Base.SetScore(0)
			challenge.Base.SetHostid("")
			challenge.Base.SetHostname("")
			challenge.Base.SetHostclanName("")
			challenge.Base.SetHostclanSymbol(0)
			challenge.Base.SetChallengeid("")
			challenge.SetMoney(0)

			//从数据库里删除该擂台
			_, err = dbclient.KVDeleteExt(common.ChallengeKeyName_MoneyChallenge, strconv.Itoa(level)+"_"+strconv.Itoa(index))
			if err != nil {
				logger.Error("删除错误", key)
			}
		}
	}
}

//战斗超时了
func (self *CenterChallengeService) OnTimeDelay(etype rpc.ChallengeType, level, index int, name string) {

	key := ChallengeKey{level: level, index: index}
	if etype == rpc.ChallengeType_Normal { //普通擂台
		if challenge, ok := self.mapAllNormalChallenge[key]; ok {
			self.OnFightResult(etype, level, index, false, challenge.GetHostid(), challenge.GetChallengeid(), name)
		}

	} else if etype == rpc.ChallengeType_Money { //赏金擂台
		if challenge, ok := self.mapAllMoneyChallenge[key]; ok {
			self.OnFightResult(etype, level, index, false, challenge.Base.GetHostid(), challenge.Base.GetChallengeid(), name)
		}
	}
}

//空擂台占领处理
func (self *CenterChallengeService) IsEmptyChallenge(etype rpc.ChallengeType, level int, index int, success bool, name, challenger string) {
	//1.把自己从挑战者列表中删除
	timeCur := uint32(time.Now().Unix())
	key := ChallengeKey{level: level, index: index}
	if etype == rpc.ChallengeType_Normal { //普通擂台
		if challenge, ok := self.mapAllNormalChallenge[key]; ok {
			//todo 防守与挑战者是否一致

			delete(self.mapPlayerId2Challenging, challenge.GetChallengeid())

			//空擂台判断
			if success && challenge.NormalChallengeInfo.GetState() == rpc.NormalChallengeInfo_None {
				challenge.SetHostid(challenger)
				challenge.SetHostname(name)
				challenge.SetState(rpc.NormalChallengeInfo_Free)
				challenge.SetStarttime(timeCur)
				challenge.SetScore(0)

				p := &PlayerDefenseInfo{}

				p.etype = rpc.ChallengeType_Normal
				p.key.level = level
				p.key.index = index

				pCenterChallengeService.mapPlayerId2Challenge[challenger] = p

				//给玩家积累分数
				var Keeper rpc.PlayerChallengeInfo
				exist, err := dbclient.KVQueryExt(common.ChallengeKeyName_PlayerChallengeInfo, challenger, &Keeper)

				//更新首擂玩家身上的数据
				if err == nil {
					if exist {
						Keeper.SetScore(Keeper.GetScore() + uint32(common.GetChallengeInfoCfg("LT_SPECIAL_LOSE_MARK")))
					}

				} else {
					Keeper.SetScore(uint32(common.GetChallengeInfoCfg("LT_SPECIAL_LOSE_MARK")))
					Keeper.SetSalarytime(0)
				}

				if _, err = dbclient.KVWriteExt(common.ChallengeKeyName_PlayerChallengeInfo, challenger, &Keeper); err != nil {
					logger.Error("写入守擂玩家到数据库是发生错误: ", err.Error())
				}

				//注册结束tick
				challenge.tickend = timer.NewTimer(time.Duration(common.GetChallengeInfoCfg("LT_DEFEND_LAST_TIME")) * time.Second)
				challenge.tickend.Start(
					func() {
						pCenterChallengeService.OnTimeEnd(rpc.ChallengeType_Normal, p.key.level, p.key.index, false)
					},
				)

				//存数据库
				self.saveNormalChallengeResult(level, index)
				return
			}
		}
	}

	if etype == rpc.ChallengeType_Money { //赏金擂台
		if challenge, ok := self.mapAllMoneyChallenge[key]; ok {

			//todo 这之前请先判断是否超时，防守与挑战者是否一致

			delete(self.mapPlayerId2Challenging, challenge.Base.GetChallengeid())

			//空擂台判断
			if success && challenge.MoneyChallengeInfo.Base.GetState() == rpc.NormalChallengeInfo_None {
				challenge.MoneyChallengeInfo.Base.SetHostid(challenger)
				challenge.MoneyChallengeInfo.Base.SetHostname(name)
				challenge.MoneyChallengeInfo.Base.SetStarttime(timeCur)
				challenge.MoneyChallengeInfo.Base.SetState(rpc.NormalChallengeInfo_Free)
				challenge.MoneyChallengeInfo.SetMoney(uint32(common.GetChallengeInfoCfg("LT_SPECIAL_EMPTY_GEM")))
				challenge.MoneyChallengeInfo.Base.SetScore(0)

				logger.Info("Empty is", challenge.MoneyChallengeInfo)

				p := &PlayerDefenseInfo{}

				p.etype = rpc.ChallengeType_Money
				p.key.level = level
				p.key.index = index

				pCenterChallengeService.mapPlayerId2Challenge[challenger] = p

				//给玩家积累分数
				var Keeper rpc.PlayerChallengeInfo
				exist, err := dbclient.KVQueryExt(common.ChallengeKeyName_PlayerChallengeInfo, challenger, &Keeper)

				//更新首擂玩家身上的数据
				if err == nil {
					if exist {
						Keeper.SetScore(Keeper.GetScore() + uint32(common.GetChallengeInfoCfg("LT_SPECIAL_WIN_MARK")))
					}
				} else {
					Keeper.SetScore(uint32(common.GetChallengeInfoCfg("LT_SPECIAL_WIN_MARK")))
					Keeper.SetSalarytime(0)
				}

				if _, err = dbclient.KVWriteExt(common.ChallengeKeyName_PlayerChallengeInfo, challenger, &Keeper); err != nil {
					logger.Error("写入守擂玩家到数据库是发生错误: ", err.Error())
				}

				//注册结束tick
				challenge.tickend = timer.NewTimer(time.Duration(common.GetChallengeInfoCfg("LT_DEFEND_LAST_TIME")) * time.Second)
				challenge.tickend.Start(
					func() {
						pCenterChallengeService.OnTimeEnd(rpc.ChallengeType_Money, p.key.level, p.key.index, false)
					},
				)

				//存数据库
				self.saveMoneyChallengeResult(level, index)

				return
			}
		}
	}
}

//cns返回战斗结束
func (self *CenterChallengeService) OnFightResult(etype rpc.ChallengeType, level int, index int, success bool, host, challenger, name string) {
	//锁定
	self.lock.Lock()
	defer self.lock.Unlock()

	self.NormalChallenge(etype, level, index, success, host, challenger, name)
	self.MoneyChallenge(etype, level, index, success, host, challenger, name)

	return
}

//从普通列表中随机几个
func (self *CenterChallengeService) RandomNormalChallenges(uid string, level int) *rpc.NormalChallenges {
	//读锁
	self.lock.RLock()
	defer self.lock.RUnlock()

	challenges := &rpc.NormalChallenges{}
	//随机取需要的个数
	//todo 判断有没有自己在守的擂台，始终排在第一位

	i2Random := common.GetChallengeInfoCfg("LT_COUNT_SPECIAL")
	iSelfIndex := -1

	if value, ok := self.mapPlayerId2Challenge[uid]; ok {
		result, _ := self.mapAllNormalChallenge[value.key]
		challenges.Challenges = append(challenges.Challenges, result.NormalChallengeInfo)
		i2Random--
		iSelfIndex = value.key.index
	}

	indexs := rand.Perm(int(common.GetChallengeInfoCfg("LT_COUNT_NORMAL")))[:uint32(common.GetChallengeInfoCfg("LT_COUNT_SPECIAL"))]
	for _, index := range indexs {
		if i2Random <= 0 {
			break
		}

		if iSelfIndex == index {
			continue
		}

		if info, ok := self.mapAllNormalChallenge[ChallengeKey{level: level, index: index}]; ok {
			challenges.Challenges = append(challenges.Challenges, info.NormalChallengeInfo)
		}

		i2Random--
	}

	return challenges
}

//取赏金列表
func (self *CenterChallengeService) GetMoneyChallenges(uid string, level int) *rpc.MoneyChallenges {
	//读锁
	self.lock.RLock()
	defer self.lock.RUnlock()

	challenges := &rpc.MoneyChallenges{}

	iSelfIndex := -1

	if value, ok := self.mapPlayerId2Challenge[uid]; ok {
		if result, ok := self.mapAllMoneyChallenge[value.key]; ok {
			challenges.Challenges = append(challenges.Challenges, result.MoneyChallengeInfo)
			iSelfIndex = value.key.index
		}
	}

	for index := 0; index < int(common.GetChallengeInfoCfg("LT_RANDOM_PER_COUNT")); index++ {
		if index == iSelfIndex {
			continue
		}
		key := ChallengeKey{index: index, level: level}
		if challenge, ok := self.mapAllMoneyChallenge[key]; ok {
			challenges.Challenges = append(challenges.Challenges, challenge.MoneyChallengeInfo)
		}
	}

	return challenges
}

func (self *CenterChallengeService) BeginNormalChallenge(index, level uint32, Challengeid, ChallengeName string) (*PlayerInfo, error) {

	player := &PlayerInfo{}

	//获取当前时间
	timeCur := uint32(time.Now().Unix())

	//1.首先从列表中找到这个擂台
	key := &ChallengeKey{
		index: int(index),
		level: int(level),
	}

	value, ok := self.mapAllNormalChallenge[*key]
	if !ok {
		player.Code = proto.NoArena
		player.Uid = value.NormalChallengeInfo.GetHostid()
		return player, nil
	}
	//2.看看擂台是不是已经有人在挑战了
	if value.GetState() == rpc.NormalChallengeInfo_Challenging {
		player.Code = proto.AlreadyChallengeing
		player.Uid = value.NormalChallengeInfo.GetHostid()
		return player, nil
	}

	//3.查看擂台的时间是不是正确
	if value.GetStarttime()+uint32(common.GetChallengeInfoCfg("LT_DEFEND_LAST_TIME")) <= timeCur && value.GetState() != rpc.NormalChallengeInfo_None {
		player.Code = proto.AlreadyTimeOut
		player.Uid = value.NormalChallengeInfo.GetHostid()
		return player, nil
	}

	//4.看看自己是不是在守擂
	if _, ok := self.mapPlayerId2Challenge[Challengeid]; ok {
		player.Code = proto.AlreadyHost
		player.Uid = value.NormalChallengeInfo.GetHostid()
		return player, nil
	}

	//5.看看自己是不是在挑战
	if _, ok := self.mapPlayerId2Challenging[Challengeid]; ok {
		player.Code = proto.AlreadyChallenger
		player.Uid = value.NormalChallengeInfo.GetHostid()
		return player, nil
	}

	//6.看看擂台是不是空的,是就设置擂台的擂主名字和id和擂台状态
	if value.GetState() == rpc.NormalChallengeInfo_None {
		player.Code = proto.Empty
		player.Uid = value.NormalChallengeInfo.GetHostid()
		self.IsEmptyChallenge(rpc.ChallengeType_Normal, int(level), int(index), true, ChallengeName, Challengeid)

		return player, nil
	}
	//7.如果都正确，就将返回值填写为ok
	p := &PlayerDefenseInfo{}

	p.etype = rpc.ChallengeType_Normal
	p.key.level = int(value.GetLevel())
	p.key.index = int(value.GetIndex())

	pCenterChallengeService.mapPlayerId2Challenging[Challengeid] = p

	//注册超时tick
	value.tickdelay = timer.NewTimer(time.Duration(60*common.GetChallengeInfoCfg("LT_BATTLE_TIME")) * time.Second)

	value.tickdelay.Start(
		func() {
			pCenterChallengeService.OnTimeDelay(rpc.ChallengeType_Normal, p.key.level, p.key.index, ChallengeName)
		},
	)

	player.Code = proto.NormalChallengeOK
	player.Uid = value.NormalChallengeInfo.GetHostid()
	value.NormalChallengeInfo.SetChallengeid(Challengeid)
	value.NormalChallengeInfo.SetChallengetimes(value.NormalChallengeInfo.GetChallengetimes() + 1)

	//存数据库
	self.saveNormalChallengeResult(int(value.GetLevel()), int(value.GetIndex()))

	return player, nil
}

//赏金擂台
func (self *CenterChallengeService) BeginMoneyChallenge(index, level uint32, Challengeid, ChallengeName string) (*PlayerInfo, error) {

	player := &PlayerInfo{}

	//获取当前时间
	timeCur := uint32(time.Now().Unix())

	//1.首先从列表中找到这个擂台
	key := &ChallengeKey{
		index: int(index),
		level: int(level),
	}

	value, ok := self.mapAllMoneyChallenge[*key]
	if !ok {
		player.Code = proto.NoArena
		player.Uid = value.MoneyChallengeInfo.Base.GetHostid()
		return player, nil
	}
	//2.看看擂台是不是已经有人在挑战了
	if value.MoneyChallengeInfo.Base.GetState() == rpc.NormalChallengeInfo_Challenging {
		player.Code = proto.AlreadyChallengeing
		player.Uid = value.MoneyChallengeInfo.Base.GetHostid()
		return player, nil
	}

	//3.查看擂台的时间是不是正确
	if value.Base.GetStarttime()+uint32(common.GetChallengeInfoCfg("LT_DEFEND_LAST_TIME")) <= timeCur && value.Base.GetState() != rpc.NormalChallengeInfo_None {
		player.Code = proto.AlreadyTimeOut
		player.Uid = value.MoneyChallengeInfo.Base.GetHostid()
		return player, nil
	}

	//4.看看自己是不是在守擂
	if _, ok := self.mapPlayerId2Challenge[Challengeid]; ok {
		player.Code = proto.AlreadyHost
		player.Uid = value.MoneyChallengeInfo.Base.GetHostid()
		return player, nil
	}

	//5.看看自己是不是在挑战
	if _, ok := self.mapPlayerId2Challenging[Challengeid]; ok {
		player.Code = proto.AlreadyChallenger
		player.Uid = value.MoneyChallengeInfo.Base.GetHostid()
		return player, nil
	}

	//6.看看擂台是不是空的,是就设置擂台的擂主名字和id和擂台状态
	if value.MoneyChallengeInfo.Base.GetState() == rpc.NormalChallengeInfo_None {
		player.Code = proto.Empty
		player.Uid = value.MoneyChallengeInfo.Base.GetHostid()
		self.IsEmptyChallenge(rpc.ChallengeType_Money, int(level), int(index), true, ChallengeName, Challengeid)

		return player, nil
	}

	//7.如果都正确，就将返回值填写为ok,并把自己的id写进挑战列表
	p := &PlayerDefenseInfo{}

	p.etype = rpc.ChallengeType_Money
	p.key.level = int(value.Base.GetLevel())
	p.key.index = int(value.Base.GetIndex())

	pCenterChallengeService.mapPlayerId2Challenging[Challengeid] = p

	player.Code = proto.NormalChallengeOK
	player.Uid = value.MoneyChallengeInfo.Base.GetHostid()
	value.MoneyChallengeInfo.Base.SetChallengeid(Challengeid)
	value.MoneyChallengeInfo.SetMoney(value.MoneyChallengeInfo.GetMoney() + uint32(common.GetChallengeInfoCfg("LT_SPECIAL_EMPTY_GEM")))
	value.Base.SetChallengetimes(value.Base.GetChallengetimes() + 1)

	//注册超时tick
	value.tickdelay = timer.NewTimer(time.Duration((60 * common.GetChallengeInfoCfg("LT_BATTLE_TIME"))) * time.Second)
	value.tickdelay.Start(
		func() {
			pCenterChallengeService.OnTimeDelay(rpc.ChallengeType_Money, p.key.level, p.key.index, ChallengeName)
		},
	)

	//存数据库
	self.saveMoneyChallengeResult(int(value.Base.GetLevel()), int(value.Base.GetIndex()))

	return player, nil
}

//cns接口--赏金
func (self *Center) StartMoneyChallenge(req *proto.MoneyChallenger, ret *proto.MoneyChallengeResult) error {
	pCenterChallengeService.lock.Lock()
	defer pCenterChallengeService.lock.Unlock()

	player, err := pCenterChallengeService.BeginMoneyChallenge(req.Index, req.Level, req.Challengeid, req.ChallengeName)

	if err != nil {
		return err
	} else {
		ret.Code = player.Code
		ret.Hostid = player.Uid
	}
	return nil
}

//cns接口--普通
func (self *Center) StartNormalChallenge(req *proto.NormalChallenger, ret *proto.NormalChallengeResult) error {
	pCenterChallengeService.lock.Lock()
	defer pCenterChallengeService.lock.Unlock()

	player, err := pCenterChallengeService.BeginNormalChallenge(req.Index, req.Level, req.Challengeid, req.ChallengeName)

	if err != nil {
		return err
	} else {
		ret.Code = player.Code
		ret.Hostid = player.Uid
	}
	return nil
}

///////////////////////////////////////////
//cns接口
func (self *Center) GetNormalChallengeList(req *proto.GetChallengeList, rst *proto.GetChallengeListResult) error {
	challenges := pCenterChallengeService.RandomNormalChallenges(req.Uid, req.Level)
	if buff, err := gp.Marshal(challenges); err != nil {
		return err
	} else {
		rst.Value = buff
	}

	return nil
}

func (self *Center) GetMoneylChallengeList(req *proto.GetChallengeList, rst *proto.GetChallengeListResult) error {
	challenges := pCenterChallengeService.GetMoneyChallenges(req.Uid, req.Level)
	if buff, err := gp.Marshal(challenges); err != nil {
		return err
	} else {
		rst.Value = buff
	}

	return nil
}

func (self *Center) ChllengeEnd(req *proto.SendtoCenter, ret *proto.SendtoCenterResult) error {
	HostInfo := pCenterChallengeService.mapPlayerId2Challenge[req.Hostid]
	ChallengerInfo := pCenterChallengeService.mapPlayerId2Challenging[req.Challengeid]

	if ChallengerInfo == nil || HostInfo == nil || HostInfo.key != ChallengerInfo.key && HostInfo.etype != ChallengerInfo.etype {
		logger.Error("挑战者已经被删除或者信息不对，返回")
		return nil
	}

	pCenterChallengeService.bIsFinishedChallenge = req.IsFinished

	if pCenterChallengeService.bIsTimeEnd {
		pCenterChallengeService.OnTimeEnd(ChallengerInfo.etype, HostInfo.key.level, HostInfo.key.index, false)
		return nil
	}

	pCenterChallengeService.OnFightResult(ChallengerInfo.etype, HostInfo.key.level, HostInfo.key.index, req.IsSuccess, req.Hostid, req.Challengeid, req.Name)

	return nil
}

func (self *CenterChallengeService) NormalChallenge(etype rpc.ChallengeType, level int, index int, success bool, host, challenger, name string) {
	key := ChallengeKey{level: level, index: index}
	timeCur := uint32(time.Now().Unix())

	if etype == rpc.ChallengeType_Normal { //普通擂台
		if challenge, ok := self.mapAllNormalChallenge[key]; ok {
			//todo 这之前请先判断是否超时，防守与挑战者是否一致

			if challenge.tickdelay == nil {
				return
			}

			challenge.tickdelay.Stop()
			challenge.tickdelay = nil

			if challenge.NormalChallengeInfo.GetHostid() != host || challenge.NormalChallengeInfo.GetChallengeid() != challenger {
				logger.Error("普通防守与挑战者不一致", challenge.NormalChallengeInfo.GetHostid(), host, challenge.NormalChallengeInfo.GetChallengeid(), challenger)
				return
			}

			//todo 根据成功失败给奖励
			//挑战成功
			if success {

				var Keeper rpc.PlayerChallengeInfo
				exist, err := dbclient.KVQueryExt(common.ChallengeKeyName_PlayerChallengeInfo, host, &Keeper)

				//更新首擂玩家身上的数据
				if err == nil {
					if exist {
						Keeper.SetScore(Keeper.GetScore() + challenge.NormalChallengeInfo.GetScore() + ((timeCur-challenge.NormalChallengeInfo.GetStarttime())/60)*uint32(common.GetChallengeInfoCfg("LT_GET_MARK_PER_MIN")))
					}
				} else {
					Keeper.SetScore(challenge.NormalChallengeInfo.GetScore() + ((timeCur-challenge.NormalChallengeInfo.GetStarttime())/60)*uint32(common.GetChallengeInfoCfg("LT_GET_MARK_PER_MIN")))
					Keeper.SetSalarytime(0)
				}

				//更新挑战者身上的数据
				var playerChallenger rpc.PlayerChallengeInfo
				exist, myerr := dbclient.KVQueryExt(common.ChallengeKeyName_PlayerChallengeInfo, challenger, &playerChallenger)
				//更新挑战者玩家身上的数据
				if myerr == nil {
					if exist {
						playerChallenger.SetScore(playerChallenger.GetScore() + uint32(float64(challenge.NormalChallengeInfo.GetScore())*float64(common.GetChallengeInfoCfg("LT_GET_MARK_SCALE"))/100))
					}
				} else {
					playerChallenger.SetScore(uint32(float64(challenge.NormalChallengeInfo.GetScore()) * float64(common.GetChallengeInfoCfg("LT_GET_MARK_SCALE")) / 100))
					playerChallenger.SetSalarytime(0)
				}

				_, err = dbclient.KVWriteExt(common.ChallengeKeyName_PlayerChallengeInfo, challenger, &playerChallenger)

				if err != nil {
					logger.Error("写入守擂玩家到数据库是发生错误: ", err.Error())
				}

				//将擂台分数置0
				challenge.NormalChallengeInfo.SetScore(0)

				//更换守擂玩家id
				challenge.NormalChallengeInfo.SetHostid(challenger)

				//从守擂者列表删除原有擂主，从挑战者列表删除原有挑战者，将挑战者写入擂主列表
				if p, exist := self.mapPlayerId2Challenge[host]; exist {
					delete(self.mapPlayerId2Challenge, host)
					self.mapAllNormalChallenge[p.key].SetHostid(challenger)
					self.mapAllNormalChallenge[p.key].SetChallengetimes(0)
					self.mapAllNormalChallenge[p.key].SetState(rpc.NormalChallengeInfo_Free)
					self.mapAllNormalChallenge[p.key].SetHostname(name)
					challenge.NormalChallengeInfo.SetChallengeid("")
				}

				p := &PlayerDefenseInfo{}
				p.etype = rpc.ChallengeType_Normal
				p.key.index = index
				p.key.level = level
				self.mapPlayerId2Challenge[challenger] = p

				if _, exist := self.mapPlayerId2Challenging[challenger]; exist {
					delete(self.mapPlayerId2Challenging, challenger)
				}
			} else {
				//擂台累积分数+10
				challenge.NormalChallengeInfo.SetScore(challenge.NormalChallengeInfo.GetScore() + uint32(common.GetChallengeInfoCfg("LT_WIN_MARK")))

				var playerChallenger rpc.PlayerChallengeInfo
				exist, myerr := dbclient.KVQueryExt(common.ChallengeKeyName_PlayerChallengeInfo, challenger, &playerChallenger)
				//更新挑战者玩家身上的数据
				if myerr == nil {
					if exist {
						playerChallenger.SetScore(playerChallenger.GetScore() + uint32(common.GetChallengeInfoCfg("LT_LOSE_MARK")))
					}
				} else {
					playerChallenger.SetScore(uint32(common.GetChallengeInfoCfg("LT_LOSE_MARK")))
					playerChallenger.SetSalarytime(0)
				}

				_, err := dbclient.KVWriteExt(common.ChallengeKeyName_PlayerChallengeInfo, challenger, &playerChallenger)

				if err != nil {
					logger.Error("写入守擂玩家到数据库是发生错误: ", err.Error())
				}

				//更新擂台的积分
				challenge.NormalChallengeInfo.SetScore(challenge.NormalChallengeInfo.GetScore() + uint32(common.GetChallengeInfoCfg("LT_WIN_MARK")))
				challenge.NormalChallengeInfo.SetState(rpc.NormalChallengeInfo_Free)
				challenge.NormalChallengeInfo.SetChallengeid("")
				//从挑战者列表删除原有挑战者
				if _, exist := self.mapPlayerId2Challenging[challenger]; exist {
					delete(self.mapPlayerId2Challenging, challenger)
				}

			}
		}

	}
}

func (self *CenterChallengeService) MoneyChallenge(etype rpc.ChallengeType, level int, index int, success bool, host, challenger, name string) {
	key := ChallengeKey{level: level, index: index}
	timeCur := uint32(time.Now().Unix())

	if etype == rpc.ChallengeType_Money { //赏金擂台
		if challenge, ok := self.mapAllMoneyChallenge[key]; ok {
			if challenge.tickdelay == nil {
				return
			}

			//todo 这之前请先判断是否超时，防守与挑战者是否一致
			challenge.tickdelay.Stop()
			challenge.tickdelay = nil

			if challenge.MoneyChallengeInfo.Base.GetHostid() != host || challenge.MoneyChallengeInfo.Base.GetChallengeid() != challenger {
				return
			}

			//挑战成功
			if success {
				var Keeper rpc.PlayerChallengeInfo
				exist, err := dbclient.KVQueryExt(common.ChallengeKeyName_PlayerChallengeInfo, host, &Keeper)

				//更新首擂玩家身上的数据
				if err == nil {
					if exist {
						Keeper.SetScore(Keeper.GetScore() + challenge.MoneyChallengeInfo.Base.GetScore() + ((timeCur-challenge.MoneyChallengeInfo.Base.GetStarttime())/60)*uint32(common.GetChallengeInfoCfg("LT_SPECIAL_LOSE_MARK")))

					}
				} else {
					Keeper.SetScore(challenge.MoneyChallengeInfo.Base.GetScore() + ((timeCur-challenge.MoneyChallengeInfo.Base.GetStarttime())/60)*uint32(common.GetChallengeInfoCfg("LT_SPECIAL_LOSE_MARK")))
					Keeper.SetSalarytime(0)
				}

				if _, err = dbclient.KVWriteExt(common.ChallengeKeyName_PlayerChallengeInfo, host, &Keeper); err != nil {
					logger.Error("写入守擂玩家到数据库是发生错误: ", err.Error())
				}

				//更新挑战者身上的数据
				var playerChallenger rpc.PlayerChallengeInfo
				exist, myerr := dbclient.KVQueryExt(common.ChallengeKeyName_PlayerChallengeInfo, challenger, &playerChallenger)
				//更新挑战者玩家身上的数据
				if myerr == nil {
					if exist {
						playerChallenger.SetScore(playerChallenger.GetScore() + uint32(float64(challenge.Base.GetScore())*float64(common.GetChallengeInfoCfg("LT_GET_MARK_SCALE"))/100))
					}
				} else {
					playerChallenger.SetScore(uint32(float64(challenge.Base.GetScore()) * float64(common.GetChallengeInfoCfg("LT_GET_MARK_SCALE")) / 100))
					playerChallenger.SetSalarytime(0)
				}

				_, err = dbclient.KVWriteExt(common.ChallengeKeyName_PlayerChallengeInfo, challenger, &playerChallenger)

				if err != nil {
					logger.Error("写入守擂玩家到数据库是发生错误: ", err.Error())
				}

				req := &proto.SendMail{}
				ret := &proto.SendMailResult{}
				req.Uid = challenger
				req.Money = uint32(float64(challenge.MoneyChallengeInfo.GetMoney()) * float64(common.GetChallengeInfoCfg("LT_SPECIAL_SCALE")) / 100.0)

				//这里要告诉cn发邮件给玩家
				conn := centerServer.cnss[0]
				if err := conn.Call("CenterService.SendMailtoplayer", req, ret); err != nil {
					logger.Error("调用发送邮件错误")
				}
				//将擂台分数置0
				challenge.MoneyChallengeInfo.Base.SetScore(0)
				challenge.MoneyChallengeInfo.SetMoney(uint32((float64(challenge.MoneyChallengeInfo.GetMoney())) * (float64(common.GetChallengeInfoCfg("LT_SPECIAL_SCALE_LOOP")) / 100.0)))

				//更换守擂玩家id
				//从守擂者列表删除原有擂主，从挑战者列表删除原有挑战者，将挑战者写入擂主列表
				if _, exist := self.mapPlayerId2Challenge[host]; exist {
					delete(self.mapPlayerId2Challenge, host)

					challenge.Base.SetHostid(challenger)
					challenge.Base.SetChallengetimes(0)
					challenge.Base.SetHostname(name)
					challenge.Base.SetState(rpc.NormalChallengeInfo_Free)
					challenge.Base.SetChallengeid("")

					//将挑战者写入擂主列表
					p := &PlayerDefenseInfo{}
					p.etype = rpc.ChallengeType_Money
					p.key.index = index
					p.key.level = level
					self.mapPlayerId2Challenge[challenger] = p
				}

				if _, exist := self.mapPlayerId2Challenging[challenger]; exist {
					delete(self.mapPlayerId2Challenging, challenger)
				}

			} else {
				var Keeper rpc.PlayerChallengeInfo
				exist, err := dbclient.KVQueryExt(common.ChallengeKeyName_PlayerChallengeInfo, host, &Keeper)

				//更新擂台分数
				challenge.Base.SetScore(challenge.Base.GetScore() + uint32(common.GetChallengeInfoCfg("LT_WIN_MARK")))

				//发邮件给玩家发放宝石
				req := &proto.SendMail{}
				ret := &proto.SendMailResult{}
				req.Uid = host
				req.Money = uint32(common.GetChallengeInfoCfg("LT_SPECIAL_SCALE"))

				//这里要告诉cn发邮件给玩家
				conn := centerServer.cnss[0]
				if err := conn.Call("CenterService.SendMailtoplayer", req, ret); err != nil {
					logger.Error("调用发送邮件错误", err)
				}

				//更新挑战者身上的数据

				var playerChallenger rpc.PlayerChallengeInfo
				exist, myerr := dbclient.KVQueryExt(common.ChallengeKeyName_PlayerChallengeInfo, challenger, &playerChallenger)
				//更新挑战者玩家身上的数据
				if myerr == nil {
					if exist {
						playerChallenger.SetScore(playerChallenger.GetScore() + uint32(common.GetChallengeInfoCfg("LT_LOSE_MARK")))
					}
				} else {
					playerChallenger.SetScore(uint32(common.GetChallengeInfoCfg("LT_LOSE_MARK")))
					playerChallenger.SetSalarytime(0)
				}

				_, err = dbclient.KVWriteExt(common.ChallengeKeyName_PlayerChallengeInfo, challenger, &playerChallenger)

				if err != nil {
					logger.Error("写入守擂玩家到数据库是发生错误: ", err.Error())
				}

				//更新擂台的积分
				challenge.MoneyChallengeInfo.Base.SetScore(challenge.MoneyChallengeInfo.Base.GetScore() + uint32(common.GetChallengeInfoCfg("LT_WIN_MARK")))
				challenge.MoneyChallengeInfo.SetMoney(uint32((float64(challenge.MoneyChallengeInfo.GetMoney())) * (float64(common.GetChallengeInfoCfg("LT_SPECIAL_SCALE_LOOP")) / 100.0)))
				challenge.Base.SetState(rpc.NormalChallengeInfo_Free)
				challenge.Base.SetChallengeid("")

				//从挑战者列表删除原有挑战者，将挑战者写入擂主列表

				if _, exist := self.mapPlayerId2Challenging[challenger]; exist {
					delete(self.mapPlayerId2Challenging, challenger)
				}
			}
		}
	}
}

func (self *Center) PlayerReturnHome(req *proto.PlayerReturnHome, ret *proto.PlayerReturnHomeResult) error {
	if req.Uid == "" {
		logger.Error("Can't find player, playerId Is nil", nil)

		return nil
	}

	if _, ok := pCenterChallengeService.mapPlayerId2Challenging[req.Uid]; ok {
		delete(pCenterChallengeService.mapPlayerId2Challenging, req.Uid)

		return nil
	}

	return nil
}

func (self *Center) PlayerQuitChallenge(req *proto.PlayerReturnHome, ret *proto.PlayerReturnHomeResult) error {
	//从列表中找到擂主
	if req.Uid == "" {
		logger.Error("Can't find Player, PlayerID is nil", nil)
		return nil
	}

	if mykey, ok := pCenterChallengeService.mapPlayerId2Challenge[req.Uid]; ok {
		//如果在普通列表中，则发放普通擂台奖励
		if normalchallenge, myok := pCenterChallengeService.mapAllNormalChallenge[mykey.key]; myok {
			//看看是不是被攻打
			if normalchallenge.NormalChallengeInfo.GetState() == rpc.NormalChallengeInfo_Challenging {
				ret.Code = proto.AlreadyHasChallenger
				return nil
			}
			pCenterChallengeService.OnTimeEnd(mykey.etype, mykey.key.level, mykey.key.index, true)

			return nil

		} else {
			if moneychallenge, myok := pCenterChallengeService.mapAllMoneyChallenge[mykey.key]; myok {
				//看看是不是被攻打
				if moneychallenge.Base.GetState() == rpc.NormalChallengeInfo_Challenging {
					ret.Code = proto.AlreadyHasChallenger
					return nil
				}
				pCenterChallengeService.OnTimeEnd(mykey.etype, mykey.key.level, mykey.key.index, true)

				return nil
			}
		}
	}

	return nil
}

func (self *CenterChallengeService) IsInTheChallenge(Uid string) bool {
	pCenterChallengeService.lock.Lock()
	defer pCenterChallengeService.lock.Unlock()

	if _, exist := self.mapPlayerId2Challenge[Uid]; exist {
		return true
	}

	return false
}

//玩家查询积分
func (self *Center) GetPlayerScore(req *proto.GetPlayerScore, ret *proto.GetPlayerScoreResult) error {

	pCenterChallengeService.lock.Lock()
	defer pCenterChallengeService.lock.Unlock()

	var player rpc.PlayerChallengeInfo
	exist, myerr := dbclient.KVQueryExt(common.ChallengeKeyName_PlayerChallengeInfo, req.Uid, &player)

	if myerr == nil {
		if exist {
			ret.Score = player.GetScore()
			ret.Salarytime = player.GetSalarytime()
		}
	} else {
		logger.Error("Get player score error", myerr.Error())
		return myerr
	}

	return nil
}

//玩家领取工资
func (self *Center) WriteToDB(req *proto.GetDailyMoney, ret *proto.GetDailyMoneyResult) error {
	logger.Info("ComeInto WriteToDB")
	pCenterChallengeService.lock.Lock()
	defer pCenterChallengeService.lock.Unlock()

	curTime := time.Now().Unix()

	var player rpc.PlayerChallengeInfo
	exist, myerr := dbclient.KVQueryExt(common.ChallengeKeyName_PlayerChallengeInfo, req.Uid, &player)
	if myerr == nil {
		if exist {
			if !common.IsTheSameDay(player.GetSalarytime(), uint32(curTime)) {
				player.SetSalarytime(uint32(curTime))
				ret.Value = proto.GetMoneyOK
			} else {
				ret.Value = proto.AlreadyGetMoney
			}
		}
	}

	_, err := dbclient.KVWriteExt(common.ChallengeKeyName_PlayerChallengeInfo, req.Uid, &player)

	if err != nil {
		logger.Error("Write player Salarytime error", err.Error())

		return err
	}

	return nil

}
