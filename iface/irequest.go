package iface

import "net"

type IRequest interface {
	GetConnection() IConnection  // 获取请求连接信息
	GetMsgID() uint32            // 获取请求消息的ID
	GetData() []byte             // 获取请求消息的数据
	GetRemoteAddr() *net.UDPAddr // 获取客户端地址(udp专用)
}
