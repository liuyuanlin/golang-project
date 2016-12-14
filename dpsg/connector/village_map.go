package connector

import (
	"golang-project/dpsg/logger"
	"golang-project/dpsg/rpc"
)

const (
	MAP_SIZE        = 44
	MAP_BLOCK_WIDTH = 4
)

func randomGetPos() (x uint32, y uint32) {
	return RandomNumber(0, MAP_SIZE-1), RandomNumber(0, MAP_SIZE-1)
}

func (v *village) mapInit() bool {
	ts("mapInit", v.vid)
	defer te("mapInit", v.vid)

	for i := uint32(rpc.BuildingId_Center); i < uint32(rpc.BuildingId_End); i++ {
		for index := uint32(0); ; index++ {
			obj := v.buildings_Get(rpc.BuildingId_IdType(i), index)
			if obj == nil {
				break
			}

			switch o := obj.(type) {
			case Movable:
				cfg := GetBuildingCfgByTypeId(rpc.BuildingId_IdType(i), 1)
				if cfg == nil {
					logger.Error("can't get cfg for building type %d", i)
					break
				}
				p := o.GetP()

				if v.mapCheckSpace(p.GetX(), p.GetY(), cfg.BuildSize, i, false) {
					v.mapInsertTo(p.GetX(), p.GetY(), cfg.BuildSize, i)
				} else {
					logger.Error("can' load map for %s, <%d:%d:<%d,%d> :%d> ", v.vid, i, index, p.GetX(), p.GetY(), cfg.BuildSize)
					//return false
				}
			}
		}
	}

	return true
}

func (v *village) mapCheckSpace(x, y, size uint32, buildType uint32, sendError bool) bool {
	if buildType >= uint32(rpc.BuildingId_Barrier1) && buildType <= uint32(rpc.BuildingId_Barrier6) {
		if x+size > MAP_SIZE || y+size > MAP_SIZE {
			if sendError {
				SyncError(v.p.conn, "mapCheckSpaceForObs: x(%d)+size(%d) >= MAP_SIZE(%d) || y(%d)+size(%d) >= MAP_SIZE(%d)", x, size, MAP_SIZE, y, size, MAP_SIZE)
			}

			return false
		}

		for idx_x := x; idx_x < x+size; idx_x++ {
			for idx_y := y; idx_y < y+size; idx_y++ {
				if v.maps[idx_x][idx_y] != 0 {
					if sendError {
						SyncError(v.p.conn, "mapCheckSpaceForObs:already has a building(type:%d) in place:<%d, %d>", v.maps[idx_x][idx_y], idx_x, idx_y)
					}

					return false
				}
			}
		}

		return true
	} else {
		if x < MAP_BLOCK_WIDTH || y < MAP_BLOCK_WIDTH {
			if sendError {
				SyncError(v.p.conn, "mapCheckSpace: x(%d) < (%d) || y(%d) < (%d)", x, MAP_BLOCK_WIDTH, y, MAP_BLOCK_WIDTH)
			}

			return false
		}

		if x+size > MAP_SIZE-MAP_BLOCK_WIDTH || y+size > MAP_SIZE-MAP_BLOCK_WIDTH {
			if sendError {
				SyncError(v.p.conn, "mapCheckSpace: x(%d)+size(%d) >= MAP_SIZE(%d - %d) || y(%d)+size(%d) >= MAP_SIZE(%d - %d)", x, size, MAP_SIZE, MAP_BLOCK_WIDTH, y, size, MAP_SIZE, MAP_BLOCK_WIDTH)
			}

			return false
		}

		for idx_x := x; idx_x < x+size; idx_x++ {
			for idx_y := y; idx_y < y+size; idx_y++ {
				if v.maps[idx_x][idx_y] != 0 {
					if sendError {
						SyncError(v.p.conn, "mapCheckSpace:already has a building(type:%d) in place:<%d, %d>", v.maps[idx_x][idx_y], idx_x, idx_y)
					}

					return false
				}
			}
		}

		return true
	}
	return true
}

func (v *village) mapRemoveFrom(x, y, size uint32) {
	for idx_x := x; idx_x < x+size; idx_x++ {
		for idx_y := y; idx_y < y+size; idx_y++ {
			v.maps[idx_x][idx_y] = 0
		}
	}
}

func (v *village) mapInsertTo(x, y, size uint32, buildType uint32) {
	for idx_x := x; idx_x < x+size; idx_x++ {
		for idx_y := y; idx_y < y+size; idx_y++ {
			v.maps[idx_x][idx_y] = buildType
		}
	}
}
