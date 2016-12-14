package connector

import (
	"golang-project/dpsg/proto"
	"golang-project/dpsg/rpc"
	"time"
	//"runtime/debug"
	//"fmt"
)

func (v *village) collect_StorageFoodTo(o CanStorageFood, total, max uint32) uint32 {
	if max > o.GetStorageFood() {
		value := max - o.GetStorageFood()
		if total > value {
			total -= value
			o.SetStorageFood(max)
		} else {
			o.SetStorageFood(o.GetStorageFood() + total)
			total = 0
		}
	}

	return total
}

func (v *village) collect_StorageFood(total uint32) uint32 {
	centers := v.buildings_GetAllOf(rpc.BuildingId_Center)
	for _, obj := range centers {
		if center, ok := obj.(*rpc.Center); ok {
			level := v.buildings_GetBuildingLevel(center)

			cfg := GetBuildingCfgByTypeId(rpc.BuildingId_Center, level)
			if cfg == nil {
				continue
			}

			total = v.collect_StorageFoodTo(center, total, cfg.MaxStoredFood)
			if total == 0 {
				return 0
			}
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

			total = v.collect_StorageFoodTo(food_storage, total, cfg.MaxStoredFood)
			if total == 0 {
				return 0
			}
		}
	}

	return total
}

func (v *village) collect_StorageGoldTo(o CanStorageGold, total, max uint32) uint32 {
	if max > o.GetStorageGold() {
		value := max - o.GetStorageGold()
		if total > value {
			total -= value
			o.SetStorageGold(max)
		} else {
			o.SetStorageGold(o.GetStorageGold() + total)
			total = 0
		}
	}

	return total
}

func (v *village) collect_StorageGold(total uint32) uint32 {
	centers := v.buildings_GetAllOf(rpc.BuildingId_Center)
	for _, obj := range centers {
		if center, ok := obj.(*rpc.Center); ok {
			level := v.buildings_GetBuildingLevel(center)

			cfg := GetBuildingCfgByTypeId(rpc.BuildingId_Center, level)
			if cfg == nil {
				continue
			}

			total = v.collect_StorageGoldTo(center, total, cfg.MaxStoredGold)
			if total == 0 {
				return 0
			}
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

			total = v.collect_StorageGoldTo(gold_storage, total, cfg.MaxStoredGold)
			if total == 0 {
				return 0
			}
		}
	}

	return total
}

func (v *village) collect_CostFood(total uint32) bool {
	food_storages := v.buildings_GetAllOf(rpc.BuildingId_FoodStorage)
	for _, obj := range food_storages {
		if food_storage, ok := obj.(*rpc.FoodStorage); ok {
			if food_storage.GetStorageFood() >= total {
				food_storage.SetStorageFood(food_storage.GetStorageFood() - total)
				return true
			} else {
				total -= food_storage.GetStorageFood()
				food_storage.SetStorageFood(0)
			}
		}
	}

	centers := v.buildings_GetAllOf(rpc.BuildingId_Center)
	for _, obj := range centers {
		if center, ok := obj.(*rpc.Center); ok {
			if center.GetStorageFood() >= total {
				center.SetStorageFood(center.GetStorageFood() - total)
				return true
			} else {
				total -= center.GetStorageFood()
				center.SetStorageFood(0)
			}
		}
	}

	return false
}

func (v *village) collect_CostGold(total uint32) bool {
	gold_storages := v.buildings_GetAllOf(rpc.BuildingId_GoldStorage)
	for _, obj := range gold_storages {
		if gold_storage, ok := obj.(*rpc.GoldStorage); ok {
			if gold_storage.GetStorageGold() >= total {
				gold_storage.SetStorageGold(gold_storage.GetStorageGold() - total)
				return true
			} else {
				total -= gold_storage.GetStorageGold()
				gold_storage.SetStorageGold(0)
			}
		}
	}

	centers := v.buildings_GetAllOf(rpc.BuildingId_Center)
	for _, obj := range centers {
		if center, ok := obj.(*rpc.Center); ok {
			if center.GetStorageGold() >= total {
				center.SetStorageGold(center.GetStorageGold() - total)
				return true
			} else {
				total -= center.GetStorageGold()
				center.SetStorageGold(0)
			}
		}
	}

	return false
}

func (v *village) collect_HasEnoughFood(need uint32) bool {
	_, total := v.collect_GetStorageFoodLimit()

	return total >= need
}

func (v *village) collect_HasEnoughGold(need uint32) bool {
	_, total := v.collect_GetStorageGoldLimit()

	return total >= need
}

