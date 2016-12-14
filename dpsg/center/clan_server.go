package center

import (
	"golang-project/dpsg/logger"
	"golang-project/dpsg/proto"
	"golang-project/dpsg/rpc"
	"math/rand"
	"time"

	"github.com/garyburd/redigo/redis"
	gp "github.com/golang/protobuf/proto"
	"github.com/golang/snappy"
	//"rpcplus"
	//"sync"
)

func (self *Center) saveClan(c *rpc.Clan) (err error) {
	cache := self.clancache.Get()
	defer cache.Recycle()

	key := c.GetInfo().GetName()

	buf, err := gp.Marshal(c)
	if err != nil {
		logger.Error("SaveClan Error On Marshal %s (%s, %v)", key, c, err.Error())
		return
	}

	dst := snappy.Encode(nil, buf)
	if err != nil {
		logger.Error("SaveClan Error On snappy.Encode %s (%s, %v)", key, c, err.Error())
		return
	}

	_, err = cache.Conn.Do("SET", key, dst)
	if err != nil {
		logger.Fatal("SaveClan error: %s (%s, %v)", err.Error(), key, dst)
	}

	return
}

func (self *Center) queryClan(key string, c *rpc.Clan) (err error) {
	cache := self.clancache.Get()
	defer cache.Recycle()

	var rst []byte

	rst, err = redis.Bytes(cache.Conn.Do("GET", key))
	if rst == nil {
		return
	}
	if err != nil {
		logger.Fatal("GetClanInfo error: %s (%s, %v)", err.Error(), key)
		return
	}

	rst, err = snappy.Decode(nil, rst)
	if err != nil {
		logger.Error("GetClanInfo Unmarshal Error On snappy.Decode %s (%s)", key, err.Error())
		return
	}

	err = gp.Unmarshal(rst, c)
	if err != nil {
		logger.Error("GetClanInfo Unmarshal Error On Query %s (%s)", key, err.Error())
		return
	}

	return
}

func (self *Center) initClans() error {
	cache := self.clancache.Get()
	defer cache.Recycle()

	cs, err := redis.Strings(cache.Conn.Do("KEYS", "*"))
	if err != nil {
		logger.Fatal("initClans error: %s", err.Error())
	}
	//logger.Info("InitClans1!!", cs)

	self.l.Lock()
	defer self.l.Unlock()
	for _, key := range cs {
		c := &rpc.Clan{}
		err := self.queryClan(key, c)
		if err != nil {
			logger.Fatal("initClans error:%s (%s, %v)", err.Error(), key, c)

			return err
		} else {
			//logger.Info("initClans %d:(%s, %v)", index, key, c)
		}
		self.clans[key] = LoadClanFromBuf(c)
		self.clanrank = append(self.clanrank, self.clans[key])

		//初始化杯数
		self.clans[key].UpdateTrophy(self)

		//logger.Info("initClans clanrank:%v", self.clanrank)
	}

	return nil
}

func (self *Center) CreateClan(req *proto.CreateClan, reply *proto.CreateClanResult) (err error) {
	c := &rpc.Clan{}
	err = gp.Unmarshal(req.Value, c)
	if err != nil {
		logger.Error("CreateClan Unmarshal Error: %s (%v)", err.Error(), c)
		return
	}

	self.l.RLock()
	_, exist := self.clans[c.GetInfo().GetName()]
	self.l.RUnlock()

	if exist {
		reply.Value = proto.CreateClanFailed_Exist
		logger.Info("CreateClan Failed: %v already exist!", c.GetInfo().GetName())

		return
	}

	err = self.saveClan(c)
	if err != nil {
		return
	}

	clan := LoadClanFromBuf(c)
	self.l.Lock()
	self.clans[c.GetInfo().GetName()] = clan
	self.clanrank = append(self.clanrank, clan)
	self.l.Unlock()

	reply.Value = proto.CreateClanOk
	logger.Info("CreateClan success: %v", c)

	return
}

