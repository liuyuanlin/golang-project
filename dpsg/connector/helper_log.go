package connector

import (
	//"logger"
	"golang-project/dpsg/proto"
	"golang-project/dpsg/rpc"
	"time"
)

//发送log
func LOG_Resources(gl rpc.GameLocation, uid string, gain bool, ResType string, ResNum uint32, ResWay uint32) bool {
	msg := proto.LogResources{
		ChannelId: uint8(gl),
		Uid:       uid,
		Gain:      gain,
		Time:      time.Now().Unix(),
		ResType:   ResType,
		ResNum:    ResNum,
		ResWay:    ResWay,
	}

	var ret proto.LogResourcesResult
	cns.logRpcConn.Go("LogServices.LogResources", msg, &ret, nil)

	return true
}
