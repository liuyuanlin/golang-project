package connector

import (
	"fmt"
	"golang-project/dpsg/accountclient"
	"golang-project/dpsg/common"
	"golang-project/dpsg/csvcfg"
	"golang-project/dpsg/dbclient"
	"golang-project/dpsg/lockclient"
	"golang-project/dpsg/logger"
	"golang-project/dpsg/proto"
	"golang-project/dpsg/rpc"
	"golang-project/dpsg/rpcplus"
	"net"
	"os"
	"path"
	"runtime/debug"
	"sync"
	"time"
)

var cns *CNServer

var buildingCfg map[string]*[]BuildingCfg
var townhallLevelsCfg map[uint32]*[]TownhallLevels
var charactorCfg map[string]*[]CharacterCfg
var spellCfg map[string]*[]SpellCfg
var taskCfg map[string]*[]TaskCfg
var expCfg map[uint32]*[]ExpCfg
var globalCfg map[string]*[]GlobalCfg
var shareawardCfg map[uint32]*[]ShareAwardCfg
var landawardCfg map[uint32]*[]LandAwardCfg
var tttCfg map[string]*[]TttCfg
var tttbuffCfg map[string]*[]TttBuffCfg
var Cfg common.CnsConfig
var gPVEStageCfg map[string]*[]PVEStageCfg

type CenterService struct {
}

type CNServer struct {
	serverForClient *rpc.Server
	center          *rpcplus.Client
	gateserver      *rpcplus.Client
	FsMgr           FServerConnMgr
	logRpcConn      *rpcplus.Client
	chatRpcConn     *rpcplus.Client
	players         map[uint64]*player
	otherplayers    map[uint64]*player
	playersbyid     map[string]*player
	centerService   *CenterService
	exit            chan bool
	l               sync.RWMutex
	id              uint8
	listenIp        string
	rankMgr         *RankMgr
	listener        net.Listener
}

func (self *CNServer) GetServerId() uint8 {
	return self.id
}

func loadConfigFiles(cfgDir string) {
	pvecfg := path.Join(cfgDir, "misson.csv")
	csvcfg.LoadCSVConfig(pvecfg, &gPVEStageCfg)

	building := path.Join(cfgDir, "buildings.csv")
	csvcfg.LoadCSVConfig(building, &buildingCfg)

	townhall_levels := path.Join(cfgDir, "townhall_levels.csv")
	csvcfg.LoadCSVConfig(townhall_levels, &townhallLevelsCfg)

	charactors := path.Join(cfgDir, "characters.csv")
	csvcfg.LoadCSVConfig(charactors, &charactorCfg)

	spells := path.Join(cfgDir, "spells.csv")
	csvcfg.LoadCSVConfig(spells, &spellCfg)

	tasks := path.Join(cfgDir, "task.csv")
	csvcfg.LoadCSVConfig(tasks, &taskCfg)

	exps := path.Join(cfgDir, "experience_levels.csv")
	csvcfg.LoadCSVConfig(exps, &expCfg)

	globals := path.Join(cfgDir, "globals.csv")
	csvcfg.LoadCSVConfig(globals, &globalCfg)

	pathShareAward := path.Join(cfgDir, "shareaward.csv")
	csvcfg.LoadCSVConfig(pathShareAward, &shareawardCfg)

	pathLandAward := path.Join(cfgDir, "landaward.csv")
	csvcfg.LoadCSVConfig(pathLandAward, &landawardCfg)

	ttts := path.Join(cfgDir, "ttt.csv")
	csvcfg.LoadCSVConfig(ttts, &tttCfg)

	tttbuffs := path.Join(cfgDir, "tttbuffs.csv")
	csvcfg.LoadCSVConfig(tttbuffs, &tttbuffCfg)

	initConfig()
}

func (self *CNServer) Quit() {
	self.listener.Close()
	self.FsMgr.Quit()
	fmt.Println("Fs Quit Over!!!!!!")
	self.serverForClient.Quit()
}

