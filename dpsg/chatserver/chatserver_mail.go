package chatserver

import (
	"golang-project/dpsg/common"
	"golang-project/dpsg/logger"
	//"fmt"
	"errors"
	"golang-project/dpsg/proto"
	"golang-project/dpsg/rpc"
	"golang-project/dpsg/rpcplus"
	"net"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
)

type ChatMailServices struct {
	maincache  *common.CachePool
	mails      []*rpc.SysMail
	curversion uint64
	l          sync.RWMutex
	db         *rpcplus.Client
}

var pChatMailServices *ChatMailServices

//创建邮件服务
func CreateMailServices(cfg common.ChatServerCfg) *ChatMailServices {
	var dbCfg common.DBConfig
	if err := common.ReadDbConfig("dbExtern.json", &dbCfg); err != nil {
		logger.Fatal("%v", err)
	}

	dbconn, err := net.Dial("tcp", dbCfg.DBHost)
	if err != nil {
		logger.Fatal("%s", err.Error())
	}

	pChatMailServices = &ChatMailServices{
		maincache:  common.NewCachePool(cfg.MainCacheProfile),
		mails:      make([]*rpc.SysMail, 0),
		curversion: uint64(0),
		db:         rpcplus.NewClient(dbconn),
	}

	if err := pChatMailServices.initSysMail(); err != nil {
		logger.Fatal("%s", err.Error())
	}

	//pMail := &rpc.SysMail {}
	//pMail.SetVersion(uint64(1))
	//pMail.SetTitle("title")
	//pMail.SetContent("content")
	//pMail.SetSendtime(uint32(0))
	//pMail.SetAttach("")
	//pChatMailServices.AddSysMail(pMail)

	return pChatMailServices
}

//初始化
func (self *ChatMailServices) initSysMail() error {
	cache := self.maincache.Get()
	defer cache.Recycle()

	self.l.Lock()
	defer self.l.Unlock()

	mailsdata, err := cache.Conn.Do("SMEMBERS", common.GetSystemTableKey_Mail())
	if err != nil {
		logger.Fatal("init sys mail failed")
	}

	arrmails, err := redis.Values(mailsdata, err)
	if err != nil {
		logger.Fatal("array sys mail failed")
	}

	uMaxVersion := uint64(0)
	for _, buf := range arrmails {
		bytes, err := redis.Bytes(buf, nil)
		if err != nil {
			logger.Fatal("Bytes sys mail failed")
			continue
		}

		var mail *rpc.SysMail = new(rpc.SysMail)
		if err = common.DecodeMessage(bytes, mail); err != nil {
			logger.Fatal("decode sys mail failed")
			continue
		}

		self.mails = append(self.mails, mail)

		if mail.GetVersion() > uMaxVersion {
			uMaxVersion = mail.GetVersion()
		}
	}
	self.curversion = uMaxVersion
	//fmt.Println(self.mails, self.curversion)
	return nil
}

//向所有玩家推送新系统邮件
func (self *ChatMailServices) pushMail2Online(mail *rpc.SysMail, version uint64) error {
	pChatServices.l.RLock()
	defer pChatServices.l.RUnlock()

	for sPlayerId, conninfo := range pChatServices.PlayerMap {
		if conninfo.ChannelId == mail.GetChannelid() {
			mailinfo := &rpc.PlayerMailInfo{}
			if exist, err := common.KVQuery(pChatMailServices.db, "playermail", sPlayerId, mailinfo); err == nil {
				if !exist {
					continue
				}

				pMail := &rpc.PlayerMail{}
				pMail.SetMailid(common.GenMailId())
				pMail.SetMailtype(rpc.PlayerMail_System)
				pMail.SetFromname("")
				pMail.SetTitle(mail.GetTitle())
				pMail.SetContent(mail.GetContent())
				pMail.SetSendtime(uint32(time.Now().Unix()))
				pMail.SetAttach(mail.GetAttach())
				pMail.SetBread(false)

				mailinfo.Maillist = append(mailinfo.Maillist, pMail)
				mailinfo.SetSysmailVersion(version)

				//存储
				ok, err := common.KVWrite(pChatMailServices.db, "playermail", sPlayerId, mailinfo)
				if err == nil && ok {
					//下发
					if conn := pChatServicesForClient.rpcServer.GetConn(conninfo.ConnId); conn != nil {
						WriteResult(conn, pMail)
					}
				}
			}
		}
	}

	return nil
}

