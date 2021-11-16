package enet

import (
	"context"
	"fmt"
	"io"
	"net"
	"runtime"
	"strings"
	"syscall"

	"github.com/impact-eintr/Zinx/iface"
	"golang.org/x/sys/unix"
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
}

func (s *Server) Start() {
	fmt.Printf("[START] Server listenner at IP: %s, Port %d, is starting\n", s.IP, s.Port)

	go func() {
		cfg := net.ListenConfig{
			Control: func(network, address string, c syscall.RawConn) error {
				return c.Control(func(fd uintptr) {
					syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, unix.SO_REUSEADDR, 1)
					syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, unix.SO_REUSEPORT, 1)
				})
			},
		}

		// 开始监听
		listenner, err := cfg.Listen(context.Background(), s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			fmt.Println("listen", s.IPVersion, listenner.Addr().String(), "err:", err)
			return
		}

		//已经监听成功
		fmt.Println("start enet server  ", s.Name, " succ, now listenning...")

		// 启动server网络连接业务
		for {
			conn, err := listenner.Accept()
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

			go func() {
				for {
					buf := make([]byte, 512)
					cnt, err := conn.Read(buf)
					if err != nil {
						if err == io.EOF {
							fmt.Println("client exit!")
							break
						} else {
							fmt.Println("recv buf err ", err)
							continue
						}
					}
					// echo
					if _, err := conn.Write(buf[:cnt]); err != nil {
						fmt.Println("write back buf err", err)
						continue
					}
				}
			}()
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

func NewServer(name, network string) iface.IServer {
	s := &Server{
		Name:      name,
		IPVersion: network, // TODO 这里注意之后的udp扩展
		IP:        "0.0.0.0",
		Port:      6430,
	}
	return s
}