func (self *CNServer) EndService() {
	self.center.Close()
	self.gateserver.Close()
	self.logRpcConn.Close()
	self.chatRpcConn.Close()
	cns = nil
}

func (self *CNServer) StartClientService(cfg *common.CnsConfig, wg *sync.WaitGroup) {

	rpcServer := rpc.NewServer()
	self.serverForClient = rpcServer

	lockclient.Init()
	accountclient.Init()

	rpcServer.Register(cns)
	rpcServer.RegCallBackOnConn(
		func(conn rpc.RpcConn) {
			self.onConn(conn)
		},
	)

	rpcServer.RegCallBackOnDisConn(
		func(conn rpc.RpcConn) {
			self.onDisConn(conn)
		},
	)

	rpcServer.RegCallBackOnCallBefore(
		func(conn rpc.RpcConn) {
			conn.Lock()
		},
	)

	rpcServer.RegCallBackOnCallAfter(
		func(conn rpc.RpcConn) {
			conn.Unlock()
		},
	)

	//开始对fightserver的RPC服务
	self.FsMgr.Init(rpcServer, cfg)
	listener, err := net.Listen("tcp", Cfg.CnsHost)
	if err != nil {
		logger.Fatal("net.Listen: %s", err.Error())
	}

	self.listener = listener
	self.listenIp = cfg.CnsHostForClient

	self.sendPlayerCountToGateServer()

	wg.Add(1) //监听client要算一个
	go func() {
		for {
			//For Client/////////////////////////////
			time.Sleep(time.Millisecond * 5)
			conn, err := self.listener.Accept()

			if err != nil {
				logger.Error("cns StartServices %s", err.Error())
				wg.Done() // 退出监听就要减去一个
				break
			}

			wg.Add(1) // 这里是给客户端增加计数
			go func() {
				rpcConn := rpc.NewProtoBufConn(rpcServer, conn, 128, 45)
				defer func() {
					if r := recover(); r != nil {
						logger.Error("player rpc runtime error begin:", r)

						rpcConn.Unlock()
						debug.PrintStack()
						self.onDisConn(rpcConn)
						rpcConn.Close()

						logger.Error("player rpc runtime error end ")
					}
					wg.Done() // 客户端退出减去计数
				}()

				rpcServer.ServeConn(rpcConn)
			}()
		}
	}()
}

func StartCenterService(self *CNServer, listener net.Listener, cfg *common.CnsConfig) {
	//连接center
	rpcCenterServer := rpcplus.NewServer()
	rpcCenterServer.Register(self.centerService)

	req := &proto.CenterConnCns{Addr: listener.Addr().String()}
	rst := &proto.CenterConnCnsResult{}
	self.center.Go("Center.CenterConnCns", req, rst, nil)

	connCenter, err := listener.Accept()
	if err != nil {
		logger.Error("StartCenterServices %s", err.Error())
		os.Exit(0)
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("StartCenterService runtime error:", r)

				debug.PrintStack()
			}
		}()
		rpcCenterServer.ServeConn(connCenter)
		connCenter.Close()
	}()

}