func (v *village) collect_GetStorageFoodLimit() (limit uint32, total uint32) {
	centers := v.buildings_GetAllOf(rpc.BuildingId_Center)
	for _, obj := range centers {
		if center, ok := obj.(*rpc.Center); ok {
			level := v.buildings_GetBuildingLevel(center)

			cfg := GetBuildingCfgByTypeId(rpc.BuildingId_Center, level)
			if cfg == nil {
				continue
			}
			limit += cfg.MaxStoredFood
			if cfg.MaxStoredFood >= center.GetStorageFood() {
				total += center.GetStorageFood()
			} else {
				total += cfg.MaxStoredFood
			}
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

			limit += cfg.MaxStoredFood
			if cfg.MaxStoredFood >= food_storage.GetStorageFood() {
				total += food_storage.GetStorageFood()
			} else {
				total += cfg.MaxStoredFood
			}
		}
	}
	return
}

func (v *village) collect_GetStorageGoldLimit() (limit uint32, total uint32) {
	centers := v.buildings_GetAllOf(rpc.BuildingId_Center)
	for _, obj := range centers {
		if center, ok := obj.(*rpc.Center); ok {

			level := v.buildings_GetBuildingLevel(center)

			cfg := GetBuildingCfgByTypeId(rpc.BuildingId_Center, level)
			if cfg == nil {
				continue
			}
			limit += cfg.MaxStoredGold
			if cfg.MaxStoredGold >= center.GetStorageGold() {
				total += center.GetStorageGold()
			} else {
				total += cfg.MaxStoredGold
			}
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
			limit += cfg.MaxStoredGold
			if cfg.MaxStoredGold >= gold_storage.GetStorageGold() {
				total += gold_storage.GetStorageGold()
			} else {
				total += cfg.MaxStoredGold
			}
		}
	}

	return
}

func (v *village) collect_GetFoodOf(index uint32) (total uint32) {
	obj := v.buildings_Get(rpc.BuildingId_Farm, index)
	if obj == nil {
		return
	}

	if farm, ok := obj.(*rpc.Farm); ok {
		cfg := GetBuildingCfgByTypeId(rpc.BuildingId_Farm, farm.GetLevel())
		if cfg == nil {
			return
		}

		hour := GetPassHour(farm.GetLastOpTime())

		count := hour*float64(cfg.ResourcePerHour) + float64(farm.GetResCount())

		if count >= float64(cfg.ResourceMax) {
			return cfg.ResourceMax
		} else {
			return uint32(count)
		}
	}

	return
}

func (v *village) collect_GetGoldOf(index uint32) (total uint32) {
	obj := v.buildings_Get(rpc.BuildingId_GoldMine, index)
	if obj == nil {
		return
	}

	if gold_mine, ok := obj.(*rpc.GoldMine); ok {
		cfg := GetBuildingCfgByTypeId(rpc.BuildingId_GoldMine, gold_mine.GetLevel())
		if cfg == nil {
			return
		}

		hour := GetPassHour(gold_mine.GetLastOpTime())

		count := hour*float64(cfg.ResourcePerHour) + float64(gold_mine.GetResCount())

		if count >= float64(cfg.ResourceMax) {
			return cfg.ResourceMax
		} else {
			return uint32(count)
		}
	}

	return
}

func (v *village) collect_SetOpTime(idType rpc.BuildingId_IdType, index uint32) bool {
	obj := v.buildings_Get(idType, index)
	if obj == nil {
		return false
	}

	if op, ok := obj.(Operable); ok {
		op.SetLastOpTime(uint32(time.Now().Unix()))
		return true
	}

	return false
}

func (v *village) collect_Opt(idType rpc.BuildingId_IdType, index uint32) bool {

	switch idType {
	case rpc.BuildingId_GoldMine:
		total := v.collect_GetGoldOf(index)

		if total == 0 {
			return false
		}

		ResCount, _ := v.p.GainResource(total, proto.ResType_Gold, proto.Gain_Gather)

		//得到金子未采集量
		obj := v.buildings_Get(rpc.BuildingId_GoldMine, index)
		if obj == nil {
			return false
		}

		if gold_mine, ok := obj.(*rpc.GoldMine); ok {
			gold_mine.SetResCount(ResCount)
		}

	case rpc.BuildingId_Farm:
		total := v.collect_GetFoodOf(index)

		if total == 0 {
			return false
		}

		ResCount, _ := v.p.GainResource(total, proto.ResType_Food, proto.Gain_Gather)

		obj := v.buildings_Get(rpc.BuildingId_Farm, index)
		if obj == nil {
			return false
		}

		if farm, ok := obj.(*rpc.Farm); ok {
			farm.SetResCount(ResCount)
		}
	default:
		return false
	}

	v.collect_SetOpTime(idType, index)

	return false
}
