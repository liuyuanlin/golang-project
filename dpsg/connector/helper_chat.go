package connector

import (
	"golang-project/dpsg/logger"
	"golang-project/dpsg/proto"
	"golang-project/dpsg/rpc"
	"time"

	gp "github.com/golang/protobuf/proto"
)

func UpdatePlayerChatInfo(uid string, level uint32, clanName string, clanPower rpc.Player_ClanPower) error {
	ts("UpdatePlayerChatInfo")
	defer te("UpdatePlayerChatInfo")

	req := &proto.UpdatePlayer{
		PlayerId:      uid,
		PlayerLevel:   uint8(level),
		AllianceName:  clanName,
		AlliancePower: uint32(clanPower),
	}
	rst := &proto.UpdatePlayerResult{}

	cns.chatRpcConn.Go("ChatServices.UpdatePlayer", req, rst, nil)

	return nil
}

func SendDonateMsg(uid string, s2cMsg rpc.S2CDonate) error {
	ts("SendDonateMsg", s2cMsg)
	defer te("SendDonateMsg", s2cMsg)

	buf, err := gp.Marshal(&s2cMsg)
	if err != nil {
		logger.Error("SendDonateMsg Error On Marshal (%s, %v)", err.Error(), s2cMsg)
		return err
	}

	req := &proto.DonateMsg{Uid: uid, Value: buf}
	rst := &proto.DonateMsgResult{}

	cns.chatRpcConn.Go("ChatServices.SendDonateMsg", req, rst, nil)

	return nil
}

func UpdateDonateMsg(uid string, s2cMsg rpc.S2CDonateUpdate) error {
	ts("UpdateDonateMsg", s2cMsg)
	defer te("UpdateDonateMsg", s2cMsg)

	buf, err := gp.Marshal(&s2cMsg)
	if err != nil {
		logger.Error("UpdateDonateMsg Error On Marshal (%s, %v)", err.Error(), s2cMsg)
		return err
	}

	req := &proto.DonateMsg{Uid: uid, Value: buf}
	rst := &proto.DonateMsgResult{}

	cns.chatRpcConn.Go("ChatServices.UpdateDonateMsg", req, rst, nil)

	return nil
}

func CastClanChatMsg(cname string, msg rpc.ClanChatMessage) error {
	ts("CastClanMsg", msg)
	defer te("CastClanMsg", msg)

	msg.SetTime(time.Now().Unix())

	buf, err := gp.Marshal(&msg)
	if err != nil {
		logger.Error("CastClanMsg Error On Marshal (%s, %v)", err.Error(), msg)
		return err
	}

	req := &proto.ClanMsg{CName: cname, Value: buf}
	rst := &proto.ClanMsgResult{}

	cns.chatRpcConn.Go("ChatServices.CastClanMsg", req, rst, nil)

	return nil
}
