package lockclient

import (
	"golang-project/dpsg/common"
	"golang-project/dpsg/logger"
	"golang-project/dpsg/proto"
	"golang-project/dpsg/rpcplusclientpool"
	"time"
)

var pPoll *rpcplusclientpool.ClientPool

//初始化加锁客户端
func Init() error {
	var lscfg common.LockServerCfg
	err := common.ReadLockServerConfig(&lscfg)
	if err != nil {
		return err
	}

	aServerHost := make([]string, 0)
	aServerHost = append(aServerHost, lscfg.LockHost)

	pPoll = rpcplusclientpool.CreateClientPool(aServerHost)
	if pPoll == nil {
		logger.Fatal("create failed")
	}

	return nil
}

//尝试加锁
func TryLock(service, key string, lid uint64) (result bool, old_value uint64, err error) {
	try := &proto.TryGetLock{Service: service, Name: key, Value: lid}
	rst := &proto.GetLockResult{}

	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return
	}

	err = conn.Call("LockServices.TryGetLock", try, rst)
	if err != nil {
		return
	}

	result = rst.Result
	old_value = rst.OldValue

	return
}

//取消加锁
func TryUnlock(service, key string, lid uint64) (result bool, err error) {
	req := &proto.FreeLock{Service: service, Name: key, Value: lid}
	rst := &proto.FreeLockResult{}

	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return
	}

	err = conn.Call("LockServices.UnLock", req, rst)
	if err != nil {
		return
	}

	result = rst.Result

	return
}

//取消加锁
func ForceUnLock(service, key string) (result bool, err error) {
	req := &proto.ForceUnLock{Service: service, Name: key}
	rst := &proto.ForceUnLockResult{}

	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return
	}

	err = conn.Call("LockServices.ForceUnLock", req, rst)
	if err != nil {
		return
	}

	result = rst.Result

	return
}

func WaitLockGet(service, key string, lid uint64) bool {
	timeLimit := 0

	for {
		successed, _, err := TryLock(service, key, lid)

		if err != nil {
			return false
		}

		if successed {
			return true
		}

		if timeLimit > 10 {
			return false
		}

		timeLimit += 1

		time.Sleep(time.Millisecond * 200)
	}

	panic("WaitLockGet unreachable")
}
