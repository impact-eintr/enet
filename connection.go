package enet

import (
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/impact-eintr/enet/iface"
)

type Connection struct {
	Conn net.Conn
	// 当前连接的ID 也可以称作为SessionID，ID全局唯一
	ConnID uint32
	// 当前连接的关闭状态
	isClosed bool

	// 该连接的处理方法router 用来替代HandFunc 实现每个路由都有自己的处理方法
	Router iface.IRouter

	// 告知该链接已经退出/停止的channel
	ExitBuffChan chan bool
}

// 创建Tcp连接的方法
func NewTcpConntion(conn *net.TCPConn, connID uint32, router iface.IRouter) *Connection {
	c := &Connection{
		Conn:         conn,
		ConnID:       connID,
		isClosed:     false,
		Router:       router,
		ExitBuffChan: make(chan bool, 1),
	}

	return c
}

//创建Udp连接的方法
func NewUdpConntion(conn *net.UDPConn, connID uint32, router iface.IRouter) *Connection {
	c := &Connection{
		Conn:         conn,
		ConnID:       connID,
		isClosed:     false,
		Router:       router,
		ExitBuffChan: make(chan bool, 1),
	}

	return c
}

/* 处理tcp conn读数据的Goroutine */
func (c *Connection) StartTcpReader() {
	fmt.Println("Reader Goroutine is  running")
	defer fmt.Println(c.RemoteAddr().String(), " conn reader exit!")
	defer c.Stop()

	for {
		// 创建包装器
		dp := NewDataPack()

		// 读取客户端的MSG Header
		headData := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(c.GetRawConnection(), headData); err != nil {
			fmt.Println("read msg head error ", err)
			c.ExitBuffChan <- true
			continue
		}

		//拆包，得到msgid 和 datalen 放在msg中
		msg, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("unpack error ", err)
			c.ExitBuffChan <- true
			continue
		}

		//根据 dataLen 读取 data，放在msg.Data中
		var data []byte
		if msg.GetDataLen() > 0 {
			data = make([]byte, msg.GetDataLen())
			if _, err := io.ReadFull(c.GetRawConnection(), data); err != nil {
				fmt.Println("read msg data error ", err)
				c.ExitBuffChan <- true
				continue
			}
		}
		msg.SetData(data)

		//得到当前客户端请求的Request数据
		req := Request{
			conn: c,
			msg:  msg,
		}
		//从路由Routers 中找到注册绑定Conn的对应Handle
		go func(request iface.IRequest) {
			//执行注册的路由方法
			c.Router.PreHandle(request)
			c.Router.Handle(request)
			c.Router.PostHandle(request)
		}(&req)
	}
}

/* 处理udp conn读数据的Goroutine */
func (c *Connection) StartUdpReader() {
	fmt.Println("Reader Goroutine is  running")
	defer c.Stop()

	udpConn, _ := c.Conn.(*net.UDPConn)
	buf := make([]byte, GlobalObject.MaxPacketSize)
	for {
		n, remoteAddr, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			fmt.Printf("error during read: %s", err)
			c.ExitBuffChan <- true
		}

		msg := &Message{Data: buf[:n]}

		//得到当前客户端请求的Request数据
		req := Request{
			conn:       c,
			msg:        msg,
			remoteAddr: remoteAddr,
		}
		fmt.Printf("<%s> %s\n", remoteAddr, buf[:n])
		//从路由Routers 中找到注册绑定Conn的对应Handle
		go func(request iface.IRequest) {
			//执行注册的路由方法
			c.Router.PreHandle(request)
			c.Router.Handle(request)
			c.Router.PostHandle(request)
		}(&req)

		//调用当前链接业务(这里执行的是当前conn的绑定的handle方法)
		//_, err = udpConn.WriteToUDP(data[:n], remoteAddr)
		//if err != nil {
		//	fmt.Printf(err.Error())
		//}
	}

}

// 启动连接
func (c *Connection) Start() {
	// 开启处理该连接
	if _, ok := c.Conn.(*net.TCPConn); ok {
		go c.StartTcpReader()
	} else if _, ok := c.Conn.(*net.UDPConn); ok {
		go c.StartUdpReader()
	} else {
		panic("invalid conn type")
	}

	for {
		select {
		case <-c.ExitBuffChan:
			// 得到消息退出
			return
		}
	}
}

// 停止连接
func (c *Connection) Stop() {
	//1. 如果当前链接已经关闭
	if c.isClosed == true {
		return
	}
	c.isClosed = true

	//TODO Connection Stop() 如果用户注册了该链接的关闭回调业务，那么在此刻应该显示调用

	// 关闭socket链接
	c.Conn.Close()

	//通知从缓冲队列读数据的业务，该链接已经关闭
	c.ExitBuffChan <- true

	//关闭该链接全部管道
	close(c.ExitBuffChan)
}

// 从当前连接中获取原始的docket
func (c *Connection) GetRawConnection() net.Conn {
	return c.Conn
}

// 获取当前连的ID
func (c *Connection) GetConnID() uint32 {
	return c.ConnID
}

// 获取远程客户端地址信息
func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

// 直接将Message数据发送数据给远程的TCP客户端
func (c *Connection) SendMsg(msgId uint32, data []byte) error {
	if c.isClosed == true {
		return errors.New("Connection closed when send msg")
	}
	// 将data封包，并且发送
	dp := NewDataPack()
	msg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack error msg id = ", msgId)
		return errors.New("Pack error msg ")
	}

	// 写回客户端
	if _, err := c.Conn.Write(msg); err != nil {
		fmt.Println("Write msg id ", msgId, " error ")
		c.ExitBuffChan <- true
		return errors.New("conn Write error")
	}

	return nil
}
