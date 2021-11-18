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
	//当前Server的消息管理模块，用来绑定MsgId和对应的处理方法
	msgHandler iface.IMsgHandle
}

func (s *Server) Start() {
	fmt.Printf("[START] Server listenner at IP: %s, Port %d, is starting\n", s.IP, s.Port)

	go func() {
		//0 启动worker工作池机制
		s.msgHandler.StartWorkerPool()

		// 1 获取一个tcp/udp的Addr
		var addr net.Addr
		var err error
		switch s.IPVersion[0] {
		case 't':
			addr, err = net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		case 'u':
			addr, err = net.ResolveUDPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		}
		if err != nil {
			fmt.Println("resolve tcp addr err: ", err)
			return
		}

		// 2 监听服务器地址/开启udp服务
		if _, ok := addr.(*net.TCPAddr); ok {
			// ========================= TCP业务 ==========================
			listener, err := net.ListenTCP(s.IPVersion, addr.(*net.TCPAddr))
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
				// TODO Server.Start() 设置服务器最大链接控制

				// TODO Server.Start() 处理该新链接请求的业务方法 每个 conn 对应一个 handler

				dealConn := NewTcpConntion(conn, cid, s.msgHandler)
				go dealConn.Start()
			}
		} else if _, ok := addr.(*net.UDPAddr); ok {
			// ========================= UDP业务 ==========================
			udpConn, err := net.ListenUDP(s.IPVersion, addr.(*net.UDPAddr))
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Printf("Local: <%s> \n", udpConn.LocalAddr().String())

			// TODO server.go 应该有一个自动生成ID的方法
			var cid uint32
			cid = 0

			dealConn := NewUdpConntion(udpConn, cid, s.msgHandler)
			go dealConn.Start()
		} else {
			panic("invalid type")
		}
	}()
}

func (s *Server) Stop() {
	fmt.Println("[STOP] Zinx server , name ", s.Name)

	//TODO  Server.Stop() 将其他需要清理的连接信息或者其他信息 也要一并停止或者清理
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

func NewServer(network string) iface.IServer {
	s := &Server{
		Name:       GlobalObject.Name,
		IPVersion:  network,
		IP:         GlobalObject.Host,
		Port:       GlobalObject.Port,
		msgHandler: NewMsgHandler(), //msgHandler 初始化
	}
	return s
}
