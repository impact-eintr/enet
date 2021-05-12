package znet

import (
	"fmt"
	"log"
	"net"

	"github.com/impact-eintr/Zinx/utils"
	"github.com/impact-eintr/Zinx/ziface"
)

type Server struct {
	// 服务器名称
	Name string
	// 服务器绑定的ip版本
	IPVersion string
	// 服务器监听的IP
	IP string
	// 服务器监听的端口
	Port int
	// 当前的Server的消息管理模块 用来绑定MsgId和对应的处理业务API
	MsgHandler ziface.IMsgHandle
	// 当前Server链接管理模块
	ConnManager ziface.IConnManager
	// 该Server的连接创建时Hook函数
	OnConnStart func(ziface.IConnection)
	// 该Server的连接断开时的Hook函数
	OnConnStop func(ziface.IConnection)
}

func (s *Server) Start() {
	log.Println("[Zinx] Conf:", utils.GlobalConf)
	fmt.Printf("[Start] Server Listening on IP: %s, Port :%d, is starting\n", s.IP, s.Port)
	// 获取一个TCP连接的Addr
	go func() {

		s.MsgHandler.StartWorkerPool()

		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			fmt.Println("处理tcp地址出错:", err)
			return
		}

		listener, err := net.ListenTCP(s.IPVersion, addr)
		if err != nil {
			fmt.Println("listen ", s.IPVersion, "err", err)
		}

		// 初始化connid
		var cid uint32 = 0

		fmt.Println("成功启动server", s.Name)

		// 阻塞地等待客户端连接处理客户端连接业务
		for {
			conn, err := listener.AcceptTCP()
			if err != nil {
				fmt.Println("Accept err", err)
				continue
			}

			fmt.Println(utils.GlobalConf.MaxConn, s.ConnManager.Len())
			if s.ConnManager.Len() > utils.GlobalConf.MaxConn {
				//TODO 给客户端相应的一个超出最大链接数量的error
				if _, err = conn.Write([]byte("链接太多了 即将断开链接 请稍后尝试")); err != nil {
					fmt.Println(err)
				}

				fmt.Println("Too Many Connection!")
				conn.Close()
				continue
			}

			connDeal := NewConnection(s, conn, cid, s.MsgHandler)

			go connDeal.Start()

			cid++
		}
	}()

}

func (s *Server) Stop() {
	// TODO
	s.ConnManager.ClearConns()
}

func (s *Server) Run() {
	s.Start()

	// TODO

	// 阻塞状态
	select {}

}

// 添加router模块
func (s *Server) AddRouter(msgId uint32, router ziface.IRouter) {
	s.MsgHandler.AddRouter(msgId, router)
	fmt.Printf("[Zinx] add Router[%v] Suss!!\n", msgId)
}

// 获取当前Server的链接管理器
func (s *Server) GetConnManager() ziface.IConnManager {
	return s.ConnManager
}

//设置该Server的连接创建时Hook函数
func (s *Server) SetOnConnStart(hook func(ziface.IConnection)) {
	s.OnConnStart = hook
}

//设置该Server的连接断开时的Hook函数
func (s *Server) SetOnConnStop(hook func(ziface.IConnection)) {
	s.OnConnStop = hook
}

//调用连接OnConnStart Hook函数
func (s *Server) CallOnConnStart(conn ziface.IConnection) {
	if s.OnConnStart != nil {
		s.OnConnStart(conn)
		fmt.Println("成功调用Hook On Start")
	}
}

//调用连接OnConnStop Hook函数
func (s *Server) CallOnConnStop(conn ziface.IConnection) {
	if s.OnConnStop != nil {
		s.OnConnStop(conn)
		fmt.Println("成功调用Hook On Stop")

	}
}

// 初始化Server模块
func NewServer(name string) ziface.IServer {
	s := &Server{
		Name:        utils.GlobalConf.Name,
		IPVersion:   "tcp4",
		IP:          utils.GlobalConf.Host,
		Port:        utils.GlobalConf.Port,
		MsgHandler:  NewMsgHandle(),
		ConnManager: NewConnManager(),
	}

	return s
}
