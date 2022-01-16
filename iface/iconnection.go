package iface

import "net"

type IConnection interface {
	// 启动连接
	Start()
	// 停止连接
	Stop(bool) // 这个bool用来判断是否要让conn自己取消管理
	// 从当前连接中获取原始的tcp socket
	GetTcpConnection() *net.TCPConn
	// 从当前连接中获取原始的udp socket
	GetUdpConnection() *net.UDPConn
	// 获取当前连的ID
	GetConnID() uint32
	// 获取远程客户端地址信息
	RemoteAddr() net.Addr
	//直接将Message数据发送数据给远程的客户端(无缓冲)
	SendTcpMsg(msgId uint32, data []byte) error
	SendUdpMsg(msgId uint32, data []byte, dst *net.UDPAddr) error
	//直接将Message数据发送数据给远程的客户端(无缓冲)
	SendBuffTcpMsg(msgId uint32, data []byte) error
	SendBuffUdpMsg(msgId uint32, data []byte, dst *net.UDPAddr) error

	//设置链接属性
	SetProperty(key string, value interface{})
	//获取链接属性
	GetProperty(key string) (interface{}, error)
	//移除链接属性
	RemoveProperty(key string)
}
