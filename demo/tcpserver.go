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

//Test Handle
func (this *PingRouter) Handle(request iface.IRequest) {
	fmt.Println("Call PingRouter Handle 🥵")
	//先读取客户端的数据，再回写ping...ping...ping
	fmt.Println("recv from client : msgId=", request.GetMsgID(), ", data=", string(request.GetData()))

	//回写数据
	err := request.GetConnection().SendMsg(1, []byte("ping...ping...ping"))
	if err != nil {
		fmt.Println(err)
	}
}

type HelloRouter struct {
	enet.BaseRouter
}

func (this *HelloRouter) Handle(request iface.IRequest) {
	fmt.Println("Call HelloZinxRouter Handle")
	//先读取客户端的数据，再回写ping...ping...ping
	fmt.Println("recv from client : msgId=", request.GetMsgID(), ", data=", string(request.GetData()))

	err := request.GetConnection().SendMsg(1, []byte("Hello Router V1.0"))
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	//1 创建一个server 句柄 s
	s := enet.NewServer("tcp4")

	// 添加路由
	s.AddRouter(0, &PingRouter{})
	s.AddRouter(1, &HelloRouter{})

	//2 开启服务
	s.Serve()
}
