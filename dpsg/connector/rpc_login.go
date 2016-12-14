package connector

import (
	//"language"
	"golang-project/dpsg/accountclient"
	"golang-project/dpsg/common"
	"golang-project/dpsg/lockclient"
	"golang-project/dpsg/logger"
	"golang-project/dpsg/proto"
	"golang-project/dpsg/rpc"
	"strings"
	"time"
)

//登陆错误处理函数
func (self *CNServer) LoginError(p *player) {
	lockclient.TryUnlock("player", *p.Uid, p.lid)
}

func (self *CNServer) Login(conn rpc.RpcConn, login rpc.Login) error {
	ts("CNServer:Login", conn.GetId(), login.GetUid())

	self.l.RLock()
	_, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if exist {
		te("CNServer:Login", conn.GetId(), login.GetUid())
		return nil
	}

	err := self.login(conn, &login)
	te("CNServer:Login", conn.GetId(), login.GetUid())
	return err
}

const LoginUnderAttackTimeLimit = 210

func (self *CNServer) login(conn rpc.RpcConn, login *rpc.Login) error {
	var temp []string
	temp = strings.Split(*login.Gatekey, ";")
	if 3 != len(temp) {
		WriteLoginResult(conn, rpc.LoginResult_GATEKEYERROR)
		return nil
	}

	Ip := temp[0]
	gatetime := temp[1]
	result := temp[2]

	tempResult := common.Base64Encode([]byte(Ip + ";" + gatetime))

	if result != string(tempResult) {
		WriteLoginResult(conn, rpc.LoginResult_GATEKEYERROR)
		return nil
	}

	uid := login.GetUid()
	if !CheckUUID(uid) {
		logger.Info("New user %s", uid)
		uid = GenUUID(self.GetServerId())
	}

	// 这里开始渠道商的介入
	channelId := login.GetChannelid()

	logger.Info("channelId:%d", channelId)

	//channelId := 1
	switch channelId {
	case rpc.GameLocation_GameTest: // 表明是直接登陆，没有走任何运营商
		{
			return self.AfterAccountServer(login, uid, conn, channelId, Ip)
		}
	case rpc.GameLocation_Robot: //机器人渠道
		{
			return self.AfterAccountServer(login, uid, conn, channelId, Ip)
		}
	case rpc.GameLocation_Tencent:
		{
			return self.TxLogin(conn, login, Ip)
		}
	}

	return nil

}

func (self *CNServer) TxLogin(conn rpc.RpcConn, login *rpc.Login, Ip string) error {

	ret, errmsg := MobileQQAuth(login, Ip)

	if ret {
		binduid, err := accountclient.QueryPlayerIdByPartnerId(common.TableName_TencentAccount, login.GetOpenid())
		if err != nil {
			logger.Info("TxLogin QueryPlayerIdByPartnerId :%v", err)
			WriteLoginResultWithErrorMsg(conn, rpc.LoginResult_TX_AUTH_FAILED, "account client failed!")
			return nil
		}

		if len(binduid) > 0 {
			self.AfterAccountServer(login, binduid, conn, login.GetChannelid(), Ip)
		} else {
			uid := GenUUID(self.GetServerId())
			err = accountclient.SetPartnerIdToPlayerId(common.TableName_TencentAccount, login.GetOpenid(), uid)
			if err == nil {
				logger.Info("TxLogin SetPartnerIdToPlayerId :%v", err)
				self.AfterAccountServer(login, uid, conn, login.GetChannelid(), Ip)
			} else {
				WriteLoginResultWithErrorMsg(conn, rpc.LoginResult_TX_AUTH_FAILED, "bind id failed!")
				return nil
			}
		}
	} else {
		WriteLoginResultWithErrorMsg(conn, rpc.LoginResult_TX_AUTH_FAILED, errmsg)
	}
	return nil
}

