package enet

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"

	"github.com/impact-eintr/enet/iface"
)

// 一个Connection 可以是一个正常的 tcp 连接 也可以是一个 c->s s->c 的 udp 通信过程
type Connection struct {
	//当前Conn属于哪个Server
	Server iface.IServer //当前conn属于哪个server，在conn初始化的时候添加即可

	// 内置连接 TCPConn / UDPConn
	Conn net.Conn

	// 当前连接的ID 也可以称作为SessionID，ID全局唯一
	ConnID uint32

	// 当前连接的关闭状态
	isClosed bool

	// 消息管理MsgId和对应处理方法的消息管理模块(多路由实现)
	MsgHandler iface.IMsgHandle

	// 告知该链接已经退出/停止的channel
	ctx    context.Context
	cancel context.CancelFunc

	//无缓冲管道，用于读、写两个goroutine之间的消息通信
	msgChan chan *connMsg

	//有缓冲管道，用于读、写两个goroutine之间的消息通信
	msgBuffChan chan *connMsg

	sync.RWMutex
	//链接属性
	property map[string]interface{}
	//保护链接属性修改的锁
	propertyLock sync.RWMutex
}

type connMsg struct {
	data []byte
	dst  *net.UDPAddr
}

// ===================== TCP Connect ==========================

// 创建Tcp连接的方法
func NewTcpConntion(server iface.IServer, conn *net.TCPConn, connID uint32, msgHandler iface.IMsgHandle) *Connection {
	c := &Connection{
		Server:      server,
		Conn:        conn,
		ConnID:      connID,
		isClosed:    false,
		MsgHandler:  msgHandler,
		msgChan:     make(chan *connMsg), //msgChan初始化
		msgBuffChan: make(chan *connMsg, GlobalObject.MaxMsgChanLen),
		property:    make(map[string]interface{}), //对链接属性map初始化
	}

	c.Server.GetConnMgr().Add(c) //将当前新创建的连接添加到ConnManager中
	return c
}

/* 处理tcp conn读数据的Goroutine */
func (c *Connection) StartTcpReader() {
	if _, ok := os.LookupEnv("enet_debug"); ok {
		fmt.Println("[Reader Goroutine is running]")
		defer fmt.Printf("[%s Reader Goroutine Exit!]\n", c.RemoteAddr().String())
	}
	defer c.Stop(true)

	// 获取包装器
	dp := GetDataPack()
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			// 读取客户端的Msg Header
			headData := make([]byte, dp.GetHeadLen())
			switch _, err := io.ReadFull(c.GetTcpConnection(), headData); err {
			case io.EOF:
				if _, ok := os.LookupEnv("enet_debug"); ok {
					fmt.Println("读取结束")
				}
				return
			case nil:
			default:
				fmt.Println("read msg head error ", err)
				return
			}
			// 拆包，得到msgid 和 datalen 放在msg中
			msg, err := dp.Unpack(headData)
			if err != nil {
				fmt.Println("unpack error ", err)
				return
			}
			// 根据 dataLen 读取 data，放在msg.Data中
			var data []byte
			if msg.GetDataLen() > 0 {
				data = make([]byte, msg.GetDataLen())
				switch _, err := io.ReadFull(c.GetTcpConnection(), data); err {
				case io.EOF:
					if _, ok := os.LookupEnv("enet_debug"); ok {
						fmt.Println("读取结束")
					}
					return
				case nil:
				default:
					fmt.Println("read msg head error ", err)
					return
				}
			}

			msg.SetData(data)

			// 得到当前客户端请求的Request数据
			req := Request{
				conn: c,
				msg:  msg,
			}

			if GlobalObject.WorkerPoolSize > 0 {
				// 将任务派发给已经存在的goroutine
				// 已经启动工作池机制，将消息交给Worker处理
				c.MsgHandler.SendMsgToTaskQueue(&req)
			} else {
				// 开启新的gouroutine 来处理这些消息
				// 从绑定好的消息和对应的处理方法中执行对应的Handle方法
				go c.MsgHandler.DoMsgHandler(&req)
			}
		}
	}
}

