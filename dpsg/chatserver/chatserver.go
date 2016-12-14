package chatserver

import (
	"fmt"
	"golang-project/dpsg/logger"

	gp "github.com/golang/protobuf/proto"
	//	"math/rand"
	"golang-project/dpsg/proto"
	"golang-project/dpsg/rpc"
	"golang-project/dpsg/rpcplus"
	"net"
	//	"strconv"
	"container/list"
	"golang-project/dpsg/common"
	"golang-project/dpsg/timer"
	"sync"
	"time"
)

type Alliance struct {
	ConnMap map[string]uint64
	Donate  *list.List
	ChatLog *list.List
	Symbol  uint32
}

type PlayerInfo struct {
	ConnId        uint64
	PlayerName    string
	AllianceName  string
	Level         uint8
	AlliancePower uint32
	ChannelId     rpc.GameLocation
}

type ChatServices struct {
	alliance  map[string]*Alliance
	PlayerMap map[string]*PlayerInfo
	l         sync.RWMutex
}

type candiinfo struct {
	CreateTime int64
	Connid     uint64
}

type ChatServicesForClient struct {
	rpcServer *rpc.Server
	candiMap  map[string]candiinfo
	l         sync.RWMutex
	t         *timer.Timer
}

var pChatServices *ChatServices
var pChatServicesForClient *ChatServicesForClient

func (self *ChatServices) AddPlayer(msg *proto.AddPlayer, result *proto.AddPlayerResult) error {

	AllianceName := msg.AllianceName
	PlayerId := msg.PlayerId
	AuthKey := msg.AuthKey
	Level := msg.PlayerLevel

	//1. check key
	ret, connid := pChatServicesForClient.CheckKey(AuthKey)
	if !ret {
		return nil
	}

	self.l.Lock()
	defer self.l.Unlock()
	//2, 开始把玩家信息放入各种表里面
	self.PlayerMap[PlayerId] = &PlayerInfo{ConnId: connid,
		PlayerName:    msg.PlayerName,
		AllianceName:  AllianceName,
		Level:         Level,
		AlliancePower: msg.AlliancePower,
		ChannelId:     msg.ChannelId}

	//3. 联盟处理
	if len(AllianceName) > 0 {
		an, ok := self.alliance[AllianceName]
		if ok {
			self.alliance[AllianceName].ConnMap[PlayerId] = connid
		} else {
			an = &Alliance{}
			an.ConnMap = make(map[string]uint64)
			an.ConnMap[PlayerId] = connid
			an.Donate = list.New()
			an.ChatLog = list.New()
			self.alliance[AllianceName] = an
		}

		conn := pChatServicesForClient.rpcServer.GetConn(connid)
		if conn == nil {
			return nil
		}

		// 如果联盟里面有卷兵信息，则要发给客户端
		if an.Donate.Len() > 0 {
			for iter := an.Donate.Front(); iter != nil; iter = iter.Next() {
				logger.Info("AddPlayer - DonateInfo:%v", *iter.Value.(*rpc.S2CDonateInfo))

				WriteResult(conn, iter.Value.(*rpc.S2CDonateInfo).GetDonateInfo())

				//如果有这个人捐兵的信息，那么发过去
				for _, ret := range iter.Value.(*rpc.S2CDonateInfo).DonateResults {
					if PlayerId == ret.GetSrcUid() {
						WriteResult(conn, ret)
					}
				}
			}
		}

		for iter := an.ChatLog.Front(); iter != nil; iter = iter.Next() {
			WriteResult(conn, iter.Value.(*rpc.ClanChatMessage))
		}
	}

	//邮件初始化
	go self.initPlayerMail(PlayerId, msg.ChannelId)

	return nil
}

func (self *ChatServices) DelPlayer(msg *proto.DelPlayer, result *proto.DelPlayerResult) error {

	PlayerId := msg.PlayerId

	self.l.Lock()
	defer self.l.Unlock()

	info, ok := self.PlayerMap[PlayerId]

	if !ok {
		return nil
	}

	AllianceName := info.AllianceName

	//2.处理联盟表
	if len(AllianceName) > 0 {
		_, ok := self.alliance[AllianceName]
		if ok {
			delete(self.alliance[AllianceName].ConnMap, PlayerId)
		}
	}

	//1.处理全局表
	delete(self.PlayerMap, PlayerId)

	return nil

}

