package center

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"golang-project/dpsg/common"
	"golang-project/dpsg/dbclient"
	"golang-project/dpsg/lockclient"
	"golang-project/dpsg/logger"
	"golang-project/dpsg/proto"
	"golang-project/dpsg/rpc"
	"golang-project/dpsg/rpcplus"
	"golang-project/dpsg/timer"
	"net"
	"runtime/debug"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const TRACE = true

func ts(name string, items ...interface{}) {
	if TRACE {
		logger.Info("+%s %v\n", name, items)
	}
}
func te(name string, items ...interface{}) {
	if TRACE {
		logger.Info("-%s %v\n", name, items)
	}
}

var nid uint32 = 0

func GenLockMessage(sid uint8, tid uint8, value uint8) uint64 {

	tmpid := uint8(atomic.AddUint32(&nid, 1))

	return uint64(time.Now().Unix()) | uint64(tmpid)<<32 | uint64(value)<<40 | uint64(tid)<<48 | uint64(sid)<<56
}

func LOG_Resources(gl rpc.GameLocation, uid string, gain bool, ResType string, ResNum uint32, ResWay uint32) bool {
	msg := proto.LogResources{
		ChannelId: uint8(gl),
		Uid:       uid,
		Gain:      gain,
		Time:      time.Now().Unix(),
		ResType:   ResType,
		ResNum:    ResNum,
		ResWay:    ResWay,
	}

	var ret proto.LogResourcesResult
	centerServer.lgs.Go("LogServices.LogResources", msg, &ret, nil)

	return true
}

var cnsConnId uint32 = 0

type Center struct {
	lgs          *rpcplus.Client
	cnss         []*rpcplus.Client
	l            sync.RWMutex
	maincache    *common.CachePool
	clancache    *common.CachePool
	clans        map[string]*clan
	clanrank     []*clan
	shields      map[string]*timer.Timer
	POnline      map[string]bool
	eveyrdaytime *timer.Timer
	updatetime   *timer.Timer //add for update rankplayers
}

var centerServer *Center

func StartServices(self *Center, listener net.Listener) {
	rpcServer := rpcplus.NewServer()

	rpcServer.Register(self)

	rpcServer.HandleHTTP("/center/rpc", "/debug/rpc")

	lockclient.Init()

	//ttt
	self.initDayTick()

	//add for save rankplayers
	self.initUpdateRankPlayers()

	//擂台赛
	startChallengeService()

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("Center StartServices %s", err.Error())
			break
		}
		go func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Info("center runtime error: %s", r)
					debug.PrintStack()
				}
			}()

			logger.Info("Center: OnCns Connected")
			rpcServer.ServeConn(conn)

			logger.Info("Center: OnCns DisConnected")
			conn.Close()
		}()
	}
}

func NewCenterServer(cfg common.CenterConfig) (server *Center) {
	//数据库服务
	dbclient.Init()

	var logCfg common.LogServerCfg
	if err := common.ReadLogConfig(&logCfg); err != nil {
		logger.Fatal("%v", err)
	}
	logConn, err := net.Dial("tcp", logCfg.LogHost)
	if err != nil {
		logger.Fatal("connect logserver failed %s", err.Error())
	}

	server = &Center{
		lgs:      rpcplus.NewClient(logConn),
		cnss:     make([]*rpcplus.Client, 0, 1),
		clans:    make(map[string]*clan),
		clanrank: make([]*clan, 0, 1),
		shields:  make(map[string]*timer.Timer),
		POnline:  make(map[string]bool),
	}

	//初始化cache
	logger.Info("Init Cache %v", cfg.MainCacheProfile)
	server.maincache = common.NewCachePool(cfg.MainCacheProfile)

	logger.Info("Init Cache %v", cfg.ClanCacheProfile)
	server.clancache = common.NewCachePool(cfg.ClanCacheProfile)

	server.initClans()

	centerServer = server

	//重名检测
	StartCenterNameService()

	return server
}

func (self *Center) CenterConnCns(req *proto.CenterConnCns, reply *proto.CenterConnCnsResult) (err error) {
	logger.Info("Center:CenterConnCns:%s", req.Addr)

	conn, err := net.Dial("tcp", req.Addr)
	if err != nil {
		logger.Fatal("%s", err.Error())
		reply.Ret = false
		return
	}

	tmp := rpcplus.NewClient(conn)
	self.l.Lock()
	self.cnss = append(self.cnss, tmp)
	self.theFirstUpdate(tmp)
	self.l.Unlock()
	reply.Ret = true

	return nil
}

func (self *Center) KickCnsPlayer(req *proto.LoginKickPlayer, rst *proto.LoginKickPlayerResult) error {
	rst.Success = false

	for _, rpccli := range self.cnss {
		if err := rpccli.Call("CenterService.LoginKickPlayer", req, rst); err == nil && rst.Success {
			rst.Success = true
			return nil
		}
	}

	/*//在登陆状态才会调到这里，都失败了就强制解锁
	ok, err := lockclient.ForceUnLock("player", req.Id)
	if err != nil {
		return err
	}

	rst.Success = ok*/

	return nil
}

type TaobaoPayDispatch struct {
}

