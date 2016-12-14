package connector

import (
	//"fmt"
	"golang-project/dpsg/logger"
	"golang-project/dpsg/proto"
	"golang-project/dpsg/rpc"
	"strconv"
	"strings"
	"time"
)

func (v *village) moveTo(idType rpc.BuildingId_IdType, index uint32, x uint32, y uint32) bool {
	logger.Info("moveTo:(%d - %d)To(%d, %d)", idType, index, x, y)
	obj := v.buildings_Get(idType, index)
	if obj == nil {
		SyncError(v.p.conn, "moveTo:Failed!!!! obj == nil (%d, %d)", idType, index)

		return false
	}

	if o, ok := obj.(Movable); ok {
		cfg := GetBuildingCfgByTypeId(idType, 1)
		if cfg == nil {
			SyncError(v.p.conn, "moveTo:can't get cfg for building type %d", idType)

			return false
		}

		if !cfg.CanMove {
			SyncError(v.p.conn, "moveToBatch:!cfg.CanMove for building type %d", idType)

			return false
		}

		v.mapRemoveFrom(o.GetP().GetX(), o.GetP().GetY(), cfg.BuildSize)
		if v.mapCheckSpace(x, y, cfg.BuildSize, uint32(idType), true) {
			v.mapInsertTo(x, y, cfg.BuildSize, uint32(idType))

			o.GetP().SetX(x)
			o.GetP().SetY(y)

			return true
		} else {
			v.mapInsertTo(o.GetP().GetX(), o.GetP().GetY(), cfg.BuildSize, uint32(idType))

			return false
		}
	}
	return false
}

type PreBatchMoveInfo struct {
	obj    interface{}
	idType uint32
	oldx   uint32
	oldy   uint32
	x      uint32
	y      uint32
	size   uint32
}

func (v *village) moveToBatch(toBatch rpc.MoveToBatch) bool {
	preMoveInfo := make([]PreBatchMoveInfo, 0, 1)

	for _, to := range toBatch.Moves {
		idType, index, x, y := to.GetId().GetType(), to.GetId().GetIndex(), to.GetP().GetX(), to.GetP().GetY()

		obj := v.buildings_Get(idType, index)
		if obj == nil {
			SyncError(v.p.conn, "moveToBatch:Failed!!!! obj == nil (%d, %d)", idType, index)

			return false
		}

		if o, ok := obj.(Movable); ok {
			cfg := GetBuildingCfgByTypeId(idType, 1)
			if cfg == nil {
				SyncError(v.p.conn, "moveToBatch:can't get cfg for building type %d", idType)

				return false
			}

			if !cfg.CanMove {
				SyncError(v.p.conn, "moveToBatch:!cfg.CanMove for building type %d", idType)

				return false
			}

			preMoveInfo = append(preMoveInfo, PreBatchMoveInfo{obj: obj, idType: uint32(idType), oldx: o.GetP().GetX(), oldy: o.GetP().GetY(), x: x, y: y, size: cfg.BuildSize})
		} else {
			return false
		}
	}

	//logger.Info("preMoveInfo:%v", preMoveInfo)

	for _, i := range preMoveInfo {
		v.mapRemoveFrom(i.oldx, i.oldy, i.size)
	}

	for _, i := range preMoveInfo {
		if !v.mapCheckSpace(i.x, i.y, i.size, i.idType, true) {
			return false
		}
	}

	for _, i := range preMoveInfo {
		v.mapInsertTo(i.x, i.y, i.size, i.idType)

		if o, ok := i.obj.(Movable); ok {
			o.GetP().SetX(i.x)
			o.GetP().SetY(i.y)
		}
	}

	return true
}

