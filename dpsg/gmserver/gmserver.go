package gmserver

import (
	"errors"
	"golang-project/dpsg/common"
	"golang-project/dpsg/dbclient"
	"golang-project/dpsg/logger"
	"golang-project/dpsg/proto"
	"golang-project/dpsg/rpc"
	"golang-project/dpsg/rpcplus"
	"net"
	"strconv"
	"strings"
)

type GmService struct {
	chat   *rpcplus.Client
	center *rpcplus.Client
}

var uGmServerId uint8 = 255
var pGmService *GmService

func CreateGmServer() {
	//chat
	var chatcfg common.ChatServerCfg
	if err := common.ReadChatConfig(&chatcfg); err != nil {
		return
	}
	chatconn, err := net.Dial("tcp", chatcfg.ListenForGm)
	if err != nil {
		logger.Fatal("%s", err.Error())
	}

	//center
	var centercfg common.CenterConfig
	if err := common.ReadCenterConfig(&centercfg); err != nil {
		return
	}
	centerconn, err := net.Dial("tcp", centercfg.CenterForGm)
	if err != nil {
		logger.Fatal("%s", err.Error())
	}

	//数据库服务
	dbclient.Init()

	var gmcfg common.GmServerCfg
	if err := common.ReadGmConfig(&gmcfg); err != nil {
		return
	}

	pGmService = &GmService{
		chat:   rpcplus.NewClient(chatconn),
		center: rpcplus.NewClient(centerconn),
	}

	//监听
	listener, err := net.Listen("tcp", gmcfg.GmServerIp)
	defer listener.Close()
	if err != nil {
		println("Listening to: ", gmcfg.GmServerIp, " failed !!")
		return
	}

	rpcServer := rpcplus.NewServer()
	rpcServer.Register(pGmService)
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("chatserver StartServices %s", err.Error())
			break
		}
		go func() {
			rpcServer.ServeConn(conn)
			conn.Close()
		}()
	}
}

//附件有效性判断
func checkAttach(info string) error {
	attachs := strings.Split(info, ",")
	for _, attach := range attachs {
		if attach == "" {
			continue
		}

		typeandnum := strings.Split(attach, ":")

		if len(typeandnum) < 2 {
			logger.Info("checkAttach:%s", attach)
			return errors.New("wrong typeandnum of attach")
		}

		attachtype, err := strconv.Atoi(typeandnum[0])
		if err != nil {
			return err
		}
		attachnum, err := strconv.Atoi(typeandnum[1])
		if err != nil {
			return err
		}

		if attachnum <= 0 {
			return errors.New("wrong attach num")
		}
		//由于角色那边的类型没有终止值就不判断了
		if (attachtype > 0 && attachtype < int(rpc.MailAttach_NorEnd)) || (attachtype > int(rpc.MailAttach_CharBegin)) {
		} else {
			return errors.New("wrong attach type")
		}
	}

	return nil
}

func (self *GmService) GmSendMail(req *proto.GmSendMail, rst *proto.GmSendMailResult) error {
	if err := checkAttach(req.Attach); err != nil {
		return err
	}

	logger.Info("GmSendMail:%+v", req)

	return self.chat.Call("ChatGmServices.GmSendMail", req, rst)
}

func (self *GmService) GmSendAllMail(req *proto.GmSendAllMail, rst *proto.GmSendAllMailResult) error {
	if err := checkAttach(req.Attach); err != nil {
		return err
	}

	return self.chat.Call("ChatGmServices.GmSendAllMail", req, rst)
}

//发送通知
func (self *GmService) GmSendNotice(req *proto.GmSendNotice, rst *proto.GmSendNoticeResult) error {
	return self.chat.Call("ChatGmServices.GmSendNotice", req, rst)
}

//锁定玩家
func (self *GmService) GmLockPlayer(req *proto.GmLockPlayer, rst *proto.GmLockPlayerResult) error {
	logger.Info("GmLockPlayer:%s", req.Uid)

	lid := common.GenLockMessage(uGmServerId, proto.MethodPlayerGmOpera, 0)

	try := &proto.TryGetLock{Service: "player", Name: req.Uid, Value: lid}
	after := &proto.GetLockResult{}

	if err := self.center.Call("CenterGmServices.GmLockPlayer", try, after); err != nil {
		logger.Error("Gm Error On LockGet : %s", err.Error())
		return err
	}

	rst.Success = after.Result
	rst.OldValue = after.OldValue

	return nil
}

//解锁玩家
func (self *GmService) GmUnLockPlayer(req *proto.GmUnLockPlayer, rst *proto.GmUnLockPlayerResult) error {
	logger.Info("GmUnLockPlayer:%s", req.Uid)

	try := &proto.FreeLock{Service: "player", Name: req.Uid, Value: 0}
	after := &proto.FreeLockResult{}

	if err := self.center.Call("CenterGmServices.GmUnLockPlayer", try, after); err != nil {
		logger.Error("Gm Error On LockFree : %s", err.Error())
		return err
	}

	rst.Success = after.Result
	return nil
}

