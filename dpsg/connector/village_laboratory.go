package connector

import (
	"golang-project/dpsg/logger"
	"golang-project/dpsg/proto"
	"golang-project/dpsg/rpc"
	"strings"
	"time"
)

func (v *village) laboratory_GetInfoLevel(chType rpc.CharacterType) uint32 {
	if v.Laboratory == nil {
		return 1
	}

	if len(v.Laboratory) == 0 {
		return 1
	}

	if v.Laboratory[0].Info == nil {
		return 1
	}

	for _, info := range v.Laboratory[0].Info {
		if info.GetType() == chType {
			return info.GetLevel()
		}
	}

	return 1
}

func (v *village) laboratory_GetSpellInfoLevel(spType rpc.SpellType) uint32 {
	if v.Laboratory == nil {
		return 1
	}

	if len(v.Laboratory) == 0 {
		return 1
	}

	if v.Laboratory[0].SpellInfo == nil {
		return 1
	}

	for _, info := range v.Laboratory[0].SpellInfo {
		if info.GetType() == spType {
			return info.GetLevel()
		}
	}

	return 1
}

func (v *village) laboratory_FinishNow() {
	if v.Laboratory == nil {
		return
	}

	if len(v.Laboratory) == 0 {
		return
	}

	if v.Laboratory[0].GetUpgradeTime() != 0 {
		return
	}

	if v.Laboratory[0].Info != nil {
		for _, info := range v.Laboratory[0].Info {
			if info.GetUpgradeTime() != 0 {
				time_now := uint32(time.Now().Unix())

				cfg := GetCharacterCfgByTypeId(info.GetType(), info.GetLevel())

				r := int32(cfg.GetUpgradeTime() - (time_now - info.GetUpgradeTime()))

				if r > 0 {
					cost := GetYuanBaoCountFromTime(uint32(r))
					if v.p.GetPlayerTotalGem() >= cost {
						//先确定扣除成功再做后面的操作
						if v.p.CostResource(cost, proto.ResType_Gem, proto.Lose_BuyTime) {
							info.SetLevel(info.GetLevel() + 1)
							info.SetUpgradeTime(0)
						}
					}
				}
				return
			}
		}
	}

	if v.Laboratory[0].SpellInfo != nil {
		for _, spellInfo := range v.Laboratory[0].SpellInfo {
			if spellInfo.GetUpgradeTime() != 0 {
				time_now := uint32(time.Now().Unix())
				//todo

				cfg := GetSpellCfgByTypeId(spellInfo.GetType(), spellInfo.GetLevel())

				r := int32(cfg.GetUpgradeTime() - (time_now - spellInfo.GetUpgradeTime()))

				if r > 0 {
					cost := GetYuanBaoCountFromTime(uint32(r))
					if v.p.GetPlayerTotalGem() >= cost {
						//先确定扣除成功再做后面的操作
						if v.p.CostResource(cost, proto.ResType_Gem, proto.Lose_BuyTime) {
							spellInfo.SetLevel(spellInfo.GetLevel() + 1)
							spellInfo.SetUpgradeTime(0)
						}
					}
				}
				return
			}
		}
	}
}

