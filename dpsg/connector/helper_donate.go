package connector

import (
	"golang-project/dpsg/lockclient"
	"golang-project/dpsg/logger"
	"golang-project/dpsg/proto"
	"golang-project/dpsg/rpc"
)

const (
	MAX_DONATE_COUNT = uint32(5)
)

//可捐献兵数量
func donate_get_freecount(info *rpc.DonateInfo, uid string) uint32 {
	total := uint32(0)
	//logger.Info("donate_get_freecount1:%v  -  %s", info.Pinfo, uid)
	for _, pinfo := range info.Pinfo {
		if pinfo.GetUid() == uid {
			for _, c := range pinfo.Characters {
				total += c.GetCount()
			}
			break
		}
	}
	//logger.Info("donate_get_freecount2:%d", total)
	if total > MAX_DONATE_COUNT {
		return 0
	}

	return MAX_DONATE_COUNT - total
}

//获取玩家捐兵信息、经验值等
func donate_get_playerinfo(info *rpc.DonateInfo, p *player, tarUid string) *rpc.C2SDonateResult {
	c2sDonateResult := &rpc.C2SDonateResult{}
	c2sDonateResult.SetTarUid(tarUid)
	c2sDonateResult.SetSrcUid(p.GetUid())
	c2sDonateResult.SetExp(0)

	for _, pinfo := range info.Pinfo {
		if pinfo.GetUid() == p.GetUid() {
			for _, c := range pinfo.Characters {
				cfg := GetCharacterCfgByTypeId(c.GetType(), 1)
				if cfg != nil {
					c2sDonateResult.SetExp(c2sDonateResult.GetExp() + cfg.HousingSpace*c.GetCount())
					c2sDonateResult.Characters = append(c2sDonateResult.Characters, c)
				}
			}

			//若捐失败返回
			if c2sDonateResult.GetExp() == 0 {
				return nil
			}

			return c2sDonateResult
		}
	}

	return nil
}

func donate_insert_playerinfo(info *rpc.DonateInfo, p *player, char *rpc.Character) {
	//logger.Info("你的%d级%d被你捐出去了！", char.GetLevel(), char.GetType())
	notifyDonate := &rpc.NotifyDonateOff{}
	notifyDonate.Character = char
	WriteResult(p.conn, notifyDonate)

	for _, pInfo := range info.Pinfo {
		if pInfo.GetUid() == p.GetUid() {
			for _, c := range pInfo.Characters {
				if c.GetType() == char.GetType() && c.GetLevel() == char.GetLevel() {
					c.SetCount(c.GetCount() + char.GetCount())
					return
				}
			}

			pInfo.Characters = append(pInfo.Characters, char)
			return
		}
	}

	pInfo := &rpc.PlayerDonateInfo{}
	pInfo.SetUid(p.GetUid())
	pInfo.Characters = append(pInfo.Characters, char)

	info.Pinfo = append(info.Pinfo, pInfo)
}

