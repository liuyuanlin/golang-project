package proto

const (
	NormalChallengeOK    = 1
	MoneyChallengeOK     = 2
	AlreadyChallengeing  = 3
	AlreadyTimeOut       = 4
	Empty                = 5
	NoArena              = 6
	AlreadyHost          = 7
	AlreadyChallenger    = 8
	GetMoneyOK           = 9
	AlreadyGetMoney      = 10
	AlreadyHasChallenger = 11
)

type NormalChallenger struct {
	Index           uint32
	Level           uint32
	Challengeid     string
	ChallengeName   string
	ChallengeClan   string
	ChallengeSymbol uint32
}

type MoneyChallenger struct {
	Index           uint32
	Level           uint32
	Money           uint32
	Challengeid     string
	ChallengeName   string
	ChallengeClan   string
	ChallengeSymbol uint32
}

type NormalChallengeResult struct {
	Hostid string
	Code   int
}

type MoneyChallengeResult struct {
	Hostid string
	Code   int
}

//取擂台列表
type GetChallengeList struct {
	Uid   string
	Level int
}

type GetChallengeListResult struct {
	Value []byte
}

//add for send mail
type SendMail struct {
	Uid   string
	Money uint32
}

type SendMailResult struct {
}

//add for call center
type SendtoCenter struct {
	Challengeid string
	Hostid      string
	Name        string
	IsSuccess   bool
	IsFinished  bool
}

type SendtoCenterResult struct {
}

type PlayerReturnHome struct {
	Uid string
}

type PlayerReturnHomeResult struct {
	Code int
}

//玩家领取爵位工资
type GetPlayerScore struct {
	Uid string
}

type GetPlayerScoreResult struct {
	Score      uint32
	Salarytime uint32
}

type GetDailyMoney struct {
	Uid string
}

type GetDailyMoneyResult struct {
	Value int
}
