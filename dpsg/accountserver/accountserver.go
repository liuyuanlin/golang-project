package accountserver

import (
	"bytes"
	"fmt"
	"golang-project/dpsg/common"
	"golang-project/dpsg/logger"
	"golang-project/dpsg/proto"
	rpc "golang-project/dpsg/rpcplus"
	"net"
	"net/http"
)

type dbGroup []*common.DbPool
type cacheGroup []*common.CachePool

type AccountServer struct {
	dbGroups    map[string]dbGroup
	cacheGroups map[string]cacheGroup
	tables      map[string]*table
	exit        chan bool
}

func StartServices(self *AccountServer, listener net.Listener) {
	rpcServer := rpc.NewServer()
	rpcServer.Register(self)

	rpcServer.HandleHTTP("/dbserver/rpc", "/debug/rpc")

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("StartServices %s", err.Error())
			break
		}
		go func() {
			rpcServer.ServeConn(conn)
			conn.Close()
		}()
	}
}

func WaitForExit(self *AccountServer) {
	<-self.exit
	close(self.exit)
}

func NewAccountServer(cfg common.DBConfig) (server *AccountServer) {
	server = &AccountServer{
		dbGroups:    map[string]dbGroup{},
		cacheGroups: map[string]cacheGroup{},
		tables:      map[string]*table{},
		exit:        make(chan bool),
	}

	http.Handle("/debug/state", debugHTTP{server})

	//初始化所有的db
	for key, pools := range cfg.DBProfiles {
		logger.Info("Init DB Profile %s", key)

		server.dbGroups = make(map[string]dbGroup)
		for _, poolCfg := range pools {
			logger.Info("Init DB %v", poolCfg)
			server.dbGroups[key] = append(server.dbGroups[key], common.NewDBPool(poolCfg))
		}
	}

	//初始化所有的cache
	for key, pools := range cfg.CacheProfiles {
		logger.Info("Init Cache Profile %s", key)

		server.cacheGroups = make(map[string]cacheGroup)
		for _, poolCfg := range pools {
			logger.Info("Init Cache %v", poolCfg)
			server.cacheGroups[key] = append(server.cacheGroups[key], common.NewCachePool(poolCfg))
		}
	}

	//初始化table
	for key, table := range cfg.Tables {
		logger.Info("Init Table: %s %v", key, table)

		server.tables[key] = NewTable(key, table, server)
	}

	return server
}

func (self *AccountServer) Write(args *proto.AccountDbWrite, reply *proto.AccountDbWriteResult) error {
	logger.Info("AccountServer.Write : %v", args)
	if table, exist := self.tables[args.Table]; exist {
		err := table.write(args.Key, args.Value)
		if err != nil {
			return err
		}
		reply.Code = proto.Ok
	} else {
		reply.Code = proto.NoExist
	}

	return nil
}

func (self *AccountServer) Query(args *proto.AccountDbQuery, reply *proto.AccountDbQueryResult) error {
	logger.Info("AccountServer.Query : %v", args)
	if table, exist := self.tables[args.Table]; exist {
		rst, err := table.get(args.Key)
		if err != nil {
			return err
		}
		if rst != "" {
			reply.Value = rst
			reply.Code = proto.Ok
		} else {
			reply.Code = proto.NoExist
		}

	} else {
		reply.Code = proto.NoExist
	}

	return nil
}

func (self *AccountServer) ReQuery(args *proto.AccountDbQuery, reply *proto.AccountDbQueryResult) error {
	logger.Info("AccountServer.ReQuery : %v", args)
	if table, exist := self.tables[args.Table]; exist {
		rst, err := table.reget(args.Key)
		if err != nil {
			return err
		}
		if rst != "" {
			reply.Value = rst
			reply.Code = proto.Ok
		} else {
			reply.Code = proto.NoExist
		}

	} else {
		reply.Code = proto.NoExist
	}

	return nil
}

func (self *AccountServer) Delete(args *proto.AccountDbDel, reply *proto.AccountDbDelResult) error {
	logger.Info("AccountServer.Delete : %v", args)
	if table, exist := self.tables[args.Table]; exist {
		err := table.del(args.Key)
		if err != nil {
			return err
		}
		reply.Code = proto.Ok
	} else {
		reply.Code = proto.NoExist
	}

	return nil
}

func (self *AccountServer) Quit(args *int32, reply *int32) error {
	self.exit <- true
	return nil
}

func (self *AccountServer) statsJSON() string {
	buf := bytes.NewBuffer(make([]byte, 0, 128))
	fmt.Fprintf(buf, "{")
	for k, v := range self.tables {

		fmt.Fprintf(buf, "\n \"Table\": {")

		fmt.Fprintf(buf, "\n   \"Name\": \"%v\",", k)
		fmt.Fprintf(buf, "\n   \"States\": %v,", v.tableStats.String())
		fmt.Fprintf(buf, "\n   \"Rates\": %v,", v.qpsRates.String())

		fmt.Fprintf(buf, "\n }")
	}

	fmt.Fprintf(buf, "\n}")
	return buf.String()
}
