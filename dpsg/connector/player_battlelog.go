package connector

import (
	"golang-project/dpsg/logger"
	"golang-project/dpsg/rpc"
	"strconv"
)

func (self *player) HasBattleLog(bid uint64) bool {
	for _, id := range self.DLogId {
		if id == 0 {
			continue
		}

		if id == bid {
			return true
		}
	}
	return false
}

func (self *player) CanRevenge(bid uint64) bool {
	for _, log := range self.d_logs.Log {
		if log.GetBid() == bid {
			if log.GetRevstate() == rpc.BattleLog_Revenged {
				return false
			}

			return true
		}
	}
	return false
}

func (self *player) SetBattleLogRevenged(bid uint64) {
	//把未阅读的标记为已阅读状态
	for _, log := range self.d_logs.Log {
		if log.GetBid() == bid {
			log.SetRevstate(rpc.BattleLog_Revenged)

			KVWriteExt("battlelog", strconv.FormatUint(bid, 16), log)

			update := rpc.UpdateBattleLog{}
			update.SetBid(bid)
			update.SetRevstate(rpc.BattleLog_Revenged)
			WriteResult(self.conn, &update)

			return
		}
	}
}

// 防御结果，存储战利品，金币和食物
func (self *player) InsertBattleLog(log *rpc.BattleLog, rep *rpc.BattleReplay) {
	bid := GenBattleId(cns.GetServerId())

	// 删除一条超过上限的部分数据
	if len(self.DLogId) >= MaxLogNumber {
		self.DLogId = append(self.DLogId[1:], bid)
	} else {
		self.DLogId = append(self.DLogId, bid)
	}
	log.SetBid(bid)

	rid := GenReplayId(cns.GetServerId())

	// 删除一条超过上限的部分数据
	if len(self.DRepId) >= MaxRepNumber {
		self.DRepId = append(self.DRepId[1:], rid)
		//todo .真正删除replay
	} else {
		self.DRepId = append(self.DRepId, rid)
	}
	log.SetRid(rid)

	if self.d_logs == nil {
		self.d_logs = &rpc.BattleLogs{}
		self.d_logs.SetType(rpc.BattleLogs_Defence)
	}
	self.d_logs.Log = append(self.d_logs.Log, log)

	//logger.Info("InsertBattleRep:%v, %s\n%v", self.DRepId, strconv.FormatUint(rid, 16), rep)

	KVWriteExt("replay", strconv.FormatUint(rid, 16), rep)

	//logger.Info("InsertBattleLog:%v, %s\n%v", self.DLogId, strconv.FormatUint(bid, 16), log)

	KVWriteExt("battlelog", strconv.FormatUint(bid, 16), log)
}

// 攻击结果，存储战利品，金币和食物
func (self *player) InsertAttackLog(log *rpc.BattleLog, rep *rpc.BattleReplay) {
	bid := GenBattleId(cns.GetServerId())

	// 删除一条超过上限的部分数据
	if len(self.ALogId) >= MaxLogNumber {
		self.ALogId = append(self.ALogId[1:], bid)
	} else {
		self.ALogId = append(self.ALogId, bid)
	}
	log.SetBid(bid)

	rid := GenReplayId(cns.GetServerId())

	// 删除一条超过上限的部分数据
	if len(self.ARepId) >= MaxRepNumber {
		self.ARepId = append(self.ARepId[1:], rid)
		//todo .真正删除replay
	} else {
		self.ARepId = append(self.ARepId, rid)
	}
	log.SetRid(rid)

	if self.a_logs == nil {
		self.a_logs = &rpc.BattleLogs{}
		self.a_logs.SetType(rpc.BattleLogs_Attack)
	}
	self.a_logs.Log = append(self.a_logs.Log, log)

	//logger.Info("InsertBattleRep:%v, %s\n%v", self.DRepId, strconv.FormatUint(rid, 16), rep)

	KVWriteExt("replay", strconv.FormatUint(rid, 16), rep)

	//logger.Info("InsertBattleLog:%v, %s\n%v", self.DLogId, strconv.FormatUint(bid, 16), log)

	KVWriteExt("attacklog", strconv.FormatUint(bid, 16), log)

	WriteResult(self.conn, self.a_logs)
}