//发全服系统邮件
func (self *ChatMailServices) AddSysMail(mail *rpc.SysMail) error {
	cache := self.maincache.Get()
	defer cache.Recycle()

	self.l.Lock()
	defer self.l.Unlock()

	//自动设置版本号
	self.curversion++
	mail.SetVersion(self.curversion)
	self.mails = append(self.mails, mail)

	vaule, err := common.EncodeMessage(mail)
	if err != nil {
		logger.Fatal("EncodeMessage sys mail failed")
		return err
	}

	_, err = cache.Conn.Do("SADD", common.GetSystemTableKey_Mail(), vaule)
	if err != nil {
		logger.Fatal("add sys mail failed")
		return err
	}

	self.pushMail2Online(mail, self.curversion)

	return nil
}

//发个人系统邮件
func (self *ChatServices) SendSysMail2Player(info *proto.SendSystemMail, result *proto.SendSystemMailResult) (err error) {
	pChatMailServices.l.Lock()
	defer pChatMailServices.l.Unlock()

	mailinfo := &rpc.PlayerMailInfo{}
	if exist, err := common.KVQuery(pChatMailServices.db, "playermail", info.ToPlayerId, mailinfo); err == nil {
		if !exist {
			return errors.New("send mail wrong player")
		}

		mail := &rpc.PlayerMail{}
		mail.SetMailid(common.GenMailId())
		mail.SetMailtype(rpc.PlayerMail_System)
		mail.SetFromname("")
		mail.SetTitle(info.Title)
		mail.SetContent(info.Content)
		mail.SetSendtime(uint32(time.Now().Unix()))
		mail.SetAttach(info.Attach)
		mail.SetBread(false)

		mailinfo.Maillist = append(mailinfo.Maillist, mail)

		//存储
		ok, err := common.KVWrite(pChatMailServices.db, "playermail", info.ToPlayerId, mailinfo)
		if err == nil && ok {
			//下发
			self.l.RLock()
			defer self.l.RUnlock()
			if conninfo, ok := self.PlayerMap[info.ToPlayerId]; ok {
				if conn := pChatServicesForClient.rpcServer.GetConn(conninfo.ConnId); conn != nil {
					WriteResult(conn, mail)
				}
			}
		}
	}

	return err
}

//玩家初始化邮件系统
func (self *ChatServices) initPlayerMail(sPlayerId string, channelId rpc.GameLocation) {
	pChatMailServices.l.Lock()
	defer pChatMailServices.l.Unlock()

	mailinfo := &rpc.PlayerMailInfo{}
	if exist, err := common.KVQuery(pChatMailServices.db, "playermail", sPlayerId, mailinfo); err == nil {
		if !exist {
			mailinfo = &rpc.PlayerMailInfo{}
		}

		//取系统邮件
		mails, version := self.playerPickupSysMail(mailinfo.GetSysmailVersion(), channelId)
		mailinfo.Maillist = append(mailinfo.Maillist, mails[:]...)
		mailinfo.SetSysmailVersion(version)

		//存储
		ok, err := common.KVWrite(pChatMailServices.db, "playermail", sPlayerId, mailinfo)
		if err == nil && ok {
			//下发
			self.l.RLock()
			defer self.l.RUnlock()
			if conninfo, ok := self.PlayerMap[sPlayerId]; ok {
				if conn := pChatServicesForClient.rpcServer.GetConn(conninfo.ConnId); conn != nil {
					WriteResult(conn, mailinfo)
				}
			}
		}
	}
}

//玩家取系统邮件
func (self *ChatServices) playerPickupSysMail(uVersion uint64, channelId rpc.GameLocation) (mails []*rpc.PlayerMail, curversion uint64) {
	cutTime := uint32(time.Now().Unix())
	for _, mail := range pChatMailServices.mails {
		if (mail.GetOverduetime() == 0 || cutTime < mail.GetOverduetime()) && mail.GetVersion() > uVersion && mail.GetChannelid() == channelId {
			pm := &rpc.PlayerMail{}
			pm.SetMailid(common.GenMailId())
			pm.SetMailtype(rpc.PlayerMail_System)
			pm.SetFromname("")
			pm.SetTitle(mail.GetTitle())
			pm.SetContent(mail.GetContent())
			pm.SetSendtime(mail.GetSendtime())
			pm.SetAttach(mail.GetAttach())
			pm.SetBread(false)

			mails = append(mails, pm)
		}
	}
	curversion = pChatMailServices.curversion

	return
}

