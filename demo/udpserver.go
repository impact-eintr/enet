package main

import (
	"fmt"
	"net"

	"github.com/impact-eintr/enet"
	"github.com/impact-eintr/enet/iface"
)

//ping test 自定义路由
type PongRouter struct {
	enet.BaseRouter //一定要先基础BaseRouter
}

//Test PreHandle
func (this *PongRouter) PreHandle(request iface.IRequest) {
	fmt.Println("Call Router PreHandle 🤤")
}

//Test Handle
func (this *PongRouter) Handle(request iface.IRequest) {
	fmt.Println("Call PongRouter Handle 🥵")

	udpConn, _ := request.GetConnection().GetRawConnection().(*net.UDPConn)
	_, err := udpConn.WriteToUDP(request.GetData(), request.GetRemoteAddr())
	if err != nil {
		fmt.Printf(err.Error())
	}
}

//Test PostHandle
func (this *PongRouter) PostHandle(request iface.IRequest) {
	fmt.Println("Call Router PostHandle 👋")
}

func main() {
	//1 创建一个server 句柄 s
	s := enet.NewServer("[enetv1.0]", "udp")

	s.AddRouter(&PongRouter{})

	//2 开启服务
	s.Serve()
}
