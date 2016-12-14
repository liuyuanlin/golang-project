package center

import (
	"encoding/csv"
	"fmt"
	"golang-project/dpsg/common"
	"golang-project/dpsg/dbclient"
	"golang-project/dpsg/logger"
	"golang-project/dpsg/proto"
	"golang-project/dpsg/rpc"
	"golang-project/dpsg/timer"
	"os"
	"path"
	"strconv"
	"time"

	gp "github.com/golang/protobuf/proto"
)

//ttt
//每周刷新的Tick do
func (self *Center) initDayTick() {
	//加载配置表
	common.LoadConfigFiles()
	//首先计算现在的时间到晚上0点时间的间隔，注册一个时间tick
	t := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 24, 0, 0, 0, time.Local)
	durtime := t.Unix() - time.Now().Unix()
	fmt.Println("现在的秒数   24点得临界秒数   看间隔的秒数 durtime Unix  ", time.Now().Unix(), t.Unix(), durtime)

	//开始定时器
	self.eveyrdaytime = timer.NewTimer(time.Duration(durtime) * time.Second)
	self.eveyrdaytime.Start(
		func() {
			self.onDayTick()
		},
	)
}

func (self *Center) onDayTick() {

	self.eveyrdaytime.Stop()
	self.eveyrdaytime = nil

	if time.Now().Weekday().String() == "Monday" {
		fmt.Println("测试是星期天 到 星期一 跨时间 开始奖励发放和数据清空")
		//在这里给玩家发放奖励
		// 开始正式加载文件
		filename := path.Join(common.GetDesignerDir(), "ttt_award.csv")
		infile, _ := os.Open(filename)
		r := csv.NewReader(infile)
		r.TrailingComma = true
		out, _ := r.ReadAll()
		fmt.Println("奖励的配置表的长度 ", len(out))

		//取出当前的redis的排名表的
		tableLength, _ := self.zcard(common.SystemTableName, common.SystemKeyName_TTTScore)
		fmt.Println("redis表的长度 ", tableLength)
		//记录一类发放人数的最大值

		var maxNumber uint32 = 0
		//循环按照段数发放(这里要改成发送一次，发送的是列表，因为如果人口太多的话，压力很大的)
		for row := 2; row < len(out); row++ {
			fmt.Println("row   index  ", row, row-1)
			cfg := common.GetTttAwardCfg(strconv.FormatInt(int64(row-1), 10))
			//第一类发放
			if cfg.RankRangeType == "INT" {
				fmt.Println("进入 一类 计算")
				if cfg.RankRangeMax > maxNumber {
					maxNumber = cfg.RankRangeMax
				}
				//fmt.Println("row  from  to 表长度 最小值", int(cfg.RankRangeMin)-1, int(cfg.RankRangeMax)-1)
				//fmt.Println("打印一类长度的最大值 ", maxNumber)
				//判断表里面的数据长度
				if tableLength > 0 && tableLength >= cfg.RankRangeMin {
					fmt.Println("分类在INT里面的 表长度 大于 1 row  from  to", int(cfg.RankRangeMin)-1, int(cfg.RankRangeMax)-1)
					//去取redis表
					buf, err := self.zrevrange(common.SystemTableName, common.SystemKeyName_TTTScore, int(cfg.RankRangeMin)-1, int(cfg.RankRangeMax)-1)
					if err != nil {
						logger.Error("get data from file error filename :", common.SystemTableName+":"+common.SystemKeyName_TTTScore)
					}

					req0 := &proto.SendTTTSystemMail{
						SenduidArray: buf,
						Awardtype1:   rpc.MailAttach_Gold,
						Awardnum1:    cfg.Award1,
						Awardtype2:   rpc.MailAttach_Gem,
						Awardnum2:    cfg.Award2,
					}
					rst0 := &proto.SendTTTSystemMailResult{}
					rst0.Success = false
					for _, rpccli := range self.cnss {
						fmt.Println("self.cnss...1类")
						if err := rpccli.Call("CenterService.SendTTTMail", req0, rst0); err == nil && rst0.Success {
							rst0.Success = true
							break
						}
					}
				}
			}
			//第二类发放
			if cfg.RankRangeType == "PERCENT" {
				fmt.Println("进入 二类 计算")
				if cfg.RankRangeMax*tableLength/100 > maxNumber {
					startNumber := cfg.RankRangeMin * tableLength / 100
					if cfg.RankRangeMin*tableLength/100 < maxNumber {
						//这里面是包含关系的情况
						startNumber = cfg.RankRangeMax*tableLength/100 - maxNumber
						fmt.Println(" 分类在PERCENT里面 并且大一段和 最大的INT 有交集")
					}
					fmt.Println(" 分类在PERCENT里面  ")
					buf, err := self.zrevrange(common.SystemTableName, common.SystemKeyName_TTTScore, int(startNumber)-1, int(cfg.RankRangeMax*tableLength/100)-1)
					if err != nil {
						logger.Error("get data from file error filename :", common.SystemTableName+":"+common.SystemKeyName_TTTScore)
					}
					req0 := &proto.SendTTTSystemMail{
						SenduidArray: buf,
						Awardtype1:   rpc.MailAttach_Gold,
						Awardnum1:    cfg.Award1,
						Awardtype2:   rpc.MailAttach_Gem,
						Awardnum2:    cfg.Award2,
					}
					rst0 := &proto.SendTTTSystemMailResult{}
					rst0.Success = false
					for _, rpccli := range self.cnss {
						fmt.Println("self.cnss..2类.")
						if err := rpccli.Call("CenterService.SendTTTMail", req0, rst0); err == nil && rst0.Success {
							rst0.Success = true
							break
						}
					}
				}
			}
		}

		fmt.Println(" 完成每周 奖励发放")
		//发完奖励清空表
		self.del(common.SystemTableName, common.SystemKeyName_TTTScore)
	}
	self.eveyrdaytime = timer.NewTimer(time.Duration(24*3600) * time.Second)
	self.eveyrdaytime.Start(
		func() {
			self.onDayTick()
		},
	)
}

