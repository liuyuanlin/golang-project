package connector

import (
	"golang-project/dpsg/logger"
	"golang-project/dpsg/proto"
)

// 增加location字段
func SetPlayerTrophy(trophy uint32, location int64, uid string) (err error) {
	//ts("SetPlayerTrophy", trophy)
	//defer te("SetPlayerTrophy", trophy)

	try := &proto.SetTrophy{Trophy: trophy, Location: location, Uid: uid}
	rst := &proto.SetTrophyResult{}

	err = cns.center.Call("Center.SetPlayerTrophy", try, rst)
	//logger.Info("SetPlayerTrophy", try, rst, err)
	if err != nil {
		logger.Error("Error On SetPlayerTrophy : %s", err.Error())

		return
	}

	return
}

func NotifyOnline(uid string) (err error) {
	//ts("NotifyOnline")
	//defer te("NotifyOnline")

	try := &proto.NotifyOnline{Uid: uid}
	rst := &proto.NotifyOnlineResult{}

	err = cns.center.Call("Center.NotifyOnline", try, rst)
	//logger.Info("NotifyOnline", try, rst, err)
	if err != nil {
		logger.Error("Error On NotifyOnline : %s", err.Error())

		return
	}

	return
}

func NotifyOffline(uid string, guidefinish bool) (err error) {
	//ts("NotifyOffline")
	//defer te("NotifyOffline")

	try := &proto.NotifyOffline{Uid: uid, GuideFinish: guidefinish}
	rst := &proto.NotifyOfflineResult{}

	err = cns.center.Call("Center.NotifyOffline", try, rst)
	//logger.Info("NotifyOffline", try, rst, err)
	if err != nil {
		logger.Error("Error On NotifyOffline : %s", err.Error())

		return
	}

	return
}

func AddPlayerShield(starttime uint32, totaltime uint32, uid string) (err error, ok bool) {
	//ts("AddPlayerShield", starttime, totaltime, uid)
	//defer te("AddPlayerShield", starttime, totaltime, uid)

	try := &proto.AddShield{StartTime: starttime, TotalTime: totaltime, Uid: uid}
	rst := &proto.AddShieldResult{}

	err = cns.center.Call("Center.AddPlayerShield", try, rst)
	//logger.Info("AddPlayerShield", try, rst, err)
	if err != nil {
		logger.Error("Error On AddPlayerShield : %s", err.Error())

		return err, false
	}

	return nil, rst.Ok
}

func RemovePlayerShield(uid string) (err error) {
	//ts("RemovePlayerShield", uid)
	//defer te("RemovePlayerShield", uid)

	try := &proto.RemoveShield{Uid: uid}
	rst := &proto.RemoveShieldResult{}

	err = cns.center.Call("Center.RemovePlayerShield", try, rst)
	//logger.Info("RemovePlayerShield", try, rst, err)
	if err != nil {
		logger.Error("Error On RemovePlayerShield : %s", err.Error())

		return
	}

	return
}

func TryMatch(trophy uint32, lastmatch string) (uid string, err error) {
	//ts("TryMatch", trophy)
	//defer te("TryMatch", trophy)

	try := &proto.Match{Trophy: trophy, Except: lastmatch}
	rst := &proto.MatchResult{}

	err = cns.center.Call("Center.TryMatch", try, rst)
	if err != nil {
		logger.Info("TryMatch Result : %s", err.Error()) //test
	}

	uid = rst.Uid

	return
}