func donate_insert_char(info *rpc.DonateInfo, p *player, char *rpc.Character) uint32 {
	//ts("donate_insert_char:%v, %s, %v", *info, p.GetUid(), *char)
	//defer te("donate_insert_char:%v, %s, %v", *info, p.GetUid(), *char)

	cfg := GetCharacterCfgByTypeId(char.GetType(), 1)
	if cfg == nil {
		logger.Error("donate_insert_char:cfg == nil, type(%d)", char.GetType())
		return 0
	}

	if info.GetTotalSpace() < info.GetUsedSpace() {
		logger.Error("donate_insert_char:info.GetTotalSpace(%d) < info.GetUsedSpace(%d)", info.GetTotalSpace(), info.GetUsedSpace())
		return 0
	}

	//根据空间所得的剩余个数
	free_space := uint32((info.GetTotalSpace() - info.GetUsedSpace()) / cfg.HousingSpace)
	//根据捐献者的名额得到的剩余个数
	free_count := donate_get_freecount(info, p.GetUid())
	//取小的那个
	count := uint32(0)
	if free_space > free_count {
		count = free_count
	} else {
		count = free_space
	}

	logger.Info("donate_insert_char:free_space(%d), free_count(%d), count(%d)", free_space, free_count, count)
	if count == 0 {
		return 0
	}

	//若超出可捐献上限
	if char.GetCount() > count {
		logger.Info("donate_insert_char:char.SetCount(%d)111", count)
		char.SetCount(count)
	}

	for i := uint32(1); i <= char.GetCount(); i++ {
		c := p.v.barrack_TroopHousingPopCharacter(char.GetType())
		if c == nil {
			if i == 1 {
				logger.Error("donate_insert_char:c(index:%d) == nil", i)
				return 0
			} else {
				char.SetCount(i)
				logger.Info("donate_insert_char:char.SetCount(%d)222", i)
				break
			}
		}
	}

	donate_insert_playerinfo(info, p, char)

	for _, c := range info.Characters {
		if c.GetType() == char.GetType() && c.GetLevel() == char.GetLevel() {
			c.SetCount(c.GetCount() + char.GetCount())

			info.SetUsedSpace(info.GetUsedSpace() + cfg.HousingSpace*char.GetCount())

			return cfg.HousingSpace * char.GetCount()
		}
	}

	info.Characters = append(info.Characters, char)

	info.SetUsedSpace(info.GetUsedSpace() + cfg.HousingSpace*char.GetCount())

	return cfg.HousingSpace * char.GetCount()
}

func Donate(p *player, c2sDonate rpc.C2SDonate) {
	//加锁
	lid := GenLockMessage(cns.GetServerId(), proto.MethodAddDonateInfo, 0)
	if !lockclient.WaitLockGet("donate", c2sDonate.GetUid(), lid) {
		logger.Error("Donate:WaitLockGet Failed!!")
		return
	}
	defer lockclient.TryUnlock("donate", c2sDonate.GetUid(), lid)

	donateInfo := rpc.DonateInfo{}

	exist, err := KVQueryExt("donate", c2sDonate.GetUid(), &donateInfo)
	if err != nil {
		return
	}

	if exist {

		total := uint32(0)

		//modify by wyc 2014-2-10 exp bug
		nNumberPer := uint32(0)

		for _, c := range c2sDonate.Characters {
			nNumberPer = donate_insert_char(&donateInfo, p, c)
			total += nNumberPer
			if nNumberPer > 0 {
				cfg := GetCharacterCfgByTypeId(c.GetType(), 1)
				nExp := cfg.HousingSpace * c.GetCount()
				p.AddExp(nExp)
			}
		}
		logger.Info("Donate:total(%d), %v", total, donateInfo)
		if total == 0 {
			return
		}

		c2sDonateResult := donate_get_playerinfo(&donateInfo, p, c2sDonate.GetUid())
		if c2sDonateResult == nil {
			return
		}

		//logger.Info("Donate:donate_get_playerinfo:%v", *c2sDonateResult)
		logger.Info("Donate:the add exp:%d", (*c2sDonateResult).GetExp())

		//存入数据库
		result, err := KVWriteExt("donate", c2sDonate.GetUid(), &donateInfo)
		//logger.Info("Donate - KVWrite:%v", donateInfo, err, result)
		if err != nil || result == false {
			return
		}

		p.SetDonateNum(p.GetDonateNum() + total)

		//广播更新捐兵信息
		s2cMsg := rpc.S2CDonateUpdate{}
		s2cMsg.SetUid(c2sDonate.GetUid())
		s2cMsg.SetInfo(donateInfo.GetInfo())
		s2cMsg.SetUsedSpace(donateInfo.GetUsedSpace())
		s2cMsg.SetTotalSpace(donateInfo.GetTotalSpace())
		s2cMsg.DonateResult = c2sDonateResult //这一条是用来暂存的，用于该捐兵玩家下线后再上线的数据恢复
		UpdateDonateMsg(p.GetUid(), s2cMsg)

		//通知接收方玩家取数据
		req := &proto.NotifyGetDonate{Uid: c2sDonate.GetUid(), Name: p.GetName()}
		rst := &proto.NotifyGetDonateResult{}
		cns.center.Go("Center.NotifyGetDonate", req, rst, nil)
	}
}
