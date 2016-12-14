package connector

import (
	"errors"
	"golang-project/dpsg/logger"
	"golang-project/dpsg/proto"
	"golang-project/dpsg/rpc"
	"strconv"
	"strings"
)

func (self *CNServer) PlayerGetMailAttach(conn rpc.RpcConn, info rpc.ClientGetMailAttach) error {
	ts("CNServer:PlayerGetMailAttach", conn.GetId(), info.GetMailid())
	defer te("CNServer:PlayerGetMailAttach", conn.GetId(), info.GetMailid())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		logger.Info("The Mail return nil!")
		return nil
	}

	req := &proto.GetMailAttach{PlayerId: p.GetUid(), MailId: info.GetMailid()}
	rst := &proto.GetMailAttachResult{}

	err := self.chatRpcConn.Call("ChatServices.PlayerGetAttach", req, rst)
	if err != nil {
		logger.Info("ChatServices.PlayerGetAttach err", err)
		return err
	}

	if rst.Attach == "" {
		logger.Info("CNServer.PlayerGetMailAttach no attach")
		return errors.New("no attach")
	}

	//附件处理
	mapAdd := make(map[int]uint32, 0)
	attachs := strings.Split(rst.Attach, ",")
	for _, attach := range attachs {
		if attach == "" {
			continue
		}

		typeandnum := strings.Split(attach, ":")
		attachtype, err := strconv.Atoi(typeandnum[0])
		if err != nil {
			logger.Info("CNServer.PlayerGetMailAttach error atoi", err)
			return err
		}
		attachnum, err := strconv.Atoi(typeandnum[1])
		if err != nil {
			logger.Info("CNServer.PlayerGetMailAttach error atoi 1", err)
			return err
		}

		if attachnum <= 0 {
			logger.Info("CNServer.PlayerGetMailAttach wrong attach num")
			return errors.New("wrong attach num")
		}

		//由于角色那边的类型没有终止值就不判断了
		if (attachtype > 0 && attachtype < int(rpc.MailAttach_NorEnd)) || (attachtype > int(rpc.MailAttach_CharBegin)) {
		} else {
			logger.Info("CNServer.PlayerGetMailAttach wrong attach type")
			return errors.New("wrong attach type")
		}

		mapAdd[attachtype] = uint32(attachnum)
	}

	//判断
	needSpace := int32(0)
	for attachtype, attachnum := range mapAdd {
		if attachtype > 0 && attachtype < int(rpc.MailAttach_NorEnd) {
		} else {
			chartype := attachtype - int(rpc.MailAttach_CharBegin)
			if charCfg := GetCharacterCfgByTypeId(rpc.CharacterType(chartype), 1); charCfg != nil {
				needSpace += int32(charCfg.HousingSpace * attachnum)
			} else {
				logger.Info("CNServer.PlayerGetMailAttach wrong char type")
				return errors.New("wrong char type")
			}
		}
	}
	if needSpace > p.v.barrack_GetTroopHousingTotalFreeSpaces() {
		logger.Info("CNServer.PlayerGetMailAttach no enougn char space")
		return errors.New("no enougn char space")
	}

	//正式加
	for attachtype, attachnum := range mapAdd {
		switch rpc.MailAttach(attachtype) {
		case rpc.MailAttach_Food:
			p.GainResource(attachnum, proto.ResType_Food, proto.Gain_SymtemMail)
		case rpc.MailAttach_Gold:
			p.GainResource(attachnum, proto.ResType_Gold, proto.Gain_SymtemMail)
		case rpc.MailAttach_Gem:
			p.GainResource(attachnum, proto.ResType_Gem, proto.Gain_SymtemMail)
		case rpc.MailAttach_Wuhun:
			p.GainResource(attachnum, proto.ResType_Wuhun, proto.Gain_SymtemMail)
		case rpc.MailAttach_Tili:
			p.GainResource(attachnum, proto.ResType_TiLi, proto.Gain_SymtemMail)
		default:
			{
				chartype := attachtype - int(rpc.MailAttach_CharBegin)
				for i := uint32(0); i < attachnum; i++ {
					p.v.barrack_TroopHousingPushCharacter(rpc.CharacterType(chartype))
				}
			}
		}
	}

	success := rpc.ClientGetMailAttachResult{}
	success.SetMailid(info.GetMailid())

	logger.Info("PlayerGetMailAttach success %v", success)

	WriteResult(conn, &success)

	return nil
}

func (self *CNServer) PlayerDeleteMail(conn rpc.RpcConn, info rpc.ClientDeleteMail) error {
	ts("CNServer:PlayerDeleteMail", conn.GetId(), info.GetMailid())
	defer te("CNServer:PlayerDeleteMail", conn.GetId(), info.GetMailid())
	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	req := &proto.DelPlayerMail{PlayerId: p.GetUid(), MailId: info.GetMailid()}
	rst := &proto.DelPlayerMailResult{}

	err := cns.chatRpcConn.Call("ChatServices.PlayerDeleteMail", req, rst)
	if err != nil {
		return err
	}

	success := rpc.ClientDeleteMailResult{}
	success.SetMailid(info.GetMailid())

	WriteResult(conn, &success)

	return nil
}

//发送邮件
func (self *CNServer) PlayerSendMail(conn rpc.RpcConn, info rpc.ClientSendMail) error {
	ts("CNServer:PlayerSendMail", conn.GetId())
	defer te("CNServer:PlayerSendMail", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	//attach的判断先不做
	req := &proto.SendPlayerMail{
		ToPlayerId:     info.GetToplayer(),
		FromUid:        p.GetUid(),
		FromName:       p.GetName(),
		FromLevel:      p.GetLevel(),
		FromClan:       p.GetClan(),
		FromClanSymbol: p.GetClanSymbol(),
		Title:          info.GetTitle(),
		Content:        info.GetContent(),
		Attach:         "",
		//Attach : info.GetAttach()
	}
	rst := &proto.SendPlayerMailResult{}

	self.chatRpcConn.Go("ChatServices.SendMail2Player", req, rst, nil)

	return nil
}

//读邮件
func (self *CNServer) PlayerReadMail(conn rpc.RpcConn, info rpc.ClientReadMail) error {
	ts("CNServer:PlayerReadMail", conn.GetId(), info.GetMailid())
	defer te("CNServer:PlayerReadMail", conn.GetId(), info.GetMailid())
	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	req := &proto.ReadPlayerMail{PlayerId: p.GetUid(), MailId: info.GetMailid()}
	rst := &proto.ReadPlayerMailResult{}

	cns.chatRpcConn.Go("ChatServices.PlayerReadMail", req, rst, nil)

	return nil
}

//func (self *CNServer) TestMail() {
//	timeBegin := time.Now().Unix()
//	for i := 0; i < 10000; i++ {
//		req := &proto.SendPlayerMail {
//			ToPlayerId : "test",
//			FromUid	: "test",
//			FromName : "test",
//			FromLevel :1,
//			FromClan : "test",
//			FromClanSymbol : 1,
//			Title : "test",
//			Content	: "test",
//			Attach : "",
//			//Attach : info.GetAttach()
//		}
//		rst := &proto.SendPlayerMailResult{}

//		self.chatRpcConn.Call("ChatServices.SendMail2Player", req, rst)
//	}
//	timeEnd := time.Now().Unix()
//	logger.Info("!!!!!!!!!!!!!time%d,%d,%d", timeBegin, timeEnd, timeEnd - timeBegin)
//}