//创建建筑
func (v *village) create(idType rpc.BuildingId_IdType, x, y uint32) bool {
	cfg := GetBuildingCfgByTypeId(idType, 1)
	if cfg == nil {
		SyncError(v.p.conn, "create:cfg == nil idType(%d)", idType)

		return false
	}

	v.buildings_ProcessUpgrade()
	//先判断是否需要工人
	if cfg.NeedWorker {
		if v.buildings_GetWorkerCnt() <= v.buildings_WokingLen() && idType != rpc.BuildingId_Worker {
			SyncError(v.p.conn, "create:not enough worker", v.buildings_GetWorkerCnt(), v.buildings_WokingLen())

			return false
		}
	}

	limit := GetBuildingCntLimitByTypeId(idType, v.getCenterLevel())
	if limit <= 0 {
		SyncError(v.p.conn, "create:limit(%d) <= 0", limit)

		return false
	}

	if v.buildings_GetCntOf(idType) >= limit {
		SyncError(v.p.conn, "create:v.buildings_GetCntOf(%d) >= limit(%d)", idType, limit)

		return false
	}

	if cfg.TownHallLevel != 0 && v.getCenterLevel() < cfg.TownHallLevel {
		SyncError(v.p.conn, "create:cfg.TownHallLevel != 0 && CenterLevel(%d) < cfg.TownHallLevel(%d)", v.getCenterLevel(), cfg.TownHallLevel)

		return false
	}

	//建造资源
	BuildCost := cfg.BuildCost
	if idType == rpc.BuildingId_Worker {
		ResNumber := []string{"200", "500", "1000", "2000"}
		Index := v.buildings_GetCntOf(idType)
		WorkerCost, _ := strconv.Atoi(ResNumber[Index-1])
		BuildCost = uint32(WorkerCost)
	}

	switch strings.ToLower(cfg.BuildResource) {
	case "gold":
		_, total := v.collect_GetStorageGoldLimit()
		if BuildCost > total {
			SyncError(v.p.conn, "create:cfg.BuildCost(%d) > total(%d) - gold", BuildCost, total)

			return false
		}
	case "food":
		_, total := v.collect_GetStorageFoodLimit()
		if BuildCost > total {
			SyncError(v.p.conn, "create:cfg.BuildCost(%d) > total(%d) - food", BuildCost, total)

			return false
		}

	case "diamonds":
		if BuildCost > v.p.GetPlayerTotalGem() {
			SyncError(v.p.conn, "create:cfg.BuildCost(%d) > total(%d) - diamonds", BuildCost, v.p.GetPlayerTotalGem())

			return false
		}
	}
	b := v.buildings_Create(idType, x, y, false)

	if nil == b {
		SyncError(v.p.conn, "create:nil == b")

		return false
	}
	v.p.CostResource(BuildCost, strings.ToLower(cfg.BuildResource), proto.Lose_CreateBuilding)

	if cfg.NeedWorker {
		if u, ok := b.(Upgradable); ok {
			cfg := GetBuildingCfgByTypeId(idType, 1)
			if cfg != nil && cfg.GetBuildingTime() != 0 {
				u.SetUpgradeTime(uint32(time.Now().Unix()))
			}
			v.buildings_NewWoking(idType, v.buildings_GetCntOf(idType)-1, u)
		}
	} else {
		if o, ok := b.(Assailable); ok {
			cfg := GetBuildingCfgByTypeId(idType, 1)
			if cfg != nil {
				o.SetHp(cfg.Hitpoints)
			}
		}
	}

	return true
}

//升级建筑
func (v *village) upgrade(idType rpc.BuildingId_IdType, index uint32) bool {
	v.buildings_ProcessUpgrade()

	if v.buildings_GetWorkerCnt() <= v.buildings_WokingLen() {
		SyncError(v.p.conn, "upgrade:not enough worker", v.buildings_GetWorkerCnt(), v.buildings_WokingLen())

		return false
	}

	obj := v.buildings_Get(idType, index)
	if obj == nil {
		SyncError(v.p.conn, "upgrade:obj == nil")

		return false
	}

	if o, ok := obj.(Upgradable); ok {
		if o.GetUpgradeTime() != 0 {
			SyncError(v.p.conn, "upgrade:o.GetUpgradeTime() != 0")

			return false
		}

		cfg := GetBuildingCfgByTypeId(idType, o.GetLevel()+1)
		if cfg == nil {
			SyncError(v.p.conn, "upgrade:cfg == nil")

			return false
		}

		if cfg.TownHallLevel != 0 && cfg.TownHallLevel > v.getCenterLevel() {
			SyncError(v.p.conn, "upgrade:cfg.TownHallLevel != 0 && CenterLevel(%d) < cfg.TownHallLevel(%d)", v.getCenterLevel(), cfg.TownHallLevel)

			return false
		}

		switch strings.ToLower(cfg.BuildResource) {
		case "gold":
			_, total := v.collect_GetStorageGoldLimit()
			if cfg.BuildCost > total {
				SyncError(v.p.conn, "upgrade:cfg.BuildCost(%d) > total(%d) - gold", cfg.BuildCost, total)

				return false
			}
		case "food":
			_, total := v.collect_GetStorageFoodLimit()
			if cfg.BuildCost > total {
				SyncError(v.p.conn, "upgrade:cfg.BuildCost(%d) > total(%d) - food", cfg.BuildCost, total)

				return false
			}
		case "diamonds":
			if cfg.BuildCost > v.p.GetPlayerTotalGem() {
				SyncError(v.p.conn, "upgrade:cfg.BuildCost(%d) > total(%d) - diamonds", cfg.BuildCost, v.p.GetPlayerTotalGem())

				return false
			}
		}

		if _, ok := o.(Operable); ok {
			v.collect_Opt(idType, index)
		}

		o.SetUpgradeTime(uint32(time.Now().Unix()))

		v.buildings_NewWoking(idType, index, o)

		v.p.CostResource(cfg.BuildCost, strings.ToLower(cfg.BuildResource), proto.Lose_UpgradeBuilding)
	}

	return true
}

