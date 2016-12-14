package center

import (
	//gp "github.com/golang/protobuf/proto"
	"golang-project/dpsg/common"
	"golang-project/dpsg/logger"
	//"math/rand"
	"golang-project/dpsg/proto"
	"golang-project/dpsg/rpc"
	//"strconv"
	//"sync"
	//"time"
	//"golang-project/dpsg/timer"
	"golang-project/dpsg/dbclient"
)

func (self *Center) FindToPlayerFromDB(req *proto.PlayerGiveGift, ret *proto.PlayerGiveGiftResult) error {
	logger.Info("ComeInto center.FindToPlayerFromDB")

	var p rpc.PlayerExtraInfo

	exist, err := dbclient.KVQueryExt(common.PlayerExtra, req.Uid, &p)
	if err == nil {
		if exist {
			logger.Info("player is exist", req.Uid)
			ret.Code = proto.FindPlayerOK
		} else {
			logger.Info("cat find player", req.Uid)
			ret.Code = proto.PlayNotExist
		}
	} else {
		logger.Error("Find player in DB error", err.Error())
		return err
	}

	return nil
}