//通天塔积分
func (self *Center) UpdatePlayerTTTScore(req *proto.UpdatePlayerTTTScore, rst *proto.UpdatePlayerTTTScoreResult) error {
	return self.zadd(common.SystemTableName, common.SystemKeyName_TTTScore,
		req.Id, req.Score)
}

func (self *Center) getRankPlayerTTTScore(req *proto.GetRankPlayerTTTScore, reply *proto.GetRankPlayerTTTScoreResult, bIsFirst bool) (err error) {
	buf, err := self.zrevrange(common.SystemTableName, common.SystemKeyName_TTTScore, req.Start, req.Stop)
	if err != nil {
		return err
	}

	playerTTTScoreStruct := &proto.PlayerTTTScoreStruct{}
	for _, uid := range buf {
		score, err := self.zscore(common.SystemTableName, common.SystemKeyName_TTTScore, uid)
		if err != nil {
			return err
		}
		playerTTTScoreStruct.Id = uid
		playerTTTScoreStruct.Score = score
		reply.Value = append(reply.Value, *playerTTTScoreStruct)
	}

	reply.Code = proto.GetRankPlayerTTTScoreResultOK

	rps := &rpc.TTTRankPlayers{}
	myreply := &proto.SaveTTTRankResult{}

	for _, PlayerTTTScoreStructinfo := range reply.Value {

		var TTTp rpc.PlayerBaseInfo
		exist, err := dbclient.KVQueryBase(common.PlayerBase, PlayerTTTScoreStructinfo.Id, &TTTp)
		if err != nil {
			continue
		}

		if exist {
			rp := rpc.TTTPlayer{}
			rp.SetType(rpc.TTTPlayer_Rank)
			rp.SetName(TTTp.GetName())
			rp.SetUid(TTTp.GetUid())
			rp.SetLevel(TTTp.GetLevel())
			rp.SetTrophy(TTTp.GetTrophy())
			rp.SetClanName(TTTp.GetClan())
			rp.SetClanSymbol(TTTp.GetClanSymbol())
			rp.SetTttSCoreQuery(PlayerTTTScoreStructinfo.Score)
			rp.SetRanknumberQuery(0)

			rps.RpsTop = append(rps.RpsTop, &rp)
		}

	}

	buff, myerr := gp.Marshal(rps)
	if err != nil {
		logger.Error("SearchClan Error On Marshal (%s, %v)", myerr.Error(), buff)
		return
	}

	tttResult := &proto.SaveTTTRank{}

	tttResult.Value = buff

	if !bIsFirst {
		for _, conn := range centerServer.cnss {
			conn.Call("CenterService.SaveTTTPlayers", tttResult, myreply)

			if !myreply.OK {
				logger.Error("CenterService.SaveTTTPlayers error")
			}
		}
	}

	//这里保存一下第一次查到的结果
	SaveAllRankPlayers.RankTTTplayers = *tttResult

	return nil
}

func (self *Center) GetPlayerTTTScore(req *proto.QueryPlayerTTTScore, rst *proto.QueryPlayerTTTScoreResult) error {
	fmt.Println("1 请求 最高积分  通天塔排行名次  ")
	score, err := self.zscore(common.SystemTableName, common.SystemKeyName_TTTScore, req.Id)
	if err != nil {
		return err
	}

	rank, err := self.zrevrank(common.SystemTableName, common.SystemKeyName_TTTScore, req.Id)
	if err != nil {
		return err
	}

	rst.Score = uint32(score)
	rst.Rank = uint32(rank)
	fmt.Println("2 请求 最高积分  通天塔排行名次  ", rst.Score, rst.Rank)
	return nil
}

//等级对应玩家id
func (self *Center) UpdatePlayerLevel2Id(req *proto.UpdaePlayerLevel2Id, rst *proto.UpdaePlayerLevel2IdResult) error {
	if req.Level > 1 { //从2级的时候才开始删除之前的
		self.srem(common.SystemTableName,
			common.SystemKeyName_Level2PlayerId+strconv.FormatUint(uint64(req.Level)-1, 10),
			req.Id)
	}

	self.sadd(common.SystemTableName,
		common.SystemKeyName_Level2PlayerId+strconv.FormatUint(uint64(req.Level), 10),
		req.Id)

	return nil
}

func (self *Center) RandomGetPlayerIdByLevel(req *proto.RandomGetPlayerIdByLevel, rst *proto.RandomGetPlayerIdByLevelResult) error {
	id, err := self.srandmember(common.SystemTableName,
		common.SystemKeyName_Level2PlayerId+strconv.FormatUint(uint64(req.Level), 10))

	if err != nil {
		return err
	}

	rst.Id = id
	fmt.Println(" 从数据库取出的对手的ID", id)
	return nil
}