/*
	写消息Goroutine， 用户将数据发送给客户端
*/
func (c *Connection) StartTcpWriter() {
	if _, ok := os.LookupEnv("enet_debug"); ok {
		fmt.Println("[Writer Goroutine is running]")
		defer fmt.Println("[Writer Goroutine Exit!]")
	}

	for {
		select {
		case data := <-c.msgChan:
			//有数据要写给客户端
			if _, err := c.Conn.(*net.TCPConn).Write(data.data); err != nil {
				fmt.Println("Send Data error:, ", err, " Conn Writer exit")
				return
			}
		case data, ok := <-c.msgBuffChan:
			if ok {
				//有数据要写给客户端
				if _, err := c.Conn.(*net.TCPConn).Write(data.data); err != nil {
					fmt.Println("Send Buff Data error:, ", err, " Conn Writer exit")
					return
				}
			} else {
				// msgBufChan 已经关闭
				break
			}
		case <-c.ctx.Done():
			return
		}
	}
}

// 直接将Message数据发送数据给远程的TCP客户端
func (c *Connection) SendTcpMsg(msgId uint32, data []byte) error {
	if _, ok := c.Conn.(*net.TCPConn); !ok {
		return errors.New("Invalid Connection type: UDP, should be: TCP")
	}

	c.Lock()
	defer c.Unlock()
	if c.isClosed == true {
		return errors.New("Connection closed when send msg")
	}
	// 将data封包，并且发送
	dp := GetDataPack()
	msg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack error msg id = ", msgId)
		return errors.New("Pack error msg ")
	}

	// 写回客户端
	c.msgChan <- &connMsg{data: msg} //将之前直接回写给conn.Write的方法 改为 发送给Channel 供Writer读取

	return nil
}

func (c *Connection) SendBuffTcpMsg(msgId uint32, data []byte) error {
	if _, ok := c.Conn.(*net.TCPConn); !ok {
		return errors.New("Invalid Connection type: UDP, should be: TCP")
	}

	c.Lock()
	defer c.Unlock()
	if c.isClosed == true {
		return errors.New("Connection closed when send buff msg")
	}

	//将data封包，并且发送
	dp := GetDataPack()
	msg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack error msg id = ", msgId)
		return errors.New("Pack error msg ")
	}

	// 写回客户端
	c.msgBuffChan <- &connMsg{data: msg}

	return nil
}

// ===================== UDP Connect ==========================

//创建Udp连接的方法
func NewUdpConntion(s iface.IServer, conn *net.UDPConn, connID uint32, msgHandler iface.IMsgHandle) *Connection {
	c := &Connection{
		Server:     s,
		Conn:       conn,
		ConnID:     connID,
		isClosed:   false,
		MsgHandler: msgHandler,
		msgChan:    make(chan *connMsg), //msgChan初始化
	}
	return c
}

/* 处理udp conn读数据的Goroutine */
func (c *Connection) StartUdpReader() {
	if _, ok := os.LookupEnv("enet_debug"); ok {
		fmt.Println("[Reader Goroutine is running]")
		defer fmt.Printf("[%s Reader Goroutine Exit!]\n", c.RemoteAddr().String())
	}

	defer c.Stop(true)

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			buf := make([]byte, GlobalObject.MaxPacketSize)
			n, remoteAddr, err := c.Conn.(*net.UDPConn).ReadFromUDP(buf)
			if err != nil {
				fmt.Printf("error during read: %s", err)
				return
			}

			// 解码 构建消息
			dp := GetDataPack()
			msg := dp.Decode(buf[:n])

			// 得到当前客户端请求的Request数据
			req := Request{
				conn:       c,
				msg:        msg,
				remoteAddr: remoteAddr,
			}

			if GlobalObject.WorkerPoolSize > 0 {
				// 将任务派发给已经存在的goroutine
				// 已经启动工作池机制，将消息交给Worker处理
				c.MsgHandler.SendMsgToTaskQueue(&req)
			} else {
				// 开启新的gouroutine 来处理这些消息
				// 从绑定好的消息和对应的处理方法中执行对应的Handle方法
				go c.MsgHandler.DoMsgHandler(&req)
			}
		}
	}
}

func (c *Connection) StartUdpWriter() {
	if _, ok := os.LookupEnv("enet_debug"); ok {
		fmt.Println("[Writer Goroutine is running]")
		defer fmt.Println("[Writer Goroutine Exit!]")
	}

	for {
		select {
		case data := <-c.msgChan:
			//有数据要写给客户端
			_, err := c.Conn.(*net.UDPConn).WriteToUDP(data.data, data.dst)
			if err != nil {
				fmt.Printf(err.Error())
			}
		case data := <-c.msgBuffChan:
			//有数据要写给客户端
			_, err := c.Conn.(*net.UDPConn).WriteToUDP(data.data, data.dst)
			if err != nil {
				fmt.Printf(err.Error())
			}
		case <-c.ctx.Done():
			//conn已经关闭
			return
		}
	}
}

