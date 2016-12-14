package connector

import (
	"golang-project/dpsg/logger"
	"golang-project/dpsg/proto"
	"golang-project/dpsg/rpc"
	"strings"
	"time"
)

func (v *village) barrack_GetQueueCnt(barrack *rpc.Barrack) (cnt uint32) {
	if barrack.Queue == nil {
		return
	}

	for _, queue := range barrack.Queue {
		c := queue.GetCharacter().GetCount()

		cfg := GetCharacterCfgByTypeId(queue.GetCharacter().GetType(), 1)

		cnt += c * cfg.HousingSpace
	}

	return
}

func (v *village) barrack_GetOrCreateQueue(barrack *rpc.Barrack, chType rpc.CharacterType) *rpc.CharacterQueue {
	if barrack.Queue == nil {
		barrack.Queue = make([]*rpc.CharacterQueue, 0)
	}

	for _, queue := range barrack.Queue {
		if chType == queue.GetCharacter().GetType() {
			return queue
		}
	}

	queue := &rpc.CharacterQueue{}

	chr := &rpc.Character{}
	chr.SetType(chType)
	chr.SetCount(0)
	queue.SetCharacter(chr)
	queue.SetStartTime(0)

	barrack.Queue = append(barrack.Queue, queue)

	return queue
}

func (v *village) barrack_GetQueue(barrack *rpc.Barrack, chType rpc.CharacterType) (*rpc.CharacterQueue, int) {
	if barrack.Queue == nil {
		barrack.Queue = make([]*rpc.CharacterQueue, 0)
	}

	for index, queue := range barrack.Queue {
		if chType == queue.GetCharacter().GetType() {
			return queue, index
		}
	}

	return nil, 0
}

func (v *village) barrack_CreateByType(index uint32, chType rpc.CharacterType, count uint32) bool {
	obj := v.buildings_Get(rpc.BuildingId_Barrack, index)
	if obj == nil {
		SyncError(v.p.conn, "barrack_CreateByType:obj == nil")

		return false
	}

	barrack, _ := obj.(*rpc.Barrack)

	if barrack.Queue == nil {
		barrack.Queue = make([]*rpc.CharacterQueue, 0)
	}

	if barrack.GetUpgradeTime() != 0 {
		SyncError(v.p.conn, "barrack_CreateByType:barrack.GetUpgradeTime() != 0")

		return false
	}

	level := barrack.GetLevel()

	l := v.laboratory_GetInfoLevel(chType)

	cfg := GetCharacterCfgByTypeId(chType, l)
	if cfg == nil {
		SyncError(v.p.conn, "barrack_CreateByType:cfg == nil")

		return false
	}

	if cfg.BarrackLevel > level {
		SyncError(v.p.conn, "barrack_CreateByType:cfg.BarrackLevel > level")

		return false
	}

	all_cnt := v.barrack_GetQueueCnt(barrack)

	bacfg := GetBuildingCfgByTypeId(rpc.BuildingId_Barrack, level)
	if bacfg == nil {
		SyncError(v.p.conn, "barrack_CreateByType:bacfg == nil")

		return false
	}

	if all_cnt+(cfg.HousingSpace*uint32(count)) > bacfg.UnitProduction {
		logger.Info("barrack_CreateByType:all_cnt(%d) > UnitProduction(%d)", all_cnt, bacfg.UnitProduction)

		//SyncError(v.p.conn, "barrack_CreateByType:all_cnt(%d) > UnitProduction(%d)", all_cnt, bacfg.UnitProduction)

		return false
	}

	switch strings.ToLower(cfg.TrainingResource) {
	case "gold":
		_, total := v.collect_GetStorageGoldLimit()
		if cfg.TrainingCost*count > total {
			SyncError(v.p.conn, "barrack_CreateByType:cfg.TrainingCost(%d)*count(%d) > total(%d) - gold", cfg.TrainingCost, count, total)

			return false
		}
	case "food":
		_, total := v.collect_GetStorageFoodLimit()
		if cfg.TrainingCost*count > total {
			SyncError(v.p.conn, "barrack_CreateByType:cfg.TrainingCost(%d)*count(%d) > total(%d) - food", cfg.TrainingCost, count, total)

			return false
		}
	}

	q := v.barrack_GetOrCreateQueue(barrack, chType)

	q.GetCharacter().SetCount(q.GetCharacter().GetCount() + count)

	v.p.CostResource(cfg.TrainingCost*count, strings.ToLower(cfg.TrainingResource), proto.Lose_CreateCharacter)

	return true
}

