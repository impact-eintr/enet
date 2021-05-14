package znet

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/impact-eintr/Zinx/utils"
	"github.com/impact-eintr/Zinx/ziface"
)

type Connection struct {
	// 当前conn隶属于哪个Server
	TcpServer ziface.IServer

	// 当前连接的socket TCP套接字
	Conn *net.TCPConn

	// 连接的ID
	ConnID uint32

	// 当前的连接状态
	isClosed bool

	// 消息处理API
	MsgHandler ziface.IMsgHandle

	// 告知当前连接已经退出/停止的channel
	ExitChan chan bool

	// 读写分离通信管道
	msgChan chan []byte

	// 链接属性
	property map[string]interface{}

	// 链接属性锁
	propertyLock sync.RWMutex
}

func NewConnection(server ziface.IServer, conn *net.TCPConn, connID uint32, msgHandler ziface.IMsgHandle) *Connection {
	c := &Connection{
		TcpServer:  server,
		Conn:       conn,
		ConnID:     connID,
		MsgHandler: msgHandler,
		isClosed:   false,
		ExitChan:   make(chan bool, 1),
		msgChan:    make(chan []byte),
		property:   make(map[string]interface{}),
	}

	server.GetConnManager().Add(c)
	fmt.Println("server: ", server)
	return c
}

// 连接读服务
func (c *Connection) StartReader() {
	fmt.Println("Reader Goroutine is running...")
	defer fmt.Println("connID:", c.ConnID, "Reader is exit, remote addr is ", c.GetRemoteAddr().String())
	defer c.Stop()

	for {
		dp := NewDataPack()

		headData := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(c.GetTCPConnection(), headData); err != nil {
			fmt.Println("read msg head error", err)
			break
		}
		// 解包 获取数据长度
		msg, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("read msg head error", err)
			break
		}

		var data []byte
		if msg.GetMsgLen() > 0 {
			data = make([]byte, msg.GetMsgLen())
			if _, err := io.ReadFull(c.GetTCPConnection(), data); err != nil {
				fmt.Println("read msg head error", err)
				break
			}
		}

		msg.SetMsgData(data)

		// 得到当前conn数据的Request请求数据
		req := Request{
			conn: c,
			msg:  msg,
		}

		// 执行注册的路由方法
		if utils.GlobalConf.WorkerPoolSize > 0 {
			go c.MsgHandler.SendMsgToTaskQueue(&req)

		} else {
			go c.MsgHandler.DoMsgHandler(&req)

		}

	}
}

// 连接写服务
func (c *Connection) StartWriter() {
	fmt.Println("Writer Goroutine is running...")
	defer fmt.Println("connID:", c.ConnID, "Writer is exit, remote addr is ", c.GetRemoteAddr().String())

	// 不同地阻塞等待channel的消息 接收到就写给客户端
	for {
		select {
		case data := <-c.msgChan:
			// 有数据要写给客户端
			if _, err := c.Conn.Write(data); err != nil {
				fmt.Println("Send data error:", err)
				return
			}
		case <-c.ExitChan:
			// 代表Reader已经退出, 此时Writer也应该退出
			return
		}
	}
}

// 启动连接 让当前的连接准备开始工作
func (c *Connection) Start() {
	fmt.Println("Conn Start ()... ConnID:", c.ConnID)
	go c.StartReader()
	go c.StartWriter()
}

// 停止连接 结束当前连接的动作
func (c *Connection) Stop() {
	fmt.Println("Conn Stop().. ConnID", c.ConnID)
	if c.isClosed {
		return
	}

	c.isClosed = true

	c.Conn.Close()

	// 告知Writer关闭 可能多余
	c.ExitChan <- true

	// 将当前链接从ConnManager中移除
	c.TcpServer.GetConnManager().Remove(c)

	close(c.ExitChan)
	close(c.msgChan)
}

// 获取当前连接模块绑定的socket conn
func (c *Connection) GetTCPConnection() *net.TCPConn {
	return c.Conn
}

// 获取当前连接模块的连接ID
func (c *Connection) GetConnID() uint32 {
	return c.ConnID
}

// 获取远程客户端的TCP状态 IP port
func (c *Connection) GetRemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

// 提供一个SendMsg方法 将我们要发送到客户端的数据 先进行封包再发送
func (c *Connection) SendMsg(msgId uint32, data []byte) error {
	if c.isClosed {
		return errors.New("Connection closed when send msg")
	}

	dp := NewDataPack()
	msgpkg := NewMsgPackage(msgId, data)

	msg, err := dp.Pack(msgpkg)
	if err != nil {
		fmt.Println("Pack error msgid:", msgId)
		return errors.New("Pack error")
	}
	// 将数据发送给写端
	c.msgChan <- msg

	return nil
}

//设置链接属性
func (c *Connection) SetProperty(key string, value interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	c.property[key] = value

}

//获取链接属性
func (c *Connection) GetProperty(key string) (interface{}, error) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	if value, ok := c.property[key]; ok {
		return value, nil
	} else {
		return nil, errors.New("未注册该链接属性")
	}

}

//移除链接属性
func (c *Connection) RemoveProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	delete(c.property, key)
}
