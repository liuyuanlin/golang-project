package proto

const (
	ResType_Gold   = "gold"
	ResType_Food   = "food"
	ResType_Wuhun  = "wuhun"
	ResType_Gem    = "diamonds"
	ResType_Trophy = "trophy"
	ResType_TiLi   = "tili"

	Gain_Gather                 = 1
	Gain_Plunder                = 2 //抢劫别人所得
	Gain_Robed                  = 3 //被别人抢劫所得
	Gain_Pve                    = 4
	Gain_Task                   = 5
	Gain_SymtemMail             = 6
	Gain_SellBuilding           = 7
	Gain_CacelCreateBuilding    = 8
	Gain_CacelUpgradeBuilding   = 9
	Gain_CacelCreateCharacter   = 10
	Gain_CacelUpgradeCharacter  = 11
	Gain_Recharge               = 12
	Gain_ClanBattle             = 13
	Gain_RemoveObstacle         = 14
	Gain_GM                     = 15
	Gain_CacelCreatespell       = 16
	Gain_Recharge_TW_GooglePlay = 17
	Gain_Recharge_TW_3Pay_Tel   = 18
	Gain_Recharge_TW_3Pay_Gash  = 19
	Gain_Recharge_TW_Ios        = 20
	Gain_Recharge_EN_GooglePlay = 21
	Gain_Recharge_EN_Ios        = 22
	Gain_Recharge_VT_GooglePlay = 23
	Gain_Recharge_VT_Ios        = 24
	Gain_Recharge_VT_Web        = 25
	Gain_LandReward             = 26
	Gain_Challenge              = 27
	Gain_AddTiLI                = 28

	Lose_Plunder          = 50 //抢劫别人损失
	Lose_Robed            = 50 //被别人抢劫损失
	Lose_Search           = 51
	Lose_CreateBuilding   = 52
	Lose_UpgradeBuilding  = 53
	Lose_CreateCharacter  = 54
	Lose_UpgradeCharacter = 55
	Lose_CreateHero       = 56
	Lose_UpgradeHero      = 57
	Lose_FinishNow        = 58
	Lose_FillXBow         = 59
	Lose_BuyRes           = 60
	Lose_BuyTime          = 61
	Lose_BuyDeco          = 62
	Lose_BuyWorker        = 63
	Lose_BuyDrill         = 64
	Lose_BuyShield        = 65
	Lose_WorldChat        = 66
	Lose_GM               = 67
	Lose_BuyFriendDrill   = 68
	Lose_CreateSpell      = 69
	Lose_UseTTTBuff       = 70
	Lose_BuyBattleAcc     = 71
	Lose_TTTBattle        = 72
	Lose_Challenge        = 73
	Lose_BuyTiLi          = 74

	Exception_RechargeFailed = 100
)

type LogPlayerLoginLogout struct {
	ChannelId uint8
	Playerid  string
	Time      int64
	Logout    bool
	Ip        string
}

type LogPlayerLoginLogoutResult struct {
}

type LogResources struct {
	ChannelId uint8
	Uid       string
	Time      int64
	Gain      bool
	ResType   string
	ResNum    uint32
	ResWay    uint32
}

type LogResourcesResult struct {
}

type TaobaoPayLog struct {
	TradeEnd    bool
	TradeError  string
	TradeNumber string
	CharId      string
	ItemName    string
	TotoalPee   string
	TradeTime   int64
	ChannelId   uint32
}

type TaobaoPayLogResult struct {
}
