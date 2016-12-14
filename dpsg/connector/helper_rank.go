package connector

import (
	"golang-project/dpsg/logger"
	"golang-project/dpsg/proto"
	"golang-project/dpsg/rpc"
	"sync"

	gp "github.com/golang/protobuf/proto"
)

type RankMgr struct {
	SaveRankClans           rpc.ClanInfos
	RankClansLock           sync.RWMutex
	SaveRankGlobalPlayers   rpc.RankPlayers
	GlobalPlayersLock       sync.RWMutex
	SaveTTTPlayers          rpc.TTTRankPlayers
	TTTPlayersLock          sync.RWMutex
	SaveRankPlayersLocation map[int64]rpc.RankPlayers
	PlayersLocationLock     sync.RWMutex
	SaveMyself              rpc.RankPlayers
	RankMySelfLock          sync.RWMutex
}

func CreateRankMgr() *RankMgr {
	return &RankMgr{SaveRankPlayersLocation: make(map[int64]rpc.RankPlayers)}
}

/*
func GetRankPlayers(uid string) {
	//ts("GetRankPlayers", uid)
	//defer te("GetRankPlayers", uid)

	try := &proto.GetRankPlayers{Uid: uid, Start: 0, Stop: 99}
	ret := &proto.GetRankPlayersResult{}

	err := cns.center.Call("Center.GetRankPlayers", try, ret)
	//logger.Info("GetRankPlayers11111", try, ret, err)
	if err != nil {
		logger.Error("Error On GetRankPlayers : %s, %s", uid, err.Error())
		return proto.CenterRpcError, nil
	}
	/*
		rps := &rpc.RankPlayers{}

		for _, uid := range ret.Value {
			var p rpc.PlayerInfo

			exist, err := KVQuery("player", uid, &p)
			if err != nil {
				continue
			}

			if exist {
				rp := rpc.Player{}
				rp.SetType(rpc.Player_Rank)
				rp.SetName(p.GetName())
				rp.SetUid(p.GetUid())
				rp.SetTrophy(p.GetTrophy())
				rp.SetLevel(p.GetLevel())
				rp.SetClanName(p.GetClan())
				rp.SetClanSymbol(p.GetClanSymbol())

				rps.RpsTop = append(rps.RpsTop, &rp)
			}
		}

	return ret.Code, rps
}

//查询location玩家
func GetRankPlayersLocation(uid string, location int64) {
	//ts("GetRankPlayers", uid)
	//defer te("GetRankPlayers", uid)

	try := &proto.GetRankPlayersLocation{Uid: uid, Location: location, Start: 0, Stop: 99}
	ret := &proto.GetRankPlayersLocationResult{}

	err := cns.center.Call("Center.GetRankPlayersLocation", try, ret)
	//logger.Info("GetRankPlayers11111", try, ret, err)
	if err != nil {
		logger.Error("Error On GetRankPlayersLocation : %s, %s", uid, err.Error())
		return proto.CenterRpcError, nil
	}

		rps := &rpc.RankPlayers{}

		for _, uid := range ret.Value {
			var p rpc.PlayerInfo

			exist, err := KVQuery("player", uid, &p)
			if err != nil {
				continue
			}

			if exist {
				rp := rpc.Player{}
				rp.SetType(rpc.Player_Rank)
				rp.SetName(p.GetName())
				rp.SetUid(p.GetUid())
				rp.SetTrophy(p.GetTrophy())
				rp.SetLevel(p.GetLevel())
				rp.SetClanName(p.GetClan())
				rp.SetClanSymbol(p.GetClanSymbol())

				rps.RpsTop = append(rps.RpsTop, &rp)
			}
		}

	return ret.Code, rps
}

func GetRankClans(clan string) {
	//ts("GetRankClans", uid)
	//defer te("GetRankClans", uid)

	try := &proto.GetRankClans{Clan: clan, Start: 0, Stop: 100}
	ret := &proto.GetRankClansResult{}

	err := cns.center.Call("Center.GetRankClans", try, ret)
	//logger.Info("GetRankClans11111", try, ret, err)
	if err != nil {
		logger.Error("Error On GetRankClans : %s, %s", clan, err.Error())

		return proto.CenterRpcError, nil
	}

		rclans := &rpc.ClanInfos{}
		err = gp.Unmarshal(ret.Value, rclans)
		if err != nil {
			logger.Error("GetRankClans Unmarshal Error: %s (%v)", err.Error(), rclans)
			return proto.CenterRpcError, nil
		}

	return ret.Code, rclans
}

//通天塔排行查询
func GetTTTRankPlayers() {

	try := &proto.GetRankPlayerTTTScore{Start: 0, Stop: 99}
	ret := &proto.GetRankPlayerTTTScoreResult{}

	err := cns.center.Call("Center.GetRankPlayerTTTScore", try, ret)
	if err != nil {
		return nil
	}

		rps := &rpc.TTTRankPlayers{}
		for _, PlayerTTTScoreStructinfo := range ret.Value {
			var TTTp rpc.PlayerInfo
			fmt.Println("*****GetTTTRankPlayers****取出的数据里面的uid******score****", PlayerTTTScoreStructinfo.Id, PlayerTTTScoreStructinfo.Score)
			exist, err := KVQuery("player", PlayerTTTScoreStructinfo.Id, &TTTp)
			if err != nil {
				continue
			}
			fmt.Println("*****GetTTTRankPlayers****有对象存在**********")
			if exist {
				rp := rpc.TTTPlayer{}
				rp.SetType(rpc.TTTPlayer_Rank)
				rp.SetName(TTTp.GetName())
				rp.SetUid(TTTp.GetUid())
				rp.SetLevel(TTTp.GetLevel())
				rp.SetTrophy(TTTp.GetTrophy())
				rp.SetClanName(TTTp.GetClan())
				rp.SetClanSymbol(TTTp.GetClanSymbol())
				rp.SetTttSCoreQuery(PlayerTTTScoreStructinfo.Score)
				rp.SetRanknumberQuery(0)

				rps.RpsTop = append(rps.RpsTop, &rp)
			}

		}
		fmt.Println("*****GetTTTRankPlayers****下发TTT排行榜x**********")

}
*/

