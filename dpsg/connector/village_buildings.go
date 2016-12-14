package connector

import (
	"golang-project/dpsg/logger"
	pt "golang-project/dpsg/proto"
	"golang-project/dpsg/rpc"
	"time"

	"github.com/golang/protobuf/proto"
)

func (v *village) buildings_Init() {
	//buildings map[rpc.BuildingId_IdType]*[]proto.Message
	ts("buildings_Init", v.vid)
	defer te("buildings_Init", v.vid)

	v.buildings = make(map[rpc.BuildingId_IdType][]proto.Message)

	v.buildings[rpc.BuildingId_Center] = make([]proto.Message, 0, 1)

	if v.Center != nil {
		v.buildings[rpc.BuildingId_Center] = append(v.buildings[rpc.BuildingId_Center], v.Center)
	}

	v.buildings[rpc.BuildingId_Barrack] = make([]proto.Message, 0)

	if v.Barrack != nil {
		for _, barrack := range v.Barrack {
			v.buildings[rpc.BuildingId_Barrack] = append(v.buildings[rpc.BuildingId_Barrack], barrack)
		}
	} else {
		v.Barrack = make([]*rpc.Barrack, 0)
	}

	v.buildings[rpc.BuildingId_Farm] = make([]proto.Message, 0)

	if v.Farm != nil {
		for _, fram := range v.Farm {
			v.buildings[rpc.BuildingId_Farm] = append(v.buildings[rpc.BuildingId_Farm], fram)
		}
	} else {
		v.Farm = make([]*rpc.Farm, 0)
	}

	v.buildings[rpc.BuildingId_Laboratory] = make([]proto.Message, 0)

	if v.Laboratory != nil {
		for _, value := range v.Laboratory {
			v.buildings[rpc.BuildingId_Laboratory] = append(v.buildings[rpc.BuildingId_Laboratory], value)
		}
	} else {
		v.Laboratory = make([]*rpc.Laboratory, 0)
	}

	v.buildings[rpc.BuildingId_Wall] = make([]proto.Message, 0)

	if v.Wall != nil {
		for _, value := range v.Wall {
			v.buildings[rpc.BuildingId_Wall] = append(v.buildings[rpc.BuildingId_Wall], value)
		}
	} else {
		v.Wall = make([]*rpc.Wall, 0)
	}

	v.buildings[rpc.BuildingId_Worker] = make([]proto.Message, 0)

	if v.Worker != nil {
		for _, value := range v.Worker {
			v.buildings[rpc.BuildingId_Worker] = append(v.buildings[rpc.BuildingId_Worker], value)
		}
	} else {
		v.Worker = make([]*rpc.Worker, 0)
	}

	v.buildings[rpc.BuildingId_FoodStorage] = make([]proto.Message, 0)

	if v.Foodstorage != nil {
		for _, value := range v.Foodstorage {
			v.buildings[rpc.BuildingId_FoodStorage] = append(v.buildings[rpc.BuildingId_FoodStorage], value)
		}
	} else {
		v.Foodstorage = make([]*rpc.FoodStorage, 0)
	}

	v.buildings[rpc.BuildingId_GoldMine] = make([]proto.Message, 0)

	if v.Goldmine != nil {
		for _, value := range v.Goldmine {
			v.buildings[rpc.BuildingId_GoldMine] = append(v.buildings[rpc.BuildingId_GoldMine], value)
		}
	} else {
		v.Goldmine = make([]*rpc.GoldMine, 0)
	}

	v.buildings[rpc.BuildingId_GoldStorage] = make([]proto.Message, 0)

	if v.Goldstorage != nil {
		for _, value := range v.Goldstorage {
			v.buildings[rpc.BuildingId_GoldStorage] = append(v.buildings[rpc.BuildingId_GoldStorage], value)
		}
	} else {
		v.Goldstorage = make([]*rpc.GoldStorage, 0)
	}

	v.buildings[rpc.BuildingId_TroopHousing] = make([]proto.Message, 0)

	if v.Troophosing != nil {
		for _, value := range v.Troophosing {
			v.buildings[rpc.BuildingId_TroopHousing] = append(v.buildings[rpc.BuildingId_TroopHousing], value)
		}
	} else {
		v.Troophosing = make([]*rpc.TroopHousing, 0)
	}

	v.buildings[rpc.BuildingId_ArcherTower] = make([]proto.Message, 0)
	if v.Archertower != nil {
		for _, value := range v.Archertower {
			v.buildings[rpc.BuildingId_ArcherTower] = append(v.buildings[rpc.BuildingId_ArcherTower], value)
		}
	} else {
		v.Archertower = make([]*rpc.ArcherTower, 0)
	}

	v.buildings[rpc.BuildingId_Cannon] = make([]proto.Message, 0)
	if v.Cannon != nil {
		for _, value := range v.Cannon {
			v.buildings[rpc.BuildingId_Cannon] = append(v.buildings[rpc.BuildingId_Cannon], value)
		}
	} else {
		v.Cannon = make([]*rpc.Cannon, 0)
	}

	v.buildings[rpc.BuildingId_WizardTower] = make([]proto.Message, 0)
	if v.Wizardtower != nil {
		for _, value := range v.Wizardtower {
			v.buildings[rpc.BuildingId_WizardTower] = append(v.buildings[rpc.BuildingId_WizardTower], value)
		}
	} else {
		v.Wizardtower = make([]*rpc.WizardTower, 0)
	}

	v.buildings[rpc.BuildingId_AirDefense] = make([]proto.Message, 0)
	if v.Airdefense != nil {
		for _, value := range v.Airdefense {
			v.buildings[rpc.BuildingId_AirDefense] = append(v.buildings[rpc.BuildingId_AirDefense], value)
		}
	} else {
		v.Airdefense = make([]*rpc.AirDefense, 0)
	}

	v.buildings[rpc.BuildingId_Mortar] = make([]proto.Message, 0)
	if v.Mortar != nil {
		for _, value := range v.Mortar {
			v.buildings[rpc.BuildingId_Mortar] = append(v.buildings[rpc.BuildingId_Mortar], value)
		}
	} else {
		v.Mortar = make([]*rpc.Mortar, 0)
	}

	v.buildings[rpc.BuildingId_TeslaTower] = make([]proto.Message, 0)
	if v.Teslatower != nil {
		for _, value := range v.Teslatower {
			v.buildings[rpc.BuildingId_TeslaTower] = append(v.buildings[rpc.BuildingId_TeslaTower], value)
		}
	} else {
		v.Teslatower = make([]*rpc.TeslaTower, 0)
	}

	v.buildings[rpc.BuildingId_XBow] = make([]proto.Message, 0)
	if v.Xbow != nil {
		for _, value := range v.Xbow {
			v.buildings[rpc.BuildingId_XBow] = append(v.buildings[rpc.BuildingId_XBow], value)
		}
	} else {
		v.Xbow = make([]*rpc.XBow, 0)
	}

	v.buildings[rpc.BuildingId_AllianceCastle] = make([]proto.Message, 0)
	if v.Alliancecastle != nil {
		for _, value := range v.Alliancecastle {
			v.buildings[rpc.BuildingId_AllianceCastle] = append(v.buildings[rpc.BuildingId_AllianceCastle], value)
		}
	} else {
		v.Alliancecastle = make([]*rpc.AllianceCastle, 0)
	}

	v.buildings[rpc.BuildingId_SpellForge] = make([]proto.Message, 0)
	if v.Spellforge != nil {
		for _, value := range v.Spellforge {
			v.buildings[rpc.BuildingId_SpellForge] = append(v.buildings[rpc.BuildingId_SpellForge], value)
		}
	} else {
		v.Spellforge = make([]*rpc.SpellForge, 0)
	}

	v.buildings[rpc.BuildingId_Deco1] = make([]proto.Message, 0)
	if v.Deco1 != nil {
		for _, value := range v.Deco1 {
			v.buildings[rpc.BuildingId_Deco1] = append(v.buildings[rpc.BuildingId_Deco1], value)
		}
	} else {
		v.Deco1 = make([]*rpc.Deco1, 0)
	}

	v.buildings[rpc.BuildingId_Deco2] = make([]proto.Message, 0)
	if v.Deco2 != nil {
		for _, value := range v.Deco2 {
			v.buildings[rpc.BuildingId_Deco2] = append(v.buildings[rpc.BuildingId_Deco2], value)
		}
	} else {
		v.Deco2 = make([]*rpc.Deco2, 0)
	}

	v.buildings[rpc.BuildingId_Deco3] = make([]proto.Message, 0)
	if v.Deco3 != nil {
		for _, value := range v.Deco3 {
			v.buildings[rpc.BuildingId_Deco3] = append(v.buildings[rpc.BuildingId_Deco3], value)
		}
	} else {
		v.Deco3 = make([]*rpc.Deco3, 0)
	}

	v.buildings[rpc.BuildingId_Deco4] = make([]proto.Message, 0)
	if v.Deco4 != nil {
		for _, value := range v.Deco4 {
			v.buildings[rpc.BuildingId_Deco4] = append(v.buildings[rpc.BuildingId_Deco4], value)
		}
	} else {
		v.Deco4 = make([]*rpc.Deco4, 0)
	}

	v.buildings[rpc.BuildingId_Deco5] = make([]proto.Message, 0)
	if v.Deco5 != nil {
		for _, value := range v.Deco5 {
			v.buildings[rpc.BuildingId_Deco5] = append(v.buildings[rpc.BuildingId_Deco5], value)
		}
	} else {
		v.Deco5 = make([]*rpc.Deco5, 0)
	}

	v.buildings[rpc.BuildingId_Deco6] = make([]proto.Message, 0)
	if v.Deco6 != nil {
		for _, value := range v.Deco6 {
			v.buildings[rpc.BuildingId_Deco6] = append(v.buildings[rpc.BuildingId_Deco6], value)
		}
	} else {
		v.Deco6 = make([]*rpc.Deco6, 0)
	}

	v.buildings[rpc.BuildingId_Deco7] = make([]proto.Message, 0)
	if v.Deco7 != nil {
		for _, value := range v.Deco7 {
			v.buildings[rpc.BuildingId_Deco7] = append(v.buildings[rpc.BuildingId_Deco7], value)
		}
	} else {
		v.Deco7 = make([]*rpc.Deco7, 0)
	}

	v.buildings[rpc.BuildingId_Deco8] = make([]proto.Message, 0)
	if v.Deco8 != nil {
		for _, value := range v.Deco8 {
			v.buildings[rpc.BuildingId_Deco8] = append(v.buildings[rpc.BuildingId_Deco8], value)
		}
	} else {
		v.Deco8 = make([]*rpc.Deco8, 0)
	}

	v.buildings[rpc.BuildingId_Deco9] = make([]proto.Message, 0)
	if v.Deco9 != nil {
		for _, value := range v.Deco9 {
			v.buildings[rpc.BuildingId_Deco9] = append(v.buildings[rpc.BuildingId_Deco9], value)
		}
	} else {
		v.Deco9 = make([]*rpc.Deco9, 0)
	}

	v.buildings[rpc.BuildingId_Deco10] = make([]proto.Message, 0)
	if v.Deco10 != nil {
		for _, value := range v.Deco10 {
			v.buildings[rpc.BuildingId_Deco10] = append(v.buildings[rpc.BuildingId_Deco10], value)
		}
	} else {
		v.Deco10 = make([]*rpc.Deco10, 0)
	}

	v.buildings[rpc.BuildingId_Deco11] = make([]proto.Message, 0)
	if v.Deco11 != nil {
		for _, value := range v.Deco11 {
			v.buildings[rpc.BuildingId_Deco11] = append(v.buildings[rpc.BuildingId_Deco11], value)
		}
	} else {
		v.Deco11 = make([]*rpc.Deco11, 0)
	}

	v.buildings[rpc.BuildingId_Deco12] = make([]proto.Message, 0)
	if v.Deco12 != nil {
		for _, value := range v.Deco12 {
			v.buildings[rpc.BuildingId_Deco12] = append(v.buildings[rpc.BuildingId_Deco12], value)
		}
	} else {
		v.Deco12 = make([]*rpc.Deco12, 0)
	}

	v.buildings[rpc.BuildingId_Deco13] = make([]proto.Message, 0)
	if v.Deco13 != nil {
		for _, value := range v.Deco13 {
			v.buildings[rpc.BuildingId_Deco13] = append(v.buildings[rpc.BuildingId_Deco13], value)
		}
	} else {
		v.Deco13 = make([]*rpc.Deco13, 0)
	}

	v.buildings[rpc.BuildingId_Deco14] = make([]proto.Message, 0)
	if v.Deco14 != nil {
		for _, value := range v.Deco14 {
			v.buildings[rpc.BuildingId_Deco14] = append(v.buildings[rpc.BuildingId_Deco14], value)
		}
	} else {
		v.Deco14 = make([]*rpc.Deco14, 0)
	}

	v.buildings[rpc.BuildingId_Deco15] = make([]proto.Message, 0)
	if v.Deco15 != nil {
		for _, value := range v.Deco15 {
			v.buildings[rpc.BuildingId_Deco15] = append(v.buildings[rpc.BuildingId_Deco15], value)
		}
	} else {
		v.Deco15 = make([]*rpc.Deco15, 0)
	}

	v.buildings[rpc.BuildingId_Deco16] = make([]proto.Message, 0)
	if v.Deco16 != nil {
		for _, value := range v.Deco16 {
			v.buildings[rpc.BuildingId_Deco16] = append(v.buildings[rpc.BuildingId_Deco16], value)
		}
	} else {
		v.Deco16 = make([]*rpc.Deco16, 0)
	}

	v.buildings[rpc.BuildingId_Deco17] = make([]proto.Message, 0)
	if v.Deco17 != nil {
		for _, value := range v.Deco17 {
			v.buildings[rpc.BuildingId_Deco17] = append(v.buildings[rpc.BuildingId_Deco17], value)
		}
	} else {
		v.Deco17 = make([]*rpc.Deco17, 0)
	}

	v.buildings[rpc.BuildingId_Deco18] = make([]proto.Message, 0)
	if v.Deco18 != nil {
		for _, value := range v.Deco18 {
			v.buildings[rpc.BuildingId_Deco18] = append(v.buildings[rpc.BuildingId_Deco18], value)
		}
	} else {
		v.Deco18 = make([]*rpc.Deco18, 0)
	}

	v.buildings[rpc.BuildingId_Deco19] = make([]proto.Message, 0)
	if v.Deco19 != nil {
		for _, value := range v.Deco19 {
			v.buildings[rpc.BuildingId_Deco19] = append(v.buildings[rpc.BuildingId_Deco19], value)
		}
	} else {
		v.Deco19 = make([]*rpc.Deco19, 0)
	}

	v.buildings[rpc.BuildingId_Deco20] = make([]proto.Message, 0)
	if v.Deco20 != nil {
		for _, value := range v.Deco20 {
			v.buildings[rpc.BuildingId_Deco20] = append(v.buildings[rpc.BuildingId_Deco20], value)
		}
	} else {
		v.Deco20 = make([]*rpc.Deco20, 0)
	}

	v.buildings[rpc.BuildingId_Deco21] = make([]proto.Message, 0)
	if v.Deco21 != nil {
		for _, value := range v.Deco21 {
			v.buildings[rpc.BuildingId_Deco21] = append(v.buildings[rpc.BuildingId_Deco21], value)
		}
	} else {
		v.Deco21 = make([]*rpc.Deco21, 0)
	}

	v.buildings[rpc.BuildingId_Deco22] = make([]proto.Message, 0)
	if v.Deco22 != nil {
		for _, value := range v.Deco22 {
			v.buildings[rpc.BuildingId_Deco22] = append(v.buildings[rpc.BuildingId_Deco22], value)
		}
	} else {
		v.Deco22 = make([]*rpc.Deco22, 0)
	}

	v.buildings[rpc.BuildingId_Deco23] = make([]proto.Message, 0)
	if v.Deco23 != nil {
		for _, value := range v.Deco23 {
			v.buildings[rpc.BuildingId_Deco23] = append(v.buildings[rpc.BuildingId_Deco23], value)
		}
	} else {
		v.Deco23 = make([]*rpc.Deco23, 0)
	}

	v.buildings[rpc.BuildingId_Deco24] = make([]proto.Message, 0)
	if v.Deco24 != nil {
		for _, value := range v.Deco24 {
			v.buildings[rpc.BuildingId_Deco24] = append(v.buildings[rpc.BuildingId_Deco24], value)
		}
	} else {
		v.Deco24 = make([]*rpc.Deco24, 0)
	}

	v.buildings[rpc.BuildingId_Bomb] = make([]proto.Message, 0)
	if v.Bomb != nil {
		for _, value := range v.Bomb {
			v.buildings[rpc.BuildingId_Bomb] = append(v.buildings[rpc.BuildingId_Bomb], value)
		}
	} else {
		v.Bomb = make([]*rpc.Bomb, 0)
	}

	v.buildings[rpc.BuildingId_GiantBomb] = make([]proto.Message, 0)
	if v.Giantbomb != nil {
		for _, value := range v.Giantbomb {
			v.buildings[rpc.BuildingId_GiantBomb] = append(v.buildings[rpc.BuildingId_GiantBomb], value)
		}
	} else {
		v.Giantbomb = make([]*rpc.GiantBomb, 0)
	}

	v.buildings[rpc.BuildingId_Eject] = make([]proto.Message, 0)
	if v.Eject != nil {
		for _, value := range v.Eject {
			v.buildings[rpc.BuildingId_Eject] = append(v.buildings[rpc.BuildingId_Eject], value)
		}
	} else {
		v.Eject = make([]*rpc.Eject, 0)
	}

	v.buildings[rpc.BuildingId_GeneralHouse] = make([]proto.Message, 0)
	if v.Generalhouse != nil {
		for _, value := range v.Generalhouse {
			v.buildings[rpc.BuildingId_GeneralHouse] = append(v.buildings[rpc.BuildingId_GeneralHouse], value)
		}
	} else {
		v.Generalhouse = make([]*rpc.GeneralHouse, 0)
	}

	v.buildings[rpc.BuildingId_Barrier1] = make([]proto.Message, 0)
	if v.Barrier1 != nil {
		for _, value := range v.Barrier1 {
			v.buildings[rpc.BuildingId_Barrier1] = append(v.buildings[rpc.BuildingId_Barrier1], value)
		}
	} else {
		v.Barrier1 = make([]*rpc.Barrier1, 0)
	}

	v.buildings[rpc.BuildingId_Barrier2] = make([]proto.Message, 0)
	if v.Barrier2 != nil {
		for _, value := range v.Barrier2 {
			v.buildings[rpc.BuildingId_Barrier2] = append(v.buildings[rpc.BuildingId_Barrier2], value)
		}
	} else {
		v.Barrier2 = make([]*rpc.Barrier2, 0)
	}

	v.buildings[rpc.BuildingId_Barrier3] = make([]proto.Message, 0)
	if v.Barrier3 != nil {
		for _, value := range v.Barrier3 {
			v.buildings[rpc.BuildingId_Barrier3] = append(v.buildings[rpc.BuildingId_Barrier3], value)
		}
	} else {
		v.Barrier3 = make([]*rpc.Barrier3, 0)
	}

	v.buildings[rpc.BuildingId_Barrier4] = make([]proto.Message, 0)
	if v.Barrier4 != nil {
		for _, value := range v.Barrier4 {
			v.buildings[rpc.BuildingId_Barrier4] = append(v.buildings[rpc.BuildingId_Barrier4], value)
		}
	} else {
		v.Barrier4 = make([]*rpc.Barrier4, 0)
	}

	v.buildings[rpc.BuildingId_Barrier5] = make([]proto.Message, 0)
	if v.Barrier5 != nil {
		for _, value := range v.Barrier5 {
			v.buildings[rpc.BuildingId_Barrier5] = append(v.buildings[rpc.BuildingId_Barrier5], value)
		}
	} else {
		v.Barrier5 = make([]*rpc.Barrier5, 0)
	}

	v.buildings[rpc.BuildingId_Barrier6] = make([]proto.Message, 0)
	if v.Barrier6 != nil {
		for _, value := range v.Barrier6 {
			v.buildings[rpc.BuildingId_Barrier6] = append(v.buildings[rpc.BuildingId_Barrier6], value)
		}
	} else {
		v.Barrier6 = make([]*rpc.Barrier6, 0)
	}
}

