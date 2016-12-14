package center

/*
import (
	"common"
	"fmt"
	"logger"
	"proto"
	"sync"
)

type Lock struct {
	name   string
	locked bool
	l      sync.Mutex
	value  uint64
	wait   sync.Mutex
}

func NewLock(name string, value uint64) (l *Lock) {
	l = &Lock{name: name, locked: true, value: value}
	l.wait.Lock()
	return
}

func (l *Lock) TryLock(value uint64) (ok bool, old_value uint64) {
	l.l.Lock()
	defer l.l.Unlock()

	if l.locked {
		return false, l.value
	}

	l.locked = true
	l.value = value

	l.wait.Lock()

	ok = true
	return
}

func (l *Lock) UnLock(value uint64) bool {
	l.l.Lock()
	defer l.l.Unlock()

	if !l.locked {
		logger.Error("Can't Unlock %s for %d", l.name, value)
		return false
	}

	if l.value != value {
		logger.Error("UnLock %s Error : %d != %d", l.name, value, l.value)
		return false
	}

	l.locked = false
	l.value = 0
	l.wait.Unlock()

	return true
}

func (l *Lock) ForceUnLock() bool {
	l.l.Lock()
	defer l.l.Unlock()

	if l.locked {
		l.locked = false
		l.value = 0
		l.wait.Unlock()

		return true
	}

	return false
}

type LockServer struct {
	name string
	m    map[string]*Lock
	l    sync.RWMutex
}

func NewLockServer(name string) *LockServer {
	return &LockServer{name: name, m: make(map[string]*Lock)}
}

func (l *LockServer) TryGetLock(name string, value uint64) (ok bool, old_value uint64) {
	l.l.RLock()
	locker, exist := l.m[name]
	l.l.RUnlock()

	if exist {
		return locker.TryLock(value)
	}

	l.l.Lock()
	defer l.l.Unlock()

	locker, exist = l.m[name]
	if exist {
		return locker.TryLock(value)
	} else {
		l.m[name] = NewLock(l.name+"."+name, value)

		ok = true
		return
	}

	panic("unreachable")
}

func (l *LockServer) UnLock(name string, value uint64) bool {
	l.l.RLock()
	locker, exist := l.m[name]
	l.l.RUnlock()

	if exist {
		return locker.UnLock(value)
	} else {
		logger.Error("UnLock Error %s Not Found! value: %d ", l.name+"."+name, value)
		return false
	}

	panic("unreachable")
}

func (l *LockServer) ForceUnLock(name string) (bool, error) {
	l.l.Lock()
	defer l.l.Unlock()

	locker, exist := l.m[name]

	if exist && locker.ForceUnLock() {
		return true, nil
	}

	return false, fmt.Errorf("lock name not found")
}

type LockServices struct {
	l sync.RWMutex
	m map[string]*LockServer
}

func NewLockServices() *LockServices {
	return &LockServices{m: make(map[string]*LockServer)}
}

func (l *LockServices) trygetlock(service string, name string, value uint64) (result bool, old_value uint64, err error) {
	l.l.RLock()
	server, exist := l.m[service]
	l.l.RUnlock()

	if exist {
		result, old_value = server.TryGetLock(name, value)
		return result, old_value, nil
	}

	l.l.Lock()
	defer l.l.Unlock()

	server, exist = l.m[service]
	if exist {
		result, old_value = server.TryGetLock(name, value)
		return result, old_value, nil
	}

	server = NewLockServer(service)
	l.m[service] = server

	result, old_value = server.TryGetLock(name, value)
	return result, old_value, nil
}

func (l *LockServices) TryGetLock(req *proto.TryGetLock, rep *proto.GetLockResult) (err error) {
	logger.Info("LockServices TryGetLock")

	logger.Info("尝试解析枷锁的类型")
	//尝试反加密信息
	_, tid, _, _, _ := common.ParseLockMessage(req.Value)

	//如果是pvp或者是玩家复仇，则去擂台的列表中看看是不是在守擂
	if tid == proto.MethodPlayerMatch || tid == proto.MethodPlayerRevenge {
		exist := pCenterChallengeService.IsInTheChallenge(req.Name)
		if exist {
			rep.Result = false
			rep.OldValue = 0
			return
		}
	}

	rep.Result, rep.OldValue, err = l.trygetlock(req.Service, req.Name, req.Value)

	return
}

func (l *LockServices) unLock(service string, name string, value uint64) (result bool, err error) {
	l.l.RLock()
	server, exist := l.m[service]
	l.l.RUnlock()
	if !exist {
		return false, fmt.Errorf("Services %s not found", service)
	}

	result = server.UnLock(name, value)

	return result, nil
}

func (l *LockServices) UnLock(req *proto.FreeLock, rep *proto.FreeLockResult) (err error) {
	rep.Result, err = l.unLock(req.Service, req.Name, req.Value)

	return
}

func (l *LockServices) forceUnLock(service string, name string) (result bool, err error) {
	l.l.RLock()
	server, exist := l.m[service]
	l.l.RUnlock()
	if !exist {
		return false, fmt.Errorf("Services %s not found", service)
	}

	return server.ForceUnLock(name)
} */