func (self *ChatServices) UpdatePlayer(msg *proto.UpdatePlayer, result *proto.UpdatePlayerResult) error {
	PlayerId := msg.PlayerId

	self.l.Lock()
	defer self.l.Unlock()
	//1, 开始把玩家信息放入各种表里面
	playerInfo, exist := self.PlayerMap[PlayerId]
	if !exist {
		logger.Error("UpdatePlayer Error:PlayerInfo(%s) Not Exist!", PlayerId)
		return nil
	}

	//2. 加入联盟处理
	if len(playerInfo.AllianceName) == 0 && len(msg.AllianceName) > 0 {
		an, ok := self.alliance[msg.AllianceName]
		if ok {
			an.ConnMap[PlayerId] = playerInfo.ConnId
		} else {
			an = &Alliance{}
			an.ConnMap = make(map[string]uint64)
			an.ConnMap[PlayerId] = playerInfo.ConnId
			an.Donate = list.New()
			an.ChatLog = list.New()
			self.alliance[msg.AllianceName] = an
		}

		conn := pChatServicesForClient.rpcServer.GetConn(playerInfo.ConnId)
		if conn == nil {
			return nil
		}
		// 如果联盟里面有卷兵信息，则要发给客户端
		if an.Donate.Len() > 0 {
			for iter := an.Donate.Front(); iter != nil; iter = iter.Next() {
				logger.Info("UpdatePlayer - DonateInfo:%v", *iter.Value.(*rpc.S2CDonateInfo))

				WriteResult(conn, iter.Value.(*rpc.S2CDonateInfo).GetDonateInfo())

				//如果有这个人捐兵的信息，那么发过去
				for _, ret := range iter.Value.(*rpc.S2CDonateInfo).DonateResults {
					if PlayerId == ret.GetSrcUid() {
						WriteResult(conn, ret)
					}
				}
			}
		}

		for iter := an.ChatLog.Front(); iter != nil; iter = iter.Next() {
			WriteResult(conn, iter.Value.(*rpc.ClanChatMessage))
		}

		logger.Info("UpdatePlayer when player(%s) join clan(%s)", PlayerId, msg.AllianceName)
	} else if len(playerInfo.AllianceName) > 0 && len(msg.AllianceName) == 0 {
		//3. 离开联盟处理
		an, ok := self.alliance[playerInfo.AllianceName]
		if !ok {
			logger.Error("UpdatePlayer Error:Alliance(%s) Not Exist!", playerInfo.AllianceName)
			return nil
		}

		_, exist := an.ConnMap[PlayerId]
		if !exist {
			logger.Error("UpdatePlayer Error:PlayerId(%s) Not Exist!", PlayerId)
			return nil
		}

		delete(an.ConnMap, PlayerId)

		logger.Info("UpdatePlayer when player(%s) leave clan(%s)", PlayerId, playerInfo.AllianceName)
	}

	//更新玩家信息
	playerInfo.Level = msg.PlayerLevel
	playerInfo.AllianceName = msg.AllianceName
	playerInfo.AlliancePower = msg.AlliancePower

	return nil
}

func (self *ChatServices) PlayerChatToPlayer(msg *proto.PlayerChatToPlayer, result *proto.PlayerChatToPlayerResult) error {

	result = &proto.PlayerChatToPlayerResult{}
	self.l.RLock()
	defer self.l.RUnlock()

	fromplayerid := msg.FromPlayerId
	toplayerid := msg.ToPlayerId

	toInfo, ok := self.PlayerMap[toplayerid]

	if !ok {
		return nil
	}

	fromInfo, ok := self.PlayerMap[fromplayerid]
	if !ok {
		return nil
	}

	connId := toInfo.ConnId

	conn := pChatServicesForClient.rpcServer.GetConn(connId)

	if conn == nil {
		return nil
	}

	cmd := &rpc.S2CChatP2P{}

	cmd.SetFromPlayerId(fromplayerid)
	cmd.SetFromPlayerName(fromInfo.PlayerName)
	cmd.SetFromPlayerLevel(uint32(fromInfo.Level))
	cmd.SetChatContent(msg.Content)

	WriteResult(conn, cmd)

	return nil
}

func (self *ChatServices) PlayerWorldChat(msg *proto.PlayerWorldChat, result *proto.PlayerWorldChatResult) error {
	self.l.RLock()
	defer self.l.RUnlock()

	fromplayerid := msg.FromPlayerId

	fromInfo, ok := self.PlayerMap[fromplayerid]
	if !ok {
		return nil
	}

	cmd := &rpc.S2CChatWorld{}
	cmd.SetFromPlayerId(fromplayerid)
	cmd.SetFromPlayerName(fromInfo.PlayerName)
	cmd.SetFromPlayerLevel(uint32(fromInfo.Level))
	cmd.SetChatContent(msg.Content)
	cmd.SetChatTime(time.Now().Unix())
	if msg.CName != "" {
		cmd.SetAllianceName(msg.CName)
		cmd.SetAllianceSymbol(msg.CSymbol)
	}

	/*
		allianceName := fromInfo.AllianceName
		alliance, ok := self.alliance[allianceName]
		if ok {
			cmd.SetAllianceName(allianceName)
			cmd.SetAllianceSymbol(alliance.Symbol)
		}
	*/
	pChatServicesForClient.rpcServer.ServerBroadcast(cmd)
	return nil
}