func (v *village) barrack_CancelCreate(index uint32, chType rpc.CharacterType, count uint32) {
	obj := v.buildings_Get(rpc.BuildingId_Barrack, index)
	if obj == nil {
		return
	}

	barrack, _ := obj.(*rpc.Barrack)

	if barrack.Queue == nil {
		return
	}

	if barrack.GetUpgradeTime() != 0 {
		return
	}

	q, qi := v.barrack_GetQueue(barrack, chType)

	if q == nil {
		return
	}

	c := q.GetCharacter().GetCount()

	if c > uint32(count) {
		q.GetCharacter().SetCount(c - uint32(count))
	} else {
		barrack.Queue = append(barrack.Queue[:qi], barrack.Queue[qi+1:]...)
		count = c
	}

	cfg := GetCharacterCfgByTypeId(chType, v.laboratory_GetInfoLevel(chType))

	v.p.GainResource(cfg.TrainingCost*count, strings.ToLower(cfg.TrainingResource), proto.Gain_CacelCreateCharacter)
}

// 获取所有兵，并且是合并了同类型的兵种
func (v *village) barrack_GetAllCharacters() []*rpc.Character {
	temp := make(map[rpc.CharacterType]*rpc.Character)
	chars := make([]*rpc.Character, 0)

	ts := v.buildings_GetAllOf(rpc.BuildingId_TroopHousing)

	//拷贝一分临时数据出来作数量整合
	for _, t := range ts {
		pt := t.(*rpc.TroopHousing)

		for _, c := range pt.Character {

			_, ok := temp[c.GetType()]

			if ok {
				temp[c.GetType()].SetCount(temp[c.GetType()].GetCount() + c.GetCount())
				//*temp[c.GetType()].Count += c.GetCount()
			} else {
				a := &rpc.Character{} //Type: c.Type, Count: c.Count, Level: c.Level
				a.SetType(c.GetType())
				a.SetCount(c.GetCount())
				a.SetLevel(0)

				temp[c.GetType()] = a
			}
		}
	}

	ghs := v.buildings_GetAllOf(rpc.BuildingId_GeneralHouse)

	for _, g := range ghs {
		gh := g.(*rpc.GeneralHouse)

		selected := gh.GetSelectedhero()

		for _, hero := range gh.Hero {
			if hero.GetUpgradetime() > 0 {
				continue
			}

			c := hero.GetCharacter()

			_, ok := temp[c.GetType()]

			if ok {
				temp[c.GetType()].SetCount(temp[c.GetType()].GetCount() + c.GetCount())
			} else if c.GetType() == selected {
				a := &rpc.Character{} //Type: c.Type, Count: c.Count, Level: c.Level
				a.SetType(c.GetType())
				a.SetCount(c.GetCount())
				a.SetLevel(c.GetLevel())

				temp[c.GetType()] = a
			}
		}
	}

	for _, c := range temp {
		//数量大于0的才发送给客户端
		if c.GetCount() > 0 {
			if c.GetLevel() == 0 {
				c.SetLevel(v.laboratory_GetInfoLevel(c.GetType()))
			}
			chars = append(chars, c)
		}
	}

	return chars
}

func (v *village) barrack_GetTroopHousing(x, y uint32, size uint32) (uint32, *rpc.TroopHousing) {
	var first_free *rpc.TroopHousing

	ts := v.buildings_GetAllOf(rpc.BuildingId_TroopHousing)

	for i, t := range ts {
		pt := t.(*rpc.TroopHousing)

		free := v.barrack_GetTroopHousingFreeSpace(pt)

		if free >= int32(size) {
			return uint32(i), pt
		}

		if free > 0 && first_free == nil {
			first_free = pt
		}
	}

	return 0, first_free
}

