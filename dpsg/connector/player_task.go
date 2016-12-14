package connector

import (
	"fmt"
	"golang-project/dpsg/common"
	"golang-project/dpsg/logger"
	"golang-project/dpsg/proto"
	"golang-project/dpsg/rpc"
	"time"
)

func (p *player) GetTask(name string) *rpc.Task {
	for _, task := range p.Tasks {
		if task.GetName() == name {
			return task
		}
	}

	return nil
}

func (p *player) UpdateTaskInfo(info *rpc.UpdateTaskInfo) {
	task := p.GetTask(info.GetName())

	if task == nil {
		task = &rpc.Task{}
		task.SetName(info.GetName())
		task.SetProgress(info.GetProgress())

		p.Tasks = append(p.Tasks, task)
	} else {
		//add by wyc 2014-2-12 add task finish check
		cfg := GetTaskCfg(task.GetName())
		if cfg == nil { //没有配置表
			logger.Error("GetTaskReward cfg(%s) == nil", task.GetName())
			return
		}
		if cfg.IsDayTask {
			if common.IsTheSameDay(uint32(time.Now().Unix()), task.GetFinishedTime()) {
				return
			}
		} else {
			if task.GetFinishedTime() != 0 {
				return
			}
		}

		task.SetProgress(info.GetProgress())
		if task.GetProgress() == 0 {
			task.SetFinishedTime(0)
		}
	}

}

func (p *player) GetTaskReward(tryget *rpc.TryGetTaskReward) bool {
	task := p.GetTask(tryget.GetName())
	if task == nil { //没有完成该任务
		logger.Error("GetTaskReward Failed, task(%s) == nil", tryget.GetName())
		return false
	}

	cfg := GetTaskCfg(tryget.GetName())
	if cfg == nil { //没有配置表
		logger.Error("GetTaskReward cfg(%s) == nil", tryget.GetName())

		return false
	}

	//没有完成该任务
	if task.GetProgress() < cfg.Progress {
		logger.Error("GetTaskReward Failed(%s), task.GetProgress(%d) < cfg.Progress(%d)", tryget.GetName(), task.GetProgress(), cfg.Progress)

		return false
	}

	needSpace := int32(0)

	if cfg.TroopType1 > 0 && cfg.TroopCount1 > 0 {
		charCfg1 := GetCharacterCfgByTypeId(rpc.CharacterType(cfg.TroopType1), 1)
		needSpace += int32(charCfg1.HousingSpace * cfg.TroopCount1)
	}

	if cfg.TroopType2 > 0 && cfg.TroopCount2 > 0 {
		charCfg2 := GetCharacterCfgByTypeId(rpc.CharacterType(cfg.TroopType2), 1)
		needSpace += int32(charCfg2.HousingSpace * cfg.TroopCount2)
	}

	if needSpace > p.v.barrack_GetTroopHousingTotalFreeSpaces() {
		logger.Error("GetTaskReward(%s):needSpace(%d),free(%d)", tryget.GetName(), needSpace, p.v.barrack_GetTroopHousingTotalFreeSpaces())

		return false
	}

	logger.Info("GetTaskReward successed:%s, Gold:%d, Food:%d, Gem:%d, Exp:%d", tryget.GetName(), cfg.Gold, cfg.Food, cfg.Gem, cfg.Exp)

	if cfg.Gold > 0 {
		p.GainResource(cfg.Gold, proto.ResType_Gold, proto.Gain_Task)
	}

	if cfg.Food > 0 {
		p.GainResource(cfg.Food, proto.ResType_Food, proto.Gain_Task)
	}

	if cfg.Gem > 0 {
		p.GainResource(cfg.Gem, proto.ResType_Gem, proto.Gain_Task)
	}

	if cfg.Exp > 0 {
		p.AddExp(cfg.Exp)
	}

	if needSpace > 0 {
		logger.Info("GetTaskReward Char Successed:%s, Type1:%d x %d, Type2:%d x %d", tryget.GetName(), cfg.TroopType1, cfg.TroopCount1, cfg.TroopType2, cfg.TroopCount2)
	}

	if cfg.TroopType1 > 0 && cfg.TroopCount1 > 0 {
		for i := uint32(0); i < cfg.TroopCount1; i++ {
			p.v.barrack_TroopHousingPushCharacter(rpc.CharacterType(cfg.TroopType1))
		}
	}

	if cfg.TroopType2 > 0 && cfg.TroopCount2 > 0 {
		for i := uint32(0); i < cfg.TroopCount2; i++ {
			p.v.barrack_TroopHousingPushCharacter(rpc.CharacterType(cfg.TroopType2))
		}
	}

	task.SetFinishedTime(uint32(time.Now().Unix()))

	return true
}

