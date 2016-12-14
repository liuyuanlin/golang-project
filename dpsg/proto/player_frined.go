package proto

const(
	PlayNotExist = 1
	FindPlayerOK = 2
)

type PlayerGiveGift struct {
	Uid string
}

type PlayerGiveGiftResult struct {
	Code uint32
}

type GiveGiftToPlayer struct {
	ToUid   string
	FromUid string
}
type GiveGiftToPlayerResult struct {
	
}