func (v *village) buildings_New(idType rpc.BuildingId_IdType) interface{} {
	switch idType {
	case rpc.BuildingId_Center:
		value := &rpc.Center{}
		v.Center = value
		v.buildings[rpc.BuildingId_Center] = append(v.buildings[rpc.BuildingId_Center], value)
		return value

	case rpc.BuildingId_Barrack:
		value := &rpc.Barrack{}
		v.Barrack = append(v.Barrack, value)
		v.buildings[rpc.BuildingId_Barrack] = append(v.buildings[rpc.BuildingId_Barrack], value)
		return value

	case rpc.BuildingId_Farm:
		value := &rpc.Farm{}
		value.SetResCount(0)
		v.Farm = append(v.Farm, value)
		v.buildings[rpc.BuildingId_Farm] = append(v.buildings[rpc.BuildingId_Farm], value)
		return value

	case rpc.BuildingId_Laboratory:
		value := &rpc.Laboratory{}
		v.Laboratory = append(v.Laboratory, value)
		v.buildings[rpc.BuildingId_Laboratory] = append(v.buildings[rpc.BuildingId_Laboratory], value)
		return value

	case rpc.BuildingId_Wall:
		value := &rpc.Wall{}
		v.Wall = append(v.Wall, value)
		v.buildings[rpc.BuildingId_Wall] = append(v.buildings[rpc.BuildingId_Wall], value)
		return value

	case rpc.BuildingId_Worker:
		value := &rpc.Worker{}
		v.Worker = append(v.Worker, value)
		v.buildings[rpc.BuildingId_Worker] = append(v.buildings[rpc.BuildingId_Worker], value)
		return value

	case rpc.BuildingId_FoodStorage:
		value := &rpc.FoodStorage{}
		v.Foodstorage = append(v.Foodstorage, value)
		v.buildings[rpc.BuildingId_FoodStorage] = append(v.buildings[rpc.BuildingId_FoodStorage], value)
		return value

	case rpc.BuildingId_GoldMine:
		value := &rpc.GoldMine{}
		value.SetResCount(0)
		v.Goldmine = append(v.Goldmine, value)
		v.buildings[rpc.BuildingId_GoldMine] = append(v.buildings[rpc.BuildingId_GoldMine], value)
		return value

	case rpc.BuildingId_GoldStorage:
		value := &rpc.GoldStorage{}
		v.Goldstorage = append(v.Goldstorage, value)
		v.buildings[rpc.BuildingId_GoldStorage] = append(v.buildings[rpc.BuildingId_GoldStorage], value)
		return value

	case rpc.BuildingId_TroopHousing:
		value := &rpc.TroopHousing{}
		v.Troophosing = append(v.Troophosing, value)
		v.buildings[rpc.BuildingId_TroopHousing] = append(v.buildings[rpc.BuildingId_TroopHousing], value)
		return value

	case rpc.BuildingId_ArcherTower:
		value := &rpc.ArcherTower{}
		v.Archertower = append(v.Archertower, value)
		v.buildings[rpc.BuildingId_ArcherTower] = append(v.buildings[rpc.BuildingId_ArcherTower], value)
		return value
	case rpc.BuildingId_Cannon:
		value := &rpc.Cannon{}
		v.Cannon = append(v.Cannon, value)
		v.buildings[rpc.BuildingId_Cannon] = append(v.buildings[rpc.BuildingId_Cannon], value)
		return value
	case rpc.BuildingId_WizardTower:
		value := &rpc.WizardTower{}
		v.Wizardtower = append(v.Wizardtower, value)
		v.buildings[rpc.BuildingId_WizardTower] = append(v.buildings[rpc.BuildingId_WizardTower], value)
		return value
	case rpc.BuildingId_AirDefense:
		value := &rpc.AirDefense{}
		v.Airdefense = append(v.Airdefense, value)
		v.buildings[rpc.BuildingId_AirDefense] = append(v.buildings[rpc.BuildingId_AirDefense], value)
		return value
	case rpc.BuildingId_Mortar:
		value := &rpc.Mortar{}
		v.Mortar = append(v.Mortar, value)
		v.buildings[rpc.BuildingId_Mortar] = append(v.buildings[rpc.BuildingId_Mortar], value)
		return value
	case rpc.BuildingId_TeslaTower:
		value := &rpc.TeslaTower{}
		v.Teslatower = append(v.Teslatower, value)
		v.buildings[rpc.BuildingId_TeslaTower] = append(v.buildings[rpc.BuildingId_TeslaTower], value)
		return value
	case rpc.BuildingId_XBow:
		value := &rpc.XBow{}
		v.Xbow = append(v.Xbow, value)
		v.buildings[rpc.BuildingId_XBow] = append(v.buildings[rpc.BuildingId_XBow], value)
		return value
	case rpc.BuildingId_AllianceCastle:
		value := &rpc.AllianceCastle{}
		v.Alliancecastle = append(v.Alliancecastle, value)
		v.buildings[rpc.BuildingId_AllianceCastle] = append(v.buildings[rpc.BuildingId_AllianceCastle], value)
		return value
	case rpc.BuildingId_SpellForge:
		value := &rpc.SpellForge{}
		v.Spellforge = append(v.Spellforge, value)
		v.buildings[rpc.BuildingId_SpellForge] = append(v.buildings[rpc.BuildingId_SpellForge], value)
		return value
	case rpc.BuildingId_Deco1:
		value := &rpc.Deco1{}
		v.Deco1 = append(v.Deco1, value)
		v.buildings[rpc.BuildingId_Deco1] = append(v.buildings[rpc.BuildingId_Deco1], value)
		return value
	case rpc.BuildingId_Deco2:
		value := &rpc.Deco2{}
		v.Deco2 = append(v.Deco2, value)
		v.buildings[rpc.BuildingId_Deco2] = append(v.buildings[rpc.BuildingId_Deco2], value)
		return value
	case rpc.BuildingId_Deco3:
		value := &rpc.Deco3{}
		v.Deco3 = append(v.Deco3, value)
		v.buildings[rpc.BuildingId_Deco3] = append(v.buildings[rpc.BuildingId_Deco3], value)
		return value
	case rpc.BuildingId_Deco4:
		value := &rpc.Deco4{}
		v.Deco4 = append(v.Deco4, value)
		v.buildings[rpc.BuildingId_Deco4] = append(v.buildings[rpc.BuildingId_Deco4], value)
		return value
	case rpc.BuildingId_Deco5:
		value := &rpc.Deco5{}
		v.Deco5 = append(v.Deco5, value)
		v.buildings[rpc.BuildingId_Deco5] = append(v.buildings[rpc.BuildingId_Deco5], value)
		return value
	case rpc.BuildingId_Deco6:
		value := &rpc.Deco6{}
		v.Deco6 = append(v.Deco6, value)
		v.buildings[rpc.BuildingId_Deco6] = append(v.buildings[rpc.BuildingId_Deco6], value)
		return value
	case rpc.BuildingId_Deco7:
		value := &rpc.Deco7{}
		v.Deco7 = append(v.Deco7, value)
		v.buildings[rpc.BuildingId_Deco7] = append(v.buildings[rpc.BuildingId_Deco7], value)
		return value
	case rpc.BuildingId_Deco8:
		value := &rpc.Deco8{}
		v.Deco8 = append(v.Deco8, value)
		v.buildings[rpc.BuildingId_Deco8] = append(v.buildings[rpc.BuildingId_Deco8], value)
		return value
	case rpc.BuildingId_Deco9:
		value := &rpc.Deco9{}
		v.Deco9 = append(v.Deco9, value)
		v.buildings[rpc.BuildingId_Deco9] = append(v.buildings[rpc.BuildingId_Deco9], value)
		return value
	case rpc.BuildingId_Deco10:
		value := &rpc.Deco10{}
		v.Deco10 = append(v.Deco10, value)
		v.buildings[rpc.BuildingId_Deco10] = append(v.buildings[rpc.BuildingId_Deco10], value)
		return value
	case rpc.BuildingId_Deco11:
		value := &rpc.Deco11{}
		v.Deco11 = append(v.Deco11, value)
		v.buildings[rpc.BuildingId_Deco11] = append(v.buildings[rpc.BuildingId_Deco11], value)
		return value
	case rpc.BuildingId_Deco12:
		value := &rpc.Deco12{}
		v.Deco12 = append(v.Deco12, value)
		v.buildings[rpc.BuildingId_Deco12] = append(v.buildings[rpc.BuildingId_Deco12], value)
		return value
	case rpc.BuildingId_Deco13:
		value := &rpc.Deco13{}
		v.Deco13 = append(v.Deco13, value)
		v.buildings[rpc.BuildingId_Deco13] = append(v.buildings[rpc.BuildingId_Deco13], value)
		return value
	case rpc.BuildingId_Deco14:
		value := &rpc.Deco14{}
		v.Deco14 = append(v.Deco14, value)
		v.buildings[rpc.BuildingId_Deco14] = append(v.buildings[rpc.BuildingId_Deco14], value)
		return value
	case rpc.BuildingId_Deco15:
		value := &rpc.Deco15{}
		v.Deco15 = append(v.Deco15, value)
		v.buildings[rpc.BuildingId_Deco15] = append(v.buildings[rpc.BuildingId_Deco15], value)
		return value
	case rpc.BuildingId_Deco16:
		value := &rpc.Deco16{}
		v.Deco16 = append(v.Deco16, value)
		v.buildings[rpc.BuildingId_Deco16] = append(v.buildings[rpc.BuildingId_Deco16], value)
		return value
	case rpc.BuildingId_Deco17:
		value := &rpc.Deco17{}
		v.Deco17 = append(v.Deco17, value)
		v.buildings[rpc.BuildingId_Deco17] = append(v.buildings[rpc.BuildingId_Deco17], value)
		return value
	case rpc.BuildingId_Deco18:
		value := &rpc.Deco18{}
		v.Deco18 = append(v.Deco18, value)
		v.buildings[rpc.BuildingId_Deco18] = append(v.buildings[rpc.BuildingId_Deco18], value)
		return value
	case rpc.BuildingId_Deco19:
		value := &rpc.Deco19{}
		v.Deco19 = append(v.Deco19, value)
		v.buildings[rpc.BuildingId_Deco19] = append(v.buildings[rpc.BuildingId_Deco19], value)
		return value
	case rpc.BuildingId_Deco20:
		value := &rpc.Deco20{}
		v.Deco20 = append(v.Deco20, value)
		v.buildings[rpc.BuildingId_Deco20] = append(v.buildings[rpc.BuildingId_Deco20], value)
		return value
	case rpc.BuildingId_Deco21:
		value := &rpc.Deco21{}
		v.Deco21 = append(v.Deco21, value)
		v.buildings[rpc.BuildingId_Deco21] = append(v.buildings[rpc.BuildingId_Deco21], value)
		return value
	case rpc.BuildingId_Deco22:
		value := &rpc.Deco22{}
		v.Deco22 = append(v.Deco22, value)
		v.buildings[rpc.BuildingId_Deco22] = append(v.buildings[rpc.BuildingId_Deco22], value)
		return value
	case rpc.BuildingId_Deco23:
		value := &rpc.Deco23{}
		v.Deco23 = append(v.Deco23, value)
		v.buildings[rpc.BuildingId_Deco23] = append(v.buildings[rpc.BuildingId_Deco23], value)
		return value
	case rpc.BuildingId_Deco24:
		value := &rpc.Deco24{}
		v.Deco24 = append(v.Deco24, value)
		v.buildings[rpc.BuildingId_Deco24] = append(v.buildings[rpc.BuildingId_Deco24], value)
		return value
	case rpc.BuildingId_Bomb:
		value := &rpc.Bomb{}
		v.Bomb = append(v.Bomb, value)
		v.buildings[rpc.BuildingId_Bomb] = append(v.buildings[rpc.BuildingId_Bomb], value)
		return value
	case rpc.BuildingId_GiantBomb:
		value := &rpc.GiantBomb{}
		v.Giantbomb = append(v.Giantbomb, value)
		v.buildings[rpc.BuildingId_GiantBomb] = append(v.buildings[rpc.BuildingId_GiantBomb], value)
		return value
	case rpc.BuildingId_Eject:
		value := &rpc.Eject{}
		v.Eject = append(v.Eject, value)
		v.buildings[idType] = append(v.buildings[idType], value)
		return value
	case rpc.BuildingId_GeneralHouse:
		value := &rpc.GeneralHouse{}
		v.Generalhouse = append(v.Generalhouse, value)
		v.buildings[rpc.BuildingId_GeneralHouse] = append(v.buildings[rpc.BuildingId_GeneralHouse], value)
		return value
	case rpc.BuildingId_Barrier1:
		value := &rpc.Barrier1{}
		v.Barrier1 = append(v.Barrier1, value)
		v.buildings[rpc.BuildingId_Barrier1] = append(v.buildings[rpc.BuildingId_Barrier1], value)
		return value
	case rpc.BuildingId_Barrier2:
		value := &rpc.Barrier2{}
		v.Barrier2 = append(v.Barrier2, value)
		v.buildings[rpc.BuildingId_Barrier2] = append(v.buildings[rpc.BuildingId_Barrier2], value)
		return value
	case rpc.BuildingId_Barrier3:
		value := &rpc.Barrier3{}
		v.Barrier3 = append(v.Barrier3, value)
		v.buildings[rpc.BuildingId_Barrier3] = append(v.buildings[rpc.BuildingId_Barrier3], value)
		return value
	case rpc.BuildingId_Barrier4:
		value := &rpc.Barrier4{}
		v.Barrier4 = append(v.Barrier4, value)
		v.buildings[rpc.BuildingId_Barrier4] = append(v.buildings[rpc.BuildingId_Barrier4], value)
		return value
	case rpc.BuildingId_Barrier5:
		value := &rpc.Barrier5{}
		v.Barrier5 = append(v.Barrier5, value)
		v.buildings[rpc.BuildingId_Barrier5] = append(v.buildings[rpc.BuildingId_Barrier5], value)
		return value
	case rpc.BuildingId_Barrier6:
		value := &rpc.Barrier6{}
		v.Barrier6 = append(v.Barrier6, value)
		v.buildings[rpc.BuildingId_Barrier6] = append(v.buildings[rpc.BuildingId_Barrier6], value)
		return value
	}

	return nil
}

