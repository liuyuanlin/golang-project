package connector

import (
	"golang-project/dpsg/lockclient"
	"golang-project/dpsg/logger"
	"golang-project/dpsg/proto"
	"golang-project/dpsg/rpc"
	"time"
)

func (v *village) castle_UpdateClanInfo(clan_name string, clan_symbol uint32) {
	obj := v.buildings_Get(rpc.BuildingId_AllianceCastle, 0)
	if obj == nil {
		return
	}

	castle := obj.(*rpc.AllianceCastle)
	castle.SetClanName(clan_name)
	castle.SetClanSymbol(clan_symbol)

	return
}

func castle_GetHousingSpace(castle *rpc.AllianceCastle) uint32 {
	cfg := GetBuildingCfgByTypeId(rpc.BuildingId_AllianceCastle, castle.GetLevel())
	if cfg == nil {
		return 0
	}

	return cfg.HousingSpace
}

func castle_GetHousingUsedSpace(castle *rpc.AllianceCastle) (s uint32) {
	if castle.Characters == nil {
		return 0
	}

	for _, c := range castle.Characters {
		cfg := GetCharacterCfgByTypeId(c.GetType(), 1)
		s += (cfg.HousingSpace * c.GetCount())
	}

	return s
}

func castle_InsertCharacter(castle *rpc.AllianceCastle, char *rpc.Character) bool {
	for _, c := range castle.Characters {
		if c.GetType() == char.GetType() && c.GetLevel() == char.GetLevel() {
			c.SetCount(c.GetCount() + char.GetCount())

			return true
		}
	}

	castle.Characters = append(castle.Characters, char)

	return true
}

func (v *village) castle_PopCharacters(chars []*rpc.Character) bool {
	if len(chars) == 0 {
		return false
	}

	obj := v.buildings_Get(rpc.BuildingId_AllianceCastle, 0)
	if obj == nil {
		logger.Error("castle_PopCharacters:obj == nil")
		return false
	}

	castle := obj.(*rpc.AllianceCastle)

	if len(castle.Characters) == 0 {
		return false
	}

	hasError := false
	for _, c := range chars {
		bFound := false
		for _, cc := range castle.Characters {
			if c.GetType() == cc.GetType() && c.GetLevel() == cc.GetLevel() {
				bFound = true
				if c.GetCount() > cc.GetCount() {
					hasError = true
					logger.Error("castle_PopCharacters:c.GetCount(%d) > cc.GetCount(%d)", c.GetCount(), cc.GetCount())
				} else {
					cc.SetCount(cc.GetCount() - c.GetCount())
				}
				break
			}
		}
		if !bFound {
			hasError = true
			logger.Error("castle_PopCharacters:not found(type:%d,level:%d)", c.GetType(), c.GetLevel())
		}
	}

	tempCharacters := make([]*rpc.Character, 0, 1)

	for _, c := range castle.Characters {
		if c.GetCount() > 0 {
			tempCharacters = append(tempCharacters, c)
		}
	}
	castle.Characters = tempCharacters

	return hasError
}
func (v *village) castle_PopSpells(spells []*rpc.Spell) bool {
	if len(spells) == 0 {
		return false
	}

	obj := v.buildings_Get(rpc.BuildingId_SpellForge, 0)
	if obj == nil {
		logger.Error("castle_PopSpells:obj == nil")
		return false
	}

	castle := obj.(*rpc.SpellForge)

	if len(castle.Spell) == 0 {
		return false
	}

	hasError := false
	for _, c := range spells {
		bFound := false
		for _, cc := range castle.Spell {
			if c.GetType() == cc.GetType() && c.GetLevel() == cc.GetLevel() {
				bFound = true
				if c.GetCount() > cc.GetCount() {
					hasError = true
					logger.Error("castle_PopSpells:c.GetCount(%d) > cc.GetCount(%d)", c.GetCount(), cc.GetCount())
				} else {
					cc.SetCount(cc.GetCount() - c.GetCount())
				}
				break
			}
		}
		if !bFound {
			hasError = true
			logger.Error("castle_PopSpells:not found(type:%d,level:%d)", c.GetType(), c.GetLevel())
		}
	}

	tempSpells := make([]*rpc.Spell, 0, 1)

	for _, c := range castle.Spell {
		if c.GetCount() > 0 {
			tempSpells = append(tempSpells, c)
		}
	}
	castle.Spell = tempSpells

	return hasError
}
func (v *village) castle_GetClanForce() *rpc.ClanForce {
	obj := v.buildings_Get(rpc.BuildingId_AllianceCastle, 0)
	if obj == nil {
		return nil
	}

	castle := obj.(*rpc.AllianceCastle)

	if len(castle.Characters) == 0 {
		return nil
	}

	force := &rpc.ClanForce{}
	force.SetSymbol(v.p.GetClanSymbol())
	force.Char = castle.Characters

	return force
}