func (self *GmService) GmGetPlayerInfo(req *proto.GmLockPlayer, rst *proto.GmPlayerInfo) error {
	//lock := proto.GmLockPlayerResult{}
	//err := self.GmLockPlayer(req, &lock)
	//if err != nil {
	//	return err
	//}

	//if !lock.Success {
	//	return errors.New("Lock fail")
	//}
	logger.Info("GmGetPlayerInfo:%d:%s", len(req.Uid), req.Uid)

	var base rpc.PlayerBaseInfo
	var extra rpc.PlayerExtraInfo

	exists, err := dbclient.KVQueryBase(common.PlayerBase, req.Uid, &base)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("GmGetPlayerInfo no player")
	}

	exists, err = dbclient.KVQueryExt(common.PlayerExtra, req.Uid, &extra)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("GmGetPlayerInfo no player extra")
	}

	rst.Uid = base.GetUid()
	rst.Name = base.GetName()   //用户名称
	rst.Clan = base.GetClan()   //所属联盟
	rst.Level = base.GetLevel() //用户等级

	rst.Wuhun = extra.GetWuhun()                                    //武魂
	rst.Trophy = base.GetTrophy()                                   //令牌
	rst.Diamonds = extra.GetDiamonds()                              //宝石数
	rst.DrillTimes = extra.GetDrillTimes()                          //演习次数
	rst.LastLogin = extra.GetLandedrewardinfo().GetLastLandedTime() //最后登录

	var v rpc.VillageInfo
	exists, err = dbclient.KVQueryExt("village", strconv.FormatUint(base.GetVillageId(), 16), &v)
	if err != nil {
		return err
	}

	if !exists {
		return nil
	}

	rst.CenterLevel = v.GetCenter().GetLevel() //主营等级

	for _, i := range v.Foodstorage {
		rst.Food += *i.StorageFood
	}
	rst.Food += *v.Center.StorageFood //食物粮草

	for _, i := range v.Goldstorage {
		rst.Gold += *i.StorageGold
	}
	rst.Gold += *v.Center.StorageGold //银币？

	return nil
}

func (self *GmService) GmSetPlayerInfo(req *proto.GmPlayerInfo, rst *proto.GmUnLockPlayerResult) error {
	logger.Info("GmSetPlayerInfo:%s", req.Uid)

	var base rpc.PlayerBaseInfo
	var extra rpc.PlayerExtraInfo

	exists, err := dbclient.KVQueryBase(common.PlayerBase, req.Uid, &base)
	if err != nil {
		return err
	}

	if !exists {
		return errors.New("GmSetPlayerInfo no player")
	}

	exists, err = dbclient.KVQueryExt(common.PlayerExtra, req.Uid, &extra)
	if err != nil {
		return err
	}

	if !exists {
		return errors.New("GmSetPlayerInfo no player extra")
	}

	if req.Name != "" && req.Name != *base.Name {
		base.SetName(req.Name)
	}
	if req.Clan != "" && req.Name != *base.Clan {
		base.SetClan(req.Clan)
	}
	if req.Level != 0 && req.Level != *base.Level {
		base.SetLevel(req.Level)
	}
	if req.Wuhun != 0 && req.Wuhun != *extra.Wuhun {
		extra.SetWuhun(req.Wuhun)
	}
	if req.Level != 0 && req.Trophy != *base.Trophy {
		base.SetTrophy(req.Trophy)
	}

	if req.Diamonds != 0 && req.Diamonds != *extra.Diamonds {
		extra.SetDiamonds(req.Diamonds)
	}

	if req.DrillTimes != 0 && req.DrillTimes != *extra.DrillTimes {
		extra.SetDrillTimes(req.DrillTimes)
	}

	if req.Wuhun != 0 && req.Wuhun != *extra.Wuhun {
		extra.SetWuhun(req.Wuhun)
	}

	_, err = dbclient.KVWriteBase(common.PlayerBase, req.Uid, &base)
	if err != nil {
		return err
	}

	_, err = dbclient.KVWriteExt(common.PlayerExtra, req.Uid, &extra)
	if err != nil {
		return err
	}

	var v rpc.VillageInfo
	exists, err = dbclient.KVQueryExt("village", strconv.FormatUint(base.GetVillageId(), 16), &v)
	if err != nil {
		return err
	}

	if !exists {
		return nil
	}

	if req.CenterLevel != 0 && req.CenterLevel != *v.Center.Level {
		v.Center.SetLevel(req.CenterLevel)
	}

	_, err = dbclient.KVWriteExt("village", strconv.FormatUint(base.GetVillageId(), 16), &v)

	if err != nil {
		return err
	}
	//unlock := &proto.GmUnLockPlayer {Uid : p.GetUid()}
	//self.GmUnLockPlayer(unlock, rst)

	return nil
}

//gm取在线玩家数量
func (self *GmService) GmGetOnlineNum(req *proto.GmGetOnlineNum, rst *proto.GmGetOnlineNumResult) error {
	return self.center.Call("CenterGmServices.GmGetOnlineNum", req, rst)
}
