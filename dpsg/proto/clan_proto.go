package proto

const (
	CenterRpcError = -1

	CreateClanOk           = 0
	CreateClanFailed_Exist = 1

	ClanNotExist   = 2
	PlayerNotExist = 3

	GetClanOk    = 4
	GetClanError = 5

	GetClanPlayerOk = 6

	JoinClanOk                     = 10
	JoinClanFailed_ClanFull        = 11
	JoinClanFailed_AleadyIn        = 12
	JoinClanFailed_Refused         = 13
	JoinClanFailed_NotEnoughTrophy = 14
	JoinClanFailed_NoCastle        = 15

	LeaveClanOk                       = 20
	LeaveClanFailed_NotFoundPlayer    = 21
	LeaveClanFailed_LeaderCannotLeave = 22

	SearchClanOk                  = 30
	SearchClanFailed_NothingFound = 31

	SaveClanOk              = 40
	SaveClanFailed_NotFound = 41

	KickPlayerOk                    = 50
	KickPlayerFailed_CannotKickSelf = 51
	KickPlayerFailed_NotEnoughPower = 52
	KickPlayerFailed_NotFoundPlayer = 53

	AppointPlayerOk                       = 60
	AppointPlayerFailed_CannotAppointSelf = 61
	AppointPlayerFailed_NotEnoughPower    = 62
	AppointPlayerFailed_NotFoundPlayer    = 63
	AppointPlayerFailed_PowerNotChange    = 64
	AppointPlayerFailed_UnKnown           = 65

	DeleteClanOK = 70
)

type CreateClan struct {
	Value []byte
}

type CreateClanResult struct {
	Value int
}

type SaveClan struct {
	Value []byte
}

type SaveClanResult struct {
	Value int
}

type GetClan struct {
	Name string
}

type GetClanResult struct {
	Value []byte
	Code  int
}

type GetClanInfoResult struct {
	Value []byte
	Code  int
}

type GetClanPlayer struct {
	Name string
	Uid  string
}

type GetClanPlayerResult struct {
	Value []byte
	Code  int
}

type RandomGetClans struct {
	Trophy uint32
	Num    int
}

type RandomGetClansResult struct {
	Value []byte
	Code  int
}

type JoinClan struct {
	Value []byte
	Name  string
}

type JoinClanResult struct {
	Value int
}

type LeaveClan struct {
	PUid string
	Name string
}

type LeaveClanResult struct {
	Value int
}

type KickPlayer struct {
	Uid    string
	CName  string
	TarUid string
}

type KickPlayerResult struct {
	Value int
	Power int32
}

type AppointPlayer struct {
	Uid    string
	CName  string
	TarUid string
	Power  int32
}

type AppointPlayerResult struct {
	Value    int
	OldPower int32
}

type SearchClan struct {
	Key string
}

type SearchClanResult struct {
	Value []byte
	Code  int
}

type NotifyGetDonate struct {
	Uid  string
	Name string
}

type NotifyGetDonateResult struct {
}

type NotifyUpdateClanInfo struct {
	Uid   string
	Type  int32
	CName string
}

type NotifyUpdateClanInfoResult struct {
}

//add for send mail to clan players
type SendClanMail struct {
	Uid       string
	OldLeader string
	NewLeader string
}

type SendClanMailResult struct {
}
