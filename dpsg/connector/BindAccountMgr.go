// BindAccountMgr

package connector

/*
import (
	//"common"
	"fmt"
	//"net"
	//	"os"
	"rpc"
	//	"strings"
	//"bytes"
	//"encoding/binary"
	//"encoding/json"
)

func (self *CNServer) GetUidFrom3rdAccount(account string, tptype rpc.ThirdPartyType) (uid string, exist bool) {
	//去数据库查找跟据账号查找uid
	var buff rpc.CommString
	switch tptype {
	case rpc.ThirdPartyType_GameCenter:
		{
			ts("ThirdPartyType_GameCenter %s", account)
			exist, _ = KVQuery("gcaccount", account, &buff)
		}
	case rpc.ThirdPartyType_GooglePlus:
		{
			ts("ThirdPartyType_GooglePlus %s", account)
			exist, _ = KVQuery("gpaccount", account, &buff)
		}
	}

	uid = buff.GetStr()

	ts("CNServer:GetUidFrom3rdAccount %s", uid)

	return
}

func (self *CNServer) Bind3rdAccount(conn rpc.RpcConn, b3a rpc.Bind3RdAccount) error {

	var err error

	account := b3a.GetAccount()
	tptype := b3a.GetTptype()

	_, exist := self.GetUidFrom3rdAccount(account, tptype)

	rps := &rpc.Bind3RDAccountResult{}

	if exist {
		//如果该账号已经绑定，返回消息该账号已经绑定，客户端重新登录即可
		rps.SetRst(rpc.Bind3RDAccountResult_ExistAccount)
	} else {

		//如果该账号未绑定，开始绑定，并返回消息绑定成功，客户端可以不做任何处理。
		buff := &rpc.CommString{}

		buff.SetStr(b3a.GetUid())

		switch tptype {
		case rpc.ThirdPartyType_GameCenter:
			{
				_, err = KVWrite("gcaccount", b3a.GetAccount(), buff)

			}
		case rpc.ThirdPartyType_GooglePlus:
			{
				_, err = KVWrite("gpaccount", b3a.GetAccount(), buff)
			}
		}

		rps.SetRst(rpc.Bind3RDAccountResult_BindOk)

		fmt.Println("Bind3rdAccount", err)

	}

	WriteResult(conn, rps)

	return err
}

func (self *CNServer) Query3rdAccount(conn rpc.RpcConn, q3a rpc.Query3RDAccount) error {
	account := q3a.GetAccount()
	tptype := q3a.GetTptype()

	_, exist := self.GetUidFrom3rdAccount(account, tptype)

	rps := &rpc.Qery3RDAccountResult{}

	if exist {
		rps.SetRst(true)
	} else {
		rps.SetRst(false)
	}

	WriteResult(conn, rps)

	return nil

} */