func (v *village) remove(idType rpc.BuildingId_IdType, index uint32) bool {
	v.buildings_ProcessUpgrade()

	if v.buildings_GetWorkerCnt() <= v.buildings_WokingLen() {
		SyncError(v.p.conn, "remove:not enough worker", v.buildings_GetWorkerCnt(), v.buildings_WokingLen())

		return false
	}

	obj := v.buildings_Get(idType, index)
	if obj == nil {
		SyncError(v.p.conn, "remove:obj == nil")

		return false
	}

	if o, ok := obj.(Removable); ok {
		cfg := GetBuildingCfgByTypeId(idType, 1)
		if cfg == nil {
			SyncError(v.p.conn, "remove:cfg == nil")

			return false
		}

		if !cfg.CanRemove {
			return false
		}

		if o.GetRemoveTime() != 0 {
			SyncError(v.p.conn, "remove:o.GetRemoveTime() != 0")

			return false
		}

		if cfg.TownHallLevel != 0 && cfg.TownHallLevel > v.getCenterLevel() {
			SyncError(v.p.conn, "remove:cfg.TownHallLevel != 0 && CenterLevel(%d) < cfg.TownHallLevel(%d)", v.getCenterLevel(), cfg.TownHallLevel)

			return false
		}

		switch strings.ToLower(cfg.BuildResource) {
		case "gold":
			_, total := v.collect_GetStorageGoldLimit()
			if cfg.BuildCost > total {
				SyncError(v.p.conn, "remove:cfg.BuildCost(%d) > total(%d) - gold", cfg.BuildCost, total)

				return false
			}
		case "food":
			_, total := v.collect_GetStorageFoodLimit()
			if cfg.BuildCost > total {
				SyncError(v.p.conn, "remove:cfg.BuildCost(%d) > total(%d) - food", cfg.BuildCost, total)

				return false
			}
		case "diamonds":
			if cfg.BuildCost > v.p.GetPlayerTotalGem() {
				SyncError(v.p.conn, "remove:cfg.BuildCost(%d) > total(%d) - diamonds", cfg.BuildCost, v.p.GetPlayerTotalGem())

				return false
			}
		}

		o.SetRemoveTime(uint32(time.Now().Unix()))

		v.buildings_NewWoking(idType, index, o)

		v.p.CostResource(cfg.BuildCost, strings.ToLower(cfg.BuildResource), proto.Lose_UpgradeBuilding)
	}

	return true
}

//收集资源
func (v *village) collect(idType rpc.BuildingId_IdType, index uint32) bool {
	v.buildings_ProcessUpgrade()

	return v.collect_Opt(idType, index)
}

