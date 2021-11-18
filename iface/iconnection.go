package iface

import "net"

type IConnection interface {
	// 启动连接
	Start()
	// 停止连接
	Stop()
	// 从当前连接中获取原始的tcp socket
	GetTcpConnection() *net.TCPConn
	// 从当前连接中获取原始的udp socket
	GetUdpConnection() *net.UDPConn
	// 获取当前连的ID
	GetConnID() uint32
	// 获取远程客户端地址信息
	RemoteAddr() net.Addr
	//直接将Message数据发送数据给远程的客户端
	SendTcpMsg(msgId uint32, data []byte) error
	SendUdpMsg(msgId uint32, data []byte, dst *net.UDPAddr) error
}
