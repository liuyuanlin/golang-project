package connector

import (
	"errors"
	"fmt"
	"golang-project/dpsg/language"
	"golang-project/dpsg/logger"
	"golang-project/dpsg/proto"
	"golang-project/dpsg/rpc"
	"strings"
)

func (self *CNServer) MoveTo(conn rpc.RpcConn, to rpc.MoveTo) (err error) {
	ts("CNServer:MoveTo", conn.GetId(), to)
	defer te("CNServer:MoveTo", conn.GetId(), to)
	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()
	if !exist {
		return
	}

	if !p.v.moveTo(to.GetId().GetType(), to.GetId().GetIndex(), to.GetP().GetX(), to.GetP().GetY()) {
		logger.Error("!!!!!!!MoveTo Failed!!!!!")
	}

	return
}

func (self *CNServer) MoveToBatch(conn rpc.RpcConn, to rpc.MoveToBatch) (err error) {
	ts("CNServer:MoveToBatch", conn.GetId(), to)
	defer te("CNServer:MoveToBatch", conn.GetId(), to)
	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()
	if !exist {
		return
	}

	if !p.v.moveToBatch(to) {
		logger.Error("!!!!!!!MoveToBatch Failed!!!!!")
	}

	return
}

func (self *CNServer) Create(conn rpc.RpcConn, to rpc.CreateTo) (err error) {
	ts("CNServer:Create", conn.GetId(), to)
	defer te("CNServer:Create", conn.GetId(), to)
	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return
	}

	if !p.v.create(to.GetId().GetType(), to.GetP().GetX(), to.GetP().GetY()) {
		logger.Error("!!!!!!!Create Failed!!!!!")
	}

	return
}

func (self *CNServer) Collect(conn rpc.RpcConn, id rpc.BuildingId) (err error) {
	ts("CNServer:Collect", conn.GetId(), id)
	defer te("CNServer:Collect", conn.GetId(), id)

	//debug.PrintStack()

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return
	}

	p.v.collect(id.GetType(), id.GetIndex())

	return
}

func (self *CNServer) Upgrade(conn rpc.RpcConn, id rpc.BuildingId) (err error) {
	ts("CNServer:Upgrade", conn.GetId(), id)
	defer te("CNServer:Upgrade", conn.GetId(), id)
	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return
	}

	p.v.upgrade(id.GetType(), id.GetIndex())

	return
}

func (self *CNServer) SellBuilding(conn rpc.RpcConn, sell rpc.SellBuilding) error {
	ts("CNServer:SellBuilding", conn.GetId())
	defer te("CNServer:SellBuilding", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()
	if !exist {
		return nil
	}

	p.SellBuilding(sell)

	return nil
}

func (self *CNServer) Remove(conn rpc.RpcConn, id rpc.BuildingId) error {
	ts("CNServer:Remove", conn.GetId(), id)
	defer te("CNServer:Remove", conn.GetId(), id)

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()
	if !exist {
		return nil
	}

	p.v.remove(id.GetType(), id.GetIndex())

	return nil
}

func (self *CNServer) Cancel(conn rpc.RpcConn, id rpc.BuildingId) (err error) {
	ts("CNServer:Cancel", conn.GetId(), id)
	defer te("CNServer:Cancel", conn.GetId(), id)
	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return
	}

	p.v.cancel(id.GetType(), id.GetIndex())

	return
}

