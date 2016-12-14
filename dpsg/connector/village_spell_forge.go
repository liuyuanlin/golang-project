package connector

import (
	"golang-project/dpsg/proto"
	"golang-project/dpsg/rpc"
	"strings"
	"time"
)

func (v *village) spellForge_GetAllCnt(sf *rpc.SpellForge) (cnt uint32) {
	if sf.Queue == nil {

	} else {
		for _, queue := range sf.Queue {
			cnt += queue.GetSpell().GetCount()
		}
	}
	//if sf.Spell == nil {
	//	return
	//}

	//for _, spell := range sf.Spell {
	//	cnt += spell.GetCount()
	//}
	return
}

func (v *village) spellForge_GetAllSpells() []*rpc.Spell {
	temp := make(map[rpc.SpellType]*rpc.Spell)
	spells := make([]*rpc.Spell, 0)

	ts := v.buildings_GetAllOf(rpc.BuildingId_SpellForge)

	for _, t := range ts {
		sf := t.(*rpc.SpellForge)

		for _, sp := range sf.Spell {
			_, ok := temp[sp.GetType()]

			if ok {
				temp[sp.GetType()].SetCount(temp[sp.GetType()].GetCount() + sp.GetCount())
			} else {
				a := &rpc.Spell{}
				a.SetType(sp.GetType())
				a.SetCount(sp.GetCount())
				a.SetLevel(sp.GetLevel())
				temp[sp.GetType()] = a
			}
		}
	}
	for _, sp := range temp {
		if sp.GetCount() > 0 {
			if sp.GetLevel() == 0 {
				sp.SetLevel(v.laboratory_GetSpellInfoLevel(sp.GetType()))
			}
			spells = append(spells, sp)
		}
	}
	return spells
}

func (v *village) spellForge_GetOrCreateQueue(sf *rpc.SpellForge, spType rpc.SpellType) *rpc.SpellQueue {
	if sf.Queue == nil {
		sf.Queue = make([]*rpc.SpellQueue, 0)
	}

	for _, queue := range sf.Queue {
		if spType == queue.GetSpell().GetType() {
			return queue
		}
	}

	queue := &rpc.SpellQueue{}

	sp := &rpc.Spell{}
	sp.SetType(spType)
	sp.SetCount(0)
	queue.SetSpell(sp)
	queue.SetStartTime(0)

	sf.Queue = append(sf.Queue, queue)

	return queue
}

func (v *village) spellForge_GetQueue(sf *rpc.SpellForge, spType rpc.SpellType) (*rpc.SpellQueue, int) {
	if sf.Queue == nil {
		sf.Queue = make([]*rpc.SpellQueue, 0)
	}

	for index, queue := range sf.Queue {
		if spType == queue.GetSpell().GetType() {
			return queue, index
		}
	}

	return nil, 0
}

func (v *village) spellForge_CreateByType(spType rpc.SpellType, count uint32) bool {
	obj := v.buildings_Get(rpc.BuildingId_SpellForge, 0)
	if obj == nil {
		SyncError(v.p.conn, "spellForge_CreateByType:obj == nil ")

		return false
	}

	sp, _ := obj.(*rpc.SpellForge)

	if sp.Queue == nil {
		sp.Queue = make([]*rpc.SpellQueue, 0)
	}

	if sp.GetUpgradeTime() != 0 {
		SyncError(v.p.conn, "spellForge_CreateByType:sp.GetUpgradeTime() != 0 ")

		return false
	}

	level := sp.GetLevel()

	l := v.laboratory_GetSpellInfoLevel(spType)

	cfg := GetSpellCfgByTypeId(spType, l)
	if cfg == nil {
		SyncError(v.p.conn, "spellForge_CreateByType:cfg == nil")

		return false
	}

	if cfg.SpellForgeLevel > level {
		SyncError(v.p.conn, "spellForge_CreateByType:cfg.SpellForgeLevel > level ")

		return false
	}

	all_cnt := v.spellForge_GetAllCnt(sp)

	spcfg := GetBuildingCfgByTypeId(rpc.BuildingId_SpellForge, level)
	if spcfg == nil {
		SyncError(v.p.conn, "spellForge_CreateByType:spcfg == nil")

		return false
	}

	if all_cnt+count > spcfg.UnitProduction {
		SyncError(v.p.conn, "spellForge_CreateByType:all_cnt(%d)+count(%d) > spcfg.UnitProduction(%d)", all_cnt, count, spcfg.UnitProduction)

		return false
	}

	switch strings.ToLower(cfg.TrainingResource) {
	case "gold":
		_, total := v.collect_GetStorageGoldLimit()
		if cfg.TrainingCost*count > total {
			SyncError(v.p.conn, "spellForge_CreateByType:cfg.TrainingCost(%d)*count(%d) > total(%d) - gold", cfg.TrainingCost, count, total)

			return false
		}
	case "food":
		_, total := v.collect_GetStorageFoodLimit()
		if cfg.TrainingCost*count > total {
			SyncError(v.p.conn, "spellForge_CreateByType:cfg.TrainingCost(%d)*count(%d) > total(%d) - gold", cfg.TrainingCost, count, total)

			return false
		}
	}

	q := v.spellForge_GetOrCreateQueue(sp, spType)

	q.GetSpell().SetCount(q.GetSpell().GetCount() + count)

	switch strings.ToLower(cfg.TrainingResource) {
	case "gold":
		v.collect_CostGold(cfg.TrainingCost * count)
	case "food":
		v.collect_CostFood(cfg.TrainingCost * count)
	}

	return true
}