func NewCNServer(cfg *common.CnsConfig) (server *CNServer) {
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

	var centerCfg common.CenterConfig
	if err := common.ReadCenterConfig(&centerCfg); err != nil {
		logger.Fatal("%v", err)
	}
	conn, err := net.Dial("tcp", centerCfg.CenterHost)
	if err != nil {
		logger.Fatal("%s", err.Error())
	}

	var gscfg common.GateServerCfg
	if err = common.ReadGateConfig(&gscfg); err != nil {
		return
	}
	gsConn, err := net.Dial("tcp", gscfg.GsIpForServer)
	if err != nil {
		logger.Fatal("%s", err.Error())
	}

	var chatcfg common.ChatServerCfg
	if err = common.ReadChatConfig(&chatcfg); err != nil {
		return
	}
	chatConn, err := net.Dial("tcp", chatcfg.ListenForServer)
	if err != nil {
		logger.Fatal("connect chatserver failed %s", err.Error())
	}

	server = &CNServer{
		center:        rpcplus.NewClient(conn),
		gateserver:    rpcplus.NewClient(gsConn),
		logRpcConn:    rpcplus.NewClient(logConn),
		players:       make(map[uint64]*player),
		otherplayers:  make(map[uint64]*player),
		playersbyid:   make(map[string]*player),
		centerService: &CenterService{},
		chatRpcConn:   rpcplus.NewClient(chatConn),
		rankMgr:       CreateRankMgr()}

	cns = server

	loadConfigFiles(common.GetDesignerDir())

	return
}

func (self *CNServer) sendPlayerCountToGateServer() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("sendPlayerCountToGateServer runtime error:", r)

				debug.PrintStack()
			}
		}()

		for {

			time.Sleep(5 * time.Second)

			self.l.RLock()
			playerCount := uint32(len(self.players))
			self.l.RUnlock()

			var ret proto.SendCnsInfoResult

			err := self.gateserver.Call("GateServices.UpdateCnsPlayerCount", proto.SendCnsInfo{999, uint16(playerCount), self.listenIp}, &ret)

			if err != nil {
				logger.Error("Error On GateServices.UpdateCnsPlayerCount : %s", err.Error())
				return
			}

		}

	}()
}

func (c *CNServer) onConn(conn rpc.RpcConn) {
}

func (self *CNServer) onDisConn(conn rpc.RpcConn) {
	ts("CNServer:onDisConn", conn.GetId())
	defer te("CNServer:onDisConn", conn.GetId())

	self.delPlayer(conn.GetId())
	self.delOtherPlayer(conn.GetId())
}

func (self *CNServer) Ping(conn rpc.RpcConn, login rpc.Ping) error {
	//ts("CNServer:Ping", conn.GetId())
	//defer te("CNServer:Ping", conn.GetId())

	rep := rpc.PingResult{}
	rep.SetServerTime(uint32(time.Now().Unix()))

	WriteResult(conn, &rep)
	return nil
}

//添加玩家到全局表中
func (self *CNServer) addPlayer(connId uint64, p *player) {
	ts("CNServer:addPlayer", connId, p.GetUid())
	defer te("CNServer:addPlayer", connId, p.GetUid())

	self.l.Lock()
	defer self.l.Unlock()

	//进入服务器全局表
	self.players[connId] = p
	self.playersbyid[p.GetUid()] = p
}

//添加被攻击玩家到全局表中
func (self *CNServer) addOtherPlayer(connId uint64, p *player) {
	ts("CNServer:addOtherPlayer", connId, p.GetUid())
	defer te("CNServer:addOtherPlayer", connId, p.GetUid())

	self.l.Lock()
	defer self.l.Unlock()

	//进入服务器全局表
	self.otherplayers[connId] = p
	self.playersbyid[p.GetUid()] = p
}

//销毁玩家
func (self *CNServer) delPlayer(connId uint64) {
	ts("CNServer:delPlayer", connId)
	defer te("CNServer:delPlayer", connId)

	p, exist := self.players[connId]
	if exist {
		p.OnQuit()

		self.l.Lock()
		delete(self.players, connId)
		delete(self.playersbyid, p.GetUid())
		self.l.Unlock()
	}
}

//销毁被攻击的玩家
func (self *CNServer) delOtherPlayer(connId uint64) {
	ts("CNServer:delOtherPlayer", connId)
	defer te("CNServer:delOtherPlayer", connId)

	self.l.Lock()
	defer self.l.Unlock()

	if p, exist := self.otherplayers[connId]; exist {
		p.OnQuit()

		delete(self.otherplayers, connId)
		delete(self.playersbyid, p.GetUid())
	}
}
