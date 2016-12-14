package connector

import (
	"fmt"
	"golang-project/dpsg/common"
	"golang-project/dpsg/logger"
	"golang-project/dpsg/proto"
	"golang-project/dpsg/rpc"
	"strconv"

	gp "github.com/golang/protobuf/proto"
)

func (self *CNServer) BeginMoneyChallenge(conn rpc.RpcConn, msg rpc.BeginNormalChallenge) error {
	ts("CNServer:BeginMoneyChallenge", conn.GetId())
	defer te("CNServer:BeginMoneyChallenge", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()
	if !exist {
		return nil
	}

	try := &proto.MoneyChallenger{
		Index:           *msg.Index,
		Level:           *msg.Level,
		Money:           p.GetPlayerTotalGem(),
		Challengeid:     p.GetUid(),
		ChallengeName:   p.GetName(),
		ChallengeClan:   p.GetClan(),
		ChallengeSymbol: p.GetClanSymbol(),
	}
	ret := &proto.MoneyChallengeResult{}

	//这里判断是否够挑战的条件
	if *msg.Index < uint32(0) || *msg.Index > uint32(10) {
		return nil
	}

	if *msg.Level > uint32(9) || *msg.Level < uint32(5) {
		return nil
	}

	if p.GetPlayerTotalGem() < 100 {
		SendMsg(conn, "TID_LT_NO_ENOUGH_GEM")
		return nil
	}

	if *msg.Level != p.getCenterLevel() {
		return nil
	}

	//判断是否有兵可用
	var count uint32
	value := p.GetVillage().barrack_GetAllCharacters()
	for _, myvalue := range value {
		count = count + myvalue.GetCount()
	}

	if count <= 1 {
		SendMsg(conn, "TID_LT_NO_TROOP")
		return nil
	}

	//这里的调用函数需要重寻在center上写
	err := cns.center.Call("Center.StartMoneyChallenge", try, ret)
	if err != nil {
		logger.Error("调用Center.StartMoneyChallenge，函数错误")
		return nil
	}

	if ret.Code == proto.AlreadyChallengeing {
		logger.Error("挑战条件出错，需要告诉客户端--已经有人在挑战了")
		SendMsg(conn, "TID_LT_ALREADY_HAS")
		return nil
	}

	if ret.Code == proto.AlreadyTimeOut {
		logger.Error("挑战条件出错，需要告诉客户端--擂台超时，已经结束")
		SendMsg(conn, "TID_LT_DEFEND_OVER")
		return nil
	}

	if ret.Code == proto.NoArena {
		logger.Error("挑战条件出错，需要告诉客户端--没有这个擂台")
		//SendMsg(conn, "没有这个擂台")
		return nil
	}

	if ret.Code == proto.AlreadyHost {
		logger.Error("挑战条件出错， 需要告诉客户端--你已经是擂主了")
		SendMsg(conn, "TID_LT_ALREADY_IS")
		return nil
	}

	if ret.Code == proto.AlreadyChallenger {
		logger.Error("挑战条件出错， 需要告诉客户端--你已经在挑战别人了")
		SendMsg(conn, "TID_LT_ALREADY_ING")
		return nil
	}

	if ret.Code == proto.Empty {
		if !p.CostResource(100, proto.ResType_Gem, proto.Lose_Challenge) {
			return nil
		}

		self.updateplayerinfo(conn, p.GetPlayerTotalGem(), p.GetWuhun())
		self.GetMoneylChallengeList(conn, rpc.GetMoneylChallengeList{})
		return nil
	}

	if ret.Code == proto.NormalChallengeOK {
		//扣钱
		if !p.CostResource(100, proto.ResType_Gem, proto.Lose_Challenge) {
			return nil
		}

		self.updateplayerinfo(conn, p.GetPlayerTotalGem(), p.GetWuhun())
		//pvp
		self.Challenge(conn, ret.Hostid, p, rpc.MatchPlayer_CHALLENGE_M)
		self.GetMoneylChallengeList(conn, rpc.GetMoneylChallengeList{})
	}

	return nil
}

func (self *CNServer) BeginNormalChallenge(conn rpc.RpcConn, msg rpc.BeginMoneyChallenge) error {
	ts("CNServer:BeginMoneyChallenge", conn.GetId())
	defer te("CNServer:BeginMoneyChallenge", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()
	if !exist {
		return nil
	}

	try := &proto.NormalChallenger{
		Index:           *msg.Index,
		Level:           *msg.Level,
		Challengeid:     p.GetUid(),
		ChallengeName:   p.GetName(),
		ChallengeClan:   p.GetClan(),
		ChallengeSymbol: p.GetClanSymbol(),
	}
	ret := &proto.NormalChallengeResult{}

	//这里判断是否够挑战的条件
	if *msg.Index < uint32(0) || *msg.Index > uint32(200) {
		return nil
	}

	if *msg.Level > uint32(9) || *msg.Level < uint32(5) {
		return nil
	}

	if *msg.Level != p.getCenterLevel() {
		return nil
	}

	//判断是否有兵可用
	var count uint32
	value := p.GetVillage().barrack_GetAllCharacters()
	for _, myvalue := range value {
		count = count + myvalue.GetCount()
	}

	if count <= 1 {
		SendMsg(conn, "TID_LT_NO_TROOP")
		return nil
	}

	//这里的调用函数
	err := cns.center.Call("Center.StartNormalChallenge", try, ret)
	if err != nil {
		logger.Error("调用Center.StartNormalChallenge")
		return nil
	}

	if ret.Code == proto.AlreadyChallengeing {
		logger.Error("挑战条件出错，需要告诉客户端--已经有人在挑战了")
		SendMsg(conn, "TID_LT_ALREADY_ING")
		return nil
	}

	if ret.Code == proto.AlreadyTimeOut {
		logger.Error("挑战条件出错，需要告诉客户端--擂台超时，已经结束")
		SendMsg(conn, "TID_LT_DEFEND_OVER")
		return nil
	}

	if ret.Code == proto.NoArena {
		logger.Error("挑战条件出错，需要告诉客户端--没有这个擂台")
		//SendMsg(conn, "没有这个擂台")
		return nil
	}

	if ret.Code == proto.AlreadyHost {
		logger.Error("挑战条件出错，需要告诉客户端--你已经是擂主了，不能再挑战")
		SendMsg(conn, "TID_LT_ALREADY_IS")
		return nil
	}

	if ret.Code == proto.AlreadyChallenger {
		logger.Error("挑战条件出错， 需要告诉客户端--你已经在挑战别人了")
		SendMsg(conn, "TID_LT_ALREADY_ING")
		return nil
	}

	if ret.Code == proto.Empty {
		self.GetNormalChallengeList(conn, rpc.GetNormalChallengeList{})
		return nil
	}

	if ret.Code == proto.NormalChallengeOK {
		self.Challenge(conn, ret.Hostid, p, rpc.MatchPlayer_CHALLENGE_N)
		self.GetNormalChallengeList(conn, rpc.GetNormalChallengeList{})
	}

	return nil
}

//取得列表
func (self *CNServer) GetNormalChallengeList(conn rpc.RpcConn, info rpc.GetNormalChallengeList) error {
	ts("CNServer:GetNormalChallengeList", conn.GetId())
	defer te("CNServer:GetNormalChallengeList", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()
	if !exist {
		return nil
	}

	req := &proto.GetChallengeList{
		Uid:   p.GetUid(),
		Level: int(p.getCenterLevel()),
	}
	rst := &proto.GetChallengeListResult{}

	if err := self.center.Call("Center.GetNormalChallengeList", req, rst); err != nil {
		logger.Error("Center.GetNormalChallengeList err", err)
		return err
	}

	msg := &rpc.NormalChallenges{}
	if err := gp.Unmarshal(rst.Value, msg); err != nil {
		logger.Error("GetNormalChallengeList Unmarshal err", err)
		return err
	}

	WriteResult(conn, msg)

	return nil
}

func (self *CNServer) GetMoneylChallengeList(conn rpc.RpcConn, info rpc.GetMoneylChallengeList) error {
	ts("CNServer:GetMoneylChallengeList", conn.GetId())
	defer te("CNServer:GetMoneylChallengeList", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()
	if !exist {
		return nil
	}

	req := &proto.GetChallengeList{
		Uid:   p.GetUid(),
		Level: int(p.getCenterLevel()),
	}
	rst := &proto.GetChallengeListResult{}

	if err := self.center.Call("Center.GetMoneylChallengeList", req, rst); err != nil {
		logger.Error("Center.GetMoneylChallengeList err", err)
		return err
	}

	msg := &rpc.MoneyChallenges{}
	if err := gp.Unmarshal(rst.Value, msg); err != nil {
		logger.Error("GetMoneylChallengeList Unmarshal err", err)
		return err
	}

	WriteResult(conn, msg)

	return nil
}

//发邮件
func (self *CenterService) SendMailtoplayer(req *proto.SendMail, ret *proto.SendMailResult) error {

	mailreq := &proto.SendSystemMail{
		ToPlayerId: req.Uid,
		Title:      fmt.Sprintf("$$L:TID_LT_MAIL_TITLE$$"),
		Content:    fmt.Sprintf("$$L:TID_LT_MAIL_CONTENT$$"),
		Attach:     fmt.Sprintf("%d:%d", rpc.MailAttach_Gem, req.Money),
	}
	mailreqReselt := &proto.SendSystemMailResult{}

	cns.chatRpcConn.Go("ChatServices.SendSysMail2Player", mailreq, mailreqReselt, nil)

	return nil
}

func (self *CNServer) updateplayerinfo(conn rpc.RpcConn, money, wuhun uint32) {
	if conn != nil {
		update := &rpc.UpdatePlayerInfo{}
		update.SetDiamonds(money)
		update.SetWuhun(wuhun)
		WriteResult(conn, update)
	}
}

func (self *CNServer) QuitChallenge(conn rpc.RpcConn, info rpc.Ping) error {
	//这里要通知center发放奖励，删除擂主
	if conn != nil {
		self.l.RLock()
		p, exist := self.players[conn.GetId()]
		self.l.RUnlock()
		if !exist {
			return nil
		}
		req := &proto.PlayerReturnHome{Uid: p.GetUid()}
		ret := &proto.PlayerReturnHomeResult{}
		cns.center.Go("Center.PlayerQuitChallenge", req, ret, nil)

		if ret.Code == proto.AlreadyHasChallenger {
			SendMsg(conn, "TID_LT_BE_ATTACKING")
		}

		return nil
	}

	return nil
}

func (self *CNServer) GetDailyMoney(conn rpc.RpcConn, info rpc.Ping) error {

	logger.Info("ComeInto GetDailyMoney")

	//玩家领取爵位工资
	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()
	if !exist {
		return nil
	}

	//发信息去center上查询玩家分数
	req := &proto.GetPlayerScore{Uid: p.GetUid()}
	ret := &proto.GetPlayerScoreResult{}

	if err := self.center.Call("Center.GetPlayerScore", req, ret); err != nil {
		logger.Error("Center.GetPlayerScore err", err)
		return err
	}

	//根据玩家分数查询配置表,根据配置表发工资
	title := self.GetTitle(ret.Score)

	if title != nil {
		//通知center将领取时间写入数据库
		myreq := &proto.GetDailyMoney{Uid: p.GetUid()}
		myret := &proto.GetDailyMoneyResult{}
		if err := self.center.Call("Center.WriteToDB", myreq, myret); err != nil {
			logger.Error("Write player get daily money time error", err.Error())
			return err
		}

		if myret.Value == proto.AlreadyGetMoney {
			SendMsg(conn, "TID_LT_ALREADY_GET_GONGZI")

			return nil
		}

		//然后在给钱
		if myret.Value == proto.GetMoneyOK {
			p.GainResource(title.AwardCount1, proto.ResType_Gem, proto.Gain_Challenge)
			p.SetWuhun(p.GetWuhun() + title.AwardCount2)
			logger.Info("player title info is", title)
		}
		self.updateplayerinfo(conn, p.GetPlayerTotalGem(), p.GetWuhun())
	}

	return nil
}

//根据表来判断玩家的爵位

func (self *CNServer) GetTitle(Score uint32) *common.GlobalInfo {

	common.LoadDailyMoney()
	size := common.GetCfgSize()

	if size <= 0 {
		logger.Error("get config size error")
		return nil
	}

	title := "TID_LT_JUEWEI"

	//这里算出表所有的行数，再减去前面没用的两行属性，如果策划变动，这里的size一定要跟着变
	for i := size - 2; i >= 1; i-- {
		title = title + strconv.Itoa(i)
		info := common.GetTitleInfoCfg(title)
		readScore, _ := strconv.Atoi(info.Mark)
		if Score >= uint32(readScore) {
			break
		}
		return info
	}

	return nil
}

func (self *CNServer) GetPlayerSocre(conn rpc.RpcConn, info rpc.Ping) error {

	logger.Info("comeinto GetPlayerSocre")
	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()
	if !exist {
		return nil
	}

	//发信息去center上查询玩家分数

	req := &proto.GetPlayerScore{Uid: p.GetUid()}
	ret := &proto.GetPlayerScoreResult{}

	if err := self.center.Call("Center.GetPlayerScore", req, ret); err != nil {
		logger.Error("Center.GetPlayerScore err", err)
		return err
	}

	result := &rpc.PlayerChallengeInfo{Score: &ret.Score, Salarytime: &ret.Salarytime}

	WriteResult(conn, result)

	return nil
}