func (v *village) laboratory_ProcessUpgrade() {
	if v.Laboratory == nil {
		return
	}

	if len(v.Laboratory) == 0 {
		return
	}

	if v.Laboratory[0].Info != nil {
		for _, info := range v.Laboratory[0].Info {
			if info.GetUpgradeTime() != 0 {
				// 正常工作中
				time_now := uint32(time.Now().Unix())

				cfg := GetCharacterCfgByTypeId(info.GetType(), info.GetLevel())

				//正常情况下， 不会有大于这个秒数的剩余时间， 所以用他来做状态切换
				if info.GetUpgradeTime() > 100000000 {
					r := int32(cfg.GetUpgradeTime() - (time_now - info.GetUpgradeTime()))
					if r <= 0 {
						info.SetLevel(info.GetLevel() + 1)
						info.SetUpgradeTime(0)
						return
					}

					if v.Laboratory[0].GetUpgradeTime() != 0 {
						// 切换到待机状态
						info.SetUpgradeTime(uint32(r))
					}
				} else {
					//待机中
					if v.Laboratory[0].GetUpgradeTime() == 0 {
						//开始工作了
						info.SetUpgradeTime(time_now - (cfg.GetUpgradeTime() - info.GetUpgradeTime()))
					}
				}
			}
		}
	}

	if v.Laboratory[0].SpellInfo != nil {
		for _, info := range v.Laboratory[0].SpellInfo {
			if info.GetUpgradeTime() != 0 {
				// 正常工作中
				time_now := uint32(time.Now().Unix())

				cfg := GetSpellCfgByTypeId(info.GetType(), info.GetLevel())

				//正常情况下， 不会有大于这个秒数的剩余时间， 所以用他来做状态切换
				if info.GetUpgradeTime() > 100000000 {
					r := int32(cfg.GetUpgradeTime() - (time_now - info.GetUpgradeTime()))
					if r <= 0 {
						info.SetLevel(info.GetLevel() + 1)
						info.SetUpgradeTime(0)
						return
					}

					if v.Laboratory[0].GetUpgradeTime() != 0 {
						// 切换到待机状态
						info.SetUpgradeTime(uint32(r))
					}
				} else {
					//待机中
					if v.Laboratory[0].GetUpgradeTime() == 0 {
						//开始工作了
						info.SetUpgradeTime(time_now - (cfg.GetUpgradeTime() - info.GetUpgradeTime()))
					}
				}
			}
		}
	}

}

func (v *village) laboratory_GetLaboratoryLevel() uint32 {
	if v.Laboratory == nil {
		return 0
	}

	if len(v.Laboratory) == 0 {
		return 0
	}

	return v.Laboratory[0].GetLevel()
}

func (v *village) laboratory_CanUpgradeInfo() bool {
	if v.Laboratory == nil {
		SyncError(v.p.conn, "laboratory_CanUpgradeInfo:v.Laboratory == nil")

		return false
	}

	if len(v.Laboratory) == 0 {
		SyncError(v.p.conn, "laboratory_CanUpgradeInfo:len(v.Laboratory) == 0")

		return false
	}

	if v.Laboratory[0].GetUpgradeTime() != 0 {
		SyncError(v.p.conn, "laboratory_CanUpgradeInfo:v.Laboratory[0].GetUpgradeTime() != 0")

		return false
	}

	if v.Laboratory[0].Info == nil && v.Laboratory[0].SpellInfo == nil {
		return true
	}

	for _, info := range v.Laboratory[0].Info {
		if info.GetUpgradeTime() != 0 {
			SyncError(v.p.conn, "laboratory_CanUpgradeInfo:info.GetUpgradeTime() != 0,  %v", info)

			return false
		}
	}

	for _, info := range v.Laboratory[0].SpellInfo {
		if info.GetUpgradeTime() != 0 {
			SyncError(v.p.conn, "laboratory_CanUpgradeInfo:spellinfo.GetUpgradeTime() != 0")

			return false
		}
	}

	return true
}

func (v *village) laboratory_NewInfo(chType rpc.CharacterType) *rpc.LaboratoryInfo {
	ret := &rpc.LaboratoryInfo{}
	ret.SetType(chType)
	ret.SetLevel(1)
	ret.SetUpgradeTime(0)
	return ret
}

