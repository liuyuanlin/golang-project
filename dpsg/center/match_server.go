package center

import (
	"golang-project/dpsg/logger"
	"golang-project/dpsg/proto"
	"golang-project/dpsg/timer"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const (
	MATCH_SEG_MIN = 0   //最小段
	MATCH_SEG_MAX = 50  //最大
	MATCH_SEG_NUM = 4   //从几个段中进行随机
	MATCH_SEG_LEN = 200 //每段长度
)

func getTrophySegment(trophy uint32) int64 {
	seg := int64(math.Floor(float64(trophy) / MATCH_SEG_LEN))
	if seg > MATCH_SEG_MAX {
		seg = MATCH_SEG_MAX
	}
	return seg
}

func getTrophySegmentStr(trophy uint32) string {
	return strconv.FormatInt(getTrophySegment(trophy), 10)
}

func (self *Center) getPlayerTrophy(uid string) uint32 {
	trophy, err := self.zscore("rank", "player", uid)
	//trophy, err := self.getInt("trophy", uid)
	if err != nil {
		return 0
	}

	return trophy
}

func (self *Center) addMatchPlayer(uid string, trophy uint32) {
	key := getTrophySegmentStr(trophy)

	self.sadd("match", key, uid)

	logger.Info("addMatchPlayer:", key, uid) //test
}

func (self *Center) removeMatchPlayer(uid string, trophy uint32) {
	key := getTrophySegmentStr(trophy)

	self.srem("match", key, uid)

	logger.Info("removeMatchPlayer:", key, uid) //test
}

func (self *Center) isOnline(uid string) bool {
	self.l.RLock()
	defer self.l.RUnlock()

	ret, exist := self.POnline[uid]
	if exist {
		return ret
	}
	return false
}

func (self *Center) hasShield(uid string) bool {
	return self.exists("shield", uid)
}

func (self *Center) setShield(uid string, start_time uint32, total_time uint32) {
	st := strconv.Itoa(int(start_time))
	tt := strconv.Itoa(int(total_time))

	//先删除，防止过期时间出错
	self.del("shield", uid)
	self.setString("shield", uid, st+"|"+tt)
	//设置过期时间，加个10秒的误差时间
	self.setexpire("shield", uid, strconv.FormatUint(uint64(total_time)+uint64(10), 10))

	return
}

func (self *Center) getShield(uid string) (start_time uint32, total_time uint32) {
	str, err := self.getString("shield", uid)
	if err != nil {
		return 0, 0
	}

	ret := strings.Split(str, "|")

	s_t, err := strconv.ParseUint(ret[0], 10, 32)
	if err != nil {
		return 0, 0
	}

	t_t, err := strconv.ParseUint(ret[1], 10, 32)
	if err != nil {
		return 0, 0
	}

	return uint32(s_t), uint32(t_t)
}

func (self *Center) delShield(uid string) {
	self.del("shield", uid)
}

func (self *Center) createShield(uid string, start_time uint32, total_time uint32) bool {
	now_time := time.Now().Unix()

	//已过期则直接返回
	if int64(start_time+total_time) <= now_time {
		return false
	}

	left_time := int64(start_time+total_time) - now_time

	s_t, t_t := self.getShield(uid)
	l_t := int64(s_t+t_t) - now_time

	if left_time < l_t {
		logger.Error("createShield:left_time(%d) < l_t(%d)", left_time, l_t)
		return false
	}

	//开始定时器
	tm := timer.NewTimer(time.Duration(left_time) * time.Second)
	tm.Start(
		func() {
			ok := self.destoryShield(uid)
			if !ok {
				return
			}

			if !self.isOnline(uid) {
				self.addMatchPlayer(uid, self.getPlayerTrophy(uid))
			}
		},
	)

	logger.Info("createShield:%v", *tm)

	//定时期管理起来
	self.l.Lock()
	self.shields[uid] = tm
	self.l.Unlock()

	//保存护盾信息
	self.setShield(uid, start_time, total_time)

	return true
}

func (self *Center) destoryShield(uid string) bool {
	self.l.Lock()
	defer self.l.Unlock()

	tm, exist := self.shields[uid]
	if exist && tm != nil {
		//若定时期存在则销毁
		tm.Stop()

		self.shields[uid] = nil
	}

	//删除护盾信息
	self.delShield(uid)

	return true
}

//*RPC*////////////////////////////////////////////////////////////////////////////
func (self *Center) SetPlayerTrophy(req *proto.SetTrophy, reply *proto.SetTrophyResult) (err error) {
	ts("Center:SetPlayerTrophy", req.Uid, req.Trophy)
	defer te("Center:SetPlayerTrophy", req.Uid, req.Trophy)

	//更新玩家杯数
	if err = self.zadd("rank", "player", req.Uid, req.Trophy); err != nil {
		//fmt.Println("全球玩家的杯数改变了", req.Uid, req.Trophy)
		return
	}

	// 增加玩家区域字段，方便查找区域玩家排名
	if err = self.zadd("rank", "PlayerLocation"+strconv.Itoa(int(req.Location)), req.Uid, req.Trophy); err != nil {
		//	fmt.Println("区域玩家的杯数改变了", req.Uid, strconv.Itoa(int(req.Location)), req.Trophy)
		return
	}

	//更新玩家所在帮会的杯数，没办法，只有遍历
	self.l.RLock()
	for _, c := range self.clans {
		if c.GetPlayer(req.Uid) != nil {
			c.UpdateTrophy(self)
			break
		}
	}
	self.l.RUnlock()

	return
}

func (self *Center) NotifyOnline(req *proto.NotifyOnline, reply *proto.NotifyOnlineResult) (err error) {
	ts("Center:NotifyOnline", req.Uid)
	defer te("Center:NotifyOnline", req.Uid)

	self.l.Lock()
	self.POnline[req.Uid] = true
	self.l.Unlock()

	//如果无盾那么就从匹配库里移去
	if !self.hasShield(req.Uid) {
		self.removeMatchPlayer(req.Uid, self.getPlayerTrophy(req.Uid))
	}

	return nil
}

func (self *Center) NotifyOffline(req *proto.NotifyOffline, reply *proto.NotifyOfflineResult) (err error) {
	ts("Center:NotifyOffline", req.Uid)
	defer te("Center:NotifyOffline", req.Uid)

	self.l.Lock()
	delete(self.POnline, req.Uid)
	self.l.Unlock()

	//如果无盾那么就放入到匹配库里去，新手完成则放入库里面
	if !self.hasShield(req.Uid) && req.GuideFinish {
		self.addMatchPlayer(req.Uid, self.getPlayerTrophy(req.Uid))
	}

	return nil
}

func (self *Center) AddPlayerShield(req *proto.AddShield, reply *proto.AddShieldResult) (err error) {
	ts("Center:AddPlayerShield", req.Uid, req.StartTime, req.TotalTime)
	defer te("Center:AddPlayerShield", req.Uid, req.StartTime, req.TotalTime)

	reply.Ok = self.createShield(req.Uid, req.StartTime, req.TotalTime)
	if !reply.Ok {
		return nil
	}

	if !self.isOnline(req.Uid) {
		self.removeMatchPlayer(req.Uid, self.getPlayerTrophy(req.Uid))
	}

	return nil
}

func (self *Center) RemovePlayerShield(req *proto.RemoveShield, reply *proto.RemoveShieldResult) (err error) {
	ts("Center:RemovePlayerShield", req.Uid)
	defer te("Center:RemovePlayerShield", req.Uid)

	ok := self.destoryShield(req.Uid)
	if !ok {
		return nil
	}

	if !self.isOnline(req.Uid) {
		self.addMatchPlayer(req.Uid, self.getPlayerTrophy(req.Uid))
	}

	return nil
}

func (self *Center) TryMatch(req *proto.Match, reply *proto.MatchResult) error {
	l := make([]string, MATCH_SEG_NUM)
	total := 0

	seg := getTrophySegment(req.Trophy)

	if seg > MATCH_SEG_MAX {
		seg = MATCH_SEG_MAX
	}

	for i, j := seg, seg; i >= MATCH_SEG_MIN || j <= MATCH_SEG_MAX; i, j = i-1, j+1 {
		if i >= MATCH_SEG_MIN {
			key := strconv.FormatInt(i, 10)
			num := self.scard("match", key)
			if num > 0 {
				l[total] = key
				total++
				if total >= MATCH_SEG_NUM {
					break
				}
			}
		}
		if j <= MATCH_SEG_MAX {
			key := strconv.FormatInt(j, 10)
			num := self.scard("match", key)
			if num > 0 {
				l[total] = key
				total++
				if total >= MATCH_SEG_NUM {
					break
				}
			}
		}
	}

	if total == 0 {
		return nil
	}

	//只尝试5次
	for i := 0; i < 5; i++ {
		key := l[rand.Intn(total)]

		uid, err := self.srandmember("match", key)

		if uid == "" || req.Except == uid || pCenterChallengeService.IsInTheChallenge(uid) {
			continue
		}

		reply.Uid = uid
		logger.Info("TryMatch success:(%s, %s)", key, uid) //test

		return err
	}

	reply.Uid = ""
	return nil
}
