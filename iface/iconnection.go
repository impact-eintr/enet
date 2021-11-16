package iface

import "net"

type IConnection interface {
	// 启动连接
	Start()
	// 停止连接
	Stop()
	// 从当前连接中获取原始的docket
	GetRawConnection() net.Conn
	// 获取当前连的ID
	GetConnID() uint32
	// 获取远程客户端地址信息
	RemoteAddr() net.Addr
}

// 定义一个统一处理连接业务的接口
type HandFunc func(net.Conn, []byte, int) error