func (self *ChatServices) CastClanMsg(req *proto.ClanMsg, result *proto.ClanMsgResult) error {
	self.l.RLock()
	defer self.l.RUnlock()

	allianceName := req.CName
	/*
		fromInfo, ok := self.PlayerMap[uid]
		if !ok {
			logger.Error("CastClanMsg:Player Not Exist:%s", uid)
			return nil
		}

		allianceName := fromInfo.AllianceName
	*/
	alliance, ok := self.alliance[allianceName]
	if !ok {
		logger.Error("CastClanMsg:Alliance Not Exist:%s", allianceName)
		return nil
	}

	msg := &rpc.ClanChatMessage{}
	err := gp.Unmarshal(req.Value, msg)
	if err != nil {
		logger.Error("CastClanMsg Unmarshal Error: %s (%v)", err.Error(), msg)
		return err
	}

	for _, connid := range alliance.ConnMap {
		conn := pChatServicesForClient.rpcServer.GetConn(connid)

		if conn != nil {
			WriteResult(conn, msg)
		}
	}

	alliance.ChatLog.PushBack(msg)

	if alliance.ChatLog.Len() > 50 {
		iter := alliance.ChatLog.Front()
		alliance.ChatLog.Remove(iter)
	}

	//pChatServicesForClient.rpcServer.ServerBroadcast(cmd)
	return nil
}

func (self *ChatServices) SendDonateMsg(req *proto.DonateMsg, result *proto.DonateMsgResult) error {
	self.l.RLock()
	defer self.l.RUnlock()

	uid := req.Uid

	fromInfo, ok := self.PlayerMap[uid]
	if !ok {
		return nil
	}

	allianceName := fromInfo.AllianceName

	alliance, ok := self.alliance[allianceName]
	if !ok {
		return nil
	}

	if alliance.Donate.Len() >= 50 {
		iter := alliance.Donate.Front()
		alliance.Donate.Remove(iter)
	}

	s2cDonateInfo := &rpc.S2CDonateInfo{}
	s2cDonateInfo.DonateInfo = &rpc.S2CDonate{}
	err := gp.Unmarshal(req.Value, s2cDonateInfo.DonateInfo)
	if err != nil {
		logger.Error("SendDonateMsg Unmarshal Error: %s (%v)", err.Error(), s2cDonateInfo)
		return err
	}

	logger.Info("SendDonateMsg:%v", *s2cDonateInfo)

	for iter := alliance.Donate.Front(); iter != nil; iter = iter.Next() {
		if iter.Value.(*rpc.S2CDonateInfo).GetDonateInfo().GetUid() == s2cDonateInfo.GetDonateInfo().GetUid() {
			alliance.Donate.Remove(iter)
			break
		}
	}

	alliance.Donate.PushBack(s2cDonateInfo)

	for _, connid := range alliance.ConnMap {
		conn := pChatServicesForClient.rpcServer.GetConn(connid)
		if conn != nil {
			WriteResult(conn, s2cDonateInfo.GetDonateInfo())
		}
	}

	return nil

}