func (v *village) laboratory_UpgradeInfoLevel(chType rpc.CharacterType) bool {
	if !v.laboratory_CanUpgradeInfo() {
		return false
	}

	v.laboratory_ProcessUpgrade()

	if v.Laboratory[0].Info == nil {
		v.Laboratory[0].Info = make([]*rpc.LaboratoryInfo, 0)
	}

	var info *rpc.LaboratoryInfo = nil
	nLevel := uint32(1)

	for _, i := range v.Laboratory[0].Info {
		if i.GetType() == chType {
			info = i
			nLevel = i.GetLevel()

			break
		}
	}

	cfg1 := GetCharacterCfgByTypeId(chType, nLevel)
	if cfg1 == nil {
		SyncError(v.p.conn, "laboratory_UpgradeInfoLevel:cfg1 == nil")

		return false
	}

	switch strings.ToLower(cfg1.UpgradeResource) {
	case "gold":
		_, total := v.collect_GetStorageGoldLimit()
		if cfg1.UpgradeCost > total {
			SyncError(v.p.conn, "laboratory_UpgradeInfoLevel:cfg1.UpgradeCost(%d) > total(%d) - gold", cfg1.UpgradeCost, total)
			return false
		}
	case "food":
		_, total := v.collect_GetStorageFoodLimit()
		if cfg1.UpgradeCost > total {
			SyncError(v.p.conn, "laboratory_UpgradeInfoLevel:cfg1.UpgradeCost(%d) > total(%d) - food", cfg1.UpgradeCost, total)
			return false
		}
	}

	cfg2 := GetCharacterCfgByTypeId(chType, nLevel+1)
	if cfg2 == nil {
		SyncError(v.p.conn, "laboratory_UpgradeInfoLevel:cfg2 == nil")

		return false
	}

	if cfg2.LaboratoryLevel > v.getLaboratoryLevel() {
		SyncError(v.p.conn, "laboratory_UpgradeInfoLevel:cfg2.LaboratoryLevel(%d) > v.getLaboratoryLevel()(%d) 1", cfg2.LaboratoryLevel, v.getLaboratoryLevel())

		return false
	}

	if info == nil {
		info = v.laboratory_NewInfo(chType)

		v.Laboratory[0].Info = append(v.Laboratory[0].Info, info)
	}
	info.SetUpgradeTime(uint32(time.Now().Unix()))

	v.p.CostResource(cfg1.UpgradeCost, strings.ToLower(cfg1.UpgradeResource), proto.Lose_UpgradeCharacter)

	logger.Info("laboratory_UpgradeInfoLevel:%v", v.Laboratory[0].Info)

	return true
}

////////////////////////////
// -_-
////////////////////////////

func (v *village) laboratory_NewSpellInfo(spType rpc.SpellType) *rpc.LaboratorySpellInfo {
	ret := &rpc.LaboratorySpellInfo{}
	ret.SetType(spType)
	ret.SetLevel(1)
	ret.SetUpgradeTime(0)
	return ret
}

//丹药升级
func (v *village) laboratory_UpgradeSpellInfoLevel(spType rpc.SpellType) bool {
	if !v.laboratory_CanUpgradeInfo() {
		return false
	}

	v.laboratory_ProcessUpgrade()

	if v.Laboratory[0].SpellInfo == nil {
		v.Laboratory[0].SpellInfo = make([]*rpc.LaboratorySpellInfo, 0)
	}

	for _, info := range v.Laboratory[0].SpellInfo {
		if info.GetType() == spType {
			cfg := GetSpellCfgByTypeId(spType, info.GetLevel()+1)

			if cfg == nil {
				return false
			}

			if cfg.LaboratoryLevel > v.getLaboratoryLevel() {
				return false
			}

			info.SetUpgradeTime(uint32(time.Now().Unix()))
			return true
		}
	}

	cfg := GetSpellCfgByTypeId(spType, 2)
	if cfg == nil {
		return false
	}

	if cfg.LaboratoryLevel > v.getLaboratoryLevel() {
		return false
	}

	info := v.laboratory_NewSpellInfo(spType)
	info.SetUpgradeTime(uint32(time.Now().Unix()))

	v.Laboratory[0].SpellInfo = append(v.Laboratory[0].SpellInfo, info)
	return true
}
