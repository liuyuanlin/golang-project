package main

import (
	db "golang-project/dpsg/accountserver"
	"golang-project/dpsg/common"
	"golang-project/dpsg/logger"
	"net"
	_ "net/http/pprof"
)

var dbServer *db.AccountServer

func main() {
	logger.Info("accountsserver start")

	var dbcfg common.DBConfig
	if err := common.ReadDbConfig("accountserver.json", &dbcfg); err != nil {
		logger.Fatal("load config failed, error is: %v", err)
		return
	}

	common.DebugInit(dbcfg.GcTime, dbcfg.DebugHost)

	dbServer = db.NewAccountServer(dbcfg)

	tsock, err := net.Listen("tcp", dbcfg.DBHost)
	if err != nil {
		logger.Fatal("net.Listen: %s", err.Error())
	}

	go db.StartServices(dbServer, tsock)

	db.WaitForExit(dbServer)

	tsock.Close()

	logger.Info("accountsserver end")
}