func (self *ChatServices) UpdateDonateMsg(req *proto.DonateMsg, result *proto.DonateMsgResult) error {
	logger.Info("ChatServices:UpdateDonateMsg")

	self.l.RLock()
	defer self.l.RUnlock()

	uid := req.Uid

	fromInfo, ok := self.PlayerMap[uid]
	if !ok {
		return nil
	}

	allianceName := fromInfo.AllianceName

	alliance, ok := self.alliance[allianceName]
	if !ok {
		return nil
	}

	s2cDonateUpdate := &rpc.S2CDonateUpdate{}
	err := gp.Unmarshal(req.Value, s2cDonateUpdate)
	if err != nil {
		logger.Error("UpdateDonateMsg Unmarshal Error: %s (%v)", err.Error(), s2cDonateUpdate)
		return err
	}

	//logger.Info("UpdateDonateMsg111:%v", *s2cDonateUpdate)

	for iter := alliance.Donate.Front(); iter != nil; iter = iter.Next() {
		s2cDonateInfo := iter.Value.(*rpc.S2CDonateInfo).GetDonateInfo()

		if s2cDonateInfo.GetUid() == s2cDonateUpdate.GetUid() {
			//直接发
			for _, connid := range alliance.ConnMap {
				conn := pChatServicesForClient.rpcServer.GetConn(connid)
				if conn != nil {
					WriteResult(conn, s2cDonateUpdate)
				}
			}

			//更新数据
			s2cDonateInfo.SetInfo(s2cDonateUpdate.GetInfo())
			s2cDonateInfo.SetUsedSpace(s2cDonateUpdate.GetUsedSpace())

			//logger.Info("UpdateDonateMsg222:%v", *(iter.Value.(*rpc.S2CDonateInfo)))

			//更新结果数据
			for _, donateRet := range iter.Value.(*rpc.S2CDonateInfo).DonateResults {
				if donateRet.GetSrcUid() == s2cDonateUpdate.DonateResult.GetSrcUid() {
					donateRet.SetExp(s2cDonateUpdate.DonateResult.GetExp())
					donateRet.Characters = make([]*rpc.Character, 0, 1)
					for _, c := range s2cDonateUpdate.DonateResult.Characters {
						donateRet.Characters = append(donateRet.Characters, c)
					}

					connid, exist := alliance.ConnMap[donateRet.GetSrcUid()]
					if exist {
						conn := pChatServicesForClient.rpcServer.GetConn(connid)
						//反馈捐兵玩家信息
						if conn != nil {
							WriteResult(conn, s2cDonateUpdate.DonateResult)
						}
					}

					//logger.Info("UpdateDonateMsg333:%v", *(iter.Value.(*rpc.S2CDonateInfo)))

					return nil
				}
			}

			iter.Value.(*rpc.S2CDonateInfo).DonateResults = append(iter.Value.(*rpc.S2CDonateInfo).DonateResults, s2cDonateUpdate.DonateResult)

			connid, exist := alliance.ConnMap[s2cDonateUpdate.DonateResult.GetSrcUid()]
			if exist {
				conn := pChatServicesForClient.rpcServer.GetConn(connid)
				//反馈捐兵玩家信息
				if conn != nil {
					WriteResult(conn, s2cDonateUpdate.DonateResult)
				}
			}
			//logger.Info("UpdateDonateMsg444:%v", *(iter.Value.(*rpc.S2CDonateInfo)))

			return nil
		}
	}

	return nil
}

func CreateChatServicesForCnserver(listener net.Listener) *ChatServices {
	pChatServices = &ChatServices{}
	pChatServices.alliance = make(map[string]*Alliance)
	pChatServices.PlayerMap = make(map[string]*PlayerInfo)

	rpcServer := rpcplus.NewServer()

	rpcServer.Register(pChatServices)

	//rpcServer.HandleHTTP("/center/rpc", "/debug/rpcdebug/rpc")

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

	return pChatServices
}

func CreateChatServicesForClient(listener net.Listener) *ChatServicesForClient {

	pChatServicesForClient = &ChatServicesForClient{rpcServer: rpc.NewServer(), candiMap: make(map[string]candiinfo)}

	pChatServicesForClient.rpcServer.Register(pChatServicesForClient)

	pChatServicesForClient.rpcServer.RegCallBackOnConn(
		func(conn rpc.RpcConn) {
			pChatServicesForClient.onConn(conn)
		},
	)

	//30秒清理一次过期的candimap
	pChatServicesForClient.t = timer.NewTimer(time.Second * 30)
	pChatServicesForClient.t.Start(
		func() {
			pChatServicesForClient.Cleanup()
		},
	)

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("chatserver StartServices %s", err.Error())
			break
		}
		go func() {
			rpcConn := rpc.NewProtoBufConn(pChatServicesForClient.rpcServer, conn, 4, 0)
			pChatServicesForClient.rpcServer.ServeConn(rpcConn)
			rpcConn.Close()
		}()
	}

	return pChatServicesForClient
}

func WriteResult(conn rpc.RpcConn, value interface{}) bool {
	return common.WriteResult(conn, value)
}

// 客户端登陆后给一个key给客户端
func (self *ChatServicesForClient) onConn(conn rpc.RpcConn) {
	rep := rpc.LoginChatServerInfo{}
	curtime := time.Now().Unix()
	authkey := fmt.Sprintf("%d%s%d%s", conn.GetId(), conn.GetRemoteIp(), curtime, "dibutdswds")

	// encode
	encodeInfo := string(common.Base64Encode([]byte(authkey)))

	rep.SetAuthKey(encodeInfo)

	self.l.Lock()
	self.candiMap[encodeInfo] = candiinfo{CreateTime: curtime, Connid: conn.GetId()}
	self.l.Unlock()

	WriteResult(conn, &rep)
}

//清除过期的key

func (self *ChatServicesForClient) CheckKey(authkey string) (bool, uint64) {
	self.l.Lock()
	defer self.l.Unlock()

	info, ok := self.candiMap[authkey]
	if !ok {
		return false, 0
	}

	delete(self.candiMap, authkey)

	return true, info.Connid
}

func (self *ChatServicesForClient) Cleanup() {
	self.l.Lock()
	defer self.l.Unlock()
	curtime := time.Now().Unix()
	for i, v := range self.candiMap {

		if curtime-v.CreateTime > 15 {
			delete(self.candiMap, i)
		}
	}
}
