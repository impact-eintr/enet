package enet

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
)

// 一个Connection 可以是一个正常的 tcp 连接 也可以是一个 c->s s->c 的 udp 通信过程
type Connection struct {
	// 当前 Conn 属于哪个Server
	Server IServer

	// 内置连接 TCPConn / UDPConn
	Conn *net.TCPConn

	// 当前连接的ID 也可以称作为SessionID，ID全局唯一
	ConnID uint32

	// 当前连接的关闭状态
	isClosed bool

	// 消息管理MsgId和对应处理方法的消息管理模块(多路由实现)
	MsgHandler IMsgHandle

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

// 创建Tcp连接的方法
func NewConntion(server IServer, conn *net.TCPConn, connID uint32, msgHandler IMsgHandle) *Connection {
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
func (c *Connection) StartReader() {
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
			switch _, err := io.ReadFull(c.GetConnection(), headData); err {
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
				switch _, err := io.ReadFull(c.GetConnection(), data); err {
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

/*写消息Goroutine， 用户将数据发送给客户端*/
func (c *Connection) StartWriter() {
	if _, ok := os.LookupEnv("enet_debug"); ok {
		fmt.Println("[Writer Goroutine is running]")
		defer fmt.Println("[Writer Goroutine Exit!]")
	}

	for {
		select {
		case data := <-c.msgChan:
			//有数据要写给客户端
			if _, err := c.Conn.Write(data.data); err != nil {
				fmt.Println("Send Data error:, ", err, " Conn Writer exit")
				return
			}
		case data, ok := <-c.msgBuffChan:
			if ok {
				//有数据要写给客户端
				if _, err := c.Conn.Write(data.data); err != nil {
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
func (c *Connection) SendMsg(msgId uint32, data []byte) error {
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

func (c *Connection) SendBuffMsg(msgId uint32, data []byte) error {
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

// 启动 Connection 的循环阻塞机制
func (c *Connection) Start() {
	c.ctx, c.cancel = context.WithCancel(context.Background())
	// 开启处理该连接
	go c.StartReader()
	go c.StartWriter()

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
	c.Conn.Close()

	// 通知 Reader / Writer，该链接已经关闭
	c.cancel()

	// 将链接从连接管理器中删除
	if remove {
		c.Server.GetConnMgr().Remove(c)
	}

	// 关闭该链接全部管道
	close(c.msgBuffChan)

	// 通知 Router 中的 goroutine 退出
	c.MsgHandler.ExitMsgHandler(c)

}

// 从当前连接中获取原始的socket
func (c *Connection) GetConnection() *net.TCPConn {
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

// 设置链接属性
func (c *Connection) SetProperty(key string, value interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	c.property[key] = value
}

// 获取链接属性
func (c *Connection) GetProperty(key string) (interface{}, error) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	if value, ok := c.property[key]; ok {
		return value, nil
	} else {
		return nil, errors.New("no property found")
	}
}

// 移除链接属性
func (c *Connection) RemoveProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	delete(c.property, key)
}
