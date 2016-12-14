package proto

const (
	GetRankPlayerOk                = 1
	GetRankClanOk                  = 2
	GetRankPlayersLocationResultOK = 3
	GetRankPlayerTTTScoreResultOK  = 4
	GetMyselfResultOK              = 5
)

type GetRankPlayers struct {
	Uid   string
	Start int
	Stop  int
}

type GetRankPlayersResult struct {
	Value []string
	Code  int
}

type GetRankClans struct {
	Clan  string
	Start int
	Stop  int
}

type GetRankClansResult struct {
	Value []byte
	Code  int
}

// 玩家区域排行结构体
type GetRankPlayersLocation struct {
	Uid      string
	Location int64
	Start    int
	Stop     int
}

type GetRankPlayersLocationResult struct {
	Value []string
	Code  int
}

//通天塔
type GetRankPlayerTTTScore struct {
	Start int
	Stop  int
}

type GetRankPlayerTTTScoreResult struct {
	Value []PlayerTTTScoreStruct
	Code  int
}

//add for saveRankPlayer

type SaveRankPlayerResult struct {
	OK bool
}

type SaveRankPlayer struct {
	Value []byte
}

type SaveRankLocationPlayers struct {
	Value    []byte
	Location int64
}

type SaveRankLocationPlayersResult struct {
	OK bool
}

type SaveRankClanResult struct {
	OK bool
}

type SaveRankClan struct {
	Value []byte
}

type SaveTTTRankResult struct {
	OK bool
}

type SaveTTTRank struct {
	Value []byte
}

//add for updatemyself

type GetMyself struct {
	Uid      string
	Location string
}

type GetMyselfResult struct {
	Rank int
	Code int
}

type SaveMyself struct {
	Value []byte
}

type SaveMyselfResult struct {
	OK bool
}