func (self *Center) SaveClan(req *proto.SaveClan, reply *proto.SaveClanResult) (err error) {
	claninfo := &rpc.ClanInfo{}
	err = gp.Unmarshal(req.Value, claninfo)
	if err != nil {
		logger.Error("SaveClan Unmarshal Error: %s (%v)", err.Error(), claninfo)
		return
	}

	self.l.RLock()
	c, exist := self.clans[claninfo.GetName()]
	self.l.RUnlock()

	if !exist {
		reply.Value = proto.SaveClanFailed_NotFound
		logger.Info("SaveClan Failed: %s not exist!", claninfo.GetName())

		return
	}

	c.UpdateInfo(claninfo)

	err = self.saveClan(c.GetClan())
	if err != nil {
		return
	}

	reply.Value = proto.SaveClanOk
	logger.Info("SaveClan success: %v", c)

	return
}

func (self *Center) GetClan(req *proto.GetClan, reply *proto.GetClanResult) (err error) {
	self.l.RLock()
	clan, exist := self.clans[req.Name]
	self.l.RUnlock()

	if !exist {
		reply.Code = proto.ClanNotExist

		return
	}

	//clan.UpdateTrophy(self)

	buf, err := gp.Marshal(clan.GetClan())
	if err != nil {
		logger.Error("GetClan Error On Marshal (%s, %v)", err.Error(), clan.GetClan())
		return
	}

	reply.Code = proto.GetClanOk
	reply.Value = buf

	return
}

func (self *Center) GetClanInfo(req *proto.GetClan, reply *proto.GetClanInfoResult) (err error) {
	self.l.RLock()
	clan, exist := self.clans[req.Name]
	self.l.RUnlock()

	if !exist {
		reply.Code = proto.ClanNotExist
		return
	}

	//clan.UpdateTrophy(self)

	buf, err := gp.Marshal(clan.GetInfo())
	if err != nil {
		logger.Error("GetClanInfo Error On Marshal (%s, %v)", err.Error(), clan.GetInfo())
		return
	}

	reply.Code = proto.GetClanOk
	reply.Value = buf

	return
}

func (self *Center) GetClanPlayer(req *proto.GetClanPlayer, reply *proto.GetClanPlayerResult) (err error) {
	self.l.RLock()
	clan, exist := self.clans[req.Name]
	self.l.RUnlock()
	if !exist {
		reply.Code = proto.ClanNotExist
		return
	}

	player := clan.GetPlayer(req.Uid)
	if player == nil {
		reply.Code = proto.PlayerNotExist
		return
	}

	buf, err := gp.Marshal(player)
	if err != nil {
		logger.Error("GetClanPlayer Error On Marshal (%s, %v)", err.Error(), *player)
		return
	}

	reply.Code = proto.GetClanPlayerOk
	reply.Value = buf

	return
}

func (self *Center) RandomGetClans(req *proto.RandomGetClans, reply *proto.RandomGetClansResult) (err error) {
	slice := self.clanrank

	lenth := len(self.clanrank)
	if req.Num < lenth {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))

		i := r.Intn(lenth - req.Num + 1)

		slice = self.clanrank[i : i+req.Num]
	}

	clans := &rpc.ClanInfos{}
	for _, clan := range slice {
		clans.Infos = append(clans.Infos, clan.GetInfo())
	}

	buf, err := gp.Marshal(clans)
	if err != nil {
		logger.Error("RandomGetClans Error On Marshal (%s, %v)", err.Error(), clans)
		return
	}

	reply.Code = proto.GetClanOk
	reply.Value = buf

	return
}

func (self *Center) TryJoinClan(req *proto.JoinClan, reply *proto.JoinClanResult) (err error) {
	self.l.RLock()
	clan, exist := self.clans[req.Name]
	self.l.RUnlock()

	if !exist {
		reply.Value = proto.ClanNotExist

		return nil
	}

	cp := &rpc.Player{}

	err = gp.Unmarshal(req.Value, cp)
	if err != nil {
		logger.Error("GetClanInfo Unmarshal Error On Query %s (%s)", req.Name, err.Error())
		return
	}

	reply.Value = clan.TryJoinClan(cp)
	if reply.Value == proto.JoinClanOk {
		err = self.saveClan(clan.Clan)
		if err != nil {
			return err
		}

		//更新杯数
		clan.UpdateTrophy(self)
	}

	return nil
}

