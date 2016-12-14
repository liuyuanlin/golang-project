package chatserver

import (
	"golang-project/dpsg/logger"
	"golang-project/dpsg/proto"
	"golang-project/dpsg/rpc"
	"golang-project/dpsg/rpcplus"
	"net"
	"strings"
	"time"
)

type ChatGmServices struct {
}

var pChatGmServices *ChatGmServices

func CreateChatServicesForGm(listener net.Listener) {
	pChatGmServices = &ChatGmServices{}

	rpcServer := rpcplus.NewServer()

	rpcServer.Register(pChatGmServices)

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

func (self *ChatGmServices) GmSendMail(req *proto.GmSendMail, rst *proto.GmSendMailResult) error {
	rst.SignId = req.SignId

	result := &proto.SendSystemMailResult{}
	users := strings.Split(req.Users, ",")
	for _, id := range users {
		if id == "" {
			continue
		}

		mail := proto.SendSystemMail{
			ToPlayerId: id,
			Title:      req.Title,
			Content:    req.Content,
			Attach:     req.Attach,
		}

		if pChatServices.SendSysMail2Player(&mail, result) == nil {
			rst.Success += id + ":1,"
		} else {
			rst.Success += id + ":0,"
		}
	}

	logger.Info("GmSendMail:%s", rst.Success)
	return nil
}

func (self *ChatGmServices) GmSendAllMail(req *proto.GmSendAllMail, rst *proto.GmSendAllMailResult) error {
	rst.SignId = req.SignId

	timeCur := uint32(time.Now().Unix())
	pMail := &rpc.SysMail{}
	pMail.SetTitle(req.Title)
	pMail.SetContent(req.Content)
	pMail.SetSendtime(timeCur)
	pMail.SetAttach(req.Attach)
	pMail.SetChannelid(rpc.GameLocation(req.Channel))
	pMail.SetOverduetime(timeCur + req.ContinueTime)

	err := pChatMailServices.AddSysMail(pMail)
	if err == nil {
		rst.Success = true
	} else {
		rst.Success = false
	}

	return err
}

//发送通知
func (self *ChatGmServices) GmSendNotice(req *proto.GmSendNotice, rst *proto.GmSendNoticeResult) error {
	rst.SignId = req.SignId

	if req.Content == "" {
		rst.Success = false
		return nil
	}

	rst.Success = true

	if req.Type == int64(proto.GmNoticeType_Login) {
		//暂时没有
	} else if req.Type == int64(proto.GmNoticeType_Marquee) {
		pNotice := &rpc.Notice{}
		pNotice.SetContent(req.Content)

		pChatServices.l.RLock()
		defer pChatServices.l.RUnlock()
		for _, conninfo := range pChatServices.PlayerMap {
			if rpc.GameLocation(req.Channel) == conninfo.ChannelId {
				if conn := pChatServicesForClient.rpcServer.GetConn(conninfo.ConnId); conn != nil {
					WriteResult(conn, pNotice)
				}
			}
		}
	}

	return nil
}