type TaobaoPayResult struct {
	TradeEnd    bool
	TradeError  string
	TradeNumber string
	CharId      string
	ItemName    string
	GemNum      uint32
	TotoalPee   string
}

type GivePlayerGemResult struct {
	Result string
}

func (self *TaobaoPayDispatch) OnPayResult(cmd *TaobaoPayResult, conn *common.PpeConn) {
	fmt.Println("center : ------------------> OnPayResult !!!! ", cmd.TradeEnd, cmd.TradeError)

	msg := proto.TaobaoPayLog{TradeNumber: cmd.TradeNumber,
		CharId:    cmd.CharId,
		TotoalPee: cmd.TotoalPee,
		TradeTime: time.Now().Unix(),
		ChannelId: uint32(rpc.GameLocation_Vietnam)}

	if cmd.TradeEnd {
		var numGem uint32

		if cmd.GemNum > 0 {
			numGem = cmd.GemNum
			msg.ItemName = "direct buy " + strconv.Itoa(int(numGem))
		} else if len(cmd.ItemName) > 0 {
			fmt.Sscanf(cmd.ItemName, "GEM_PRICE_%d", &numGem)
			msg.ItemName = cmd.ItemName
		}

		ret := GivePlayerGem(cmd.CharId, uint32(numGem), true)

		cmd := &GivePlayerGemResult{}
		if ret == nil {
			cmd.Result = "success"
			msg.TradeEnd = true
			msg.TradeError = ""
		} else {
			cmd.Result = "failed"
			msg.TradeEnd = false
			msg.TradeError = ret.Error()
		}

		fmt.Println("GivePlayerGem Result : ", cmd.Result, msg.TradeError)

		data, _ := json.Marshal(cmd)

		msgId := uint16(1)
		datalen := uint16(len(data))

		sendCmd := bytes.Buffer{}
		binary.Write(&sendCmd, binary.LittleEndian, msgId)
		binary.Write(&sendCmd, binary.LittleEndian, datalen)
		sendCmd.Write(data)
		conn.Send(sendCmd.Bytes())

		fmt.Println("send data to taobaoserver : ", sendCmd.String())
	} else {
		msg.TradeEnd = false
		msg.TradeError = cmd.TradeError
	}

	var ret proto.TaobaoPayLogResult
	centerServer.lgs.Go("LogServices.LogTaobaoPayResult", msg, &ret, nil)
	//
	conn.ShutDown()

	fmt.Println("center ------ OnPayResult end -------")
}

var taobaoPayDispatch TaobaoPayDispatch

func StartTaobaoServices(host string) error {
	logger.Info("Center:StartTaobaoServices:%s", host)

	listener, err := net.Listen("tcp", host)
	if err != nil {
		logger.Fatal("net.Listen: %s", err.Error())
		return err
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				logger.Error("Center StartTaobaoServices %s", err.Error())
				break
			}
			common.CreatePpeConn(conn, &taobaoPayDispatch)
		}
	}()

	return nil
}

func GivePlayerGem(uid string, num uint32, bNotify bool) (err error) {
	lid := GenLockMessage(0, proto.MethodPlayerGiveGem, 0)

	successed, _, err := lockclient.TryLock("player", uid, lid)

	if err != nil {
		return err
	}

	if successed {
		logger.Info("Center:GivePlayerGem:%s, %d", uid, num)

		defer lockclient.TryUnlock("player", uid, lid)

		var base rpc.PlayerBaseInfo
		var p rpc.PlayerExtraInfo

		exist, err := dbclient.KVQueryBase(common.PlayerBase, uid, &base)
		if err != nil {
			return err
		}

		if exist {
			exist, err = dbclient.KVQueryExt(common.PlayerExtra, uid, &p)
			if err != nil {
				return err
			}
		}

		if exist {
			p.SetDiamonds(p.GetDiamonds() + num)

			LOG_Resources(base.GetGamelocation(), uid, true, proto.ResType_Gem, num, proto.Gain_Recharge)

			_, err = dbclient.KVWriteBase(common.PlayerBase, uid, &p)

			logger.Info("Center:GivePlayerGem success:%s, %d", uid, num)
		}

	} else {
		if bNotify {
			logger.Info("Center:NotifyGivePlayerGem:%s, %d", uid, num)

			req := &proto.NotifyGivePlayerGem{Uid: uid, Num: num}
			reply := &proto.NotifyGivePlayerGemResult{}

			ok := false

			for _, conn := range centerServer.cnss {
				conn.Call("CenterService.NotifyGivePlayerGem", req, reply)

				if reply.Ok {
					ok = true

					logger.Info("Center:NotifyGivePlayerGem success:%s, %d", uid, num)

					break
				}
			}

			if !ok {
				time.Sleep(time.Second)

				return GivePlayerGem(uid, num, false)
			}
		} else { //异常
			LOG_Resources(rpc.GameLocation_InvaildChannel, uid, true, proto.ResType_Gem, num, proto.Exception_RechargeFailed)

			logger.Info("RechargeFailed! Player<%s> not found!", uid)

			return errors.New("RechargeFailed! Player not found!")
		}
	}

	return nil
}
