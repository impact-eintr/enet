package main

import (
	"fmt"

	"github.com/impact-eintr/enet"
	"github.com/impact-eintr/enet/iface"
)

//ping test 自定义路由
type PingRouter struct {
	enet.BaseRouter //一定要先基础BaseRouter
}

//Test PreHandle
func (this *PingRouter) PreHandle(request iface.IRequest) {
	fmt.Println("Call Router PreHandle 🤤")
	_, err := request.GetConnection().GetRawConnection().Write([]byte("before ping ....\n"))
	if err != nil {
		fmt.Println("call back ping ping ping error")
	}
}

//Test Handle
func (this *PingRouter) Handle(request iface.IRequest) {
	fmt.Println("Call PingRouter Handle 🥵")
	_, err := request.GetConnection().GetRawConnection().Write([]byte("ping...ping...ping\n"))
	if err != nil {
		fmt.Println("call back ping ping ping error")
	}
}

//Test PostHandle
func (this *PingRouter) PostHandle(request iface.IRequest) {
	fmt.Println("Call Router PostHandle 👋")
	_, err := request.GetConnection().GetRawConnection().Write([]byte("After ping .....\n"))
	if err != nil {
		fmt.Println("call back ping ping ping error")
	}
}

func main() {
	//1 创建一个server 句柄 s
	s := enet.NewServer("[enetv1.0]", "tcp4")

	// 添加路由
	s.AddRouter(&PingRouter{})

	//2 开启服务
	s.Serve()
}
