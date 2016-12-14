package connector

import (
	"golang-project/dpsg/logger"
	"golang-project/dpsg/proto"
	"golang-project/dpsg/rpc"
	"strings"
	"time"
)

func Hero_New(gh *rpc.GeneralHouse, ctype rpc.CharacterType) *rpc.Hero {
	ct := &rpc.Character{}
	ct.SetType(ctype)
	ct.SetCount(1)
	ct.SetLevel(1)

	hero := &rpc.Hero{}
	hero.SetCharacter(ct)
	hero.SetUpgradetime(0)

	gh.Hero = append(gh.Hero, hero)

	return hero
}

func Hero_Get(gh *rpc.GeneralHouse, ctype rpc.CharacterType) *rpc.Hero {
	for _, h := range gh.Hero {
		if h.GetCharacter().GetType() == ctype {
			return h
		}
	}
	return nil
}

func Hero_Has(gh *rpc.GeneralHouse, ctype rpc.CharacterType) bool {
	for _, h := range gh.Hero {
		if h.GetCharacter().GetType() == ctype {
			return true
		}
	}
	return false
}

func (v *village) hero_create(index uint32, ctype rpc.CharacterType) bool {
	ts("hero_create", v.vid)
	defer te("hero_create", v.vid)

	obj := v.buildings_Get(rpc.BuildingId_GeneralHouse, index)
	if obj == nil {
		SyncError(v.p.conn, "hero_create:generalHouse(%d) == nil", index)

		return false
	}

	gh := obj.(*rpc.GeneralHouse)

	cfg := GetCharacterCfgByTypeId(ctype, 1)
	if cfg == nil {
		SyncError(v.p.conn, "hero_create:cfg(type:%d) == nil", ctype)

		return false
	}

	switch strings.ToLower(cfg.UpgradeResource) {
	case "wuhun":
		if cfg.UpgradeCost > v.p.GetWuhun() {
			SyncError(v.p.conn, "hero_create:cfg.UpgradeCost(%d) > total(%d) - wuhun", cfg.UpgradeCost, v.p.GetWuhun())

			return false
		}
	}

	Hero_New(gh, ctype)

	v.p.CostResource(cfg.UpgradeCost, strings.ToLower(cfg.UpgradeResource), proto.Lose_CreateHero)

	return true
}

func (v *village) hero_upgrade(index uint32, ctype rpc.CharacterType) bool {
	ts("hero_upgrade", v.vid)
	defer te("hero_upgrade", v.vid)

	obj := v.buildings_Get(rpc.BuildingId_GeneralHouse, index)
	if obj == nil {
		SyncError(v.p.conn, "hero_upgrade:generalHouse(%d) == nil", index)

		return false
	}

	gh := obj.(*rpc.GeneralHouse)

	hero := Hero_Get(gh, ctype)
	if hero == nil {
		SyncError(v.p.conn, "hero_upgrade:hero(type:%d) == nil", ctype)

		return false
	}

	if hero.GetUpgradetime() != 0 {
		SyncError(v.p.conn, "hero_upgrade:hero.GetUpgradetime(%d) != 0", hero.GetUpgradetime())

		return false
	}

	cfg := GetCharacterCfgByTypeId(ctype, hero.GetCharacter().GetLevel()+1)
	if cfg == nil {
		SyncError(v.p.conn, "hero_upgrade:cfg(type:%d, level:%d) == nil", ctype, hero.GetCharacter().GetLevel())

		return false
	}

	switch strings.ToLower(cfg.UpgradeResource) {
	case "wuhun":
		if cfg.UpgradeCost > v.p.GetWuhun() {
			SyncError(v.p.conn, "hero_upgrade:cfg.UpgradeCost(%d) > total(%d) - wuhun", cfg.UpgradeCost, v.p.GetWuhun())

			return false
		}
	}

	hero.SetUpgradetime(uint32(time.Now().Unix()))

	v.p.CostResource(cfg.UpgradeCost, strings.ToLower(cfg.UpgradeResource), proto.Lose_UpgradeHero)

	//如果是当前出战的则取消当前出战
	if gh.GetSelectedhero() == ctype {
		gh.SetSelectedhero(0)
	}

	return true
}

