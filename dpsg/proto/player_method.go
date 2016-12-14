package proto

const (
	MethodPlayerLogin = iota
	MethodPlayerMatch
	MethodPlayerRevenge
	MethodPlayerGiveGem
	MethodPlayerGmOpera
)

type NotifyGivePlayerGem struct {
	Uid string
	Num uint32
}

type NotifyGivePlayerGemResult struct {
	Ok bool
}
