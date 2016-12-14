package connector

import (
	"container/list"
	"golang-project/dpsg/logger"
	"golang-project/dpsg/rpc"
	"strconv"

	"github.com/golang/protobuf/proto"
)

type woking struct {
	idType rpc.BuildingId_IdType
	index  uint32
	i      interface{}
	finish uint32
}

type village struct {
	vid uint64
	*rpc.VillageInfo

	maps [MAP_SIZE][MAP_SIZE]uint32

	buildings map[rpc.BuildingId_IdType][]proto.Message
	p         *player
	w         *list.List
}

func (v *village) getCenterLevel() uint32 {
	if v.Center != nil {
		return v.Center.GetLevel()
	}

	return 0
}

func (v *village) getLaboratoryLevel() uint32 {
	if v.Laboratory != nil {
		if len(v.Laboratory) == 0 {
			return 0
		}
		return v.Laboratory[0].GetLevel()
	}

	return 0
}

func (v *village) getBarrack(index uint32) (b *rpc.Barrack) {
	if v.Barrack != nil {
		if len(v.Barrack) == 0 || len(v.Barrack) >= int(index) {
			return
		}

		if v.Barrack[index].GetUpgradeTime() == 0 {
			return v.Barrack[index]
		}

		return
	}

	return
}

func (v *village) Save() (err error) {
	ts("Save", v.vid)
	defer te("Save", v.vid)

	_, err = KVWriteExt("village", strconv.FormatUint(v.vid, 16), v.VillageInfo)

	return
}

func (v *village) OnTick() {
	v.buildings_ProcessUpgrade()
	v.laboratory_ProcessUpgrade()
	v.barrack_Process()
	v.spellForge_Process()
	v.hero_ProcessUpgrade()
}

func (v *village) OnQuit() {
	v.Save()
}

//得到当前金子和当前最大量金子
func (v *village) GetGoldStorage() (Cur uint32, Max uint32) {
	logger.Info("GetGoldStorage! \n")
	var CurGold uint32 = 0
	var MaxGold uint32 = 0

	centers := v.buildings_GetAllOf(rpc.BuildingId_Center)
	for _, obj := range centers {
		if center, ok := obj.(*rpc.Center); ok {
			level := v.buildings_GetBuildingLevel(center)

			cfg := GetBuildingCfgByTypeId(rpc.BuildingId_Center, level)
			if cfg == nil {
				continue
			}
			CurGold += center.GetStorageGold()
			MaxGold += cfg.MaxStoredGold
		}
	}

	gold_storages := v.buildings_GetAllOf(rpc.BuildingId_GoldStorage)
	for _, obj := range gold_storages {
		if gold_storage, ok := obj.(*rpc.GoldStorage); ok {

			level := v.buildings_GetBuildingLevel(gold_storage)

			cfg := GetBuildingCfgByTypeId(rpc.BuildingId_GoldStorage, level)
			if cfg == nil {
				continue
			}
			CurGold += gold_storage.GetStorageGold()
			MaxGold += cfg.MaxStoredGold
		}
	}

	logger.Info("CurGold = ", CurGold, "MaxGold = ", MaxGold, "\n")
	return CurGold, MaxGold
}

//得到当前食物和当前最大量食物
func (v *village) GetFoodStorage() (Cur uint32, Max uint32) {
	logger.Info("GetFoodStorage! \n")
	var CurFood uint32 = 0
	var MaxFood uint32 = 0

	centers := v.buildings_GetAllOf(rpc.BuildingId_Center)
	for _, obj := range centers {
		if center, ok := obj.(*rpc.Center); ok {
			level := v.buildings_GetBuildingLevel(center)

			cfg := GetBuildingCfgByTypeId(rpc.BuildingId_Center, level)
			if cfg == nil {
				continue
			}
			CurFood += center.GetStorageFood()
			MaxFood += cfg.MaxStoredFood
		}
	}

	food_storages := v.buildings_GetAllOf(rpc.BuildingId_FoodStorage)
	for _, obj := range food_storages {
		if food_storage, ok := obj.(*rpc.FoodStorage); ok {
			level := v.buildings_GetBuildingLevel(food_storage)

			cfg := GetBuildingCfgByTypeId(rpc.BuildingId_FoodStorage, level)
			if cfg == nil {
				continue
			}
			CurFood += food_storage.GetStorageFood()
			MaxFood += cfg.MaxStoredFood
		}
	}

	logger.Info("CurFood = ", CurFood, "MaxFood = ", MaxFood, "\n")
	return CurFood, MaxFood
}

//更新新建筑数据
func (v *village) UpdetaBuildingNewData() {
	logger.Info("UpdetaBuildingNewData! \n")
	for i := uint32(rpc.BuildingId_Center); i < uint32(rpc.BuildingId_End); i++ {
		for index := uint32(0); ; index++ {
			obj := v.buildings_Get(rpc.BuildingId_IdType(i), index)
			if obj == nil {
				break
			}
			logger.Info("BuildingId_IdType = ", rpc.BuildingId_IdType(i), "Index = ", index, "\n")
			cfg := GetBuildingCfgByTypeId(rpc.BuildingId_IdType(i), index)
			if cfg == nil {
				break
			}
			Ammo := cfg.AmmoCount
			if Ammo != 0 {
				xbow := obj.(*rpc.XBow)
				if xbow.GetAltAttackRange() == 0 {
					xbow.SetAltAttackRange(1)
					xbow.SetAmmoCount(Ammo)
				}
			} else {
				break
			}
		}
	}
}