func (v *village) buildings_Remove(idType rpc.BuildingId_IdType, index uint32) {
	switch idType {

	case rpc.BuildingId_Barrack:

		v.Barrack = append(v.Barrack[:index], v.Barrack[index+1:]...)
		v.buildings[rpc.BuildingId_Barrack] = append(v.buildings[rpc.BuildingId_Barrack][:index], v.buildings[rpc.BuildingId_Barrack][index+1:]...)

	case rpc.BuildingId_Farm:

		v.Farm = append(v.Farm[:index], v.Farm[index+1:]...)
		v.buildings[rpc.BuildingId_Farm] = append(v.buildings[rpc.BuildingId_Farm][:index], v.buildings[rpc.BuildingId_Farm][index+1:]...)

	case rpc.BuildingId_Laboratory:

		v.Laboratory = append(v.Laboratory[:index], v.Laboratory[index+1:]...)
		v.buildings[rpc.BuildingId_Laboratory] = append(v.buildings[rpc.BuildingId_Laboratory][:index], v.buildings[rpc.BuildingId_Laboratory][index+1:]...)

	case rpc.BuildingId_Wall:

		v.Wall = append(v.Wall[:index], v.Wall[index+1:]...)
		v.buildings[rpc.BuildingId_Wall] = append(v.buildings[rpc.BuildingId_Wall][:index], v.buildings[rpc.BuildingId_Wall][index+1:]...)

	case rpc.BuildingId_Worker:

		v.Worker = append(v.Worker[:index], v.Worker[index+1:]...)
		v.buildings[rpc.BuildingId_Worker] = append(v.buildings[rpc.BuildingId_Worker][:index], v.buildings[rpc.BuildingId_Worker][index+1:]...)

	case rpc.BuildingId_FoodStorage:

		v.Foodstorage = append(v.Foodstorage[:index], v.Foodstorage[index+1:]...)
		v.buildings[rpc.BuildingId_FoodStorage] = append(v.buildings[rpc.BuildingId_FoodStorage][:index], v.buildings[rpc.BuildingId_FoodStorage][index+1:]...)

	case rpc.BuildingId_GoldMine:

		v.Goldmine = append(v.Goldmine[:index], v.Goldmine[index+1:]...)

		v.buildings[rpc.BuildingId_GoldMine] = append(v.buildings[rpc.BuildingId_GoldMine][:index], v.buildings[rpc.BuildingId_GoldMine][index+1:]...)

	case rpc.BuildingId_GoldStorage:

		v.Goldstorage = append(v.Goldstorage[:index], v.Goldstorage[index+1:]...)

		v.buildings[rpc.BuildingId_GoldStorage] = append(v.buildings[rpc.BuildingId_GoldStorage][:index], v.buildings[rpc.BuildingId_GoldStorage][index+1:]...)

	case rpc.BuildingId_TroopHousing:

		v.Troophosing = append(v.Troophosing[:index], v.Troophosing[index+1:]...)
		v.buildings[rpc.BuildingId_TroopHousing] = append(v.buildings[rpc.BuildingId_TroopHousing][:index], v.buildings[rpc.BuildingId_TroopHousing][index+1:]...)

	case rpc.BuildingId_ArcherTower:

		v.Archertower = append(v.Archertower[:index], v.Archertower[index+1:]...)
		v.buildings[rpc.BuildingId_ArcherTower] = append(v.buildings[rpc.BuildingId_ArcherTower][:index], v.buildings[rpc.BuildingId_ArcherTower][index+1:]...)

	case rpc.BuildingId_Cannon:

		v.Cannon = append(v.Cannon[:index], v.Cannon[index+1:]...)
		v.buildings[rpc.BuildingId_Cannon] = append(v.buildings[rpc.BuildingId_Cannon][:index], v.buildings[rpc.BuildingId_Cannon][index+1:]...)

	case rpc.BuildingId_WizardTower:

		v.Wizardtower = append(v.Wizardtower[:index], v.Wizardtower[index+1:]...)
		v.buildings[rpc.BuildingId_WizardTower] = append(v.buildings[rpc.BuildingId_WizardTower][:index], v.buildings[rpc.BuildingId_WizardTower][index+1:]...)

	case rpc.BuildingId_AirDefense:

		v.Airdefense = append(v.Airdefense[:index], v.Airdefense[index+1:]...)
		v.buildings[rpc.BuildingId_AirDefense] = append(v.buildings[rpc.BuildingId_AirDefense][:index], v.buildings[rpc.BuildingId_AirDefense][index+1:]...)

	case rpc.BuildingId_Mortar:

		v.Mortar = append(v.Mortar[:index], v.Mortar[index+1:]...)
		v.buildings[rpc.BuildingId_Mortar] = append(v.buildings[rpc.BuildingId_Mortar][:index], v.buildings[rpc.BuildingId_Mortar][index+1:]...)

	case rpc.BuildingId_TeslaTower:

		v.Teslatower = append(v.Teslatower[:index], v.Teslatower[index+1:]...)
		v.buildings[rpc.BuildingId_TeslaTower] = append(v.buildings[rpc.BuildingId_TeslaTower][:index], v.buildings[rpc.BuildingId_TeslaTower][index+1:]...)

	case rpc.BuildingId_XBow:

		v.Xbow = append(v.Xbow[:index], v.Xbow[index+1:]...)
		v.buildings[rpc.BuildingId_XBow] = append(v.buildings[rpc.BuildingId_XBow][:index], v.buildings[rpc.BuildingId_XBow][index+1:]...)

	//case rpc.BuildingId_AllianceCastle:

	//	v.Alliancecastle = append(v.Alliancecastle[:index], v.Alliancecastle[index+1:]...)
	//	v.buildings[rpc.BuildingId_AllianceCastle] = append(v.buildings[rpc.BuildingId_AllianceCastle][:index], v.buildings[rpc.BuildingId_AllianceCastle][index+1:]...)

	case rpc.BuildingId_SpellForge:

		v.Spellforge = append(v.Spellforge[:index], v.Spellforge[index+1:]...)
		v.buildings[rpc.BuildingId_SpellForge] = append(v.buildings[rpc.BuildingId_SpellForge][:index], v.buildings[rpc.BuildingId_SpellForge][index+1:]...)

	case rpc.BuildingId_Deco1:
		v.Deco1 = append(v.Deco1[:index], v.Deco1[index+1:]...)
		v.buildings[rpc.BuildingId_Deco1] = append(v.buildings[rpc.BuildingId_Deco1][:index], v.buildings[rpc.BuildingId_Deco1][index+1:]...)

	case rpc.BuildingId_Deco2:
		v.Deco2 = append(v.Deco2[:index], v.Deco2[index+1:]...)
		v.buildings[rpc.BuildingId_Deco2] = append(v.buildings[rpc.BuildingId_Deco2][:index], v.buildings[rpc.BuildingId_Deco2][index+1:]...)

	case rpc.BuildingId_Deco3:
		v.Deco3 = append(v.Deco3[:index], v.Deco3[index+1:]...)
		v.buildings[rpc.BuildingId_Deco3] = append(v.buildings[rpc.BuildingId_Deco3][:index], v.buildings[rpc.BuildingId_Deco3][index+1:]...)

	case rpc.BuildingId_Deco4:
		v.Deco4 = append(v.Deco4[:index], v.Deco4[index+1:]...)
		v.buildings[rpc.BuildingId_Deco4] = append(v.buildings[rpc.BuildingId_Deco4][:index], v.buildings[rpc.BuildingId_Deco4][index+1:]...)

	case rpc.BuildingId_Deco5:
		v.Deco5 = append(v.Deco5[:index], v.Deco5[index+1:]...)
		v.buildings[rpc.BuildingId_Deco5] = append(v.buildings[rpc.BuildingId_Deco5][:index], v.buildings[rpc.BuildingId_Deco5][index+1:]...)

	case rpc.BuildingId_Deco6:
		v.Deco6 = append(v.Deco6[:index], v.Deco6[index+1:]...)
		v.buildings[rpc.BuildingId_Deco6] = append(v.buildings[rpc.BuildingId_Deco6][:index], v.buildings[rpc.BuildingId_Deco6][index+1:]...)

	case rpc.BuildingId_Deco8:
		v.Deco8 = append(v.Deco8[:index], v.Deco8[index+1:]...)
		v.buildings[rpc.BuildingId_Deco8] = append(v.buildings[rpc.BuildingId_Deco8][:index], v.buildings[rpc.BuildingId_Deco8][index+1:]...)

	case rpc.BuildingId_Deco9:
		v.Deco9 = append(v.Deco9[:index], v.Deco9[index+1:]...)
		v.buildings[rpc.BuildingId_Deco9] = append(v.buildings[rpc.BuildingId_Deco9][:index], v.buildings[rpc.BuildingId_Deco9][index+1:]...)

	case rpc.BuildingId_Deco10:
		v.Deco10 = append(v.Deco10[:index], v.Deco10[index+1:]...)
		v.buildings[rpc.BuildingId_Deco10] = append(v.buildings[rpc.BuildingId_Deco10][:index], v.buildings[rpc.BuildingId_Deco10][index+1:]...)

	case rpc.BuildingId_Deco11:
		v.Deco11 = append(v.Deco11[:index], v.Deco11[index+1:]...)
		v.buildings[rpc.BuildingId_Deco11] = append(v.buildings[rpc.BuildingId_Deco11][:index], v.buildings[rpc.BuildingId_Deco11][index+1:]...)

	case rpc.BuildingId_Deco12:
		v.Deco12 = append(v.Deco12[:index], v.Deco12[index+1:]...)
		v.buildings[rpc.BuildingId_Deco12] = append(v.buildings[rpc.BuildingId_Deco12][:index], v.buildings[rpc.BuildingId_Deco12][index+1:]...)

	case rpc.BuildingId_Deco13:
		v.Deco13 = append(v.Deco13[:index], v.Deco13[index+1:]...)
		v.buildings[rpc.BuildingId_Deco13] = append(v.buildings[rpc.BuildingId_Deco13][:index], v.buildings[rpc.BuildingId_Deco13][index+1:]...)

	case rpc.BuildingId_Deco14:
		v.Deco14 = append(v.Deco14[:index], v.Deco14[index+1:]...)
		v.buildings[rpc.BuildingId_Deco14] = append(v.buildings[rpc.BuildingId_Deco14][:index], v.buildings[rpc.BuildingId_Deco14][index+1:]...)

	case rpc.BuildingId_Deco15:
		v.Deco15 = append(v.Deco15[:index], v.Deco15[index+1:]...)
		v.buildings[rpc.BuildingId_Deco15] = append(v.buildings[rpc.BuildingId_Deco15][:index], v.buildings[rpc.BuildingId_Deco15][index+1:]...)

	case rpc.BuildingId_Deco16:
		v.Deco16 = append(v.Deco16[:index], v.Deco16[index+1:]...)
		v.buildings[rpc.BuildingId_Deco16] = append(v.buildings[rpc.BuildingId_Deco16][:index], v.buildings[rpc.BuildingId_Deco16][index+1:]...)

	case rpc.BuildingId_Deco17:
		v.Deco17 = append(v.Deco17[:index], v.Deco17[index+1:]...)
		v.buildings[rpc.BuildingId_Deco17] = append(v.buildings[rpc.BuildingId_Deco17][:index], v.buildings[rpc.BuildingId_Deco17][index+1:]...)

	case rpc.BuildingId_Deco18:
		v.Deco18 = append(v.Deco18[:index], v.Deco18[index+1:]...)
		v.buildings[rpc.BuildingId_Deco18] = append(v.buildings[rpc.BuildingId_Deco18][:index], v.buildings[rpc.BuildingId_Deco18][index+1:]...)

	case rpc.BuildingId_Deco19:
		v.Deco19 = append(v.Deco19[:index], v.Deco19[index+1:]...)
		v.buildings[rpc.BuildingId_Deco19] = append(v.buildings[rpc.BuildingId_Deco19][:index], v.buildings[rpc.BuildingId_Deco19][index+1:]...)

	case rpc.BuildingId_Deco20:
		v.Deco20 = append(v.Deco20[:index], v.Deco20[index+1:]...)
		v.buildings[rpc.BuildingId_Deco20] = append(v.buildings[rpc.BuildingId_Deco20][:index], v.buildings[rpc.BuildingId_Deco20][index+1:]...)

	case rpc.BuildingId_Deco21:
		v.Deco21 = append(v.Deco21[:index], v.Deco21[index+1:]...)
		v.buildings[rpc.BuildingId_Deco21] = append(v.buildings[rpc.BuildingId_Deco21][:index], v.buildings[rpc.BuildingId_Deco21][index+1:]...)

	case rpc.BuildingId_Deco22:
		v.Deco22 = append(v.Deco22[:index], v.Deco22[index+1:]...)
		v.buildings[rpc.BuildingId_Deco22] = append(v.buildings[rpc.BuildingId_Deco22][:index], v.buildings[rpc.BuildingId_Deco22][index+1:]...)

	case rpc.BuildingId_Deco23:
		v.Deco23 = append(v.Deco23[:index], v.Deco23[index+1:]...)
		v.buildings[rpc.BuildingId_Deco23] = append(v.buildings[rpc.BuildingId_Deco23][:index], v.buildings[rpc.BuildingId_Deco23][index+1:]...)

	case rpc.BuildingId_Deco24:
		v.Deco24 = append(v.Deco24[:index], v.Deco24[index+1:]...)
		v.buildings[rpc.BuildingId_Deco24] = append(v.buildings[rpc.BuildingId_Deco24][:index], v.buildings[rpc.BuildingId_Deco24][index+1:]...)

	case rpc.BuildingId_Bomb:
		v.Bomb = append(v.Bomb[:index], v.Bomb[index+1:]...)
		v.buildings[rpc.BuildingId_Bomb] = append(v.buildings[rpc.BuildingId_Bomb][:index], v.buildings[rpc.BuildingId_Bomb][index+1:]...)

	case rpc.BuildingId_GiantBomb:
		v.Giantbomb = append(v.Giantbomb[:index], v.Giantbomb[index+1:]...)
		v.buildings[rpc.BuildingId_GiantBomb] = append(v.buildings[rpc.BuildingId_GiantBomb][:index], v.buildings[rpc.BuildingId_GiantBomb][index+1:]...)

	case rpc.BuildingId_Eject:
		v.Eject = append(v.Eject[:index], v.Eject[index+1:]...)
		v.buildings[idType] = append(v.buildings[idType][:index], v.buildings[idType][index+1:]...)

	case rpc.BuildingId_GeneralHouse:
		v.Generalhouse = append(v.Generalhouse[:index], v.Generalhouse[index+1:]...)
		v.buildings[rpc.BuildingId_GeneralHouse] = append(v.buildings[rpc.BuildingId_GeneralHouse][:index], v.buildings[rpc.BuildingId_GeneralHouse][index+1:]...)

	case rpc.BuildingId_Barrier1:
		v.Barrier1 = append(v.Barrier1[:index], v.Barrier1[index+1:]...)
		v.buildings[rpc.BuildingId_Barrier1] = append(v.buildings[rpc.BuildingId_Barrier1][:index], v.buildings[rpc.BuildingId_Barrier1][index+1:]...)

	case rpc.BuildingId_Barrier2:
		v.Barrier2 = append(v.Barrier2[:index], v.Barrier2[index+1:]...)
		v.buildings[rpc.BuildingId_Barrier2] = append(v.buildings[rpc.BuildingId_Barrier2][:index], v.buildings[rpc.BuildingId_Barrier2][index+1:]...)

	case rpc.BuildingId_Barrier3:
		v.Barrier3 = append(v.Barrier3[:index], v.Barrier3[index+1:]...)
		v.buildings[rpc.BuildingId_Barrier3] = append(v.buildings[rpc.BuildingId_Barrier3][:index], v.buildings[rpc.BuildingId_Barrier3][index+1:]...)

	case rpc.BuildingId_Barrier4:
		v.Barrier4 = append(v.Barrier4[:index], v.Barrier4[index+1:]...)
		v.buildings[rpc.BuildingId_Barrier4] = append(v.buildings[rpc.BuildingId_Barrier4][:index], v.buildings[rpc.BuildingId_Barrier4][index+1:]...)

	case rpc.BuildingId_Barrier5:
		v.Barrier5 = append(v.Barrier5[:index], v.Barrier5[index+1:]...)
		v.buildings[rpc.BuildingId_Barrier5] = append(v.buildings[rpc.BuildingId_Barrier5][:index], v.buildings[rpc.BuildingId_Barrier5][index+1:]...)

	case rpc.BuildingId_Barrier6:
		v.Barrier6 = append(v.Barrier6[:index], v.Barrier6[index+1:]...)
		v.buildings[rpc.BuildingId_Barrier6] = append(v.buildings[rpc.BuildingId_Barrier6][:index], v.buildings[rpc.BuildingId_Barrier6][index+1:]...)

	}

	for e := v.w.Front(); e != nil; e = e.Next() {
		w, _ := e.Value.(*woking)
		if w.idType == idType && w.index > index {
			w.index--
		}
	}
	return
}