//分享
func (p *player) shareFinish(conn rpc.RpcConn, share *rpc.ShareFinish) bool {
	curTime := time.Now().Unix()

	info := p.GetShareinfo()
	if !common.IsTheSameDay(uint32(time.Now().Unix()), uint32(info.GetSharetime())) {
		if share.GetStep() != 1 {
			logger.Error("shareFinish Failed : not begin with 1")
			return false
		}
	} else {
		if share.GetStep() != info.GetStep()+1 {
			logger.Error("shareFinish Failed : wrong step(%d/%d)", info.GetStep(), share.GetStep())
			return false
		}
	}

	awardcfg := GetShareAwardCfg(share.GetStep())
	if awardcfg == nil {
		logger.Error("shareFinish Failed : no award(%d)", share.GetStep())
		return false
	}

	info.SetStep(share.GetStep())
	info.SetSharetime(curTime)
	p.SetShareinfo(info)

	//奖励邮件
	req := &proto.SendSystemMail{
		ToPlayerId: p.GetUid(),
		Title:      fmt.Sprintf("$$L:TID_SHARE_08$$"),
		Content:    fmt.Sprintf("$$L:TID_SHARE_09$$"),
		Attach:     fmt.Sprintf("%d:%d", rpc.MailAttach_Gem, awardcfg.GiveGem),
	}
	rst := &proto.SendSystemMailResult{}
	cns.chatRpcConn.Go("ChatServices.SendSysMail2Player", req, rst, nil)

	common.WriteResult(conn, info)

	return true
}

//连续登陆奖励
func (p *player) LandedReceiveAward(conn rpc.RpcConn, LandedReceiveAward *rpc.LandedReceiveAward) bool {
	//取得当前需要领取的奖励的天数，利用天数来对应数字
	//这里服务器直接加宝石就可以了
	info := p.GetLandedrewardinfo()
	if info.GetLandCount() != LandedReceiveAward.GetCurrentCount() {
		logger.Error("Landreward Failed : wrong curCount(%d)  and sendCount (%d)", info.GetLandCount(), LandedReceiveAward.GetCurrentCount())
		return false
	}
	curCount := LandedReceiveAward.GetCurrentCount()
	if curCount >= 7 {
		curCount = 7
	}
	awardcfg := GetLandAwardCfg(curCount)
	if awardcfg == nil {
		logger.Error("Landreward Failed : no award(%d)", LandedReceiveAward.GetCurrentCount())
		return false
	}
	//这里判断一天之内不可以再次领取
	if info.GetLastEndGameTime() != 0 {
		if common.IsTheSameDay(uint32(time.Now().Unix()), uint32(info.GetLastEndGameTime())) {
			return false
		}
	}
	info.SetNeedGetReward(false)
	preCount := p.GetPlayerTotalGem()
	p.GainResource(awardcfg.AddDiamondCount, proto.ResType_Gem, proto.Gain_LandReward)
	fmt.Println("当前连续天数   领取前的宝石   要奖励的宝石 领取后的宝石  ", curCount, preCount, awardcfg.AddDiamondCount, p.GetPlayerTotalGem())

	WriteResult(conn, info)
	return true
}