func (self *Center) TryLeaveClan(req *proto.LeaveClan, reply *proto.LeaveClanResult) (err error) {
	cache := self.clancache.Get()
	defer cache.Recycle()

	self.l.RLock()
	clan, exist := self.clans[req.Name]
	self.l.RUnlock()

	if !exist {
		reply.Value = proto.ClanNotExist

		return nil
	}

	reply.Value = clan.TryLeaveClan(req.PUid)
	if reply.Value == proto.LeaveClanOk {
		err = self.saveClan(clan.Clan)
		if err != nil {
			return err
		}

		//更新杯数
		clan.UpdateTrophy(self)
	} else {
		//删除redis
		if reply.Value == proto.DeleteClanOK {
			logger.Info("delete clan from redis and map!")
			_, err = cache.Conn.Do("DEL", req.Name)
			if err != nil {
				logger.Error("delete clan error: %s, %s", err.Error(), req.Name)
			}
			//删除map表
			delete(self.clans, req.Name)
			//删除切片数组
			for i, v := range self.clanrank {
				if v.GetInfo().GetName() == req.Name {
					self.clanrank = append(self.clanrank[:i], self.clanrank[i+1:]...)
				}
			}
		}
	}

	return nil
}

func (self *Center) TryKickPlayer(req *proto.KickPlayer, reply *proto.KickPlayerResult) (err error) {
	self.l.RLock()
	clan, exist := self.clans[req.CName]
	self.l.RUnlock()

	if !exist {
		reply.Value = proto.ClanNotExist

		return nil
	}

	reply.Value, reply.Power = clan.TryKickPlayer(req.Uid, req.TarUid)
	if reply.Value == proto.KickPlayerOk {
		err = self.saveClan(clan.Clan)
		if err != nil {
			return err
		}

		//更新杯数
		clan.UpdateTrophy(self)
	}

	return nil
}

func (self *Center) TryAppointPlayer(req *proto.AppointPlayer, reply *proto.AppointPlayerResult) (err error) {
	self.l.RLock()
	clan, exist := self.clans[req.CName]
	self.l.RUnlock()

	if !exist {
		reply.Value = proto.ClanNotExist

		return nil
	}

	reply.Value, reply.OldPower = clan.TryAppointPlayer(req.Uid, req.TarUid, rpc.Player_ClanPower(req.Power))
	if reply.Value == proto.AppointPlayerOk {
		err = self.saveClan(clan.Clan)
		if err != nil {
			return err
		}
	}

	return nil
}

func (self *Center) SearchClan(req *proto.SearchClan, reply *proto.SearchClanResult) (err error) {
	cache := self.clancache.Get()
	defer cache.Recycle()

	cs, err := redis.Strings(cache.Conn.Do("KEYS", "*"+req.Key+"*"))
	if err != nil {
		logger.Fatal("SearchClan error: %s", err.Error())
		return
	}

	clans := &rpc.ClanInfos{}

	self.l.RLock()
	for _, cname := range cs {
		if clan, exist := self.clans[cname]; exist {
			clans.Infos = append(clans.Infos, clan.GetInfo())
		}
	}
	self.l.RUnlock()

	buf, err := gp.Marshal(clans)
	if err != nil {
		logger.Error("SearchClan Error On Marshal (%s, %v)", err.Error(), clans)
		return
	}

	reply.Code = proto.SearchClanOk
	reply.Value = buf

	return nil
}

func (self *Center) NotifyGetDonate(req *proto.NotifyGetDonate, reply *proto.NotifyGetDonateResult) (err error) {
	logger.Info("Center:NotifyGetDonate:%s, %s", req.Uid, req.Name)

	for _, conn := range self.cnss {
		conn.Go("CenterService.NotifyGetDonate", req, reply, nil)
	}

	return nil
}

func (self *Center) NotifyUpdateClanInfo(req *proto.NotifyUpdateClanInfo, reply *proto.NotifyUpdateClanInfoResult) (err error) {
	logger.Info("Center:NotifyUpdateClanInfo:%s, %s", req.Uid, req.CName)

	for _, conn := range self.cnss {
		conn.Go("CenterService.NotifyUpdateClanInfo", req, reply, nil)
	}

	return nil
}
