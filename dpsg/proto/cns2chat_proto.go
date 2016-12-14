package proto

import (
	"golang-project/dpsg/rpc"
)

type AddPlayer struct {
	AllianceName   string
	AllianceSymbol uint32
	AlliancePower  uint32
	PlayerId       string
	AuthKey        string
	PlayerName     string
	PlayerLevel    uint8
	ChannelId      rpc.GameLocation
}

type AddPlayerResult struct {
}

type DelPlayer struct {
	PlayerId string
}

type DelPlayerResult struct {
}

type UpdatePlayer struct {
	PlayerId      string
	PlayerLevel   uint8
	AllianceName  string
	AlliancePower uint32
}

type UpdatePlayerResult struct {
}

type PlayerChatToPlayer struct {
	FromPlayerId string
	ToPlayerId   string
	Content      string
}

type PlayerChatToPlayerResult struct {
}

type PlayerWorldChat struct {
	FromPlayerId string
	Content      string
	CName        string
	CSymbol      uint32
}

type PlayerWorldChatResult struct {
}

type ClanMsg struct {
	CName string
	Value []byte
}

type ClanMsgResult struct {
}

type DonateMsg struct {
	Uid   string
	Value []byte
}

type DonateMsgResult struct {
}

type GetMailAttach struct {
	PlayerId string
	MailId   string
}

type GetMailAttachResult struct {
	Attach string
}

type DelPlayerMail struct {
	PlayerId string
	MailId   string
}

type DelPlayerMailResult struct {
}

type SendPlayerMail struct {
	ToPlayerId     string
	FromUid        string
	FromName       string
	FromLevel      uint32
	FromClan       string
	FromClanSymbol uint32
	Title          string
	Content        string
	Attach         string
}

type SendPlayerMailResult struct {
}

type SendSystemMail struct {
	ToPlayerId string
	Title      string
	Content    string
	Attach     string
}

type SendSystemMailResult struct {
}

type ReadPlayerMail struct {
	PlayerId string
	MailId   string
}

type ReadPlayerMailResult struct {
}