// 好友演习结果，存储战利品，金币和食物
func (self *player) InsertFriendExeciseLog(log *rpc.BattleLog, rep *rpc.BattleReplay) {
	bid := GenBattleId(cns.GetServerId())

	// 删除一条超过上限的部分数据
	if len(self.FLogId) >= MaxLogNumber {
		self.FLogId = append(self.FLogId[1:], bid)
	} else {
		self.FLogId = append(self.FLogId, bid)
	}
	log.SetBid(bid)

	rid := GenReplayId(cns.GetServerId())

	// 删除一条超过上限的部分数据
	if len(self.FRepId) >= MaxRepNumber {
		self.FRepId = append(self.FRepId[1:], rid)
		//todo .真正删除replay
	} else {
		self.FRepId = append(self.FRepId, rid)
	}
	log.SetRid(rid)

	if self.f_logs == nil {
		self.f_logs = &rpc.BattleLogs{}
		self.f_logs.SetType(rpc.BattleLogs_FriendExecise)
	}
	self.f_logs.Log = append(self.f_logs.Log, log)

	//logger.Info("InsertBattleRep:%v, %s\n%v", self.DRepId, strconv.FormatUint(rid, 16), rep)

	KVWriteExt("replay", strconv.FormatUint(rid, 16), rep)

	//logger.Info("InsertBattleLog:%v, %s\n%v", self.DLogId, strconv.FormatUint(bid, 16), log)

	KVWriteExt("friendexeciselog", strconv.FormatUint(bid, 16), log)
	KVWriteExt("friendexecise", self.GetUid(), self.FriendsExeciseInfo)
	ts("FriendExecise save", len(self.f_logs.Log), bid, len(self.FLogId))
}

func (self *player) InitBattleLog() {
	self.d_logs = &rpc.BattleLogs{}
	self.d_logs.SetType(rpc.BattleLogs_Defence)

	logger.Info("InitBattleLog:%v", self.DLogId)

	for _, bid := range self.DLogId {
		if bid == 0 {
			continue
		}

		var def_log rpc.BattleLog

		//logger.Info("InitBattleLog111:%s", strconv.FormatUint(bid, 16))
		exist, err := KVQueryExt("battlelog", strconv.FormatUint(bid, 16), &def_log)
		if err != nil {
			continue
		}

		if exist {
			bHasReplay := false
			for _, rid := range self.DRepId {
				if def_log.GetRid() == rid {
					bHasReplay = true
					break
				}
			}
			if !bHasReplay {
				def_log.Rid = nil
			}

			self.d_logs.Log = append(self.d_logs.Log, &def_log)
			//logger.Info("InitBattleLog222:%s, %v", strconv.FormatUint(bid, 16), def_log)
		}
	}

	WriteResult(self.conn, self.d_logs)

	//把未阅读的标记为已阅读状态
	for _, log := range self.d_logs.Log {
		if log.GetState() == rpc.BattleLog_UnRead {
			log.SetState(rpc.BattleLog_Read)

			KVWriteExt("battlelog", strconv.FormatUint(log.GetBid(), 16), log)
		}
	}
}

func (self *player) InitAttackLog() {
	self.a_logs = &rpc.BattleLogs{}
	self.a_logs.SetType(rpc.BattleLogs_Attack)

	logger.Info("InitAttackLog:%v", self.ALogId)

	for _, bid := range self.ALogId {
		if bid == 0 {
			continue
		}

		var atc_log rpc.BattleLog

		//logger.Info("InitBattleLog111:%s", strconv.FormatUint(bid, 16))
		exist, err := KVQueryExt("attacklog", strconv.FormatUint(bid, 16), &atc_log)
		if err != nil {
			continue
		}

		if exist {
			bHasReplay := false
			for _, rid := range self.ARepId {
				if atc_log.GetRid() == rid {
					bHasReplay = true
					break
				}
			}
			if !bHasReplay {
				atc_log.Rid = nil
			}

			self.a_logs.Log = append(self.a_logs.Log, &atc_log)
			//logger.Info("InitBattleLog222:%s, %v", strconv.FormatUint(bid, 16), def_log)
		}
	}

	WriteResult(self.conn, self.a_logs)
}

func (self *player) InitFriendExeciseLog(bSendUpdate bool) {
	var f rpc.FriendsExeciseInfo
	KVQueryExt("friendexecise", self.GetUid(), &f)
	self.FriendsExeciseInfo = &f

	self.f_logs = &rpc.BattleLogs{}
	self.f_logs.SetType(rpc.BattleLogs_FriendExecise)

	logger.Info("InitFriendExeciseLog:%v", self.FLogId)

	for _, bid := range self.FLogId {
		if bid == 0 {
			continue
		}

		var friend_log rpc.BattleLog

		//logger.Info("InitBattleLog111:%s", strconv.FormatUint(bid, 16))
		exist, err := KVQueryExt("friendexeciselog", strconv.FormatUint(bid, 16), &friend_log)
		if err != nil {
			continue
		}

		if exist {
			bHasReplay := false
			for _, rid := range self.FRepId {
				if friend_log.GetRid() == rid {
					bHasReplay = true
					break
				}
			}
			if !bHasReplay {
				friend_log.Rid = nil
			}

			self.f_logs.Log = append(self.f_logs.Log, &friend_log)
			//logger.Info("InitBattleLog222:%s, %v", strconv.FormatUint(bid, 16), def_log)
		}
	}

	if bSendUpdate == true {
		WriteResult(self.conn, self.f_logs)
	}
}
