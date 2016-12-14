package connector

type BuildingCfg struct {
	Name              string
	TID               string
	InfoTID           string
	BuildingClass     string
	ResName           string
	BuildSize         uint32
	BuildTimeD        uint32 `csv:",d"`
	BuildTimeH        uint32 `csv:",d"`
	BuildTimeM        uint32 `csv:",d"`
	BuildTimeS        uint32 `csv:",d"`
	BuildResource     string
	BuildCost         uint32
	Icon              string
	Hitpoints         uint32
	ProducesResource  string
	ResourcePerHour   uint32
	ResourceMax       uint32
	ResourceIconLimit uint32
	UnitProduction    uint32
	BoostCost         uint32
	MaxStoredGold     uint32
	MaxStoredFood     uint32
	TownHallLevel     uint32
	HousingSpace      uint32
	NeedWorker        bool
	CanMove           bool
	CanSell           bool
	CanRemove         bool
	SellPrice         uint32
	CleanAwardExp     uint32
	CleanAwardGemMin  uint32
	CleanAwardGemMax  uint32
	GenerateWeight    uint32
	AmmoCount         uint32
	AmmoCost          uint32
	AmmoResource      string
}

type TownhallLevels struct {
	Barrack        uint32
	AttackCost     uint32
	Farm           uint32
	Laboratory     uint32
	Walls          uint32
	Worker         uint32
	FoodStorage    uint32
	GoldMine       uint32
	GoldStorage    uint32
	TroopHousing   uint32
	ArcherTower    uint32
	Cannon         uint32
	WizardTower    uint32
	AirDefense     uint32
	Mortar         uint32
	TeslaTower     uint32
	XBow           uint32
	AllianceCastle uint32
	SpellForge     uint32
	Deco1          uint32
	Deco2          uint32
	Deco3          uint32
	Deco4          uint32
	Deco5          uint32
	Deco6          uint32
	Deco7          uint32
	Deco8          uint32
	Deco9          uint32
	Deco10         uint32
	Deco11         uint32
	Deco12         uint32
	Deco13         uint32
	Deco14         uint32
	Deco15         uint32
	Deco16         uint32
	Deco17         uint32
	Deco18         uint32
	Deco19         uint32
	Deco20         uint32
	Deco21         uint32
	Deco22         uint32
	Deco23         uint32
	Deco24         uint32
	Bomb           uint32
	GiantBomb      uint32
	Eject          uint32
	GeneralHouse   uint32
}

type CharacterCfg struct {
	HousingSpace            uint32
	BarrackLevel            uint32
	LaboratoryLevel         uint32
	Speed                   uint32
	Hitpoints               uint32
	TrainingTime            uint32
	TrainingResource        string
	TrainingCost            uint32
	UpgradeTimeH            uint32
	UpgradeResource         string
	UpgradeCost             uint32
	AttackRange             uint32
	AttackSpeed             uint32
	Damage                  int32
	PreferedTargetDamageMod uint32
	DamageRadius            uint32
	PreferedTargetBuilding  string
	IsFlying                bool
	AirTargets              bool
	GroundTargets           bool
	AttackCount             uint32
	IsHero                  bool
}

type SpellCfg struct {
	DisableProduction  bool
	SpellForgeLevel    uint32
	LaboratoryLevel    uint32
	TrainingResource   string
	TrainingCost       uint32
	HousingSpace       uint32
	TrainingTime       uint32
	DeployTimeMS       uint32
	ChargingTimeMS     uint32
	HitTimeMS          uint32
	UpgradeTimeH       uint32
	UpgradeResource    string
	UpgradeCost        uint32
	BoostTimeMS        uint32
	SpeedBoost         uint32
	SpeedBoost2        uint32
	JumpHousingLimit   uint32
	JumpBoostMS        uint32
	DamageBoostPercent uint32
	Damage             int32
	Radius             uint32
	NumberOfHits       uint32
	RandomRadius       uint32
	TimeBetweenHitsMS  uint32
}

type TaskCfg struct {
	Name        string
	InfoTID     string
	Progress    uint32
	Gold        uint32
	Food        uint32
	Gem         uint32
	Exp         uint32
	TroopType1  uint32
	TroopCount1 uint32
	TroopType2  uint32
	TroopCount2 uint32
	IsDayTask   bool
}
type TttCfg struct {
	ID            string
	TownhallLevel uint32
	Mark          uint32
	AwardType1    string
	Award1        uint32
	AwardType2    string
	Award2        uint32
	AwardType3    string
	Award3        uint32
}
type TttBuffCfg struct {
	BuffType string
	NameTID  string
	InfoID   string
	Icon     string
	CostType string
	Cost     uint32
	Arg1     uint32
}
type ExpCfg struct {
	ExpPoints uint32
}

type GlobalCfg struct {
	Value uint32
}

func (s *SpellCfg) GetUpgradeTime() uint32 {
	ts("this is s.UpgradeTimeH:%d", s.UpgradeTimeH)
	return s.UpgradeTimeH * 60 * 60
}

func (c *CharacterCfg) GetUpgradeTime() uint32 {
	return c.UpgradeTimeH * 60 * 60
}

func (b *BuildingCfg) GetBuildingTime() uint32 {
	return b.BuildTimeD*24*60*60 + b.BuildTimeH*60*60 + b.BuildTimeM*60 + b.BuildTimeS
}

//分享奖励
type ShareAwardCfg struct {
	Step    uint32
	GiveGem uint32
}

//连续登陆奖励
type LandAwardCfg struct {
	CurCount        uint32
	AddDiamondCount uint32
}

//pve数据
type PVEStageCfg struct {
	ID			string
	GoldStorage uint32
	FoodStorage uint32
	GuideStage	bool
}