func (v *village) buildings_Create(idType rpc.BuildingId_IdType, x, y uint32, bFinishRightNow bool) interface{} {
	cfg := GetBuildingCfgByTypeId(idType, 1)
	if cfg == nil {
		SyncError(v.p.conn, "buildings_Create:Not exist config for type %d level 1", idType)

		return nil
	}

	if !v.mapCheckSpace(x, y, cfg.BuildSize, uint32(idType), false) {
		return nil
	}
	v.mapInsertTo(x, y, cfg.BuildSize, uint32(idType))

	b := v.buildings_New(idType)
	if b == nil {
		SyncError(v.p.conn, "buildings_Create:Create Building Error: type %d, x: %d y: %d", idType, x, y)

		return nil
	}

	if m, ok := b.(Movable); ok {
		m.SetP(NewPosition(x, y))
	}

	if a, ok := b.(Assailable); ok {
		if cfg != nil && bFinishRightNow {
			a.SetHp(cfg.Hitpoints)
		} else {
			a.SetHp(0)
		}
	}

	if u, ok := b.(Upgradable); ok {
		if bFinishRightNow {
			u.SetLevel(1)
		} else {
			u.SetLevel(0)
		}
		u.SetUpgradeTime(0)
	}

	if r, ok := b.(Removable); ok {
		r.SetRemoveTime(0)
		r.SetGemNum(RandomNumber(cfg.CleanAwardGemMin, cfg.CleanAwardGemMax))
	}

	if o, ok := b.(Operable); ok {
		o.SetLastOpTime(uint32(time.Now().Unix()))
	}

	if c, ok := b.(CanStorageFood); ok {
		c.SetStorageFood(0)
	}

	if c, ok := b.(CanStorageGold); ok {
		c.SetStorageGold(0)
	}

	return b
}

