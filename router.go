package enet

import (
	"sync"
)

//实现router时，先嵌入这个基类，然后根据需要对这个基类的方法进行重写
type BaseRouter struct {
	sync.RWMutex
	ExitMap map[uint32]chan bool // 连接ID

	*WaitGroupWrapper
}

func NewBaseRouter() *BaseRouter {
	return &BaseRouter{
		ExitMap:          make(map[uint32]chan bool),
		WaitGroupWrapper: &WaitGroupWrapper{},
	}
}

// 这里之所以BaseRouter的方法都为空，
// 是因为有的Router不希望有PreHandle或PostHandle
// 所以Router全部继承BaseRouter的好处是，不需要实现PreHandle和PostHandle也可以实例化

// 预处理函数 不要阻塞
func (br *BaseRouter) PreHandle(req IRequest) {}

// 处理函数 可以阻塞
func (br *BaseRouter) Handle(req IRequest) {}

// 后处理函数 不要阻塞
func (br *BaseRouter) PostHandle(req IRequest) {}

// 不要重写
func (br *BaseRouter) WrapHandle(connID uint32, cb func()) {
	br.Lock()
	br.ExitMap[connID] = make(chan bool)
	br.Unlock()

	br.Wrap(cb)
}

// 不要重写
func (br *BaseRouter) ExitHandle(connID uint32) {
	ch, ok := make(chan bool), true

	br.Lock()
	if ch, ok = br.ExitMap[connID]; !ok {
		br.Unlock()
		return
	}
	//delete(br.ExitMap, connID)
	br.Unlock()

	close(ch)

	br.Wait()
}

// 不要重写
func (br *BaseRouter) Exit(connID uint32) <-chan bool {
	br.RLock()
	defer br.RUnlock()
	// TODO 在什么时候删除这个键值对呢 delete(br.ExitMap, connID)
	return br.ExitMap[connID]
}
