package enet

import (
	"net"

	"github.com/impact-eintr/enet/iface"
)

type Request struct {
	conn       iface.IConnection //已经和客户端建立好的 链接
	data       []byte            //客户端请求的数据
	remoteAddr *net.UDPAddr      // udp通信时 远端地址
}

//获取请求连接信息
func (r *Request) GetConnection() iface.IConnection {
	return r.conn
}

//获取请求消息的数据
func (r *Request) GetData() []byte {
	return r.data
}

//获取客户端的地址(udp专用)
func (r *Request) GetRemoteAddr() *net.UDPAddr {
	return r.remoteAddr
}