func (v *village) buildings_Get(idType rpc.BuildingId_IdType, index uint32) interface{} {
	if s, exist := v.buildings[idType]; exist {
		if index >= uint32(len(s)) {
			return nil
		}

		return s[index]
	}
	return nil
}

func (v *village) buildings_GetCntOf(idType rpc.BuildingId_IdType) uint32 {
	if s, exist := v.buildings[idType]; exist {
		return uint32(len(s))
	}
	return 0
}

func (v *village) buildings_GetAllOf(idType rpc.BuildingId_IdType) []proto.Message {
	return v.buildings[idType]
}

func (v *village) buildings_GetBuildingLevel(o Upgradable) (level uint32) {
	return o.GetLevel()
}

func (v *village) buildings_Cancel(idType rpc.BuildingId_IdType, index uint32) {
	for e := v.w.Front(); e != nil; e = e.Next() {
		w, _ := e.Value.(*woking)
		if w.idType == idType && index == w.index {
			v.w.Remove(e)
			return
		}
	}
}

func (v *village) buildings_NewWoking(idType rpc.BuildingId_IdType, index uint32, i interface{}) {
	if u, ok := i.(Upgradable); ok {
		cfg := GetBuildingCfgByTypeId(idType, u.GetLevel()+1)
		if cfg == nil {
			SyncError(v.p.conn, "buildings_NewWoking(Upgrade):Can't get cfg for building type:%d, level:%d", idType, u.GetLevel()+1)

			return
		}

		v.w.PushBack(&woking{
			idType: idType,
			index:  index,
			i:      i,
			finish: u.GetUpgradeTime() + cfg.GetBuildingTime(),
		})

		logger.Info("buildings_NewWoking(Upgrade):(%d, %d), Level:%d, Time:%d", idType, index, u.GetLevel(), u.GetUpgradeTime()+cfg.GetBuildingTime())
	} else if r, ok := i.(Removable); ok {
		cfg := GetBuildingCfgByTypeId(idType, 1)
		if cfg == nil {
			SyncError(v.p.conn, "buildings_NewWoking(Remove):Can't get cfg for building type:%d", idType)

			return
		}

		v.w.PushBack(&woking{
			idType: idType,
			index:  index,
			i:      i,
			finish: r.GetRemoveTime() + cfg.GetBuildingTime(),
		})

		logger.Info("buildings_NewWoking(Remove):(%d, %d), Time:%d", idType, index, r.GetRemoveTime()+cfg.GetBuildingTime())
	}
}

