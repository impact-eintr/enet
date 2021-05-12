package znet

import (
	"errors"
	"fmt"
	"sync"

	"github.com/impact-eintr/Zinx/ziface"
)

type ConnManager struct {
	connections map[uint32]ziface.IConnection
	connLock    sync.RWMutex // 保护链接集合的读写锁ConnManager
}

func NewConnManager() *ConnManager {
	return &ConnManager{
		connections: make(map[uint32]ziface.IConnection),
	}
}

// 添加链接
func (cm *ConnManager) Add(conn ziface.IConnection) {
	cm.connLock.Lock()
	defer cm.connLock.Unlock()

	// 将conn加入ConnManager中
	cm.connections[conn.GetConnID()] = conn
	fmt.Printf("conn[%v] Add to ConnManager OK!\n", conn.GetConnID())
}

// 删除链接
func (cm *ConnManager) Remove(conn ziface.IConnection) {
	cm.connLock.Lock()
	defer cm.connLock.Unlock()

	// 将conn移除ConnManager
	conn.Stop()
	delete(cm.connections, conn.GetConnID())
	fmt.Printf("conn[%v] Del to ConnManager OK!\n", conn.GetConnID())
}

// 根据ConnID获取链接
func (cm *ConnManager) Get(connId uint32) (ziface.IConnection, error) {
	cm.connLock.RLock()
	defer cm.connLock.RUnlock()

	if conn, ok := cm.connections[connId]; ok {
		fmt.Printf("conn[%v] is in ConnManager!\n", conn.GetConnID())
		return conn, nil
	} else {
		return nil, errors.New("Connection NOT FOUND!")
	}

}

// 得到当前链接总数
func (cm *ConnManager) Len() int {
	cm.connLock.RLock()
	defer cm.connLock.RUnlock()

	return len(cm.connections)

}

// 清除并终止所有链接
func (cm *ConnManager) ClearConns() {
	cm.connLock.Lock()
	defer cm.connLock.Unlock()

	// 将conn移除ConnManager
	for connId, conn := range cm.connections {
		conn.Stop()
		delete(cm.connections, connId)
	}

	fmt.Println("ConnManager Clean OK!")

}