//向玩家发送邮件
func (self *ChatServices) SendMail2Player(info *proto.SendPlayerMail, result *proto.SendPlayerMailResult) (err error) {
	pChatMailServices.l.Lock()
	defer pChatMailServices.l.Unlock()

	mailinfo := &rpc.PlayerMailInfo{}
	if exist, err := common.KVQuery(pChatMailServices.db, "playermail", info.ToPlayerId, mailinfo); err == nil {
		if !exist {
			return errors.New("send mail wrong player")
		}

		mail := &rpc.PlayerMail{}
		mail.SetMailid(common.GenMailId())
		mail.SetMailtype(rpc.PlayerMail_Normal)
		mail.SetFromname(info.FromName)
		mail.SetFromuid(info.FromUid)
		mail.SetFromlevel(info.FromLevel)
		mail.SetFromclan(info.FromClan)
		mail.SetFromclansymbol(info.FromClanSymbol)
		mail.SetTitle(info.Title)
		mail.SetContent(info.Content)
		mail.SetSendtime(uint32(time.Now().Unix()))
		mail.SetAttach(info.Attach)
		mail.SetBread(false)
		mailinfo.Maillist = append(mailinfo.Maillist, mail)

		//存储
		ok, err := common.KVWrite(pChatMailServices.db, "playermail", info.ToPlayerId, mailinfo)
		if err == nil && ok {
			//下发
			self.l.RLock()
			defer self.l.RUnlock()
			if conninfo, ok := self.PlayerMap[info.ToPlayerId]; ok {
				if conn := pChatServicesForClient.rpcServer.GetConn(conninfo.ConnId); conn != nil {
					WriteResult(conn, mail)
				}
			}

			//给发送者发消息
			if conninfo, ok := self.PlayerMap[info.FromUid]; ok {
				if conn := pChatServicesForClient.rpcServer.GetConn(conninfo.ConnId); conn != nil {
					common.SendMsg(conn, "TID_MAIL_SEND_SUCCESS")
				}
			}
		}
	}

	return err
}

//玩家取附件
func (self *ChatServices) PlayerGetAttach(info *proto.GetMailAttach, result *proto.GetMailAttachResult) (err error) {
	pChatMailServices.l.Lock()
	defer pChatMailServices.l.Unlock()

	mailinfo := &rpc.PlayerMailInfo{}
	if exist, err := common.KVQuery(pChatMailServices.db, "playermail", info.PlayerId, mailinfo); err == nil {
		if !exist {
			return errors.New("no mail")
		}

		for _, mail := range mailinfo.Maillist {
			if mail.GetMailid() == info.MailId {
				result.Attach = mail.GetAttach()

				//存储
				mail.SetAttach("")
				ok, err := common.KVWrite(pChatMailServices.db, "playermail", info.PlayerId, mailinfo)
				if err == nil && ok {
					return nil
				}

				return errors.New("save mail error")
			}
		}

		return errors.New("no mail")
	}

	return
}

//玩家删除邮件
func (self *ChatServices) PlayerDeleteMail(info *proto.DelPlayerMail, result *proto.DelPlayerMailResult) (err error) {
	pChatMailServices.l.Lock()
	defer pChatMailServices.l.Unlock()

	mailinfo := &rpc.PlayerMailInfo{}
	if exist, err := common.KVQuery(pChatMailServices.db, "playermail", info.PlayerId, mailinfo); err == nil {
		if !exist {
			return errors.New("no mail")
		}

		for index, mail := range mailinfo.Maillist {
			if mail.GetMailid() == info.MailId {
				//删除
				mailinfo.Maillist = append(mailinfo.Maillist[:index], mailinfo.Maillist[index+1:]...)
				ok, err := common.KVWrite(pChatMailServices.db, "playermail", info.PlayerId, mailinfo)
				if err == nil && ok {
					return nil
				}

				return errors.New("del mail failed")
			}
		}
	}

	return errors.New("del mail failed")
}

//玩家读取邮件
func (self *ChatServices) PlayerReadMail(info *proto.ReadPlayerMail, result *proto.ReadPlayerMailResult) (err error) {
	pChatMailServices.l.Lock()
	defer pChatMailServices.l.Unlock()

	mailinfo := &rpc.PlayerMailInfo{}
	if exist, err := common.KVQuery(pChatMailServices.db, "playermail", info.PlayerId, mailinfo); err == nil {
		if !exist {
			return errors.New("no mail")
		}

		for _, mail := range mailinfo.Maillist {
			if mail.GetMailid() == info.MailId {
				//删除
				mail.SetBread(true)
				ok, err := common.KVWrite(pChatMailServices.db, "playermail", info.PlayerId, mailinfo)
				if err == nil && ok {
					return nil
				}

				return errors.New("read mail failed")
			}
		}
	}

	return errors.New("del mail failed")
}
