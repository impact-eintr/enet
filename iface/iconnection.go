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
	//直接将Message数据发送数据给远程的TCP客户端
	SendMsg(msgId uint32, data []byte) error
}
