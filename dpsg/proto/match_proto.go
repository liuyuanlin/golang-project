package proto

type SetTrophy struct {
	Trophy   uint32
	Uid      string
	Location int64 //add for location
}

type SetTrophyResult struct {
	Value int
}

type NotifyOnline struct {
	Uid string
}

type NotifyOnlineResult struct {
	Value int
}

type NotifyOffline struct {
	Uid string
	GuideFinish bool
}

type NotifyOfflineResult struct {
	Value int
}

type AddShield struct {
	StartTime uint32
	TotalTime uint32
	Uid       string
}

type AddShieldResult struct {
	Value int
	Ok    bool
}

type RemoveShield struct {
	Uid string
}

type RemoveShieldResult struct {
	Value int
}

type Match struct {
	Trophy uint32
	Except string
}

type MatchResult struct {
	Uid string
}
