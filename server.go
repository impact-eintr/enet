package enet

import (
	"fmt"
	"net"
	"runtime"
	"strings"

	"github.com/impact-eintr/enet/iface"
)

type Server struct {
	// 服务器名称
	Name string
	// tcp4 udp4 或者 其他
	IPVersion string
	// 服务器绑定的IP
	IP string
	// 服务器绑定的端口
	Port int
	// 当前Server的消息管理模块，用来绑定MsgId和对应的处理方法
	msgHandler iface.IMsgHandle
	// 当前Server的链接管理器
	ConnMgr iface.IConnManager
	// =======================
	//新增两个hook函数原型

	// 该Server的连接创建时Hook函数
	OnConnStart func(conn iface.IConnection)
	// 该Server的连接断开时的Hook函数
	OnConnStop func(conn iface.IConnection)
}

func NewServer(network string) iface.IServer {
	s := &Server{
		Name:       GlobalObject.Name,
		IPVersion:  network,
		IP:         GlobalObject.Host,
		Port:       GlobalObject.Port,
		msgHandler: NewMsgHandler(),  //msgHandler 初始化
		ConnMgr:    NewConnManager(), //创建ConnManager
	}
	return s
}

func (s *Server) Start() {
	fmt.Printf("[START] Server listenner at IP: %s, Port %d, is starting\n", s.IP, s.Port)
	fmt.Printf("[LOG] Version: %s, MaxConn: %d, MaxPacketSize: %d\n",
		GlobalObject.Version,
		GlobalObject.MaxConn,
		GlobalObject.MaxPacketSize)
	go func() {
		//0 启动worker工作池机制
		s.msgHandler.StartWorkerPool()

		// 2 监听服务器地址/开启udp服务
		if s.IPVersion[0] == 't' {
			// ========================= TCP业务 ==========================
			listener, err := net.ListenTCP(s.IPVersion,
				&net.TCPAddr{IP: net.ParseIP(s.IP), Port: s.Port})
			if err != nil {
				fmt.Println("listen", s.IPVersion, "err", err)
				return
			}

			//已经监听成功
			fmt.Println("start enet server  ", s.Name, " succ, now listenning...")

			//TODO server.go 应该有一个自动生成ID的方法
			var cid uint32
			cid = 0

			// 启动server网络连接业务
			for {
				conn, err := listener.AcceptTCP()
				if err != nil {
					if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
						fmt.Printf("temporary Accept() failure - %s", err)
						runtime.Gosched()
						continue
					}
					// theres no direct way to detect this error because it is not exposed
					if !strings.Contains(err.Error(), "use of closed network connection") {
						fmt.Printf("listener.Accept() error - %s", err)
					}
					break
				}
				//  设置服务器最大链接控制
				if s.ConnMgr.Len() >= GlobalObject.MaxConn {
					conn.Close()
					continue
				}

				dealConn := NewTcpConntion(s, conn, cid, s.msgHandler)
				cid++

				// 开始处理业务
				go dealConn.Start()
			}
		} else if s.IPVersion[0] == 'u' {
			// ========================= UDP业务 ==========================
			udpConn, err := net.ListenUDP(s.IPVersion,
				&net.UDPAddr{IP: net.ParseIP(s.IP), Port: s.Port})
			if err != nil {
				fmt.Println(err)
				return
			}
			//fmt.Printf("Local: <%s> \n", udpConn.LocalAddr().String())

			// udp不设置连接控制
			dealConn := NewUdpConntion(s, udpConn, 0, s.msgHandler)

			// 开始处理业务
			go dealConn.Start()
		} else {
			panic("invalid type")
		}
	}()
}

func (s *Server) Stop() {
	fmt.Println("[STOP] enet server , name ", s.Name)

	// 将其他需要清理的连接信息或者其他信息 也要一并停止或者清理
	s.ConnMgr.ClearConn()
}

func (s *Server) Serve() {
	s.Start()

	//TODO Server.Serve() 是否在启动服务的时候 还要处理其他的事情呢 可以在这里添加

	//阻塞,否则main goroutine退出， listenner的goroutine将会退出
	for {
		select {}
	}
}

func (s *Server) AddRouter(msgId uint32, router iface.IRouter) {
	s.msgHandler.AddRouter(msgId, router)
}

// 得到链接管理
func (s *Server) GetConnMgr() iface.IConnManager {
	return s.ConnMgr
}

//设置该Server的连接创建时Hook函数
func (s *Server) SetOnConnStart(hookFunc func(iface.IConnection)) {
	s.OnConnStart = hookFunc
}

//设置该Server的连接断开时的Hook函数
func (s *Server) SetOnConnStop(hookFunc func(iface.IConnection)) {
	s.OnConnStop = hookFunc
}

//调用连接OnConnStart Hook函数
func (s *Server) CallOnConnStart(conn iface.IConnection) {
	if s.OnConnStart != nil {
		fmt.Println("---> CallOnConnStart....")
		s.OnConnStart(conn)
	}
}

//调用连接OnConnStop Hook函数
func (s *Server) CallOnConnStop(conn iface.IConnection) {
	if s.OnConnStop != nil {
		fmt.Println("---> CallOnConnStop....")
		s.OnConnStop(conn)
	}
}