func (s *CenterService) SaveRankPlayers(req *proto.SaveRankPlayer, reply *proto.SaveRankPlayerResult) (err error) {
	logger.Info("CenterService.SaveRankPlayers")

	if req == nil {
		return nil
	}

	cns.rankMgr.RankClansLock.Lock()

	globalPlayers := &cns.rankMgr.SaveRankGlobalPlayers

	err = gp.Unmarshal(req.Value, globalPlayers)
	if err != nil {
		logger.Error("GetRankClans Unmarshal Error: %s (%v)", err.Error(), globalPlayers)
		return nil
	}

	reply.OK = true

	cns.rankMgr.RankClansLock.Unlock()

	return nil
}

func (s *CenterService) SaveRankLocationPlayers(req *proto.SaveRankLocationPlayers, reply *proto.SaveRankLocationPlayersResult) (err error) {
	logger.Info("CenterService.SaveRankLocationPlayers")

	if req == nil {
		return nil
	}

	cns.rankMgr.GlobalPlayersLock.Lock()

	locationPlayers := &rpc.RankPlayers{}
	err = gp.Unmarshal(req.Value, locationPlayers)
	if err != nil {
		logger.Error("GetRankClans Unmarshal Error: %s (%v)", err.Error(), locationPlayers)
		return nil
	}

	cns.rankMgr.SaveRankPlayersLocation[req.Location] = *locationPlayers
	reply.OK = true

	cns.rankMgr.GlobalPlayersLock.Unlock()

	return nil
}

func (s *CenterService) SaveRankClans(req *proto.SaveRankClan, reply *proto.SaveRankClanResult) (err error) {
	logger.Info("CenterService.SaveRankClans")

	if req == nil {
		return nil
	}

	cns.rankMgr.RankClansLock.Lock()
	rclans := &cns.rankMgr.SaveRankClans

	err = gp.Unmarshal(req.Value, rclans)
	if err != nil {
		logger.Error("GetRankClans Unmarshal Error: %s (%v)", err.Error(), rclans)
		return nil
	}

	reply.OK = true

	cns.rankMgr.RankClansLock.Unlock()

	return nil
}

func (s *CenterService) SaveTTTPlayers(req *proto.SaveTTTRank, reply *proto.SaveTTTRankResult) (err error) {
	logger.Info("CenterService.SaveRankTTTPlayers")

	if req == nil {
		return nil
	}

	cns.rankMgr.TTTPlayersLock.Lock()

	tttPlayers := &cns.rankMgr.SaveTTTPlayers

	err = gp.Unmarshal(req.Value, tttPlayers)
	if err != nil {
		logger.Error("GetRankClans Unmarshal Error: %s (%v)", err.Error(), tttPlayers)
		return nil
	}

	reply.OK = true

	cns.rankMgr.TTTPlayersLock.Unlock()

	return nil
}

func (s *CenterService) SaveRankMyself(req *proto.SaveMyself, reply *proto.SaveMyselfResult) error {
	logger.Info("CenterService.SaveRankMyself")

	if req == nil {
		return nil
	}

	cns.rankMgr.RankMySelfLock.Lock()

	myself := &cns.rankMgr.SaveMyself
	err := gp.Unmarshal(req.Value, myself)
	if err != nil {
		logger.Error("SaveRankMyself Unmarshal Error: %s (%v)", err.Error(), myself)
		return nil
	}

	reply.OK = true

	cns.rankMgr.RankMySelfLock.Unlock()

	return nil
}
