package enet

import (
	"fmt"
	"log"
	"net"

	"github.com/impact-eintr/Zinx/iface"
)

type Connection struct {
	Conn net.Conn
	//当前连接的ID 也可以称作为SessionID，ID全局唯一
	ConnID uint32
	//当前连接的关闭状态
	isClosed bool

	//该连接的处理方法api
	handleAPI iface.HandFunc

	//告知该链接已经退出/停止的channel
	ExitBuffChan chan bool
}

//创建Tcp连接的方法
func NewTcpConntion(conn *net.TCPConn, connID uint32, callback_api iface.HandFunc) *Connection {
	c := &Connection{
		Conn:         conn,
		ConnID:       connID,
		isClosed:     false,
		handleAPI:    callback_api,
		ExitBuffChan: make(chan bool, 1),
	}

	return c
}

//创建Udp连接的方法
func NewUdpConntion(conn *net.UDPConn, connID uint32, callback_api iface.HandFunc) *Connection {
	c := &Connection{
		Conn:         conn,
		ConnID:       connID,
		isClosed:     false,
		handleAPI:    callback_api,
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
		//读取我们最大的数据到buf中
		buf := make([]byte, 512)
		cnt, err := c.Conn.Read(buf)
		if err != nil {
			fmt.Println("recv buf err ", err)
			c.ExitBuffChan <- true
			continue
		}
		//调用当前链接业务(这里执行的是当前conn的绑定的handle方法)
		if err := c.handleAPI(c.Conn, buf, cnt); err != nil {
			fmt.Println("connID ", c.ConnID, " handle is error")
			c.ExitBuffChan <- true
			return
		}
	}
}

/* 处理udp conn读数据的Goroutine */
func (c *Connection) StartUdpReader() {
	fmt.Println("Reader Goroutine is  running")
	defer c.Stop()

	udpConn, _ := c.Conn.(*net.UDPConn)
	data := make([]byte, 1024)
	for {
		log.Println(c.Conn.LocalAddr(), c.Conn.RemoteAddr())
		n, remoteAddr, err := udpConn.ReadFromUDP(data)
		if err != nil {
			fmt.Printf("error during read: %s", err)
			c.ExitBuffChan <- true
		}

		//调用当前链接业务(这里执行的是当前conn的绑定的handle方法)
		fmt.Printf("<%s> %s\n", remoteAddr, data[:n])
		_, err = udpConn.WriteToUDP(data[:n], remoteAddr)
		if err != nil {
			fmt.Printf(err.Error())
		}
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