func (v *village) buildings_WokingLen() uint32 {
	return uint32(v.w.Len())
}

func (v *village) buildings_GetWorkerCnt() uint32 {
	return v.buildings_GetCntOf(rpc.BuildingId_Worker)
}

func (v *village) buildings_FinishNow(idType rpc.BuildingId_IdType, index uint32) {

	finishCount := 0

	defer func() {
		if finishCount > 0 {
			v.p.Save()
			v.Save()
		}
	}()

	logger.Info("Buildings FinishNow(%d, %d):Start!\n", idType, index)
	for e := v.w.Front(); e != nil; e = e.Next() {
		w, _ := e.Value.(*woking)
		if w.idType == idType && index == w.index {
			time_now := uint32(time.Now().Unix())
			logger.Info("Time(%d, %d)\n", w.finish, time_now)
			if w.finish > time_now {
				cost := GetYuanBaoCountFromTime(w.finish - time_now)
				logger.Info("Cost = ", cost, "Current Diamonds =", v.p.GetPlayerTotalGem(), "\n")
				if v.p.GetPlayerTotalGem() >= cost {
					v.p.CostResource(cost, pt.ResType_Gem, pt.Lose_FinishNow)
					finishCount++

					level := uint32(1)

					if o, ok := w.i.(Upgradable); ok {
						//当前等级
						var CurLeve uint32
						CurLeve = o.GetLevel()

						o.SetUpgradeTime(0)
						o.SetLevel(o.GetLevel() + 1)

						level = o.GetLevel()

						//通过配置表判断是否是xbow
						cfg := GetBuildingCfgByTypeId(w.idType, 1)
						Ammo := cfg.AmmoCount
						if Ammo != 0 {
							if level > CurLeve {
								buildingObj := v.buildings_Get(w.idType, w.index)
								xbow := buildingObj.(*rpc.XBow)
								xbow.SetAmmoCount(Ammo)
								if CurLeve == 0 {
									xbow.SetAltAttackRange(1)
								}
							}
						}

						//建筑物等级改变
						v.on_Building_Level_Change(w.idType, level, false)

						logger.Info("Buildings FinishNow Upgraded(%d, %d):Ok!\n", idType, index)
					}

					if o, ok := w.i.(Operable); ok {
						o.SetLastOpTime(time_now)
					}

					cfg := GetBuildingCfgByTypeId(idType, level)
					if o, ok := w.i.(Assailable); ok {
						if cfg != nil {
							o.SetHp(cfg.Hitpoints)

							v.p.AddExp(level)
						}
					}

					if o, ok := w.i.(Removable); ok {
						o.SetRemoveTime(0)

						if cfg != nil {
							v.p.AddExp(cfg.CleanAwardExp)
							v.p.GainResource(o.GetGemNum(), pt.ResType_Gem, pt.Gain_RemoveObstacle)
						}

						if m, ok := w.i.(Movable); ok {
							v.mapRemoveFrom(m.GetP().GetX(), m.GetP().GetY(), cfg.BuildSize)
						}

						v.buildings_Remove(idType, index)

						logger.Info("Buildings FinishNow Removed(%d, %d):Ok!\n", idType, index)
					}
					rm := e

					e = e.Next()

					v.w.Remove(rm)
				} else {
					SyncError(v.p.conn, "Buildings FinishNow(%d, %d):Failed! Not Enough Diamonds.\n", idType, index)
				}
			}
			return
		}
	}
}

