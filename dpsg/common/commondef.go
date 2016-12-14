package common

const (
	SystemTableName = "system"
)

const DbTableKeySplit = ":"

const (
	SystemKeyName_Mail           = "mail"
	SystemKeyName_Name2Id        = "playername2id"
	SystemKeyName_TTTScore       = "tttscore"
	SystemKeyName_Level2PlayerId = "level2playerid"

	//add for challenge
	ChallengeKeyName_NormalChallenge     = "normalchallenge"
	ChallengeKeyName_MoneyChallenge      = "moneychallenge"
	ChallengeKeyName_PlayerChallengeInfo = "playerchallengeinfo"
	
	PlayerBase = "playerbase"
	PlayerExtra = "playerextra"
)

const (
	//账号服务器
	TableName_TencentAccount = "tencentid2playerid"
)

func GetSystemTableKey_Mail() string {
	return SystemTableName + DbTableKeySplit + SystemKeyName_Mail
}

//通天塔分数表
func GetSystemKeyName_TTTScore() string {
	return SystemTableName + DbTableKeySplit + SystemKeyName_TTTScore
}

//玩家等级对应村庄id
func GetSystemKeyName_Level2PlayerId() string {
	return SystemTableName + DbTableKeySplit + SystemKeyName_Level2PlayerId
}