func (v *village) barrack_GetTroopHousingSpace(ts *rpc.TroopHousing) (s uint32) {
	cfg := GetBuildingCfgByTypeId(rpc.BuildingId_TroopHousing, ts.GetLevel())

	if cfg == nil {
		return 0
	}

	return cfg.HousingSpace
}

func (v *village) barrack_GetTroopHousingUsedSpace(ts *rpc.TroopHousing) (s uint32) {

	if ts.Character == nil {
		return
	}

	for _, b := range ts.Character {
		cfg := GetCharacterCfgByTypeId(b.GetType(), 1)
		s += (cfg.HousingSpace * b.GetCount())
	}

	return
}

func (v *village) barrack_GetTroopHousingFreeSpace(ts *rpc.TroopHousing) (s int32) {
	return int32(v.barrack_GetTroopHousingSpace(ts) - v.barrack_GetTroopHousingUsedSpace(ts))
}

func (v *village) barrack_GetTroopHousingTotalSpaces() (s uint32) {
	ts := v.buildings_GetAllOf(rpc.BuildingId_TroopHousing)

	for _, t := range ts {
		pt := t.(*rpc.TroopHousing)

		s += v.barrack_GetTroopHousingSpace(pt)
	}

	return
}

func (v *village) barrack_GetTroopHousingTotalUsedSpaces() (s uint32) {
	ts := v.buildings_GetAllOf(rpc.BuildingId_TroopHousing)

	for _, t := range ts {
		pt := t.(*rpc.TroopHousing)

		s += v.barrack_GetTroopHousingUsedSpace(pt)
	}

	return
}

func (v *village) barrack_GetTroopHousingTotalFreeSpaces() (s int32) {
	return int32(v.barrack_GetTroopHousingTotalSpaces() - v.barrack_GetTroopHousingTotalUsedSpaces())
}

func (v *village) barrack_TroopHousingPushCharacterTo(t *rpc.TroopHousing, chType rpc.CharacterType) bool {
	if t.Character == nil {
		t.Character = make([]*rpc.Character, 0)

		goto NotExist
	}

	for _, chr := range t.Character {
		if chr.GetType() == chType {
			chr.SetCount(chr.GetCount() + 1)
			return true
		}
	}

NotExist:

	chr := &rpc.Character{}
	chr.SetType(chType)
	chr.SetCount(1)

	t.Character = append(t.Character, chr)

	return true
}

func (v *village) barrack_TroopHousingPushCharacter(chType rpc.CharacterType) bool {
	cfg := GetCharacterCfgByTypeId(chType, 1)
	if cfg == nil {
		return false
	}

	index, t := v.barrack_GetTroopHousing(0, 0, cfg.HousingSpace)

	logger.Info("Push Character<%d> into <%d> TroopHousing!", chType, index)

	return v.barrack_TroopHousingPushCharacterTo(t, chType)
}

func (v *village) barrack_TroopHousingPopCharacterFrom(t *rpc.TroopHousing, chType rpc.CharacterType) *rpc.Character {
	for _, c := range t.Character {
		if c.GetType() == chType {
			if c.GetCount() > 0 {
				c.SetCount(c.GetCount() - 1)

				cc := rpc.Character{}
				cc.SetType(c.GetType())
				cc.SetCount(1)
				cc.SetLevel(c.GetLevel())
				return &cc
			}
		}
	}
	return nil
}
func (v *village) spell_SpellForgePopSpell(chType rpc.SpellType) *rpc.Spell {
	obj := v.buildings_Get(rpc.BuildingId_SpellForge, 0)
	if obj == nil {
		return nil
	}
	sp, _ := obj.(*rpc.SpellForge)
	for _, c := range sp.Spell {
		if c.GetType() == chType {
			if c.GetCount() > 0 {
				c.SetCount(c.GetCount() - 1)
				cc := rpc.Spell{}
				cc.SetType(c.GetType())
				cc.SetCount(1)
				cc.SetLevel(c.GetLevel())
				return &cc
			}
		}
	}
	return nil
}
func (v *village) barrack_TroopHousingPopCharacter(chType rpc.CharacterType) *rpc.Character {
	ts := v.buildings_GetAllOf(rpc.BuildingId_TroopHousing)
	for _, t := range ts {
		c := v.barrack_TroopHousingPopCharacterFrom(t.(*rpc.TroopHousing), chType)

		if c != nil {
			return c
		}
	}
	return nil
}