func (v *village) hero_FinishNow(index uint32, ctype rpc.CharacterType) bool {
	obj := v.buildings_Get(rpc.BuildingId_GeneralHouse, index)
	if obj == nil {
		SyncError(v.p.conn, "hero_FinishNow:generalHouse(%d) == nil", index)

		return false
	}

	gh := obj.(*rpc.GeneralHouse)

	hero := Hero_Get(gh, ctype)
	if hero == nil {
		SyncError(v.p.conn, "hero_FinishNow:hero(type:%d) == nil", ctype)

		return false
	}

	if hero.GetUpgradetime() == 0 {
		SyncError(v.p.conn, "hero_FinishNow:hero.GetUpgradetime() == 0")

		return false
	}

	cfg := GetCharacterCfgByTypeId(ctype, hero.GetCharacter().GetLevel())
	if cfg == nil {
		SyncError(v.p.conn, "hero_FinishNow:cfg(type:%d, level:%d) == nil", ctype, hero.GetCharacter().GetLevel())

		return false
	}

	time_now := uint32(time.Now().Unix())

	if time_now > (hero.GetUpgradetime() + cfg.GetUpgradeTime()) {
		SyncError(v.p.conn, "hero_FinishNow:time_now(%d) > (hero.GetUpgradetime(%d) + cfg.GetUpgradeTime(%d))", time_now, hero.GetUpgradetime(), cfg.GetUpgradeTime())

		return false
	}

	cost := GetYuanBaoCountFromTime(hero.GetUpgradetime() + cfg.GetUpgradeTime() - time_now)

	if v.p.GetPlayerTotalGem() >= cost {
		//先确定扣除成功再做后面的操作
		if !v.p.CostResource(cost, proto.ResType_Gem, proto.Lose_BuyTime) {
			return false
		}

		hero.GetCharacter().SetLevel(hero.GetCharacter().GetLevel() + 1)
		hero.SetUpgradetime(0)

		return true
	}

	return false
}

func (v *village) hero_choose(index uint32, ctype rpc.CharacterType) bool {
	ts("hero_choose", index)
	defer te("hero_choose", index)

	obj := v.buildings_Get(rpc.BuildingId_GeneralHouse, index)
	if obj == nil {
		SyncError(v.p.conn, "hero_choose:obj(%d, %d) == nil", index, ctype)

		return false
	}

	gh := obj.(*rpc.GeneralHouse)

	if !Hero_Has(gh, ctype) {
		SyncError(v.p.conn, "hero_choose:!Hero_Has(index, ctype)", index, ctype)

		return false
	}

	if gh.GetSelectedhero() == ctype {
		gh.SetSelectedhero(0)
	} else {
		gh.SetSelectedhero(ctype)
	}

	return true
}

func (v *village) hero_ProcessUpgrade() {
	obj := v.buildings_GetAllOf(rpc.BuildingId_GeneralHouse)
	if obj == nil {
		return
	}

	time_now := uint32(time.Now().Unix())

	for _, o := range obj {
		gh := o.(*rpc.GeneralHouse)
		for _, hero := range gh.Hero {
			if hero.GetUpgradetime() != 0 {
				cfg := GetCharacterCfgByTypeId(hero.GetCharacter().GetType(), hero.GetCharacter().GetLevel())
				if cfg == nil {
					logger.Error("hero_upgrade:cfg(type:%d, level:%d) == nil", hero.GetCharacter().GetType(), hero.GetCharacter().GetLevel())
					continue
				}
				//finish
				if time_now > (hero.GetUpgradetime() + cfg.GetUpgradeTime()) {
					hero.GetCharacter().SetLevel(hero.GetCharacter().GetLevel() + 1)

					hero.SetUpgradetime(0)
				}
			}
		}
	}
}
