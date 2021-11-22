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
	err := request.GetConnection().SendBuffTcpMsg(1, []byte("ping...ping...ping"))
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

	err := request.GetConnection().SendBuffTcpMsg(1, []byte("Hello Router V1.0"))
	if err != nil {
		fmt.Println(err)
	}
}

//创建连接的时候执行
func DoConnectionBegin(conn iface.IConnection) {
	fmt.Println("DoConnecionBegin is Called ... ")

	fmt.Println("Set conn Name, Home done!")
	conn.SetProperty("Name", "Aceld")
	conn.SetProperty("Home", "https://www.jianshu.com/u/35261429b7f1")

	err := conn.SendBuffTcpMsg(2, []byte("DoConnection BEGIN..."))
	if err != nil {
		fmt.Println(err)
	}
}

//连接断开的时候执行
func DoConnectionLost(conn iface.IConnection) {
	//============在连接销毁之前，查询conn的Name，Home属性=====
	if name, err := conn.GetProperty("Name"); err == nil {
		fmt.Println("Conn Property Name = ", name)
	}

	if home, err := conn.GetProperty("Home"); err == nil {
		fmt.Println("Conn Property Home = ", home)
	}

	fmt.Println("DoConneciotnLost is Called ... ")
}

func main() {
	//1 创建一个server 句柄 s
	s := enet.NewServer("tcp4")

	//注册链接hook回调函数
	s.SetOnConnStart(DoConnectionBegin)
	s.SetOnConnStop(DoConnectionLost)

	// 添加路由
	s.AddRouter(0, &PingRouter{})
	s.AddRouter(1, &HelloRouter{})

	//2 开启服务
	s.Serve()
}
