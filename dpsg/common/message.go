package common

import (
	"fmt"
	"golang-project/dpsg/logger"
	"golang-project/dpsg/rpc"
)

func WriteResult(conn rpc.RpcConn, value interface{}) bool {
	err := conn.WriteObj(value)
	if err != nil {
		logger.Info("WriteResult Error %s", err.Error())
		return false
	}
	return true
}

func SyncError(conn rpc.RpcConn, format string, args ...interface{}) {
	//tArgs := make([]interface{}, len(args))
	//for i, arg := range args {
	//	tArgs[i] = arg
	//}

	msg := rpc.SyncError{}
	msg.SetText(fmt.Sprintf(format, args...))

	WriteResult(conn, &msg)

	logger.Error(format, args...)
}

func SendMsg(conn rpc.RpcConn, code string) {
	msg := rpc.Msg{}
	msg.SetCode(code)

	WriteResult(conn, &msg)
}

func SendText(conn rpc.RpcConn, text string) {
	msg := rpc.Msg{}
	msg.SetText(text)

	WriteResult(conn, &msg)
}