func (self *CNServer) AfterAccountServer(login *rpc.Login, uid string, conn rpc.RpcConn, gl rpc.GameLocation, ip string) error {
	lid := GenLockMessage(self.GetServerId(), proto.MethodPlayerLogin, 0)

	bFirstRound := true
	//顶号次数，做限制，最多5次，因为其它逻辑会出现一直顶不下的情况
	iKickTimes := 0

	for {
		successed, old_value, err := lockclient.TryLock("player", uid, lid)

		if err != nil {
			WriteLoginResult(conn, rpc.LoginResult_SERVERERROR)
			return nil
		}

		if successed {
			break
		}

		_, tid, _, t, _ := ParseLockMessage(old_value)

		switch tid {
		case proto.MethodPlayerLogin:
			{
				iKickTimes++
				bFailed := true
				if iKickTimes <= 5 {
					//顶号
					logger.Info("begin kick self...")
					req := &proto.LoginKickPlayer{
						Id: uid,
					}
					rst := &proto.LoginKickPlayerResult{}
					if err := self.center.Call("Center.KickCnsPlayer", req, rst); err == nil && rst.Success {
						logger.Info("kick self success!")
						//不处理，等下面等待1秒后重试
						bFailed = false
					} else {
						logger.Info("kick self error:%v", err)
					}
				}

				if bFailed {
					logger.Info("login kick self failed")
					WriteLoginResult(conn, rpc.LoginResult_ISONFIRE)
					return nil
				}
			}
		case proto.MethodPlayerMatch:
			{
				// 剩余时间
				timeleft := LoginUnderAttackTimeLimit - (uint32(time.Now().Unix()) - t)

				if timeleft > LoginUnderAttackTimeLimit {
					logger.Info("Login wait timeout!")
					WriteLoginResult(conn, rpc.LoginResult_ISONFIRE)
					return nil
				}

				if bFirstRound {
					wait := &rpc.WaitLogin{}
					wait.SetTime(timeleft)

					if !WriteResult(conn, wait) {
						logger.Error("MethodPlayerMatch:!WriteResult(conn, wait)")
						return nil
					}
					bFirstRound = false
				}
			}
		case proto.MethodPlayerRevenge:
			{
				// 剩余时间
				timeleft := LoginUnderAttackTimeLimit - (uint32(time.Now().Unix()) - t)

				if timeleft > LoginUnderAttackTimeLimit {
					logger.Info("Login wait timeout!")
					WriteLoginResult(conn, rpc.LoginResult_ISONFIRE)
					return nil
				}

				if bFirstRound {
					wait := &rpc.WaitLogin{}
					wait.SetTime(timeleft)

					if !WriteResult(conn, wait) {
						logger.Error("MethodPlayerRevenge:!WriteResult(conn, wait)")
						return nil
					}
					bFirstRound = false
				}
			}
		default:
			logger.Info("default, tid(%d)", tid)
			return nil
		}

		time.Sleep(time.Second)
	}

	p := LoadPlayer(uid, lid, gl)
	if p == nil {
		lockclient.TryUnlock("player", uid, lid)
		WriteLoginResult(conn, rpc.LoginResult_SERVERERROR)
		return nil
	}
	//应该放在判断非空的后面
	p.OnInit(conn)

	//腾讯
	info := &MobileQQInfo{
		Openid:    login.GetOpenid(),
		Openkey:   login.GetOpenkey(),
		Pay_token: login.GetPayToken(),
		Pf:        login.GetPf(),
		Pfkey:     login.GetPfkey(),
		Balance:   uint32(0),
	}
	p.mobileqqinfo = info
	logger.Info("MobileQQInfo : %v", info)

	v := p.GetVillage()

	if v == nil {
		lockclient.TryUnlock("player", uid, lid)
		WriteLoginResult(conn, rpc.LoginResult_SERVERERROR)
		return nil
	}

	ret := WriteLoginResult(conn, rpc.LoginResult_OK)
	if !ret {
		self.LoginError(p)
		return nil
	}

	//在下发之前更新qq昵称
	if ok, _, nickname, gender, picture40, _ := MobileQQQuery(p); ok {
		p.SetName(nickname)
		p.SetGender(gender)
		p.SetHeadurl(picture40)
	}

	//下发玩家数据
	playerinfo := &rpc.PlayerInfo{Base: p.PlayerBaseInfo, Extra: p.PlayerExtraInfo}
	ret = WriteResult(conn, playerinfo)
	if !ret {
		self.LoginError(p)
		return nil
	}

	ret = WriteResult(conn, v.VillageInfo)
	if !ret {
		self.LoginError(p)
		return nil
	}

	if p.pve != nil {
		ret = WriteResult(conn, p.pve)
		if !ret {
			self.LoginError(p)
			return nil
		}
	}

	c := p.GetClanInfo()
	if c != nil {
		ret = WriteResult(conn, c)
		if !ret {
			self.LoginError(p)
			return nil
		}
	}

	//初始化战斗日志
	p.InitBattleLog()
	p.InitAttackLog()
	p.InitFriendExeciseLog(true)

	//复原村庄血量
	v.ResetBuildingHp()

	//进入服务器全局表
	self.addPlayer(conn.GetId(), p)

	//查询玩家金币数量，放在最后，可能会踢玩家下线
	p.UpdateThirdGem()
	//最后下发玩家元宝数据
	p.SyncPlayerGem()

	//发送登陆log
	msg := proto.LogPlayerLoginLogout{
		ChannelId: uint8(p.GetGamelocation()),
		Playerid:  uid,
		Time:      time.Now().Unix(),
		Logout:    false,
		Ip:        ip,
	}

	var logret proto.LogPlayerLoginLogoutResult
	self.logRpcConn.Go("LogServices.LogPlayerLoginLogoutGame", msg, &logret, nil)
	return nil
}

//顶号
func (self *CenterService) LoginKickPlayer(req *proto.LoginKickPlayer, rst *proto.LoginKickPlayerResult) error {
	if p, ok := cns.playersbyid[req.Id]; ok {
		if err := p.conn.Close(); err == nil {
			rst.Success = true
			return nil
		} else {
			return err
		}
	}

	rst.Success = false
	return nil
}