func (v *village) spellForge_CancelCreate(spType rpc.SpellType, count uint32) {
	obj := v.buildings_Get(rpc.BuildingId_SpellForge, 0)
	if obj == nil {
		return
	}

	sp, _ := obj.(*rpc.SpellForge)

	if sp.Queue == nil {
		return
	}

	if sp.GetUpgradeTime() != 0 {
		return
	}

	q, qi := v.spellForge_GetQueue(sp, spType)

	if q == nil {
		return
	}
	c := q.GetSpell().GetCount()

	if c > uint32(count) {
		q.GetSpell().SetCount(c - uint32(count))
	} else {
		sp.Queue = append(sp.Queue[:qi], sp.Queue[qi+1:]...)
		count = c
	}

	cfg := GetSpellCfgByTypeId(spType, v.laboratory_GetSpellInfoLevel(spType))

	switch strings.ToLower(cfg.TrainingResource) {
	case "gold":
		//v.collect_CostGold(cfg.TrainingCost * count / 2)
		v.p.GainResource(cfg.TrainingCost*count, strings.ToLower(cfg.TrainingResource), proto.Gain_CacelCreatespell)
	case "food":
		//v.collect_CostFood(cfg.TrainingCost * count / 2)
		v.p.GainResource(cfg.TrainingCost*count, strings.ToLower(cfg.TrainingResource), proto.Gain_CacelCreatespell)
	}
}

func (v *village) spellForge_PushSpell(spType rpc.SpellType) {

	obj := v.buildings_Get(rpc.BuildingId_SpellForge, 0)
	if obj == nil {
		return
	}

	sp, _ := obj.(*rpc.SpellForge)

	if sp.Spell == nil {
		sp.Spell = make([]*rpc.Spell, 0)

		goto NotExist
	}

	for _, spell := range sp.Spell {
		if spell.GetType() == spType {
			spell.SetCount(spell.GetCount() + 1)
			return
		}
	}

NotExist:

	spell := &rpc.Spell{}
	spell.SetType(spType)
	spell.SetCount(1)

	sp.Spell = append(sp.Spell, spell)
}

func (v *village) spellForge_FinishNow() {
	obj := v.buildings_Get(rpc.BuildingId_SpellForge, 0)
	if obj == nil {
		return
	}

	sp, _ := obj.(*rpc.SpellForge)

	if sp.Queue == nil {
		return
	}

	var time_left uint32 = 0

	for _, queue := range sp.Queue {

		l := v.laboratory_GetSpellInfoLevel(queue.GetSpell().GetType())

		cfg := GetSpellCfgByTypeId(queue.GetSpell().GetType(), l)

		if queue.GetStartTime() == 0 {
			time_left += cfg.TrainingTime * queue.GetSpell().GetCount()
		} else {
			time_left += cfg.TrainingTime*(queue.GetSpell().GetCount()-1) + cfg.TrainingTime - (uint32(time.Now().Unix()) - queue.GetStartTime())
		}
	}

	if time_left > 0 {
		cost := GetYuanBaoCountFromTime(time_left)
		if v.p.GetPlayerTotalGem() >= cost {
			//先确定扣除成功再做后面的操作
			if v.p.CostResource(cost, proto.ResType_Gem, proto.Lose_BuyTime) {
				for _, queue := range sp.Queue {
					for i := uint32(0); i < queue.GetSpell().GetCount(); i++ {
						v.spellForge_PushSpell(queue.GetSpell().GetType())
					}
				}

				sp.Queue = make([]*rpc.SpellQueue, 0)
			}
		}
	}
}

func (v *village) spellForge_GetTroopHousingTotalSpaces(level uint32) (s uint32) {
	spcfg := GetBuildingCfgByTypeId(rpc.BuildingId_SpellForge, level)
	if spcfg == nil {
		return 0
	}
	s = spcfg.HousingSpace
	return s
}
func (v *village) spellForge_Process() {
	obj := v.buildings_Get(rpc.BuildingId_SpellForge, 0)
	if obj == nil {
		return
	}

	sp, _ := obj.(*rpc.SpellForge)

	if sp.Queue == nil {
		return
	}

ONE_BARRACK:
	for {
		for _, queue := range sp.Queue {
			if queue.GetStartTime() == 0 {

				queue.SetStartTime(uint32(time.Now().Unix()))

			} else {

				l := v.laboratory_GetSpellInfoLevel(queue.GetSpell().GetType())

				cfg := GetSpellCfgByTypeId(queue.GetSpell().GetType(), l)

				time_now := uint32(time.Now().Unix())

				if queue.GetStartTime()+cfg.TrainingTime <= time_now {

					c := queue.GetSpell().GetCount()

					if c > 1 {
						queue.GetSpell().SetCount(c - 1)
						queue.SetStartTime(queue.GetStartTime() + cfg.TrainingTime)

					} else {

						sp.Queue = sp.Queue[1:]

					}

					v.spellForge_PushSpell(queue.GetSpell().GetType())

				} else {
					break ONE_BARRACK
				}

			}

			break
		}

		if len(sp.Queue) == 0 {
			break
		}
	}

}
