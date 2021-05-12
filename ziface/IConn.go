package ziface

import "net"

// 定义连接模块的抽象层
type IConnection interface {
	// 启动连接 让当前的连接准备开始工作
	Start()

	// 停止连接 结束当前连接的动作
	Stop()

	// 获取当前连接模块绑定的socket conn
	GetTCPConnection() *net.TCPConn

	// 获取当前连接模块的连接ID
	GetConnID() uint32

	// 获取远程客户端的TCP状态 IP port
	GetRemoteAddr() net.Addr

	// 发送数据 将数据发送给远程的客户端
	SendMsg(msgId uint32, data []byte) error

	//设置链接属性
	SetProperty(key string, value interface{})

	//获取链接属性
	GetProperty(key string) (interface{}, error)

	//移除链接属性
	RemoveProperty(key string)
}

// 定义一个处理连接业务的方法 连接 内容 长度
type HandleFunc func(*net.TCPConn, []byte, int) error