func (v *village) cancel(idType rpc.BuildingId_IdType, index uint32) bool {
	v.buildings_ProcessUpgrade()

	obj := v.buildings_Get(idType, index)
	if obj == nil {
		return false
	}

	if o, ok := obj.(Upgradable); ok {
		if o.GetUpgradeTime() == 0 {
			return false
		}

		cfg := GetBuildingCfgByTypeId(idType, o.GetLevel()+1)
		if cfg == nil {
			SyncError(v.p.conn, "cancel:cfg1 == nil(level:%d)", o.GetLevel()+1)
			return false
		}

		v.p.GainResource(cfg.BuildCost/2, strings.ToLower(cfg.BuildResource), proto.Gain_CacelUpgradeBuilding)

		v.buildings_Cancel(idType, index)

		o.SetUpgradeTime(0)

		logger.Info("cancel:Now Level %d", o.GetLevel())

		if o.GetLevel() == 0 {
			v.buildings_Remove(idType, index)
		}
	}

	if o, ok := obj.(Removable); ok {
		if o.GetRemoveTime() == 0 {
			return false
		}

		cfg := GetBuildingCfgByTypeId(idType, 1)
		if cfg == nil {
			SyncError(v.p.conn, "cancel:cfg2 == nil(level:%d)", 1)
			return false
		}

		v.p.GainResource(cfg.BuildCost, strings.ToLower(cfg.BuildResource), proto.Gain_CacelUpgradeBuilding)

		v.buildings_Cancel(idType, index)

		o.SetRemoveTime(0)

		logger.Info("cancel:Remove")
	} else if o, ok := obj.(Movable); ok {
		cfg := GetBuildingCfgByTypeId(idType, 1)
		if cfg == nil {
			SyncError(v.p.conn, "cancel:cfg3 == nil")

			return false
		}

		v.mapRemoveFrom(o.GetP().GetX(), o.GetP().GetY(), cfg.BuildSize)
	}

	return true
}

//填充子弹
func (v *village) Reloading(idType rpc.BuildingId_IdType, index uint32) bool {
	logger.Info("Reloading! buildingType = ", idType, "Index = ", index, " \n")

	cfg := GetBuildingCfgByTypeId(idType, 1)
	logger.Info("CostType = ", cfg.AmmoResource, "\n")

	buildingObj := v.buildings_Get(idType, index)
	xbow := buildingObj.(*rpc.XBow)
	AmmoCount := xbow.GetAmmoCount()
	logger.Info("AmmoCount = ", AmmoCount, "\n")

	cfgAmmoCount := cfg.AmmoCount
	cfgAmmoCost := cfg.AmmoCost

	ScaleFactor := float32(AmmoCount) / float32(cfgAmmoCount)
	logger.Info("ScaleFactor = ", ScaleFactor, "\n")

	Factor := float32(cfgAmmoCost) * ScaleFactor
	logger.Info("Factor = ", Factor, "\n")

	SurplusAmmo := cfgAmmoCost - uint32(Factor)
	logger.Info("SurplusAmmo = ", SurplusAmmo, "\n")

	switch cfg.AmmoResource {
	case "Food":
		{
			foodCount, _ := v.GetFoodStorage()
			logger.Info("foodCount = ", foodCount, "\n")

			if SurplusAmmo == 0 {
				return false
			} else if SurplusAmmo > foodCount {
				return false
			} else {
				v.collect_CostFood(SurplusAmmo)
				logger.Info("CostFood = ", SurplusAmmo, "\n")

				xbow.SetAmmoCount(cfg.AmmoCount)
				logger.Info("AmmoCount = ", cfg.AmmoCount, "\n")
				return true
			}
		}
	case "Gold":
		{
			goldCount, _ := v.GetGoldStorage()
			logger.Info("goldCount = ", goldCount, "\n")

			if SurplusAmmo == 0 {
				return false
			} else if SurplusAmmo > goldCount {
				return false
			} else {
				v.collect_CostGold(SurplusAmmo)
				logger.Info("CostGold = ", SurplusAmmo, "\n")

				xbow.SetAmmoCount(cfg.AmmoCount)
				logger.Info("AmmoCount = ", cfg.AmmoCount, "\n")
				return true
			}
		}
	}
	return true
}

//转换攻击模式
func (v *village) ChangeMode(idType rpc.BuildingId_IdType, index uint32) bool {
	logger.Info("ChangeMode! buildingType = ", idType, "Index = ", index, " \n")

	buildingObj := v.buildings_Get(idType, index)
	xbow := buildingObj.(*rpc.XBow)
	AttackRange := xbow.GetAltAttackRange()
	logger.Info("AttackRangeCount = ", AttackRange, "\n")

	switch AttackRange {
	case 1:
		{
			AttackRange++
			xbow.SetAltAttackRange(AttackRange)
			logger.Info("AttackRangeEnd = ", AttackRange, "\n")
		}
	case 2:
		{
			AttackRange--
			xbow.SetAltAttackRange(AttackRange)
			logger.Info("AttackRangeEnd = ", AttackRange, "\n")
		}
	}
	return true
}