func (v *village) castle_GetCapacity() (uint32, uint32) {
	obj := v.buildings_Get(rpc.BuildingId_AllianceCastle, 0)
	if obj == nil {
		return 0, 0
	}

	castle := obj.(*rpc.AllianceCastle)

	return castle_GetHousingUsedSpace(castle), castle_GetHousingSpace(castle)
}

/////////////////////////////////////////////////
func (v *village) getDonateFromDB(notify *proto.NotifyGetDonate) {
	donateInfo := rpc.DonateInfo{}

	exist, err := KVQueryExt("donate", v.p.GetUid(), &donateInfo)
	if err != nil {
		return
	}

	if exist {
		obj := v.buildings_Get(rpc.BuildingId_AllianceCastle, 0)
		if obj == nil {
			return
		}

		castle := obj.(*rpc.AllianceCastle)

		for _, c := range donateInfo.Characters {
			castle_InsertCharacter(castle, c)

			if notify != nil {
				//logger.Info("你获得了来自%s的%d级%d(数量：%d)", notify.Name, c.GetLevel(), c.GetType(), c.GetCount())
				notifyDonate := rpc.NotifyDonateGet{}
				notifyDonate.SetName(notify.Name)
				notifyDonate.Character = c

				if v.p.conn != nil {
					WriteResult(v.p.conn, &notifyDonate)
				}
			} else {
				//logger.Info("你获得了%d级%d(数量：%d)", c.GetLevel(), c.GetType(), c.GetCount())
			}
		}

		//若满了则删除
		if donateInfo.GetUsedSpace() == donateInfo.GetTotalSpace() {
			KVDeleteExt("donate", v.p.GetUid())
			//logger.Info("getDonateFromDB - KVDelete:%s, %v", v.p.GetUid(), donateInfo)
		} else {
			donateInfo.Characters = make([]*rpc.Character, 0, 1)

			result, err := KVWriteExt("donate", v.p.GetUid(), &donateInfo)
			//logger.Info("getDonateFromDB - KVWrite:%s, %v", v.p.GetUid(), donateInfo)
			if err != nil || result == false {
				logger.Error("getDonateFromDB:KVWrite Failed!!", err, result)
				return
			}
		}
	}
}

func (v *village) castle_GetDonate(notify *proto.NotifyGetDonate) {
	ts("castle_GetDonate", v.vid)
	defer te("castle_GetDonate", v.vid)

	//加锁
	lid := GenLockMessage(cns.GetServerId(), proto.MethodGetDonateInfo, 0)
	if !lockclient.WaitLockGet("donate", v.p.GetUid(), lid) {
		logger.Error("castle_GetDonate:WaitLockGet Failed!!")
		return
	}
	defer lockclient.TryUnlock("donate", v.p.GetUid(), lid)

	//尝试取
	v.getDonateFromDB(notify)
}

func (v *village) castle_RequestDonate() bool {
	obj := v.buildings_Get(rpc.BuildingId_AllianceCastle, 0)
	if obj == nil {
		return false
	}

	castle := obj.(*rpc.AllianceCastle)

	if castle.GetDonateTime() > 0 && uint32(time.Now().Unix())-castle.GetDonateTime() < 30 {
		logger.Error("castle_RequestDonate:DonateTime is not cooldown!!")
		return false
	}

	//加锁
	lid := GenLockMessage(cns.GetServerId(), proto.MethodGetDonateInfo, 0)
	if !lockclient.WaitLockGet("donate", v.p.GetUid(), lid) {
		logger.Error("castle_RequestDonate:WaitLockGet Failed!!")
		return false
	}
	defer lockclient.TryUnlock("donate", v.p.GetUid(), lid)

	//先尝试取一次
	v.getDonateFromDB(nil)

	used, total := castle_GetHousingUsedSpace(castle), castle_GetHousingSpace(castle)
	if used == total {
		//已满
		return false
	}

	castle.SetDonateTime(uint32(time.Now().Unix()))

	donateInfo := rpc.DonateInfo{}
	donateInfo.SetInfo("Need troops!!")
	donateInfo.SetUsedSpace(used)
	donateInfo.SetTotalSpace(total)

	result, err := KVWriteExt("donate", v.p.GetUid(), &donateInfo)
	//logger.Info("castle_RequestDonate:%v", donateInfo, err, result)
	if err != nil || result == false {
		return false
	}

	s2cMsg := rpc.S2CDonate{}
	s2cMsg.SetUid(v.p.GetUid())
	s2cMsg.SetName(v.p.GetName())
	s2cMsg.SetLevel(v.p.GetLevel())
	s2cMsg.SetPower(v.p.GetClanPlayerPower())
	s2cMsg.SetTime(castle.GetDonateTime())
	s2cMsg.SetInfo("Need troops!!")
	s2cMsg.SetUsedSpace(used)
	s2cMsg.SetTotalSpace(total)

	SendDonateMsg(v.p.GetUid(), s2cMsg)

	return true
}
