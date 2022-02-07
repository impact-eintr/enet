package enet

import (
	"errors"
	"fmt"
	"os"
	"sync"
)

/*
	连接管理模块
*/
type ConnManager struct {
	connections map[uint32]IConnection //管理的连接信息
	connLock    sync.RWMutex           //读写连接的读写锁
}

/*
	创建一个链接管理
*/
func NewConnManager() *ConnManager {
	return &ConnManager{
		connections: make(map[uint32]IConnection),
	}
}

//添加链接
func (connMgr *ConnManager) Add(conn IConnection) {
	//保护共享资源Map 加写锁
	connMgr.connLock.Lock()
	defer connMgr.connLock.Unlock()

	//将conn连接添加到ConnMananger中
	connMgr.connections[conn.GetConnID()] = conn

	if _, ok := os.LookupEnv("enet_debug"); ok {
		fmt.Println("connection add to ConnManager successfully: conn num = ", connMgr.Len())
	}
}

//删除连接
func (connMgr *ConnManager) Remove(conn IConnection) {
	//保护共享资源Map 加写锁
	connMgr.connLock.Lock()
	defer connMgr.connLock.Unlock()

	//删除连接信息
	delete(connMgr.connections, conn.GetConnID())

	if _, ok := os.LookupEnv("enet_debug"); ok {
		fmt.Println("connection Remove ConnID=", conn.GetConnID(), " successfully: conn num = ", connMgr.Len())
	}
}

//利用ConnID获取链接
func (connMgr *ConnManager) Get(connID uint32) (IConnection, error) {
	//保护共享资源Map 加读锁
	connMgr.connLock.RLock()
	defer connMgr.connLock.RUnlock()

	if conn, ok := connMgr.connections[connID]; ok {
		return conn, nil
	} else {
		return nil, errors.New("connection not found")
	}
}

//获取当前连接
func (connMgr *ConnManager) Len() int {
	//保护共享资源Map 加读锁
	connMgr.connLock.RLock()
	defer connMgr.connLock.RUnlock()
	return len(connMgr.connections)
}

//清除并停止所有连接
func (connMgr *ConnManager) ClearConn() {
	//保护共享资源Map 加写锁
	connMgr.connLock.Lock()
	defer connMgr.connLock.Unlock()

	//停止并删除全部的连接信息
	for connID, conn := range connMgr.connections {
		//停止
		conn.Stop(false)
		//删除
		delete(connMgr.connections, connID)
	}

	if _, ok := os.LookupEnv("enet_debug"); ok {
		fmt.Println("Clear All Connections successfully: conn num = ", connMgr.Len())
	}
}
