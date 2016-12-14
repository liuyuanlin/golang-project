package connector

import (
	"container/list"
	//"fmt"
	"golang-project/dpsg/logger"
	"golang-project/dpsg/rpc"
	"strconv"
)

func LoadVillage(id uint64, userguide bool) *village {
	ts("LoadVillage", id)
	defer te("LoadVillage", id)

	var v rpc.VillageInfo

	if id == 0 {
		return NewVillage(&v)
	}

	result, err := KVQueryExt("village", strconv.FormatUint(id, 16), &v)

	if err != nil || result != true {
		if err == nil {
			logger.Error("load village failed : (id %d)", id)
		} else {
			logger.Error("load village failed : (id %d)(%s)", id, err.Error())
		}
		return nil
	}

	vi := &village{vid: id, VillageInfo: &v}

	vi.buildings_Init()

	if !vi.mapInit() {
		logger.Error("Init Map Fail")

		return nil
	}

	vi.OnInit(userguide)

	return vi
}

func NewVillage(v *rpc.VillageInfo) *village {
	vid := GenVillageId(cns.GetServerId())
	ts("NewVillage", vid)
	defer te("NewVillage", vid)

	vi := &village{vid: vid, VillageInfo: v}

	vi.buildings_Init()
	vi.mapInit()

	if nil == vi.buildings_Create(rpc.BuildingId_Center, 22, 22, true) {
		return nil
	}
	vi.Center.SetStorageFood(2000)
	vi.Center.SetStorageGold(2000)
	vi.Center.SetLevel(1)

	if nil == vi.buildings_Create(rpc.BuildingId_Worker, 22, 27, true) {
		return nil
	}

	//add for worker

	if nil == vi.buildings_Create(rpc.BuildingId_Worker, 25, 27, true) {
		return nil
	}

	if nil == vi.buildings_Create(rpc.BuildingId_GoldMine, 28, 21, true) {
		return nil
	}
	vi.Goldmine[0].SetResCount(0)

	if nil == vi.buildings_Create(rpc.BuildingId_TroopHousing, 22, 16, true) {
		return nil
	}

	if nil == vi.buildings_Create(rpc.BuildingId_Cannon, 18, 23, true) {
		return nil
	}

	if nil == vi.buildings_Create(rpc.BuildingId_Farm, 18, 17, true) {
		return nil
	}

	if nil == vi.buildings_Create(rpc.BuildingId_FoodStorage, 16, 26, true) {
		return nil
	}

	if nil == vi.buildings_Create(rpc.BuildingId_GoldStorage, 14, 22, true) {
		return nil
	}

	//策划要求去掉
	/*if nil == vi.buildings_Create(rpc.BuildingId_Barrack, 19, 26, true) {
		return nil
	}*/

	g := vi.buildings_Create(rpc.BuildingId_GeneralHouse, 19, 29, true)
	if g == nil {
		return nil
	} else {
		Hero_New(g.(*rpc.GeneralHouse), rpc.CharacterType_Yuanfang)
		vi.hero_choose(0, rpc.CharacterType_Yuanfang)
	}

	if nil == vi.buildings_Create(rpc.BuildingId_Barrier1, 5, 5, true) {
		return nil
	}

	if nil == vi.buildings_Create(rpc.BuildingId_Barrier1, 10, 5, true) {
		return nil
	}

	if nil == vi.buildings_Create(rpc.BuildingId_Barrier1, 5, 20, true) {
		return nil
	}

	if nil == vi.buildings_Create(rpc.BuildingId_Barrier2, 30, 30, true) {
		return nil
	}

	if nil == vi.buildings_Create(rpc.BuildingId_Barrier3, 30, 6, true) {
		return nil
	}

	if nil == vi.buildings_Create(rpc.BuildingId_Barrier3, 20, 5, true) {
		return nil
	}

	if nil == vi.buildings_Create(rpc.BuildingId_Barrier4, 22, 1, true) {
		return nil
	}

	if nil == vi.buildings_Create(rpc.BuildingId_Barrier4, 35, 10, true) {
		return nil
	}

	if nil == vi.buildings_Create(rpc.BuildingId_Barrier5, 9, 3, true) {
		return nil
	}

	if nil == vi.buildings_Create(rpc.BuildingId_Barrier5, 41, 8, true) {
		return nil
	}

	if nil == vi.buildings_Create(rpc.BuildingId_Barrier5, 3, 29, true) {
		return nil
	}

	if nil == vi.buildings_Create(rpc.BuildingId_Barrier5, 3, 10, true) {
		return nil
	}

	if nil == vi.buildings_Create(rpc.BuildingId_Barrier5, 11, 40, true) {
		return nil
	}

	if nil == vi.buildings_Create(rpc.BuildingId_Barrier6, 33, 33, true) {
		return nil
	}

	if nil == vi.buildings_Create(rpc.BuildingId_Barrier6, 20, 36, true) {
		return nil
	}
	//if nil == vi.buildings_Create(rpc.BuildingId_AllianceCastle, 10, 30, true) {
	//	return nil
	//}

	result, err := KVWriteExt("village", strconv.FormatUint(vid, 16), v)

	if err != nil || result == false {
		return nil
	}

	vi.OnInit(false)
	return vi
}

func (v *village) OnInit(userguide bool) {
	v.w = list.New()

	for i := uint32(rpc.BuildingId_Center); i < uint32(rpc.BuildingId_End); i++ {
		for index := uint32(0); ; index++ {
			obj := v.buildings_Get(rpc.BuildingId_IdType(i), index)
			if obj == nil {
				break
			}

			if o, ok := obj.(Upgradable); ok {
				if o.GetUpgradeTime() != 0 {
					//新手下线再上秒完
					if userguide {
						o.SetUpgradeTime(1)
					}

					v.buildings_NewWoking(rpc.BuildingId_IdType(i), index, obj)
				}
			} else if o, ok := obj.(Removable); ok {
				if o.GetRemoveTime() != 0 {
					//新手下线再上秒完
					if userguide {
						o.SetRemoveTime(1)
					}

					v.buildings_NewWoking(rpc.BuildingId_IdType(i), index, obj)
				}
			} else {
				break
			}
		}
	}
}

// 恢复村庄建筑的血量
// type Assailable interface {
//	GetHp() uint32
//	SetHp(value uint32)
//}
func (v *village) ResetBuildingHp() {
	for i := uint32(rpc.BuildingId_Center); i < uint32(rpc.BuildingId_End); i++ {
		for index := uint32(0); ; index++ {
			obj := v.buildings_Get(rpc.BuildingId_IdType(i), index)
			if obj == nil {
				break
			}
			if o, ok := obj.(Assailable); ok {
				var leve uint32 = 1
				if oo, ok := o.(Upgradable); ok {
					leve = oo.GetLevel()
				}
				cfg := GetBuildingCfgByTypeId(rpc.BuildingId_IdType(i), leve)
				if cfg != nil {
					o.SetHp(cfg.Hitpoints)
				} else {
					break
				}
			} else {
				break
			}
		}
	}
}
