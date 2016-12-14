package connector

import (
//"lockclient"
//"logger"
//"proto"
//"time"
)

/*func LockGet(service, key string, lid uint64) (result bool, old_value uint64, err error) {
	ts("LockGet", service, key, lid)
	defer te("LockGet", service, key, lid)

	try := &proto.TryGetLock{Service: service, Name: key, Value: lid}
	rst := &proto.GetLockResult{}

	err = cns.center.Call("LockServices.TryGetLock", try, rst)

	if err != nil {
		logger.Error("Error On LockGet : %s", err.Error())
		return
	}

	result = rst.Result
	old_value = rst.OldValue

	return
}*/

/*func LockFree(service, key string, lid uint64) (result bool, err error) {
	ts("LockFree", service, key, lid)
	defer te("LockFree", service, key, lid)

	ul := &proto.FreeLock{Service: service, Name: key, Value: lid}
	fl := &proto.FreeLockResult{}

	err = cns.center.Call("LockServices.UnLock", ul, fl)

	if err != nil {
		logger.Error("Error On LockFree : %s", err.Error())
		return
	}

	result = true
	return
}*/

/*func WaitLockGet(service, key string, lid uint64) bool {
	timeLimit := 0

	for {
		successed, _, err := lockclient.TryLock(service, key, lid)

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
}*/