func (v *village) barrack_FinishNow(index uint32) {
	barracks := v.buildings_GetAllOf(rpc.BuildingId_Barrack)
	if barracks == nil {
		return
	}

	if int(index) >= len(barracks) {
		return
	}

	barrack := barracks[index].(*rpc.Barrack)

	if barrack.Queue == nil {
		return
	}

	var needSpace uint32 = 0
	var time_left uint32 = 0

	for _, cq := range barrack.Queue {
		c := cq.GetCharacter()

		l := v.laboratory_GetInfoLevel(c.GetType())

		cfg := GetCharacterCfgByTypeId(c.GetType(), l)

		needSpace += cfg.HousingSpace * c.GetCount()

		if cq.GetStartTime() == 0 {
			time_left += cfg.TrainingTime * c.GetCount()
		} else {
			time_left += cfg.TrainingTime*(c.GetCount()-1) + cfg.TrainingTime - (uint32(time.Now().Unix()) - cq.GetStartTime())
		}
	}

	if int32(needSpace) > v.barrack_GetTroopHousingTotalFreeSpaces() {
		SyncError(v.p.conn, "barrack_FinishNow:needSpace(%d),free(%d)", int32(needSpace), v.barrack_GetTroopHousingTotalFreeSpaces())

		return
	}

	if time_left > 0 {
		cost := GetYuanBaoCountFromTime(time_left)
		if v.p.GetPlayerTotalGem() >= cost {
			//先确定扣除成功再做后面的操作
			if v.p.CostResource(cost, proto.ResType_Gem, proto.Lose_BuyTime) {
				for _, cq := range barrack.Queue {
					c := cq.GetCharacter()

					for i := uint32(0); i < c.GetCount(); i++ {
						v.barrack_TroopHousingPushCharacter(c.GetType())
					}
				}

				barrack.Queue = make([]*rpc.CharacterQueue, 0)
			}
		}
	}

}

func (v *village) barrack_Process() {
	barracks := v.buildings_GetAllOf(rpc.BuildingId_Barrack)
	if barracks == nil {
		return
	}

	free := v.barrack_GetTroopHousingTotalFreeSpaces()
	time_now := uint32(time.Now().Unix())

	for _, obj := range barracks {
		barrack := obj.(*rpc.Barrack)
		if barrack.Queue == nil {
			continue
		}

		last_ok := uint32(0)
		newQueue := make([]*rpc.CharacterQueue, 0)
		for index, queue := range barrack.Queue {
			newQueue = append(newQueue, queue)

			if queue.GetStartTime() == 0 {
				if last_ok > 0 {
					queue.SetStartTime(last_ok)
					last_ok = 0
				} else {
					queue.SetStartTime(time_now)
					continue
				}
			}

			l := v.laboratory_GetInfoLevel(queue.GetCharacter().GetType())
			cfg := GetCharacterCfgByTypeId(queue.GetCharacter().GetType(), l)

			total := queue.GetCharacter().GetCount()
			bQuitQueue := false

			for c := uint32(1); c <= total; c++ {
				if queue.GetStartTime()+cfg.TrainingTime <= time_now && free >= int32(cfg.HousingSpace) {
					queue.GetCharacter().SetCount(total - c)
					free -= int32(cfg.HousingSpace)
					v.barrack_TroopHousingPushCharacter(queue.GetCharacter().GetType())
					queue.SetStartTime(queue.GetStartTime() + cfg.TrainingTime)
				} else {
					bQuitQueue = true
					break
				}
			}

			total = queue.GetCharacter().GetCount()
			if total <= 0 {
				newQueue = newQueue[:len(newQueue)-1]
				last_ok = queue.GetStartTime()
				continue
			}

			if bQuitQueue {
				if index < len(barrack.Queue)-1 {
					newQueue = append(newQueue, barrack.Queue[index+1:]...)
				}
				break
			}
		}

		barrack.Queue = newQueue

	}
}