// 直接将Message数据发送数据给远程的UDP客户端
func (c *Connection) SendUdpMsg(msgId uint32, data []byte, dst *net.UDPAddr) error {
	if _, ok := c.Conn.(*net.UDPConn); !ok {
		return errors.New("Invalid Connection type: TCP, should be: UDP")
	}

	if c.isClosed == true {
		return errors.New("Connection closed when send msg")
	}

	// 编码 构建数据包
	pkg := NewMsgPackage(msgId, data)
	dp := GetDataPack()
	msg := dp.Encode(pkg)

	// 写回客户端
	c.msgChan <- &connMsg{data: msg, dst: dst} //将之前直接回写给conn.Write的方法 改为 发送给Channel 供Writer读取
	return nil
}

func (c *Connection) SendBuffUdpMsg(msgId uint32, data []byte, dst *net.UDPAddr) error {
	if _, ok := c.Conn.(*net.UDPConn); !ok {
		return errors.New("Invalid Connection type: TCP, should be: UDP")
	}

	if c.isClosed == true {
		return errors.New("Connection closed when send buff msg")
	}

	// 编码 构建数据包
	pkg := NewMsgPackage(msgId, data)
	dp := GetDataPack()
	msg := dp.Encode(pkg)

	// 写回客户端
	c.msgChan <- &connMsg{data: msg, dst: dst} //将之前直接回写给conn.Write的方法 改为 发送给Channel 供Writer读取
	return nil
}

// 启动连接
func (c *Connection) Start() {
	c.ctx, c.cancel = context.WithCancel(context.Background())
	// 开启处理该连接
	if _, ok := c.Conn.(*net.TCPConn); ok {
		go c.StartTcpReader()
		go c.StartTcpWriter()
	} else if _, ok := c.Conn.(*net.UDPConn); ok {
		go c.StartUdpReader()
		go c.StartUdpWriter()
	} else {
		panic("invalid conn type")
	}

	//按照用户传递进来的创建连接时需要处理的业务，执行钩子方法
	c.Server.CallOnConnStart(c)
}

// 停止连接
func (c *Connection) Stop(remove bool) {
	c.Lock()
	defer c.Unlock()

	// 如果用户注册了该链接的关闭回调业务，那么在此刻应该显示调用
	c.Server.CallOnConnStop(c)

	// 如果当前链接已经关闭
	if c.isClosed == true {
		return
	}
	c.isClosed = true

	// 关闭socket链接
	if _, ok := c.Conn.(*net.TCPConn); ok {
		c.Conn.Close()
	} else if _, ok := c.Conn.(*net.UDPConn); ok {
		c.Conn.Close()
	}

	// 通知 Reader / Writer，该链接已经关闭
	c.cancel()

	// 将链接从连接管理器中删除
	if remove {
		c.Server.GetConnMgr().Remove(c)
	}

	// 关闭该链接全部管道
	close(c.msgBuffChan)
}

// 从当前连接中获取原始的socket
func (c *Connection) GetTcpConnection() *net.TCPConn {
	return c.Conn.(*net.TCPConn)
}

// 从当前连接中获取原始的socket
func (c *Connection) GetUdpConnection() *net.UDPConn {
	return c.Conn.(*net.UDPConn)
}

// 获取当前连的ID
func (c *Connection) GetConnID() uint32 {
	return c.ConnID
}

// 获取远程客户端地址信息
func (c *Connection) RemoteAddr() net.Addr {
	if _, ok := c.Conn.(*net.TCPConn); ok {
		return c.Conn.(*net.TCPConn).RemoteAddr()
	} else if _, ok := c.Conn.(*net.UDPConn); ok {
		return c.Conn.(*net.UDPConn).RemoteAddr()
	} else {
		panic("invalid net connect")
	}
}

// 设置链接属性
// 注意 连接属性均不为 UDP 开放
func (c *Connection) SetProperty(key string, value interface{}) {
	if _, ok := c.Conn.(*net.UDPConn); ok {
		return
	}

	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	c.property[key] = value
}

// 获取链接属性
// 注意 连接属性均不为 UDP 开放
func (c *Connection) GetProperty(key string) (interface{}, error) {
	if _, ok := c.Conn.(*net.UDPConn); ok {
		return nil, errors.New("invalid connection type")
	}

	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	if value, ok := c.property[key]; ok {
		return value, nil
	} else {
		return nil, errors.New("no property found")
	}
}

// 移除链接属性
// 注意 连接属性均不为 UDP 开放
func (c *Connection) RemoveProperty(key string) {
	if _, ok := c.Conn.(*net.UDPConn); ok {
		return
	}

	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	delete(c.property, key)
}
