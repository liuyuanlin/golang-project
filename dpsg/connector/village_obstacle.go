package connector

import (
	"golang-project/dpsg/logger"
	"golang-project/dpsg/rpc"
	"time"
)

var ObstacleWeight map[interface{}]uint32 = make(map[interface{}]uint32, 0)

func initConfig() {
	cfg := GetBuildingCfgByTypeId(rpc.BuildingId_Barrier1, 1)
	if cfg != nil {
		ObstacleWeight[rpc.BuildingId_Barrier1] = cfg.GenerateWeight
	}

	cfg = GetBuildingCfgByTypeId(rpc.BuildingId_Barrier2, 1)
	if cfg != nil {
		ObstacleWeight[rpc.BuildingId_Barrier2] = cfg.GenerateWeight
	}

	cfg = GetBuildingCfgByTypeId(rpc.BuildingId_Barrier3, 1)
	if cfg != nil {
		ObstacleWeight[rpc.BuildingId_Barrier3] = cfg.GenerateWeight
	}

	cfg = GetBuildingCfgByTypeId(rpc.BuildingId_Barrier4, 1)
	if cfg != nil {
		ObstacleWeight[rpc.BuildingId_Barrier4] = cfg.GenerateWeight
	}

	cfg = GetBuildingCfgByTypeId(rpc.BuildingId_Barrier5, 1)
	if cfg != nil {
		ObstacleWeight[rpc.BuildingId_Barrier5] = cfg.GenerateWeight
	}

	cfg = GetBuildingCfgByTypeId(rpc.BuildingId_Barrier6, 1)
	if cfg != nil {
		ObstacleWeight[rpc.BuildingId_Barrier6] = cfg.GenerateWeight
	}
}

func (v *village) obstacle_GetObstacleNum() uint32 {
	total := uint32(0)

	obs := v.buildings_GetAllOf(rpc.BuildingId_Barrier1)
	for _, _ = range obs {
		total += 1
	}

	obs = v.buildings_GetAllOf(rpc.BuildingId_Barrier2)
	for _, _ = range obs {
		total += 1
	}

	obs = v.buildings_GetAllOf(rpc.BuildingId_Barrier3)
	for _, _ = range obs {
		total += 1
	}

	obs = v.buildings_GetAllOf(rpc.BuildingId_Barrier4)
	for _, _ = range obs {
		total += 1
	}

	obs = v.buildings_GetAllOf(rpc.BuildingId_Barrier5)
	for _, _ = range obs {
		total += 1
	}

	obs = v.buildings_GetAllOf(rpc.BuildingId_Barrier6)
	for _, _ = range obs {
		total += 1
	}
	return total
}

func (v *village) obstacle_TryGenerate() bool {
	idType := RandomWeightTable(ObstacleWeight).(rpc.BuildingId_IdType)

	cfg := GetBuildingCfgByTypeId(idType, 1)
	if cfg == nil {
		return false
	}

	for i := 0; i < 10; i++ {
		x, y := randomGetPos()

		logger.Info("TryGenerateObs:type:%d, <%d, %d>(size:%d)", idType, x, y, cfg.BuildSize)

		if nil != v.buildings_Create(idType, x, y, true) {
			logger.Info("GenerateObs successed:type:%d, <%d, %d>(size:%d)", idType, x, y, cfg.BuildSize)
			return true
		}
	}

	return false
}

func (v *village) obstacle_ProcessGenerate() {
	if v.p.GetObstacleTime() == 0 {
		v.p.SetObstacleTime(uint32(time.Now().Unix()))
		return
	}

	deltaTime := uint32(time.Now().Unix()) - v.p.GetObstacleTime()

	logger.Info("obstacle_ProcessGenerate:deltaTime:%d (%d - %d)", deltaTime, uint32(time.Now().Unix()), v.p.GetObstacleTime())

	if deltaTime/GetGlobalCfg("BARRIER_REBORN_TIME") > 1 {
		v.p.SetObstacleTime(uint32(time.Now().Unix()))

		for i := uint32(0); i < deltaTime/GetGlobalCfg("BARRIER_REBORN_TIME"); i++ {
			if v.obstacle_GetObstacleNum() >= GetGlobalCfg("BARRIER_COUNT_LIMIT") {
				return
			}
			v.obstacle_TryGenerate()
		}
	}

	return
}
