package proto

import (
	"golang-project/dpsg/rpc"
)

//查询是否重名
type QueryName struct {
	Name   string
	Id     string
	BQuery bool //为true时表示根据名字查询返回值，false表示要插入名字与id的关联
}

type QueryNameResult struct {
	Success bool //返回值根据上面的BQuery不同
	Id      string
}

//cns请求center处理，center再分别发给每个cns
//登陆踢人
type LoginKickPlayer struct {
	Id string
}

type LoginKickPlayerResult struct {
	Success bool
}

//通天塔相关的功能
type UpdatePlayerTTTScore struct {
	Id    string
	Score uint32
}

type UpdatePlayerTTTScoreResult struct {
}

type QueryPlayerTTTScore struct {
	Id string
}

type QueryPlayerTTTScoreResult struct {
	Score uint32
	Rank  uint32
}

/*
type GetRankPlayerTTTScore struct {
	Start int
	Stop  int
}
*/
type PlayerTTTScoreStruct struct {
	Id    string
	Score uint32
}

/*
type GetRankPlayerTTTScoreResult struct {
	Value []PlayerTTTScoreStruct
}
*/
type SendTTTSystemMail struct {
	SenduidArray []string
	Awardtype1   rpc.MailAttach
	Awardnum1    uint32
	Awardtype2   rpc.MailAttach
	Awardnum2    uint32
}

type SendTTTSystemMailResult struct {
	Success bool
}

//等级对应玩家id
type UpdaePlayerLevel2Id struct {
	Id    string
	Level uint32
}

type UpdaePlayerLevel2IdResult struct {
}

type RandomGetPlayerIdByLevel struct {
	Level uint32
}

type RandomGetPlayerIdByLevelResult struct {
	Id string
}