//当建筑物等级改变
func (v *village) on_Building_Level_Change(idType rpc.BuildingId_IdType, level uint32, init bool) {
	if idType == rpc.BuildingId_Center {
		//向center发送消息
		req := &pt.UpdaePlayerLevel2Id{Id: v.p.GetUid(), Level: level}
		rst := &pt.UpdaePlayerLevel2IdResult{}
		cns.center.Go("Center.UpdatePlayerLevel2Id", req, rst, nil)
		logger.Info("Center on_Building_Level_Change(%d, %d):Ok!\n", idType, level)
	}

}

func (v *village) buildings_ProcessUpgrade() {
	time_now := uint32(time.Now().Unix())
	for e := v.w.Front(); e != nil; {
		w, _ := e.Value.(*woking)

		if time_now >= w.finish {
			level := uint32(1)

			if o, ok := w.i.(Upgradable); ok {

				//当前等级
				var CurLeve uint32
				CurLeve = o.GetLevel()

				o.SetUpgradeTime(0)
				o.SetLevel(o.GetLevel() + 1)

				level = o.GetLevel()

				//通过配置表判断是否是xbow
				cfg := GetBuildingCfgByTypeId(w.idType, 1)
				Ammo := cfg.AmmoCount
				if Ammo != 0 {
					if level > CurLeve {
						buildingObj := v.buildings_Get(w.idType, w.index)
						xbow := buildingObj.(*rpc.XBow)
						xbow.SetAmmoCount(Ammo)
						if CurLeve == 0 {
							xbow.SetAltAttackRange(1)
						}
					}
				}

				//建筑物等级改变
				v.on_Building_Level_Change(w.idType, level, false)

				logger.Info("Buildings Finished(%d, %d):Ok!\n", w.idType, w.index)
			}

			if o, ok := w.i.(Operable); ok {
				o.SetLastOpTime(w.finish)
			}

			cfg := GetBuildingCfgByTypeId(w.idType, level)
			if o, ok := w.i.(Assailable); ok {
				if cfg != nil {
					o.SetHp(cfg.Hitpoints)

					v.p.AddExp(level)
				}
			}

			if o, ok := w.i.(Removable); ok {
				o.SetRemoveTime(0)

				if cfg != nil {
					v.p.AddExp(cfg.CleanAwardExp)
					v.p.GainResource(o.GetGemNum(), pt.ResType_Gem, pt.Gain_RemoveObstacle)
				}

				if m, ok := w.i.(Movable); ok {
					v.mapRemoveFrom(m.GetP().GetX(), m.GetP().GetY(), cfg.BuildSize)
				}

				v.buildings_Remove(w.idType, w.index)

				logger.Info("Buildings Removed(%d, %d):Ok!\n", w.idType, w.index)
			}
			rm := e

			e = e.Next()

			v.w.Remove(rm)
		} else {
			e = e.Next()
		}
	}
}
