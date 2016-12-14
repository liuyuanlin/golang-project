package router

import (
	"io"
	"rpcplus"
	"sync"
	"sync/atomic"
)

type Router struct {
	servers map[uint64]*rpcplus.Client
	id      uint64
	l       sync.Mutex
}

func (r *Router) OnServerConnected(c io.ReadWriteCloser) {
	new_id := atomic.AddUint64(r.id, 1)
	client := rpcplus.NewClient(c)
	r.l.Lock()
	r.servers[new_id] = client
	r.l.Unlock()

	client.SetCloseCallBack(
		func(err Error) {
			r.l.Lock()
			delete(r.servers, new_id)
			r.l.Unlock()
		},
	)
}