func (self *CNServer) Training(conn rpc.RpcConn, t rpc.Training) error {
	ts("CNServer:Training %v %v", conn.GetId(), t, t.GetCharacter())
	defer te("CNServer:Training", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	p.v.barrack_CreateByType(t.GetIndex(), t.GetCharacter().GetType(), t.GetCharacter().GetCount())

	return nil
}

func (self *CNServer) CancelTraining(conn rpc.RpcConn, t rpc.CancelTraining) error {
	ts("CNServer:CancelTraining", conn.GetId())
	defer te("CNServer:CancelTraining", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	p.v.barrack_CancelCreate(t.GetIndex(), t.GetCharacter().GetType(), t.GetCharacter().GetCount())

	return nil
}

func (self *CNServer) TrainingSpell(conn rpc.RpcConn, t rpc.Spell) error {
	ts("CNServer:TrainingSpell %v %v", conn.GetId(), t)
	defer te("CNServer:TrainingSpell", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	p.v.spellForge_CreateByType(t.GetType(), t.GetCount())

	return nil
}

func (self *CNServer) CancelTrainingSpell(conn rpc.RpcConn, t rpc.Spell) error {
	ts("CNServer:CancelTrainingSpell", conn.GetId())
	defer te("CNServer:CancelTrainingSpell", conn.GetId())
	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()
	if !exist {
		return nil
	}
	p.v.spellForge_CancelCreate(t.GetType(), t.GetCount())

	return nil
}

func (self *CNServer) Barrack_FinishNow(conn rpc.RpcConn, id rpc.BuildingId) error {
	ts("CNServer:Barrack_FinishNow", conn.GetId())
	defer te("CNServer:Barrack_FinishNow", conn.GetId())
	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	p.v.barrack_FinishNow(id.GetIndex())

	return nil
}

func (self *CNServer) Buildings_FinishNow(conn rpc.RpcConn, id rpc.BuildingId) error {
	ts("CNServer:Buildings_FinishNow", conn.GetId())
	defer te("CNServer:Buildings_FinishNow", conn.GetId())
	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	p.v.buildings_FinishNow(id.GetType(), id.GetIndex())

	return nil
}

func (self *CNServer) SpellForge_FinishNow(conn rpc.RpcConn, id rpc.BuildingId) error {
	ts("CNServer:SpellForge_FinishNow", conn.GetId())
	defer te("CNServer:SpellForge_FinishNow", conn.GetId())
	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	p.v.spellForge_FinishNow()

	return nil
}

func (self *CNServer) Laboratory_FinishNow(conn rpc.RpcConn, id rpc.BuildingId) error {
	ts("CNServer:Laboratory_FinishNow", conn.GetId())
	defer te("CNServer:Laboratory_FinishNow", conn.GetId())
	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	p.v.laboratory_FinishNow()

	return nil
}

//丹药升级协议
func (self *CNServer) UpgradeSpellInfoLevel(conn rpc.RpcConn, s rpc.Spell) error {
	t := s.GetType()

	ts("CNServer:UpgradeSpellInfoLevel", conn.GetId())
	defer te("CNServer:UpgradeSpellInfoLevel", conn.GetId())
	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	p.v.laboratory_UpgradeSpellInfoLevel(t)

	return nil
}

func (self *CNServer) UpgradeInfoLevel(conn rpc.RpcConn, chr rpc.Character) error {
	t := chr.GetType()

	ts("CNServer:laboratory_UpgradeInfoLevel", conn.GetId(), t)
	defer te("CNServer:laboratory_UpgradeInfoLevel", conn.GetId(), t)
	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		logger.Info("UpgradeInfoLevel FFFuck!!")
		return nil
	}

	p.v.laboratory_UpgradeInfoLevel(t)

	return nil
}

func (self *CNServer) SetPlayerName(conn rpc.RpcConn, update rpc.UpdatePlayerInfo) error {
	ts("CNServer:SetPlayerName", conn.GetId(), update.GetName())
	defer te("CNServer:SetPlayerName", conn.GetId(), update.GetName())
	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	if strings.Contains(update.GetName(), " ") || len(update.GetName()) < 2 || len(update.GetName()) > 24 {
		SyncError(conn, "wrong name")
		return errors.New("wrong name")
	}

	//已经设置过的
	if p.GetName() != "" {
		update.SetName(p.GetName())
		WriteResult(conn, &update)
		return nil
	}

	//判断重名
	req := &proto.QueryName{
		Name:   update.GetName(),
		Id:     p.GetUid(),
		BQuery: false,
	}
	rst := &proto.QueryNameResult{Success: false}
	if err := self.center.Call("Center.CheckPlayerName", req, rst); err != nil || !rst.Success {
		SendMsg(conn, "TID_SAME_NAME")
		return err
	}

	p.SetName(update.GetName())
	WriteResult(conn, &update)

	return nil
}

//ttt积分榜发送
func (self *CenterService) SendTTTMail(req *proto.SendTTTSystemMail, rst *proto.SendTTTSystemMailResult) error {
	//fmt.Println(" CenterService  count   good  yuanbao", len(req.SenduidArray), req.Awardnum1, req.Awardnum2)
	//var number uint32 = 0
	for _, uid := range req.SenduidArray {
		req1 := &proto.SendSystemMail{
			ToPlayerId: uid,
			Title:      language.GetLanguage("TID_TTT_AWARD_MAIL_TITLE"),
			Content:    language.GetLanguage("TID_TTT_AWARD_MAIL_CONTENT"),
			Attach:     fmt.Sprintf("%d:%d,%d:%d", req.Awardtype1, req.Awardnum1, req.Awardtype2, req.Awardnum2),
		}
		rst1 := &proto.SendSystemMailResult{}
		err := cns.chatRpcConn.Go("ChatServices.SendSysMail2Player", req1, rst1, nil)
		if err != nil {
			fmt.Println(" 发放失败 ")
			rst.Success = false
		}
	}
	//fmt.Println(" 这个阶段奖励发放的人数 ")
	rst.Success = true
	return nil
}

//子弹填充
func (self *CNServer) Reloading(conn rpc.RpcConn, id rpc.BuildingId) (err error) {
	ts("CNServer:Reloading", conn.GetId(), id)
	defer te("CNServer:Reloading", conn.GetId(), id)

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return
	}

	p.v.Reloading(id.GetType(), id.GetIndex())

	return
}

//转换攻击模式
func (self *CNServer) ChangeMode(conn rpc.RpcConn, id rpc.BuildingId) (err error) {
	ts("CNServer:ChangeMode", conn.GetId(), id)
	defer te("CNServer:ChangeMode", conn.GetId(), id)

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return
	}

	p.v.ChangeMode(id.GetType(), id.GetIndex())

	return
}

//新手步骤
func (self *CNServer) SetGuideFinishedStep(conn rpc.RpcConn, msg rpc.GuideFinishedStep) (err error) {
	ts("CNServer:SetGuideFinishedStep", conn.GetId())
	defer te("CNServer:SetGuideFinishedStep", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return
	}

	p.SetGuideFinishedStep(msg.GetStepId())

	return nil
}
